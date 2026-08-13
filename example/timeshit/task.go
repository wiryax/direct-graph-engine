package timeshit

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	dge "github.com/wiryax/dage/core"
)

type IssuesTask struct {
	token string
	email string
}

func NewIssuesTask(email, token string) *IssuesTask {
	return &IssuesTask{
		token: token,
		email: email,
	}
}

func (t *IssuesTask) ProducerTask(buff dge.WriteOnlyBuffer, gCtx *dge.GraphContext) error {
	body := []byte(`{
    "jql":"assignee = currentUser() AND status CHANGED TO (\"To do\", \"In progress\", \"Done\") DURING (\"2026-07-05\", \"2026-08-04\") ORDER BY updated DESC",
    "maxResults": 50,
    "fields": [
      "summary",
      "status",
      "assignee",
      "created"
    ]
  }`)
	req, err := http.NewRequest(http.MethodPost, "https://it-baf.atlassian.net/rest/api/3/search/jql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(t.email, t.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	tasks, err := parseJSONTask(responseBody)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err = buff.WriteBuff([]dge.Variable{
			dge.ParseVariable([]byte(task.id)),
			dge.ParseVariable([]byte(task.title)),
			dge.ParseVariable([]byte(task.link)),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type DetailIssueTask struct {
	token string
	email string
}

func NewDetailIssueTask(email, token string) *DetailIssueTask {
	return &DetailIssueTask{
		token: token,
		email: email,
	}
}

func (d *DetailIssueTask) TransformerTask(buffReader dge.ReadOnlyBuffer, buffWriter dge.WriteOnlyBuffer, gCtx *dge.GraphContext) error {
	ch := buffReader.Read()
	for data := range ch {
		if len(data) != 3 {
			return fmt.Errorf("expected 3 cols got %d: %v", len(data), data)
		}

		task := Task{
			id:    data[0].String(),
			title: data[1].String(),
			link:  data[2].String(),
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://it-baf.atlassian.net/rest/api/3/issue/%s/changelog", task.id), nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(d.email, d.token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		err = d.parseResponse(resp, task, buffWriter)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*DetailIssueTask) parseResponse(resp *http.Response, task Task, buffWriter dge.WriteOnlyBuffer) error {
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	taskDetail, err := parseJSONTaskDetail(task, respData)
	if err != nil {
		return err
	}

	ttl := parser(taskDetail)
	result := parseTaskTimeline(ttl)

	for k, v := range result {
		err = buffWriter.WriteBuff([]dge.Variable{
			dge.ParseVariable([]byte(k)),
			dge.ParseVariable([]byte(v)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type GenerateReport struct {
	exceptionalDate map[string]struct{}
}

func NewGenerateReport(exceptionalDate ...string) *GenerateReport {
	gr := &GenerateReport{
		exceptionalDate: make(map[string]struct{}),
	}
	for _, e := range exceptionalDate {
		gr.exceptionalDate[e] = struct{}{}
	}
	return gr
}

func (g *GenerateReport) ConsumerTask(buff dge.ReadOnlyBuffer, gCtx *dge.GraphContext) error {
	result := make(map[string][]string)
	for data := range buff.Read() {
		if len(data) != 2 {
			return fmt.Errorf("expected 2 cols got %d: %v", len(data), data)
		}

		var (
			date        = data[0].String()
			description = data[1].String()
		)

		if _, ok := g.exceptionalDate[date]; ok {
			continue
		}

		result[date] = append(result[date], description)
	}

	for k, v := range result {
		fmt.Printf("%s: %s\n", k, strings.Join(v, "\n"))
	}

	return nil
}
