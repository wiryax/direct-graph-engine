package csv

import (
	"encoding/csv"
	"fmt"
	"strings"

	engine "github.com/wiryax/direct-graph-engine"
)

type Clause struct {
	column,
	value string
	index int
}

type CsvFilter struct {
	storageId string
	clauses   []Clause
}

func NewCsvFilter(connectorId string, clauses ...Clause) *CsvFilter {
	return &CsvFilter{
		clauses:   clauses,
		storageId: connectorId,
	}
}

func (cf *CsvFilter) Execute(gCtx *engine.GraphContext) error {
	t, err := gCtx.GetTabularStorage(cf.storageId)
	if err != nil {
		return err
	}

	for c := range cf.clauses {
		i := t.GetColIndex(cf.clauses[c].column)
		if i == -1 {
			return fmt.Errorf("column with key %s not found", cf.clauses[c].column)
		}
		cf.clauses[c].index = i
	}

	fn := func(v []engine.Variable) bool {
		for _, c := range cf.clauses {
			if string(v[c.index].GetRaw()) != c.value {
				return false
			}
		}

		return true
	}

	tResult := t.FilterTabular(cf.storageId, fn)

	gCtx.SetTabularStorage(cf.storageId, tResult)
	return nil
}

type MockCsvReader struct {
	b,
	storageId string
}

func NewMockCsvReader(storageId, content string) *MockCsvReader {
	return &MockCsvReader{
		storageId: storageId,
		b:         content,
	}
}

func (cr *MockCsvReader) Execute(gCtx *engine.GraphContext) error {
	csvReader := csv.NewReader(strings.NewReader(string(cr.b)))

	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	columns := records[0]
	rows := records[1:]

	t := engine.MakeTabular(columns)

	for _, r := range rows {
		var tempRow []engine.Variable

		for _, record := range r {
			tempRow = append(tempRow, engine.ParseVariable([]byte(record)))
		}

		t.AddRow(tempRow...)
	}

	gCtx.SetTabularStorage(cr.storageId, *t)

	return nil
}
