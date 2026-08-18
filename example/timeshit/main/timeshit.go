package main

import (
	"flag"
	"os"
	"time"

	"github.com/wiryax/dage/connector/file"
	dge "github.com/wiryax/dage/core"
	"github.com/wiryax/dage/example/timeshit"
)

func panicOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	var (
		email       = flag.String("email", "", "jira user email")
		token       = flag.String("token", "", "jira user security token")
		from        = flag.String("from", "", "report start")
		to          = flag.String("to", "", "report end")
		template    = flag.String("template", "", "HTML report template file")
		destination = flag.String("destination", "", "HTML report result file")
	)
	flag.Parse()

	sDate, err := time.Parse("2006-01-02", *from)
	panicOnError(err)

	eDate, err := time.Parse("2006-01-02", *to)
	panicOnError(err)

	getAllIssueTask := dge.NewBufferProducer("getAllIssueTask", timeshit.NewIssuesTask(*email, *token))
	getDetailIssueTask := dge.NewBufferTransformerTask("getDetailIssueTask", 100, timeshit.NewDetailIssueTask(*email, *token))
	generateReportTask := dge.NewBufferConsumerVertex("generateReportTask", 100, timeshit.NewGenerateReport("htmlTemplate", "result", sDate, eDate, nil))

	e1, err := dge.NewEdge(dge.OnCompilation, getAllIssueTask, getDetailIssueTask)
	panicOnError(err)

	e2, err := dge.NewEdge(dge.OnCompilation, getDetailIssueTask, generateReportTask)
	panicOnError(err)

	graph := dge.NewGraph(*template)

	panicOnError(graph.Connect(e1))
	panicOnError(graph.Connect(e2))

	htmlTemplate := file.NewHTMLTemplateConnector(*template)
	result := file.NewFileConnector(*destination, os.O_CREATE|os.O_TRUNC, 0644)

	runtime := dge.NewEngine(graph)
	runtime.SetConnection(map[string]dge.Connection{
		"htmlTemplate": htmlTemplate,
		"result":       result,
	})
	panicOnError(runtime.Run())
}
