package dge

import (
	"context"
	"sync"
	"sync/atomic"
)

type TaskResult int

const (
	Pending TaskResult = iota
	Running
	Success
	Fail
	Skipped
)

type evalResult int

const (
	Ready evalResult = iota
	Pending100
	Skip
)

type vertex interface {
	processor
	ID() string
	State() TaskResult
	Child() []*Edge
	Parents() []*Edge
	validate() evalResult
	setExecutionStatus(TaskResult)
	addChild(*Edge)
	addParent(*Edge)
	notify(evaluateStatus)
	done() <-chan struct{}
}

type processor interface {
	process(gCtx *GraphContext) error
	preProcess(*GraphContext) error
	postProcess()
}

type vertexState struct {
	id          string
	isFinish    bool
	execStatus  TaskResult
	pendingEdge atomic.Int64
	failEdge    atomic.Int64
	doneSign    chan struct{}
	in          []*Edge
	out         []*Edge
}

func newVertexState(id string) *vertexState {
	return &vertexState{
		id:       id,
		doneSign: make(chan struct{}),
	}
}

func (v *vertexState) ID() string {
	context.Background()
	return v.id
}

func (v *vertexState) State() TaskResult {
	return v.execStatus
}

func (v *vertexState) Child() []*Edge {
	return v.out
}

func (v *vertexState) Parents() []*Edge {
	return v.Parents()
}

func (v *vertexState) addChild(child *Edge) {
	v.out = append(v.out, child)
}

func (v *vertexState) addParent(parent *Edge) {
	v.pendingEdge.Add(1)
	v.in = append(v.in, parent)
}

func (v *vertexState) setExecutionStatus(s TaskResult) {
	v.execStatus = s
}

func (v *vertexState) notify(e evaluateStatus) {
	if v.pendingEdge.Load() == 0 {
		return
	}

	if e == Compilation100 {
		v.pendingEdge.Store(0)
		v.failEdge.Store(0)
		return
	}

	if e == Fail100 {
		v.failEdge.Add(1)
	}

	v.pendingEdge.Add(-1)
}

func (v *vertexState) validate() evalResult {
	pending := v.pendingEdge.Load()
	fail := v.failEdge.Load()
	if pending == 0 && fail == 0 {
		return Ready
	}

	if pending > 0 {
		return Pending100
	}

	return Skip
}

func (v *vertexState) finish() {
	if v.isFinish {
		return
	}
	v.isFinish = true
	close(v.doneSign)
}

func (v *vertexState) done() <-chan struct{} {
	return v.doneSign
}

type Task interface {
	RunTask(*GraphContext) error
}

type BasicVertex struct {
	*vertexState
	task Task
}

func NewVertexState(id string, task Task) *BasicVertex {
	return &BasicVertex{
		vertexState: newVertexState(id),
		task:        task,
	}
}

func (b *BasicVertex) preProcess(gCtx *GraphContext) error {
	return nil
}

func (b *BasicVertex) process(gCtx *GraphContext) error {
	return b.task.RunTask(gCtx)
}

func (b *BasicVertex) postProcess() {
	b.finish()
}

type WriteOnlyBufferVertex interface {
	vertex
	SetSenderBuffer(WriteOnlyBuffer)
}

type ReadOnlyBufferVertex interface {
	vertex
	GetBuffer() WriteOnlyBuffer
}

type BufferProducerTask interface {
	ProducerTask(*GraphContext, WriteOnlyBuffer) error
}

type BufferProducerVertex struct {
	*vertexState
	outBuff WriteOnlyBuffer
	task    BufferProducerTask
}

func NewBufferProducer(id string, task BufferProducerTask) *BufferProducerVertex {
	return &BufferProducerVertex{
		vertexState: newVertexState(id),
		task:        task,
	}
}

func (bp *BufferProducerVertex) SetSenderBuffer(buff WriteOnlyBuffer) {
	bp.outBuff = buff
}

func (bp *BufferProducerVertex) preProcess(gCtx *GraphContext) error {
	bp.outBuff.open()
	return nil
}

func (bp *BufferProducerVertex) process(gCtx *GraphContext) error {
	return bp.task.ProducerTask(gCtx, bp.outBuff)
}

func (bp *BufferProducerVertex) postProcess() {
	bp.finish()
	bp.outBuff.done()
}

type BufferConsumerTask interface {
	ConsumerTask(*GraphContext, ReadOnlyBuffer) error
}

type BufferConsumerVertex struct {
	*vertexState
	buffSize int
	task     BufferConsumerTask
	buff     *BufferVariables
	once     sync.Once
}

func NewBufferConsumerVertex(id string, buffSize int, task BufferConsumerTask) *BufferConsumerVertex {
	return &BufferConsumerVertex{
		task:        task,
		buffSize:    buffSize,
		vertexState: newVertexState(id),
	}
}

func (bc *BufferConsumerVertex) GetBuffer() WriteOnlyBuffer {
	bc.once.Do(func() {
		bc.buff = NewBufferVariables(bc.buffSize)
	})
	return bc.buff
}

func (bc *BufferConsumerVertex) preProcess(gCtx *GraphContext) error {
	return nil
}

func (bc *BufferConsumerVertex) process(gCtx *GraphContext) error {
	return bc.task.ConsumerTask(gCtx, bc.buff)
}

func (bc *BufferConsumerVertex) postProcess() {
	bc.finish()
}

type BufferTransformerTask interface {
	TransformerTask(*GraphContext, ReadOnlyBuffer, WriteOnlyBuffer) error
}

type BufferTransformerVertex struct {
	*vertexState
	buffOutSize int
	outBuff     WriteOnlyBuffer
	buff        *BufferVariables
	once        sync.Once
	task        BufferTransformerTask
}

func NewBufferTransformerTask(id string, buffOutSize int, task BufferTransformerTask) *BufferTransformerVertex {
	return &BufferTransformerVertex{
		task:        task,
		vertexState: newVertexState(id),
		buffOutSize: buffOutSize,
	}
}

func (bt *BufferTransformerVertex) GetBuffer() WriteOnlyBuffer {
	bt.once.Do(func() {
		bt.buff = NewBufferVariables(bt.buffOutSize)
	})
	return bt.buff
}

func (bt *BufferTransformerVertex) SetSenderBuffer(buff WriteOnlyBuffer) {
	bt.outBuff = buff
}

func (bt *BufferTransformerVertex) preProcess(gCtx *GraphContext) error {
	bt.outBuff.open()
	return nil
}

func (bt *BufferTransformerVertex) process(gCtx *GraphContext) error {
	return bt.task.TransformerTask(gCtx, bt.buff, bt.outBuff)
}

func (bt *BufferTransformerVertex) postProcess() {
	bt.finish()
	bt.outBuff.done()
}
