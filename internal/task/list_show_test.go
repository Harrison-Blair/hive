package task

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func seedSimple(t *testing.T, home, id, status string, deps []string) {
	t.Helper()
	seedTask(t, home, records.Task{
		ID: id, Project: "one", Title: id,
		ActingAgent: "worker-1",
		Status:      status, DependsOn: deps,
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
		ID: "3-d", Project: "two", Title: "d",
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
		ID: "3-c", Project: "two", Title: "c",
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
