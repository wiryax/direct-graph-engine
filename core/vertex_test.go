package dge

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func newAtomic64(v int64) *atomic.Int64 {
	a := atomic.Int64{}
	a.Add(v)
	return &a
}

type mockBuffTask struct {
	readBuff [][]Variable
	dataLen  int
}

func (m *mockBuffTask) TransformerTask(buffReader ReadOnlyBuffer, buffWriter WriteOnlyBuffer, _ *GraphContext) error {
	var (
		wg  sync.WaitGroup
		err error
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = errors.Join(m.ProducerTask(buffWriter, nil))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = errors.Join(m.ConsumerTask(buffReader, nil))
	}()
	wg.Wait()
	return err
}

func (m *mockBuffTask) ConsumerTask(buff ReadOnlyBuffer, _ *GraphContext) error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	ch := buff.Read()
	for item := range ch {
		m.readBuff = append(m.readBuff, item)
	}
	return nil
}

func (m *mockBuffTask) ProducerTask(buff WriteOnlyBuffer, _ *GraphContext) error {
	for range m.dataLen {
		if err := buff.WriteBuff([]Variable{
			{
				code: 0,
				raw:  []byte("1"),
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func TestNotifyVertex(t *testing.T) {
	v := vertexState{
		id:          "test notify",
		execStatus:  Pending,
		pendingEdge: *newAtomic64(3),
		failEdge:    *newAtomic64(0),
	}

	v.notify(Success100)
	v.notify(Success100)
	v.notify(Fail100)

	status := v.validate()

	assertEqual(t, int64(0), v.pendingEdge.Load(), "")
	assertEqual(t, int64(1), v.failEdge.Load(), "")
	assertEqual(t, Skipped100, status, "compare status")
}

func TestOverflowNotify(t *testing.T) {
	v := vertexState{
		id:          "test notify",
		execStatus:  Pending,
		pendingEdge: *newAtomic64(3),
		failEdge:    *newAtomic64(0),
	}

	v.notify(Success100)
	v.notify(Success100)
	v.notify(Success100)
	v.notify(Success100)

	status := v.validate()

	assertEqual(t, int64(0), v.pendingEdge.Load(), "")
	assertEqual(t, int64(0), v.failEdge.Load(), "")
	assertEqual(t, Ready, status, "")
}

func TestConsumerVertex(t *testing.T) {
	var (
		dataLen = 16
		wg      sync.WaitGroup
		buff    *BufferVariables
	)

	mockTask := &mockBuffTask{
		readBuff: make([][]Variable, 0, 8),
	}

	v := NewBufferConsumerVertex("mockTask", 5, mockTask)
	_ = v.GetBuffer()
	buff = v.buff
	buff.open()

	wg.Add(1)
	go func() {
		for range dataLen {
			assertShouldNotErr(t, buff.WriteBuff([]Variable{{
				code: 0,
				raw:  []byte{'0'},
			}}))
		}
		wg.Done()
		assertShouldNotErr(t, buff.done())
	}()

	v.preProcess(nil)
	v.process(nil)
	v.postProcess()

	wg.Wait()

	assertEqual(t, dataLen, len(mockTask.readBuff), "")
}

func TestVertexStateCycle(t *testing.T) {
	vA := newVertexState("A")

	vA.finish()
	vA.finish()

	d := <-vA.done()

	assertEqual(t, struct{}{}, d, "")
	assertEqual(t, true, vA.isFinish, "")

}

func TestTransformerVertex(t *testing.T) {
	var (
		buffReader = NewBufferVariables(5)
		buffWriter = NewBufferVariables(5)
		dataLen    = 100
		result     [][]Variable
		wg         sync.WaitGroup
	)

	buffReader.open()
	buffWriter.open()
	mock := &mockBuffTask{}

	transformer := NewBufferTransformerTask("", 5, mock)
	transformer.SetSenderBuffer(buffWriter)
	transformer.buff = buffReader

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range dataLen {
			assertShouldNotErr(t, buffReader.WriteBuff([]Variable{
				{
					code: 0,
					raw:  []byte{},
				},
			}))
		}
		assertShouldNotErr(t, buffReader.done())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for ch := range buffWriter.Read() {
			result = append(result, ch)
		}
	}()

	assertShouldNotErr(t, transformer.process(nil))
	transformer.postProcess()
	wg.Wait()

	assertEqual(t, dataLen, len(mock.readBuff), "")
}

func TestSingleProducerToSingleConsumer(t *testing.T) {
	var (
		wg   sync.WaitGroup
		mock = &mockBuffTask{
			dataLen: 100,
		}
	)

	consumer := NewBufferConsumerVertex("consumer", 10, mock)
	producer := NewBufferProducer("producer", mock)

	producer.SetSenderBuffer(consumer.GetBuffer())

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, producer.preProcess(nil))
		assertShouldNotErr(t, producer.process(nil))
		producer.postProcess()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, consumer.preProcess(nil))
		assertShouldNotErr(t, consumer.process(nil))
		consumer.postProcess()
	}()

	wg.Wait()

	assertEqual(t, mock.dataLen, len(mock.readBuff), "")
	assertEqual(t, true, consumer.buff.closed, "")
}

func TestMultipleProducerToSingleConsumer(t *testing.T) {
	var (
		wg   sync.WaitGroup
		mock = &mockBuffTask{
			dataLen: 100,
		}
		producer = 1
	)

	consumer := NewBufferConsumerVertex("consumer", 10, mock)

	for range producer {
		mockP := &mockBuffTask{
			dataLen: 100,
		}
		producer := NewBufferProducer("producer", mockP)
		producer.SetSenderBuffer(consumer.GetBuffer())
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer producer.postProcess()
			assertShouldNotErr(t, producer.preProcess(nil))
			assertShouldNotErr(t, producer.process(nil))
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer consumer.postProcess()
		assertShouldNotErr(t, consumer.preProcess(nil))
		assertShouldNotErr(t, consumer.process(nil))
	}()

	wg.Wait()

	assertEqual(t, mock.dataLen*producer, len(mock.readBuff), "")
	assertEqual(t, true, consumer.buff.closed, "")
}

func TestSingleProducerToSingleConsumerToSingleTransformer(t *testing.T) {
	var (
		wg   sync.WaitGroup
		mock = &mockBuffTask{
			dataLen: 3,
		}
	)

	consumer := NewBufferConsumerVertex("consumer", 10, mock)
	producer := NewBufferProducer("producer", mock)
	transformer := NewBufferTransformerTask("transformer", 10, mock)

	producer.SetSenderBuffer(transformer.GetBuffer())
	transformer.SetSenderBuffer(consumer.GetBuffer())

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, producer.preProcess(nil))
		assertShouldNotErr(t, producer.process(nil))
		producer.postProcess()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, consumer.preProcess(nil))
		assertShouldNotErr(t, consumer.process(nil))
		consumer.postProcess()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, transformer.preProcess(nil))
		assertShouldNotErr(t, transformer.process(nil))
		transformer.postProcess()
	}()

	wg.Wait()

	assertEqual(t, mock.dataLen*2, len(mock.readBuff), "")
	assertEqual(t, true, consumer.buff.closed, "")
	assertEqual(t, true, transformer.buff.closed, "")
}
