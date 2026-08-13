package timeshit

import (
	"reflect"
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	mock := TaskDetail{
		changeLogs: []ChangeLog{
			{
				time: time.Now().Add(-4 * (time.Hour * 24)),
				from: "To do",
				to:   "In progress",
			}, {
				time: time.Now().Add(-3 * (time.Hour * 24)),
				from: "In progress",
				to:   "Hold",
			}, {
				time: time.Now().Add(-2 * (time.Hour * 24)),
				from: "Hold",
				to:   "In progress",
			}, {
				time: time.Now().Add(-1 * (time.Hour * 24)),
				from: "In progress",
				to:   "Done",
			},
		},
	}

	expected := TaskTimeline{
		holds: []HoldInterval{
			{
				from: time.Now().Add(-3 * (time.Hour * 24)),
				to:   time.Now().Add(-2 * (time.Hour * 24)),
			},
		},
		from: time.Now().Add(-4 * (time.Hour * 24)),
		to:   time.Now().Add(-1 * (time.Hour * 24)),
	}

	result := parser(mock)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("unexpected error, want %#v got %#v", expected, result)
	}
}

func TestParseJSON(t *testing.T) {
	data := `
		{
		"issues": [
				{
					"expand": "renderedFields,names,schema,operations,editmeta,changelog,versionedRepresentations",
					"id": "26226",
					"self": "https://it-baf.atlassian.net/rest/api/3/issue/26226",
					"key": "SSIS-181",
					"fields": {
						"summary": "OPA - Adding Field Name BPKB (Truncate Table)",
						"assignee": {
							"self": "https://it-baf.atlassian.net/rest/api/3/user?accountId=712020%3Aa2e7ae12-6904-4eb8-aa05-2cb32e7b423e",
							"accountId": "712020:a2e7ae12-6904-4eb8-aa05-2cb32e7b423e",
							"emailAddress": "wirya.nugraha@baf.id",
							"avatarUrls": {
								"48x48": "https://secure.gravatar.com/avatar/460c02484192af66bca222ba81452705?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FWN-0.png",
								"24x24": "https://secure.gravatar.com/avatar/460c02484192af66bca222ba81452705?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FWN-0.png",
								"16x16": "https://secure.gravatar.com/avatar/460c02484192af66bca222ba81452705?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FWN-0.png",
								"32x32": "https://secure.gravatar.com/avatar/460c02484192af66bca222ba81452705?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FWN-0.png"
							},
							"displayName": "Wirya Muhammad Nugraha",
							"active": true,
							"timeZone": "Asia/Jakarta",
							"accountType": "atlassian"
						},
						"created": "2024-10-10T11:17:05.027+0700",
						"status": {
							"self": "https://it-baf.atlassian.net/rest/api/3/status/10114",
							"description": "",
							"iconUrl": "https://it-baf.atlassian.net/",
							"name": "Done",
							"id": "10114",
							"statusCategory": {
								"self": "https://it-baf.atlassian.net/rest/api/3/statuscategory/3",
								"id": 3,
								"key": "done",
								"colorName": "green",
								"name": "Done"
							}
						}
					}
				}
			]
		}`

	expected := []Task{
		{
			id:    "26226",
			title: "OPA - Adding Field Name BPKB (Truncate Table)",
		},
	}

	result, err := parseJSONTask([]byte(data))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("unexpected result. want %#v, got %#v", expected, result)
	}
}

func TestParseTaskTimeline(t *testing.T) {
	ttl := TaskTimeline{
		from: time.Now().AddDate(0, 0, -7),
		to:   time.Now(),
		holds: []HoldInterval{
			{
				from: time.Now().AddDate(0, 0, -6),
				to:   time.Now().AddDate(0, 0, -5),
			},
		},
	}

	expected := map[string]string{
		time.Now().AddDate(0, 0, -7).Format("01-02-2006"): "start doing",
		time.Now().AddDate(0, 0, -5).Format("01-02-2006"): "continue doing",
		time.Now().AddDate(0, 0, -4).Format("01-02-2006"): "continue doing",
		time.Now().AddDate(0, 0, -3).Format("01-02-2006"): "continue doing",
		time.Now().AddDate(0, 0, -2).Format("01-02-2006"): "continue doing",
		time.Now().AddDate(0, 0, -1).Format("01-02-2006"): "continue doing",
		time.Now().Format("01-02-2006"):                   "continue doing",
	}

	result := parseTaskTimeline(ttl)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("unexpected error. want %v, got %v", expected, result)
	}
}
