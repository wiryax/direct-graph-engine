package csv

import (
	"os"
	"reflect"
	"testing"

	engine "github.com/wiryax/direct-graph-engine"
)

func TestCsvFilterWorkflow(t *testing.T) {
	logger := engine.NewLogger(os.Stdout)
	mockCsvStorageId := "mockCsvConn"
	rState := engine.NewRuntimeState(make(map[string]string))
	storage := engine.NewStorage()
	gCtx := engine.NewGraphContext(logger, rState, storage)
	graph := engine.NewGraph("TestCsvWorkflow")

	csvReader := graph.Add("Csv Reader", NewMockCsvReader(mockCsvStorageId, "first_name,middle_name,last_name\nwirya,muhammad,nugraha\nnugraha,muhammad,wirya"))

	csvFilter := graph.Add("Csv Filter", NewCsvFilter(mockCsvStorageId,
		Clause{
			column: "first_name",
			value:  "wirya",
		}, Clause{
			column: "middle_name",
			value:  "muhammad",
		}, Clause{
			column: "last_name",
			value:  "nugraha",
		}),
	)

	graph.Connect(csvReader, csvFilter, engine.Success, engine.ExpAnd, nil)

	graph.RunWithContext(gCtx)

	expectedTabular := engine.MakeTabular([]string{"first_name", "middle_name", "last_name"})
	expectedTabular.AddRow(
		engine.ParseVariable([]byte("wirya")),
		engine.ParseVariable([]byte("muhammad")),
		engine.ParseVariable([]byte("nugraha")))

	tResult, err := gCtx.GetTabularStorage(mockCsvStorageId)
	if err != nil {
		t.Fatalf("unexpected error %v", tResult)
	}

	if !reflect.DeepEqual(tResult, *expectedTabular) {
		t.Errorf("unexpected result. want %v got %v", *expectedTabular, tResult)
	}
}
