package dge

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
)

type StorageType int

const (
	TypeTabular StorageType = iota
	TypeBlob
)

type VariableType int

const (
	VNull VariableType = iota
	VRaw
	VInt
	VString
	VTrue
	VFalse
)

type Variable struct {
	code VariableType
	raw  []byte
}

func (v *Variable) String() string {
	return string(v.raw)
}

func (v *Variable) GetRaw() []byte {
	return v.raw
}

func ParseVariable(b []byte) Variable {
	return Variable{
		code: VRaw,
		raw:  b,
	}
}

func NewIntVar(data int) (Variable, error) {
	b := new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, data); err != nil {
		return Variable{}, err
	}

	return Variable{
		code: VInt,
		raw:  b.Bytes(),
	}, nil
}

type ReadOnlyBuffer interface {
	Read() <-chan []Variable
}

type WriteOnlyBuffer interface {
	WriteBuff([]Variable) error
	open()
	done() error
}

type BufferVariables struct {
	closed bool
	buff   chan []Variable
	once   sync.Once
	wg     sync.WaitGroup
}

func NewBufferVariables(size int) *BufferVariables {
	return &BufferVariables{
		buff: make(chan []Variable, size),
	}
}

func (b *BufferVariables) open() {
	b.wg.Add(1)
	go b.once.Do(func() {
		b.wg.Wait()
		close(b.buff)
		b.closed = true
	})
}

func (b *BufferVariables) WriteBuff(v []Variable) error {
	if b.closed {
		return errors.New("cannot write to closed buffer")
	}
	b.buff <- v
	return nil
}

func (b *BufferVariables) Read() <-chan []Variable {
	return b.buff
}

func (b *BufferVariables) done() error {
	if b.closed {
		return errors.New("buffer already closed")
	}
	b.wg.Done()
	return nil
}
