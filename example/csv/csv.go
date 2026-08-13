package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"slices"

	dge "github.com/wiryax/dage/core"
	engine "github.com/wiryax/dage/core"
)

type Clause struct {
	column,
	value string
	index int
}

type CsvFilter struct {
	connectorId string
	clauses     []Clause
}

func NewCsvFilter(columns []string, clauses ...Clause) *CsvFilter {
	for c := range clauses {
		i := slices.Index(columns, clauses[c].column)
		clauses[c].index = i
	}

	return &CsvFilter{
		clauses: clauses,
	}
}

func (cf *CsvFilter) TransformerTask(buffReader dge.ReadOnlyBuffer, buffWriter dge.WriteOnlyBuffer, gCtx *engine.GraphContext) error {
	fn := func(v []dge.Variable) bool {
		for _, c := range cf.clauses {
			if v[c.index].String() != c.value {
				return false
			}
		}

		return true
	}

	for item := range buffReader.Read() {
		if fn(item) {
			if err := buffWriter.WriteBuff(item); err != nil {
				return err
			}
		}
	}

	return nil
}

type MockCsvReader struct {
	connId string
}

func NewMockCsvReader(connId string) *MockCsvReader {
	return &MockCsvReader{
		connId: connId,
	}
}

func (cr *MockCsvReader) ProducerTask(buffWriter dge.WriteOnlyBuffer, gCtx *engine.GraphContext) error {
	conn, err := gCtx.GetConnection(cr.connId)
	if err != nil {
		return err
	}

	r, ok := conn.Acquire(nil).(*csv.Reader)
	if !ok {
		return fmt.Errorf("cannot cast connector")
	}

	for {
		row, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		var temp []dge.Variable
		for _, c := range row {
			temp = append(temp, dge.ParseVariable([]byte(c)))
		}
		buffWriter.WriteBuff(temp)
	}

	return nil
}

type MockCsvWriter struct {
	connId string
}

func NewMockCsvWriter(connId string) *MockCsvWriter {
	return &MockCsvWriter{
		connId: connId,
	}
}

func (cw *MockCsvWriter) ConsumerTask(buff dge.ReadOnlyBuffer, gCtx *dge.GraphContext) error {
	conn, err := gCtx.GetConnection(cw.connId)
	if err != nil {
		return err
	}

	csvWriter, ok := conn.Acquire(nil).(*csv.Writer)
	if !ok {
		return fmt.Errorf("cannot casting connection")
	}
	ch := buff.Read()
	for item := range ch {
		if item == nil {
			continue
		}
		var prow []string
		for _, c := range item {
			prow = append(prow, c.String())
		}

		if err := csvWriter.Write(prow); err != nil {
			return err
		}

		csvWriter.Flush()
	}

	return nil
}
