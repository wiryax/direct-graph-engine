package dge

import (
	"reflect"
	"testing"
)

func TestAddNewColToEmptyTabular(t *testing.T) {
	result := Tabular{}
	expected := Tabular{
		columns: []Column{
			{
				name: "A",
				data: []Variable{
					{
						code: 0,
						raw:  []byte("1"),
					},
				},
			}, {
				name: "B",
				data: []Variable{
					{
						code: 0,
						raw:  []byte("2"),
					},
				},
			},
		},
	}

	result.AddOrSetColumn("A", Variable{
		code: 0,
		raw:  []byte("1"),
	})

	result.AddOrSetColumn("B", Variable{
		code: 0,
		raw:  []byte("2"),
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("unexpected result. want %v, got %v", expected, result)
	}

	t.Logf("result : %v", result)

}

func TestAddRowToEmptyTabular(t *testing.T) {
	result := Tabular{}
	err := result.AddData(Column{
		name: "A",
		data: []Variable{
			{
				code: 0,
				raw:  []byte("1"),
			},
		},
	})

	if err == nil {
		t.Error("err should be appear")
	}

	t.Logf("result: %v", result)
}

func TestAddNewColAgainstExistingCol(t *testing.T) {
	result := Tabular{
		columns: []Column{
			{
				name: "A",
				data: []Variable{
					{
						code: 0,
						raw:  []byte("1"),
					},
				},
			},
		},
	}

	result.AddOrSetColumn("A")

	t.Logf("result: %v", result)
}
