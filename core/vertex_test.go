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
	readBuff      [][]Variable
	dataLen       int
	consumerErrAt int
	producerErrAt int
}

func (m *mockBuffTask) TransformerTask(_ *GraphContext, buffReader ReadOnlyBuffer, buffWriter WriteOnlyBuffer) error {
	var (
		wg  sync.WaitGroup
		err error
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = errors.Join(m.ProducerTask(nil, buffWriter))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = errors.Join(m.ConsumerTask(nil, buffReader))
	}()
	wg.Wait()
	return err
}

func (m *mockBuffTask) ConsumerTask(_ *GraphContext, buff ReadOnlyBuffer) error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	ch := buff.Read()
	i := 0
	for item := range ch {
		if m.consumerErrAt == i {
			return mockErr
		}
		m.readBuff = append(m.readBuff, item)
		i++
	}
	return nil
}

func (m *mockBuffTask) ProducerTask(_ *GraphContext, buff WriteOnlyBuffer) error {
	i := 0
	for range m.dataLen {
		if m.producerErrAt == i {
			return mockErr
		}
		if err := buff.WriteBuff([]Variable{
			{
				code: 0,
				raw:  []byte("1"),
			},
		}); err != nil {
			return err
		}
		i++
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
	assertEqual(t, Skip, status, "compare status")
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
		consumerErrAt: -1,
		producerErrAt: -1,
		readBuff:      make([][]Variable, 0, 8),
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
	mock := &mockBuffTask{
		consumerErrAt: -1,
		producerErrAt: -1,
	}

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
			consumerErrAt: -1,
			producerErrAt: -1,
			dataLen:       100,
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
			consumerErrAt: -1,
			producerErrAt: -1,
			dataLen:       100,
		}
		producer = 1
	)

	consumer := NewBufferConsumerVertex("consumer", 10, mock)

	for range producer {
		mockP := &mockBuffTask{
			consumerErrAt: -1,
			producerErrAt: -1,
			dataLen:       100,
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
			consumerErrAt: -1,
			producerErrAt: -1,
			dataLen:       3,
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

	assertEqual(t, mock.dataLen, len(mock.readBuff), "")
	assertEqual(t, true, consumer.buff.closed, "")
	assertEqual(t, true, transformer.buff.closed, "")
}

func TestTransformerTaskFail(t *testing.T) {
	mockTransformer := &mockBuffTask{
		dataLen:       100,
		consumerErrAt: 9,
		producerErrAt: 10,
	}

	mockProducer := &mockBuffTask{
		dataLen:       100,
		producerErrAt: 10,
	}

	mockConsumer := &mockBuffTask{
		dataLen:       100,
		consumerErrAt: 9,
	}

	producer := NewBufferProducer("", mockProducer)
	transformer := NewBufferTransformerTask("", 10, mockTransformer)
	consumer := NewBufferConsumerVertex("", 10, mockConsumer)

	producer.SetSenderBuffer(transformer.GetBuffer())
	transformer.SetSenderBuffer(consumer.GetBuffer())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, producer.preProcess(nil))
		assertShouldErr(t, producer.process(nil), mockErr)
		producer.postProcess()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, consumer.preProcess(nil))
		assertShouldErr(t, consumer.process(nil), mockErr)
		consumer.postProcess()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		assertShouldNotErr(t, transformer.preProcess(nil))
		assertShouldErr(t, transformer.process(nil), mockErr)
		transformer.postProcess()
	}()

	wg.Wait()

	assertEqual(t, true, transformer.GetBuffer().(*BufferVariables).closed, "")
	assertEqual(t, true, consumer.GetBuffer().(*BufferVariables).closed, "")
}
