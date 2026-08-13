package timeshit

import (
	"encoding/json"
	"errors"
	"time"
)

type TSEvent int

const (
	Skip TSEvent = iota
	Start
	Hold
	Continue
	End
)

type (
	ChangeLog struct {
		time     time.Time
		from, to string
	}

	Task struct {
		id    string
		link  string
		title string
	}

	TaskDetail struct {
		Task
		changeLogs []ChangeLog
	}

	HoldInterval struct {
		from, to time.Time
	}

	TaskTimeline struct {
		Task
		holds    []HoldInterval
		from, to time.Time
	}
)

func tokenizer(from, to string) TSEvent {
	if from == "To do" && to == "In progress" {
		return Start
	} else if from == "In progress" && to == "Hold" {
		return Hold
	} else if from == "Hold" && to == "In progress" {
		return Continue
	} else if from == "In progress" && to == "Done" {
		return End
	} else {
		return Skip
	}
}

func parser(task TaskDetail) TaskTimeline {
	var (
		result = TaskTimeline{
			Task: task.Task,
		}
	)

	for _, t := range task.changeLogs {
		tk := tokenizer(t.from, t.to)
		switch tk {
		case Start:
			result.from = t.time
		case Hold:
			result.holds = append(result.holds, HoldInterval{
				from: t.time,
			})
		case End:
			result.to = t.time
		case Continue:
			result.holds[len(result.holds)-1].to = t.time
		default:
			continue
		}
	}

	return result
}

func parseJSONTask(data []byte) ([]Task, error) {
	obj := make(map[string]any)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	issues, ok := obj["issues"].([]any)
	if !ok || len(issues) == 0 {
		return nil, errors.New("fail casting issue")
	}

	var result []Task

	for _, issue := range issues {
		var task Task
		cIssue, ok := issue.(map[string]any)
		if !ok {
			return nil, errors.New("fail casting issue")
		}

		task.id, ok = cIssue["id"].(string)
		if !ok {
			return nil, errors.New("fail casting id")
		}

		fields, ok := cIssue["fields"].(map[string]any)
		if !ok {
			return nil, errors.New("fail casting fields")
		}

		task.title, ok = fields["summary"].(string)
		if !ok {
			return nil, errors.New("fail casting summary")
		}
		result = append(result, task)
	}

	return result, nil
}

func parseTaskTimeline(task TaskTimeline) map[string]string {
	var (
		result = make(map[string]string)
		holds  = make(map[string]struct{})
		begin  = true
	)

	for _, interval := range task.holds {
		for ; interval.from.Before(interval.to); interval.from = interval.from.AddDate(0, 0, 1) {
			holds[interval.from.Format("01-02-2006")] = struct{}{}
		}
	}

	for ; task.from.Before(task.to); task.from = task.from.AddDate(0, 0, 1) {
		date := task.from.Format("01-02-2006")
		if _, ok := holds[date]; ok {
			continue
		}
		if begin {
			result[date] = "start doing " + task.title
			begin = false
		} else {
			result[date] = "continue doing " + task.title
		}
	}

	return result
}

func parseJSONTaskDetail(task Task, data []byte) (TaskDetail, error) {
	var (
		result = TaskDetail{
			Task: task,
		}
		obj = make(map[string]any)
	)

	err := json.Unmarshal(data, &obj)
	if err != nil {
		return TaskDetail{}, err
	}

	values, ok := obj["values"].([]any)
	if !ok {
		return TaskDetail{}, errors.New("fail casting values")
	}

	for _, v := range values {
		value, ok := v.(map[string]any)
		if !ok {
			return TaskDetail{}, errors.New("fail casting values to map")
		}

		items, ok := value["items"].([]any)
		if !ok {
			return TaskDetail{}, errors.New("fail casting items")
		}

		for _, i := range items {
			item, ok := i.(map[string]any)
			if !ok {
				return TaskDetail{}, errors.New("field casting item to map")
			}

			field, ok := item["field"].(string)
			if !ok {
				return TaskDetail{}, errors.New("field casting field")
			}

			if field != "status" {
				continue
			}

			fromString, ok := item["fromString"].(string)
			if !ok {
				return TaskDetail{}, errors.New("fail casting fromString")
			}

			toString, ok := item["toString"].(string)
			if !ok {
				return TaskDetail{}, errors.New("fail casting toString")
			}

			date, ok := value["created"].(string)
			if !ok {
				return TaskDetail{}, errors.New("fail casting created")
			}

			create, err := time.Parse("2006-01-02T15:04:05.000-0700", date)
			if err != nil {
				return TaskDetail{}, err
			}

			result.changeLogs = append(result.changeLogs, ChangeLog{
				time: create,
				from: fromString,
				to:   toString,
			})
		}
	}

	return result, nil
}
