package dge

import (
	"sync"
	"testing"
)

func TestBuffReadAndWriteCycle(t *testing.T) {
	var (
		buff    = NewBufferVariables(4)
		dataLen = 100
		result  [][]Variable
		wg      sync.WaitGroup
	)

	buff.open()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range dataLen {
			assertShouldNotErr(t, buff.WriteBuff([]Variable{
				{
					code: 0,
					raw:  []byte{},
				},
			}))
		}
		assertShouldNotErr(t, buff.done())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ch := buff.Read()
		for item := range ch {
			result = append(result, item)
		}
	}()

	wg.Wait()
	assertEqual(t, dataLen, len(result), "")
}
