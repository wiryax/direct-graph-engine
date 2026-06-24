package csv

import (
	"encoding/csv"
	"fmt"
	"strings"

	dge "github.com/wiryax/direct-graph-engine"
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
		i := t.GetColumnIndex(cf.clauses[c].column)
		if i == -1 {
			return fmt.Errorf("column with key %s not found", cf.clauses[c].column)
		}
		cf.clauses[c].index = i
	}

	fn := func(v []engine.Column) bool {
		for _, c := range cf.clauses {
			data, _ := v[c.index].GetFirst()
			if string(data.GetRaw()) != c.value {
				return false
			}
		}

		return true
	}

	var (
		tResult = t.CloneStructure()
		// columns dge.Column
	)
	for ri := range t.CountRows() {
		rows, err := t.GetRows(ri)
		if err != nil {
			return err
		}

		if fn(rows) {
			for ri := range rows {
				tResult.AddOrSetColumn(rows[ri].GetColumnName(), rows[ri].GetAllData()...)
			}
		}
	}

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

	t := engine.MakeTabular()

	for c := range records[0] {
		var temp []dge.Variable
		for d := range records[1:] {
			temp = append(temp, dge.ParseVariable([]byte(records[d][c])))
		}
		t.AddOrSetColumn(records[0][c], temp...)
	}

	gCtx.SetTabularStorage(cr.storageId, *t)

	return nil
}
