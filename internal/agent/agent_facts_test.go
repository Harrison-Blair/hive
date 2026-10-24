package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func readAgentFields(t *testing.T, home, key string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "agents", key+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

// TestRegisterWritesAgentFacts covers AC1 and AC7 for a new record of each
// role: the record gets its facts, and the role mirror follows the facts.
func TestRegisterWritesAgentFacts(t *testing.T) {
	cases := []struct {
		name         string
		role         string
		project      string
		worktree     string
		wantKey      string
		wantProfile  string
		wantKind     string
		wantOwns     string
		wantWorktree string
		wantRole     string
	}{
		{name: "coordinator", role: "coordinator", wantKey: "coordinator-g1", wantProfile: "coordinator", wantKind: "coordinator", wantRole: "coordinator"},
		{name: "supervisor", role: "supervisor", wantKey: "supervisor-g1", wantKind: "supervisor", wantRole: "supervisor"},
		{name: "manager", role: "manager", project: "hive", wantKey: "manager-hive-g1", wantProfile: "planner", wantOwns: "hive", wantRole: "manager"},
		{name: "worker", role: "worker", project: "hive", worktree: "/tmp/worktrees/1-facts", wantKey: "worker-hive-1", wantProfile: "builder", wantWorktree: "/tmp/worktrees/1-facts", wantRole: "worker"},
		{name: "verifier", role: "verifier", project: "hive", worktree: "/tmp/worktrees/1-facts", wantKey: "verifier-hive-1", wantProfile: "verifier", wantWorktree: "/tmp/worktrees/1-facts", wantRole: "verifier"},
		{name: "agent", role: "agent", wantKey: "agent", wantProfile: "general", wantRole: "agent"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := testHome(t)
			registry, err := store.Open(home)
			if err != nil {
				t.Fatal(err)
			}
			if c.project != "" {
				if err := registry.WriteIn([]string{"projects"}, c.project, records.Project{ID: c.project, Path: "/tmp/" + c.project, Base: "dev"}); err != nil {
					t.Fatal(err)
				}
			}
			task := ""
			if c.worktree != "" {
				task = testTaskID
				record := records.Task{ID: task, Project: c.project, Title: "Task", Worktree: &records.Worktree{Path: c.worktree}}
				if err := registry.WriteIn([]string{"tasks"}, task, record); err != nil {
					t.Fatal(err)
				}
			}
			envelope := Register(home, RegisterOptions{Role: c.role, Project: c.project, Task: task, PaneID: "w1:p1"})
			if envelope.Outcome != "ok" {
				t.Fatalf("Register() = %#v; want ok", envelope)
			}
			record := readAgentFields(t, home, c.wantKey)
			for field, want := range map[string]string{
				"role":     c.wantRole,
				"profile":  c.wantProfile,
				"kind":     c.wantKind,
				"owns":     c.wantOwns,
				"worktree": c.wantWorktree,
			} {
				if got, ok := record[field]; !ok || got != want {
					t.Errorf("%s = %#v, present %t; want %q", field, got, ok, want)
				}
			}
		})
	}
}

// TestListAndShowExposeAgentFacts covers AC5.
func TestListAndShowExposeAgentFacts(t *testing.T) {
	home := testHome(t)
	writeProject(t, home, "hive")
	if envelope := Register(home, RegisterOptions{Role: "manager", Project: "hive", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	listed := List(home, ListOptions{All: true})
	if listed.Outcome != "ok" {
		t.Fatalf("List() = %#v; want ok", listed)
	}
	data, err := json.Marshal(listed.Result)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"profile":"planner"`, `"kind":""`, `"owns":"hive"`, `"role":"manager"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("List JSON = %s; missing %s", data, want)
		}
	}
	shown := Show(home, "manager-hive-g1")
	if shown.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", shown)
	}
	text := ShowText(shown)
	for _, want := range []string{"profile: planner", "kind: ", "owns: hive", "role: manager"} {
		if !strings.Contains(text, want) {
			t.Errorf("ShowText() = %q; missing %q", text, want)
		}
	}
}
