package dge

import "log/slog"

type GraphContext struct {
	rState  *RuntimeState
	storage *Storage
	Log     *slog.Logger
}

func NewGraphWithLogContext(gLog *slog.Logger, rState *RuntimeState, storage *Storage) *GraphContext {
	return &GraphContext{
		rState:  rState,
		storage: storage,
		Log:     gLog,
	}
}

func (gCtx *GraphContext) WithVertex(id string) *GraphContext {
	return gCtx.withNewLog("vertex_id", id)
}

func (gCtx *GraphContext) WithGraph(id string) *GraphContext {
	return gCtx.withNewLog("graph_id", id)
}

func (gCtx *GraphContext) withNewLog(newField, id string) *GraphContext {
	return &GraphContext{
		rState:  gCtx.rState,
		storage: gCtx.storage,
		Log:     gCtx.Log.With(newField, id),
	}
}

func (gCtx *GraphContext) GetVariable(key string) (string, error) {
	return gCtx.rState.GetVariable(key)
}

func (gCtx *GraphContext) SetVariable(key, value string) {
	gCtx.rState.SetVariable(key, value)
}

func (gCtx *GraphContext) SetTabularStorage(key string, data Tabular) {
	gCtx.storage.SetTabular(key, data)
}

func (gCtx *GraphContext) GetTabularStorage(key string) (Tabular, error) {
	return gCtx.storage.GetTabular(key)
}

func (gCtx *GraphContext) GetBlob(key string) (Blob, error) {
	return gCtx.storage.GetBlob(key)
}

func (gCtx *GraphContext) SetBlob(key string, b Blob) {
	gCtx.storage.SetBlob(key, b)
}
