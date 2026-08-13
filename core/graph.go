package dge

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

var ErrCyclicRelation = errors.New("cannot add cyclic relationship")
var ErrDuplicateRelation = errors.New("duplicate relation")
var ErrOutsideScopeRelation = errors.New("cannot connect add relation outside relation")
var ErrMissingRelation = errors.New("cannot connect empty relations")

type Relation struct {
	from,
	to vertex
	kind Edge
}

type Graph struct {
	vertexState
	vertices   map[vertex]struct{}
	edge       []*Edge
	vertexPool chan vertex
	donePool   chan vertex
}

func NewGraph(id string) *Graph {
	return &Graph{
		vertexState: *newVertexState(id),
		vertices:    make(map[vertex]struct{}),
	}
}

func (g *Graph) preProcess(gCtx *GraphContext) error {
	for e := range g.edge {
		if err := g.edge[e].wired(gCtx); err != nil {
			return err
		}
	}

	root := g.getRoot()
	g.vertexPool = make(chan vertex, len(g.vertices))
	g.donePool = make(chan vertex, len(g.vertices))
	for i := range root {
		g.vertexPool <- root[i]
	}
	return nil
}

func (g *Graph) process(gCtx *GraphContext) error {
	var (
		counter int
		wg      sync.WaitGroup
	)

	for counter < len(g.vertices) {
		select {
		case v := <-g.vertexPool:
			err := v.preProcess(gCtx)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				v.setExecutionStatus(Running)
				err = v.process(gCtx)
				if err != nil {
					v.setExecutionStatus(Fail)
				} else {
					v.setExecutionStatus(Success)
				}
				v.postProcess()
				g.donePool <- v
			}()
			g.getReadyVertex(gCtx, v)
		case <-g.donePool:
			counter++
		}
	}
	wg.Wait()

	return nil
}

func (g *Graph) postProcess() {
	g.finish()
	close(g.donePool)
	close(g.vertexPool)
}

func (g *Graph) RunWithContext(gCtx *GraphContext) error {
	defer g.postProcess()
	err := g.preProcess(gCtx)
	if err != nil {
		return err
	}
	return g.process(gCtx)
}

func (g *Graph) Run() error {
	gCtx := NewGraphContext(context.Background(), slog.With("graph_id", g.id), nil)
	return g.RunWithContext(gCtx)
}

func (e *Graph) getRoot() []vertex {
	var queue []vertex
	for v := range e.vertices {
		if v.validate() == Ready {
			queue = append(queue, v)
		}
	}
	return queue
}

func (e *Graph) getReadyVertex(gCtx *GraphContext, v vertex) {
	for _, child := range v.Child() {
		go func() {
			vChild := child.EvaluateConstrain(gCtx)
			if vChild != nil {
				e.vertexPool <- vChild
			}
		}()
	}
}

func (e *Graph) isContain(v vertex) bool {
	_, ok := e.vertices[v]
	return ok
}

func (e *Graph) Connect(r *Edge) error {
	if r.In() == nil && r.Out() == nil {
		return ErrMissingRelation
	}

	if r.Out() != nil && r.In() == r.Out() {
		return ErrCyclicRelation
	}

	if err := isCyclicOrDuplicate(r, e.edge); err != nil {
		return err
	}

	for v := range e.vertices {
		if g, ok := v.(*Graph); ok {
			if g.isContain(r.Out()) {
				return ErrOutsideScopeRelation
			}
		}
	}

	e.edge = append(e.edge, r)
	r.In().addChild(r)
	e.vertices[r.In()] = struct{}{}
	if r.Out() != nil {
		r.Out().addParent(r)
		e.vertices[r.Out()] = struct{}{}
	}
	return nil
}

func isCyclicOrDuplicate(target *Edge, relation []*Edge) error {
	for _, e := range relation {
		if e.In() == target.In() && e.Out() == target.Out() {
			return ErrDuplicateRelation
		}

		if e.In() == target.Out() {
			return ErrCyclicRelation
		}
	}
	return nil
}
