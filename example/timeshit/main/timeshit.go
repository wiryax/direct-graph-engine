package main

import (
	dge "github.com/wiryax/dage/core"
	"github.com/wiryax/dage/example/timeshit"
)

var (
	email = ""
	token = ""
)

func main() {
	getAllIssueTask := dge.NewBufferProducer("getAllIssueTask", timeshit.NewIssuesTask(email, token))
	getDetailIssueTask := dge.NewBufferTransformerTask("getDetailIssueTask", 100, timeshit.NewDetailIssueTask(email, token))
	generateReportTask := dge.NewBufferConsumerVertex("generateReportTask", 100, timeshit.NewGenerateReport())

	e1, err := dge.NewEdge(dge.OnCompilation, getAllIssueTask, getDetailIssueTask)
	if err != nil {
		panic(err)
	}

	e2, err := dge.NewEdge(dge.OnCompilation, getDetailIssueTask, generateReportTask)
	if err != nil {
		panic(err)
	}

	graph := dge.NewGraph("timeshit")

	if err := graph.Connect(e1); err != nil {
		panic(err)
	}

	if err := graph.Connect(e2); err != nil {
		panic(err)
	}

	runtime := dge.NewEngine(graph)
	if err := runtime.Run(); err != nil {
		panic(err)
	}
}
