package dge

import (
	"context"
	"errors"
	"log/slog"
)

type Engine struct {
	conn     map[string]Connection
	variable map[string]Variable
	cancelFn context.CancelFunc
	ctx      context.Context
	graph    *Graph
}

func NewEngine(g *Graph) *Engine {
	ctx, cancelFn := context.WithCancel(context.Background())
	return &Engine{
		cancelFn: cancelFn,
		ctx:      ctx,
		graph:    g,
	}
}

func NewEngineWithContext(ctx context.Context, g *Graph) *Engine {
	return &Engine{
		ctx:   ctx,
		graph: g,
	}
}

func (e *Engine) SetVariable(variable map[string]Variable) {
	e.variable = variable
}

func (e *Engine) SetConnection(conn map[string]Connection) {
	e.conn = conn
}

func (e *Engine) Run() error {
	var err error
	gCtx := NewGraphContext(e.ctx, slog.Default(), e.conn)
	defer func(cause error) {
		err = errors.Join(gCtx.releaseConnections())
	}(err)

	err = gCtx.validateConnections()
	if err != nil {
		return errors.Join(err)
	}
	err = e.graph.RunWithContext(gCtx)
	if err != nil {
		return errors.Join(err)
	}

	return nil
}
