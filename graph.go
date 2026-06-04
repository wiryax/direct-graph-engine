package dge

import (
	"fmt"
)

type Task interface {
	Execute(gCtx *GraphContext) error
}

type Graph interface {
	Run()
	RunWithContext(gCtx *GraphContext)
}

type state int

const (
	Pending state = iota
	Success
	Fail
	Running
	Skipped
)

type Edge struct {
	from, to *BasicVertex
	exp      expression
	lOp      tokenType
	pConst   state
}

func (e *Edge) evalConst() bool {
	return e.from.state == e.pConst
}

type Vertex interface {
	ExecuteTask(gCtx *GraphContext) error
}

type BasicVertex struct {
	id                    string
	state                 state
	task                  Task
	pendingEdge, failEdge int
	in, out               []*Edge
}

func (v *BasicVertex) String() string {
	return fmt.Sprintf("id=%s state=%d pendingEdge=%d failEdge=%d,", v.id, v.state, v.pendingEdge, v.failEdge)
}

func (v *BasicVertex) GetId() string {
	return v.id
}

func (v *BasicVertex) SetState(s state) {
	v.state = s
}

func (v *BasicVertex) GetState() state {
	return v.state
}

func (v *BasicVertex) ExecuteTask(gCtx *GraphContext) error {
	return v.task.Execute(gCtx)
}

type BasicGraph struct {
	id     string
	vertex []*BasicVertex
}

func NewGraph(id string) *BasicGraph {
	return &BasicGraph{
		id: id,
	}
}

func (g *BasicGraph) RunWithContext(gCtx *GraphContext) {
	g.run(gCtx)
}

func (g *BasicGraph) Run() {
	rState := &RuntimeState{
		variable: make(map[string]string),
	}

	storage := NewStorage()

	gCtx := NewGraphContext(NewLogger(nil), rState, storage)
	g.run(gCtx)
}

func (g *BasicGraph) run(gCtx *GraphContext) {
	if gCtx == nil {
		panic("graph context cannot nil")
	}

	var queue []*BasicVertex
	for _, v := range g.vertex {
		if v.pendingEdge == 0 && v.failEdge == 0 {
			queue = append(queue, v)
		}
	}

	g.execute(gCtx, queue)
}

func (g *BasicGraph) execute(gCtx *GraphContext, queue []*BasicVertex) {
	for {
		if len(queue) == 0 {
			break
		}

		v := queue[0]
		queue = queue[1:]

		v.state = Running

		gCtx.Log(EventStart, LevelInfo, "Start execute vertex", v.GetId(), g.id)
		err := v.ExecuteTask(gCtx)
		if err != nil {
			gCtx.Log(EventFailed, LevelInfo, err.Error(), v.GetId(), g.id)
			v.state = Fail
		} else {
			gCtx.Log(EventSuccess, LevelInfo, "Finish execute vertex", v.GetId(), g.id)
			v.state = Success
		}
		g.getReadyVertex(gCtx, v, &queue)
	}
}

func (g *BasicGraph) getReadyVertex(gCtx *GraphContext, v *BasicVertex, queue *[]*BasicVertex) {
	for _, child := range v.out {
		if !child.evalConst() && child.lOp == ExpAnd {
			child.to.failEdge++
			child.to.pendingEdge--
		} else {
			rEvaluate := evaluate(gCtx, child.exp.tokens)
			if rEvaluate == False {
				child.to.failEdge++
				child.to.pendingEdge--
			} else if rEvaluate == True {
				child.to.pendingEdge--
			}
		}

		if child.to.pendingEdge == 0 && child.to.failEdge > 0 {
			child.to.state = Skipped
		}

		g.getReadyVertex(gCtx, child.to, queue)

		if child.to.pendingEdge == 0 && child.to.failEdge == 0 {
			*queue = append(*queue, child.to)
		}
	}
}

func (g *BasicGraph) AddVertex(vertex ...*BasicVertex) {
	g.vertex = append(g.vertex, vertex...)
}

func (g *BasicGraph) Connect(from, to *BasicVertex, op state, lOp tokenType, tk []token) {
	edge := &Edge{
		from:   from,
		to:     to,
		pConst: op,
		lOp:    lOp,
	}

	to.pendingEdge++

	for i := range tk {
		edge.exp.push(tk[i])
	}

	from.out = append(from.out, edge)
}

func (g *BasicGraph) Add(id string, task Task) *BasicVertex {
	v := &BasicVertex{
		id:    id,
		task:  task,
		state: Pending,
	}

	g.vertex = append(g.vertex, v)
	return v
}

func (g *BasicGraph) GetVertex(id string) *BasicVertex {
	for i := range g.vertex {
		if g.vertex[i].id == id {
			return g.vertex[i]
		}
	}
	return nil
}
