package task

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func seedSimple(t *testing.T, home, id, status string, deps []string) {
	t.Helper()
	seedTask(t, home, records.Task{
		ID: id, Project: "one", Title: id,
		ControllingRole: "manager-one", ActingAgent: "worker-1",
		Status: status, DependsOn: deps,
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
}

func listIDs(t *testing.T, envelope output.Envelope) []string {
	t.Helper()
	tasks, ok := envelope.Result.([]records.Task)
	if !ok {
		t.Fatalf("result = %#v; want a task list", envelope.Result)
	}
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func TestListSortsByProjectThenNumber(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	registerProject(t, home, "two")
	seedSimple(t, home, "10-c", "created", []string{})
	seedTask(t, home, records.Task{
		ID: "3-d", Project: "two", Title: "d", ControllingRole: "manager-two",
		Status: "created", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	seedSimple(t, home, "2-b", "created", []string{})
	seedSimple(t, home, "1-a", "created", []string{})

	envelope := List(home, ListOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("List() = %#v; want ok", envelope)
	}
	tasks, ok := envelope.Result.([]records.Task)
	if !ok {
		t.Fatalf("result = %#v; want a task list", envelope.Result)
	}
	var ids []string
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	want := "1-a,2-b,10-c,3-d"
	if got := strings.Join(ids, ","); got != want {
		t.Fatalf("order = %s; want %s", got, want)
	}
}

func TestListFilters(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	registerProject(t, home, "two")
	seedSimple(t, home, "1-a", "created", []string{})
	seedSimple(t, home, "2-b", "verified", []string{})
	seedTask(t, home, records.Task{
		ID: "3-c", Project: "two", Title: "c", ControllingRole: "manager-two",
		Status: "created", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})

	byProject := List(home, ListOptions{Project: "one"})
	if len(byProject.Result.([]records.Task)) != 2 {
		t.Fatalf("project filter = %#v; want two tasks", byProject.Result)
	}
	byStatus := List(home, ListOptions{Status: "verified"})
	verified := byStatus.Result.([]records.Task)
	if len(verified) != 1 || verified[0].ID != "2-b" {
		t.Fatalf("status filter = %#v; want 2-b", byStatus.Result)
	}
}

func TestListEmpty(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := List(home, ListOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("List() = %#v; want ok", envelope)
	}
	if text := ListText(envelope); text != "no tasks\n" {
		t.Fatalf("ListText() = %q; want no tasks", text)
	}
}

func TestReadyFilter(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	seedSimple(t, home, "1-a", "created", []string{})
	seedSimple(t, home, "2-b", "created", []string{"1-a"})
	seedSimple(t, home, "3-c", "assigned", []string{})
	seedSimple(t, home, "4-d", "created", []string{"1-a", "2-b"})

	first := List(home, ListOptions{Ready: true})
	ids := listIDs(t, first)
	if strings.Join(ids, ",") != "1-a" {
		t.Fatalf("ready = %v; want only 1-a", ids)
	}

	seedSimple(t, home, "1-a", "verified", []string{})
	second := List(home, ListOptions{Ready: true})
	ids = listIDs(t, second)
	if strings.Join(ids, ",") != "2-b" {
		t.Fatalf("ready = %v; want only 2-b after 1-a is verified", ids)
	}
}

func TestShowRecordAndBrief(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Second step"})

	envelope := Show(home, task.ID, ShowOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(ShowResult)
	if !ok {
		t.Fatalf("result = %#v; want a show result", envelope.Result)
	}
	if result.Brief != "Second step" {
		t.Fatalf("brief = %q; want the title text", result.Brief)
	}
	text := ShowText(envelope)
	for _, want := range []string{
		"id: " + task.ID,
		"project: one",
		"status: created",
		"brief: .hive/state/briefs/" + task.ID + ".md",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("ShowText() = %q; want %q", text, want)
		}
	}
	if briefText := ShowBriefText(Show(home, task.ID, ShowOptions{Brief: true})); briefText != "Second step" {
		t.Fatalf("ShowBriefText() = %q; want the brief text only", briefText)
	}
}

func TestShowJSONShape(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Second step"})
	envelope := Show(home, task.ID, ShowOptions{})
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Result struct {
			ID           string                `json:"id"`
			Status       string                `json:"status"`
			Verification *records.Verification `json:"verification"`
			Brief        string                `json:"brief"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Result.ID != task.ID || decoded.Result.Status != "created" || decoded.Result.Brief != "Second step" {
		t.Fatalf("JSON result = %#v; want the record fields and the brief", decoded.Result)
	}
}

func TestShowMissing(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	for _, id := range []string{"9-missing", "..", "a/b"} {
		envelope := Show(home, id, ShowOptions{})
		if envelope.Outcome != "failed" || envelope.Error != "task not found: "+id {
			t.Fatalf("Show(%q) = %#v; want task not found", id, envelope)
		}
	}
}
