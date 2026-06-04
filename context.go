package dge

type GraphContext struct {
	rState  *RuntimeState
	logger  GraphLogger
	storage *Storage
}

func NewGraphContext(gLog GraphLogger, rState *RuntimeState, storage *Storage) *GraphContext {
	return &GraphContext{
		logger:  gLog,
		rState:  rState,
		storage: storage,
	}
}

func (gCtx *GraphContext) GetVariable(key string) (string, error) {
	return gCtx.rState.GetVariable(key)
}

func (gCtx *GraphContext) SetVariable(key, value string) {
	gCtx.rState.SetVariable(key, value)
}

func (gCtx *GraphContext) Log(et EvenType, logLv LogLevel, msg, vId, gId string) {
	gCtx.logger.FlushLog(et, logLv, msg, vId, gId)
}

func (gCtx *GraphContext) SetTabularStorage(key string, data Tabular) {
	gCtx.storage.SetTabular(key, data)
}

func (gCtx *GraphContext) GetTabularStorage(key string) (Tabular, error) {
	return gCtx.storage.GetTabular(key)
}
