package csv

import (
	"testing"

	dge "github.com/wiryax/dage/core"
	"github.com/wiryax/dage/example/connector"
)

func TestCsvFilterWorkflow(t *testing.T) {
	csvColumns := []string{"Index", "Customer Id", "First Name", "Last Name", "Company", "City", "Country", "Phone 1", "Phone 2", "Email", "Subscription Date", "Website"}
	csvWriter := connector.NewCSVWriterConnector(csvColumns)
	csvReader := connector.NewCSVReaderConnector("customers-10000.csv")
	conn := map[string]dge.Connection{
		"csvConn":   csvReader,
		"csvWriter": csvWriter,
	}

	g := dge.NewGraph("csv filter")

	csvFilterTask := NewCsvFilter(csvColumns, Clause{
		column: "First Name",
		value:  "Heather",
	})
	csvReaderTask := NewMockCsvReader("csvConn")
	csvWriterTask := NewMockCsvWriter("csvWriter")

	writerTask := dge.NewBufferConsumerVertex("csvWriterTask", 1000, csvWriterTask)
	filterTask := dge.NewBufferTransformerTask("csvFilterTask", 1000, csvFilterTask)

	readerTask := dge.NewBufferProducer("csvReaderTask", csvReaderTask)
	e, err := dge.NewEdge(dge.OnCompilation, readerTask, filterTask)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = g.Connect(e)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	e1, err := dge.NewEdge(dge.OnCompilation, filterTask, writerTask)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if err := g.Connect(e1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	engine := dge.NewEngine(g)
	engine.SetConnection(conn)
	if err := engine.Run(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	t.Logf("%s", string(csvWriter.GetData()))
}
