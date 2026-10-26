package show

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/create"
	"github.com/Harrison-Blair/hive/internal/task/list"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestShowAndListRenderNeedsUser(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := records.Task{
		ID: "8-needs-user", Project: "one", Title: "Needs a decision",
		Status: "verified", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, struct {
		records.Task
		NeedsUser string `json:"needs_user"`
	}{Task: task, NeedsUser: "choose an option"}); err != nil {
		t.Fatal(err)
	}
	briefPath := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	if err := os.MkdirAll(filepath.Dir(briefPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(briefPath, []byte("Needs a decision"), 0o644); err != nil {
		t.Fatal(err)
	}

	shown := Run(home, task.ID, Options{})
	if text := Text(shown); !strings.Contains(text, "needs user: choose an option") {
		t.Fatalf("ShowText() = %q; want the needs-user text", text)
	}
	listed := tasklib.List(home, tasklib.ListOptions{})
	if text := list.Text(listed); !strings.Contains(text, "[needs user]") {
		t.Fatalf("ListText() = %q; want the needs-user marker", text)
	}
}

func TestShowRecordAndBrief(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := createOK(t, home, create.Options{Project: "one", Title: "Second step"})

	envelope := Run(home, task.ID, Options{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(Result)
	if !ok {
		t.Fatalf("result = %#v; want a show result", envelope.Result)
	}
	if result.Brief != "Second step" {
		t.Fatalf("brief = %q; want the title text", result.Brief)
	}
	text := Text(envelope)
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
	if briefText := BriefText(Run(home, task.ID, Options{Brief: true})); briefText != "Second step" {
		t.Fatalf("ShowBriefText() = %q; want the brief text only", briefText)
	}
}

func TestShowJSONShape(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := createOK(t, home, create.Options{Project: "one", Title: "Second step"})
	envelope := Run(home, task.ID, Options{})
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
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	for _, id := range []string{"9-missing", "..", "a/b"} {
		envelope := Run(home, id, Options{})
		if envelope.Outcome != "failed" || envelope.Error != "task not found: "+id {
			t.Fatalf("Show(%q) = %#v; want task not found", id, envelope)
		}
	}
}

func TestShowCombinesTheActingAgentAndKeepsUnknownDisplayable(t *testing.T) {
	home, task := tasktest.NewSeed(t, "assigned")
	briefPath := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	if err := os.MkdirAll(filepath.Dir(briefPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(briefPath, []byte("Seed"), 0o644); err != nil {
		t.Fatal(err)
	}
	envelope := Run(home, task.ID, Options{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(Result)
	if !ok || result.Agent == nil || result.Agent.Key != "worker-1" {
		t.Fatalf("Show() = %#v; want the acting agent record", envelope.Result)
	}
	if text := Text(envelope); !strings.Contains(text, "acting_agent_status: live") {
		t.Fatalf("ShowText() = %q; want the acting agent status", text)
	}

	task.ActingAgent = "missing-agent"
	tasktest.SeedTask(t, home, task)
	envelope = Run(home, task.ID, Options{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show(unknown agent) = %#v; want the task to stay displayable", envelope)
	}
	result = envelope.Result.(Result)
	if result.Agent != nil {
		t.Fatalf("agent = %#v; want nil for an unknown name", result.Agent)
	}
	if text := Text(envelope); !strings.Contains(text, "acting_agent_status: no record") {
		t.Fatalf("ShowText() = %q; want the absent status", text)
	}
}

// createOK creates one task and checks a success outcome.
func createOK(t *testing.T, home string, options create.Options) records.Task {
	t.Helper()
	envelope := create.Run(home, options)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Create(%+v) = %#v; want ok", options, envelope)
	}
	task, ok := envelope.Result.(records.Task)
	if !ok {
		t.Fatalf("Create(%+v) result = %#v; want a task record", options, envelope.Result)
	}
	return task
}
