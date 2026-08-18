package dge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var ErrUndefinedVariable = errors.New("variable is undefined")
var ErrUndefinedConnection = errors.New("connection is undefined")

type GraphContext struct {
	sync.RWMutex
	Log      *slog.Logger
	variable map[string]Variable
	conn     map[string]Connection
	context.Context
}

func NewGraphContext(ctx context.Context, gLog *slog.Logger, conn map[string]Connection) *GraphContext {
	gCtx := &GraphContext{
		Log:     gLog,
		conn:    conn,
		Context: ctx,
	}
	return gCtx
}

func (gCtx *GraphContext) releaseConnections() error {
	for k := range gCtx.conn {
		if err := gCtx.conn[k].Release(); err != nil {
			return err
		}
	}

	return nil
}
func (gCtx *GraphContext) validateConnections() error {
	for k := range gCtx.conn {
		if err := gCtx.conn[k].Validate(gCtx); err != nil {
			return err
		}
	}

	return nil
}

func (gCtx *GraphContext) GetConnection(key string) (Acquirer, error) {
	gCtx.Lock()
	defer gCtx.Unlock()

	conn, ok := gCtx.conn[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUndefinedConnection, key)
	}
	return conn, nil
}

func (gCtx *GraphContext) SetVariable(key string, v Variable) {
	gCtx.Lock()
	defer gCtx.Unlock()

	gCtx.variable[key] = v
}

func (gCtx *GraphContext) GetVariable(key string) (Variable, error) {
	gCtx.Lock()
	defer gCtx.Unlock()

	v, ok := gCtx.variable[key]
	if !ok {
		return Variable{}, fmt.Errorf("%w: %s", ErrUndefinedVariable, key)
	}
	return v, nil
}
