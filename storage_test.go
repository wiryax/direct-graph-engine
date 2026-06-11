package dge

import (
	"reflect"
	"testing"
)

func makeBulkVars(vType VariableType, rawData ...[]byte) []Variable {
	var r []Variable
	for _, rd := range rawData {
		r = append(r, Variable{vType, rd})
	}
	return r
}

func TestAddTabularCols(t *testing.T) {
	expected := MakeTabular([]string{"A", "B", "C", "D"})
	err := expected.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"), []byte("D"))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := MakeTabular([]string{"A", "B", "C"})
	err = result.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = result.AddColumn(func(rows []Variable) [][]Variable {
		var temp [][]Variable
		rows = append(rows, Variable{
			code: VString,
			raw:  []byte("D"),
		})
		temp = append(temp, rows)
		return temp
	}, "D")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("unexpected result. want %v, got %v", expected, result)
	}
}

func TestAddTabularWithDifferentLength(t *testing.T) {
	expected := MakeTabular([]string{"A", "B", "C", "D"})
	expected.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"), []byte("D"))...)

	result := MakeTabular([]string{"A", "B", "C"})
	result.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"))...)

	err := result.AddColumn(func(_ []Variable) [][]Variable {
		return [][]Variable{{{
			code: 0,
			raw:  []byte("D"),
		}}}
	}, "D")

	if err == nil {
		t.Errorf("unexpected result, should be err")
	}
}

func TestJoinTabular(t *testing.T) {
	expected := Tabular{
		column: []string{"a", "b", "c", "d", "e"},
	}

	err := expected.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"), []byte("D"), []byte("E"))...)
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}

	t1 := MakeTabular([]string{"a", "b", "c"})
	err = t1.AddRow(makeBulkVars(VString, []byte("A"), []byte("B"), []byte("C"))...)

	t2 := MakeTabular([]string{"d", "e"})
	err = t2.AddRow(makeBulkVars(VString, []byte("D"), []byte("E"))...)

	jT, err := t1.Join(*t2, func(rows []Variable) [][]Variable {
		var temp [][]Variable
		for _, t2r := range t2.rows {
			t2rt := append(rows, t2r...)
			temp = append(temp, t2rt)
		}
		return temp
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(jT, expected) {
		t.Errorf("unexpected result. want %v, got %v", expected, jT)
	}
}
