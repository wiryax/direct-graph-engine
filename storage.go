package dge

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"
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

func NewBlob(code int, data []byte) Blob {
	return Blob{
		code: code,
		raw:  data,
	}
}

func (b *Blob) GetRaw() []byte {
	return b.raw
}

func (b *Blob) GetBlobCode() int {
	return b.code
}

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

type Column struct {
	name string
	data []Variable
}

func (c Column) GetAllData() []Variable {
	return slices.Clone(c.data)
}

func (c Column) GetColumnName() string {
	return c.name
}

func (c Column) GetFirst() (Variable, error) {
	if len(c.data) == 0 {
		return Variable{}, errors.New("empty column")
	}
	return c.data[0], nil
}

type Tabular struct {
	columns []Column
}

func MakeTabular() *Tabular {
	return &Tabular{}
}

func (t *Tabular) Clone() Tabular {
	newT := MakeTabular()
	for i := range t.CountColumns() {
		newT.AddOrSetColumn(t.columns[i].name, t.columns[i].data...)
	}
	return *newT
}

func (t *Tabular) CloneStructure() Tabular {
	var col []Column
	for _, c := range t.columns {
		col = append(col, Column{
			name: c.name,
		})
	}
	return Tabular{
		columns: col,
	}
}

func (t *Tabular) CountRows() int {
	if len(t.columns) == 0 {
		return 0
	}

	if !t.isCartesian() {
		return len(t.columns[0].data)
	}

	total := 1
	for i := range t.columns {
		total *= len(t.columns[i].data)
	}
	return total
}

func (t *Tabular) CountColumns() int {
	return len(t.columns)
}

func (t *Tabular) GetColumnIndex(name string) int {
	for i := range t.columns {
		if t.columns[i].name == name {
			return i
		}
	}

	return -1
}

func (t *Tabular) AddOrSetColumn(name string, data ...Variable) {
	if t.AddData(Column{name: name, data: data}) == nil {
		return
	}
	t.columns = append(t.columns, Column{name: name, data: data})
}

func (t *Tabular) AddData(data Column) error {
	var index int
	if index = t.GetColumnIndex(data.name); index == -1 {
		return fmt.Errorf("column with %s key not exist", data.name)
	}

	t.columns[index].data = append(t.columns[index].data, data.data...)
	return nil
}

func (t *Tabular) GetRows(ri int) ([]Column, error) {
	var temp []Column
	for ci := range t.columns {
		cell, err := t.GetCell(ci, ri)
		if err != nil {
			return nil, err
		}

		temp = append(temp, Column{
			name: t.columns[ci].name,
			data: []Variable{cell},
		})
	}
	return temp, nil
}

func (t *Tabular) GetCell(ci, ri int) (Variable, error) {
	if (ci < 0 || ri < 0) || ci > len(t.columns) || ri > t.CountRows() {
		return Variable{}, fmt.Errorf("invalid index")
	}
	if t.isCartesian() {
		return t.cartesianProduct(ci, ri)
	}
	return t.pairedRows(ci, ri)
}

func (t *Tabular) isCartesian() bool {
	if len(t.columns) == 0 {
		return false
	}

	firstColLen := len(t.columns[0].data)
	for _, c := range t.columns {
		if len(c.data) != firstColLen {
			return true
		}
	}
	return false
}

func (t *Tabular) pairedRows(ci, ri int) (Variable, error) {
	return t.columns[ci].data[ri], nil
}

func (t *Tabular) cartesianProduct(ci, ri int) (Variable, error) {
	c := t.columns[ci]
	clen := len(c.data)

	divisor := 1
	for i := 0; i < ci; i++ {
		divisor *= len(t.columns[i].data)
	}
	actual := (ri / divisor) % clen
	return c.data[actual], nil
}

func (t Tabular) String() string {
	var buff bytes.Buffer
	w := tabwriter.NewWriter(&buff, 1, 1, 3, ' ', 0)

	if len(t.columns) > 0 {
		var columns []string
		for _, c := range t.columns {
			columns = append(columns, c.name)
		}
		fmt.Fprintln(w, strings.Join(columns, "\t"))
	} else {
		return ""
	}

	for ri := 0; ri < t.CountRows(); ri++ {
		for ci := range t.columns {
			v, _ := t.GetCell(ci, ri)
			w.Write(v.raw)
			w.Write([]byte("\t"))
		}
		w.Write([]byte("\n"))
	}

	w.Flush()
	return "\n" + buff.String() + "\n"
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

	return t.tabular.Clone(), nil
}

func (s *Storage) SetBlob(key string, data Blob) {
	s.item[key] = StorageItem{
		key:   key,
		sType: TypeBlob,
		blob:  data,
	}
}

func (s *Storage) GetBlob(key string) (Blob, error) {
	si, ok := s.item[key]
	if !ok {
		return Blob{}, fmt.Errorf("cannot find storage item with key %s", key)
	}

	if si.sType != TypeBlob {
		return Blob{}, fmt.Errorf("storage item with key %s are not blob type", key)

	}
	return si.blob, nil
}
