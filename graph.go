package dge

import (
	"fmt"
	"log/slog"
)

type Task interface {
	Execute(gCtx *GraphContext) error
}

type graph interface {
	Run()
	RunWithContext(gCtx *GraphContext)
	Copy() graph
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

func (g *BasicGraph) Copy() graph {
	newGraph := &BasicGraph{
		id: g.id,
	}

	for i := range g.vertex {
		newGraph.Add(g.vertex[i].id, g.vertex[i].task)
	}

	for i := range g.vertex {
		for j := range g.vertex[i].in {
			newGraph.Connect(g.vertex[i].in[j].from, g.vertex[i].in[j].from, g.vertex[i].in[j].pConst, g.vertex[i].in[j].lOp, g.vertex[i].in[j].exp.tokens)
		}

		for j := range g.vertex[i].out {
			newGraph.Connect(g.vertex[i].out[j].from, g.vertex[i].out[j].from, g.vertex[i].out[j].pConst, g.vertex[i].out[j].lOp, g.vertex[i].out[j].exp.tokens)
		}
	}

	return newGraph

}

func (g *BasicGraph) RunWithContext(gCtx *GraphContext) {
	g.run(gCtx)
}

func (g *BasicGraph) Run() {
	rState := &RuntimeState{
		variable: make(map[string]string),
	}

	storage := NewStorage()

	gCtx := NewGraphWithLogContext(slog.With("graph_id", g.id), rState, storage)
	g.run(gCtx)
}

func (g *BasicGraph) run(gCtx *GraphContext) {
	if gCtx == nil {
		panic("graph context cannot nil")
	}

	queue := getRoot(g.vertex)

	execute(gCtx, queue)
}

func getRoot(vertex []*BasicVertex) []*BasicVertex {
	var queue []*BasicVertex
	for _, v := range vertex {
		if v.pendingEdge == 0 && v.failEdge == 0 {
			queue = append(queue, v)
		}
	}
	return queue
}

func execute(gCtx *GraphContext, queue []*BasicVertex) {
	for {
		if len(queue) == 0 {
			break
		}

		v := queue[0]
		queue = queue[1:]
		v.state = Running

		childCtx := gCtx.WithVertex(v.GetId())
		err := v.ExecuteTask(childCtx)
		if err != nil {
			gCtx.Log.Error(err.Error(), "vertex_id", v.id)
			v.state = Fail
		} else {
			gCtx.Log.Info("success execute vertex", "vertex_id", v.id)
			v.state = Success
		}
		getReadyVertex(gCtx, v, &queue)
	}
}

func getReadyVertex(gCtx *GraphContext, v *BasicVertex, queue *[]*BasicVertex) {
	for _, child := range v.out {
		if child.from.state == Pending {
			continue
		}
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

		getReadyVertex(gCtx, child.to, queue)

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

type TabularLoop struct {
	id               string
	tabularStorageId string
	graph            graph
	vState           state
	maxLoop          int
}

func NewTabularLoop(id string, storageId string, maxLoop int) *TabularLoop {
	return &TabularLoop{
		id:               id,
		tabularStorageId: storageId,
		vState:           Pending,
		maxLoop:          maxLoop,
	}
}

func (t *TabularLoop) ExecuteTask(gCtx *GraphContext) error {
	tabular, err := gCtx.GetTabularStorage(t.tabularStorageId)
	if err != nil {
		return err
	}

	t.vState = Running
	defer func() {
		if err != nil {
			t.vState = Fail
		} else {
			t.vState = Success
		}
	}()

	t.RunWithTabular(gCtx, tabular)
	return nil
}

func (t *TabularLoop) RunWithContext(gCtx *GraphContext) {
	if gCtx == nil {
		panic("graph context cannot nil")
	}
	t.ExecuteTask(gCtx)
}

func (t *TabularLoop) RunWithTabular(gCtx *GraphContext, tabular Tabular) {
	var (
		state = make(map[string]string)
	)

	for i := 0; i < t.maxLoop && i < tabular.CountRows(); i++ {
		var childCtx *GraphContext
		for ci := range tabular.columns {
			cell, err := tabular.GetCell(ci, i)
			if err != nil {
				gCtx.Log.Error("error while mapping tabular data:" + err.Error())
				break
			}
			state[tabular.columns[ci].name] = cell.String()
		}

		rState := NewRuntimeState(state)
		storage := NewStorage()
		childCtx = NewGraphWithLogContext(slog.With("sub_graph", t.id, "iteration", i), rState, storage)
		g := t.graph.Copy()
		g.RunWithContext(childCtx)
	}
}
