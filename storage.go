package dge

import (
	"fmt"
)

type StorageType int

const (
	TypeTabular StorageType = iota
	TypeBlob
)

type StorageItem struct {
	key     string
	sType   StorageType
	blob    Blob
	tabular Tabular
}

type VariableType int

const (
	VNull VariableType = iota
	VRaw
	VInt
	VString
	VBool
)

type Blob struct {
	code int
	raw  []byte
}

type Variable struct {
	code int
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
		code: int(VRaw),
		raw:  b,
	}
}

type Tabular struct {
	rows   [][]Variable
	column []string
}

func MakeTabular(column []string) *Tabular {
	return &Tabular{
		rows:   make([][]Variable, 0),
		column: column,
	}
}

func (t *Tabular) AddRow(v ...Variable) error {
	if len(t.column) != len(v) {
		return fmt.Errorf("column not match")
	}
	t.rows = append(t.rows, v)
	return nil
}

func (t *Tabular) GetColIndex(key string) int {
	for i := range t.column {
		if t.column[i] == key {
			return i
		}
	}

	return -1
}

func (t *Tabular) FilterTabular(key string, fn func(v []Variable) bool) Tabular {
	var temp [][]Variable
	for i := range t.rows {
		if fn(t.rows[i]) {
			temp = append(temp, t.rows[i])
		}
	}

	return Tabular{
		rows:   temp,
		column: t.column,
	}
}

type Storage struct {
	item map[string]StorageItem
}

func NewStorage() *Storage {
	return &Storage{
		item: make(map[string]StorageItem),
	}
}

func (s *Storage) SetTabular(key string, data Tabular) {
	s.item[key] = StorageItem{
		key:     key,
		sType:   TypeTabular,
		tabular: data,
	}
}

func (s *Storage) GetTabular(key string) (Tabular, error) {
	t, ok := s.item[key]
	if !ok {
		return Tabular{}, fmt.Errorf("cannot find data with key %s", key)
	}

	if t.sType != TypeTabular {
		return Tabular{}, fmt.Errorf("cannot find tabular data with key %s", key)
	}

	return t.tabular, nil
}

func (s *Storage) SetBlob(key string, data []byte) {
	s.item[key] = StorageItem{
		key:   key,
		sType: TypeBlob,
		blob: Blob{
			code: 0,
			raw:  data,
		},
	}
}

func (s *Storage) GetBlob(key string) []byte {
	return s.item[key].blob.raw
}
