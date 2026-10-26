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

// TestRegisterWritesAgentFacts covers AC1 and AC7: a registration writes the
// facts directly and no role field.
func TestRegisterWritesAgentFacts(t *testing.T) {
	cases := []struct {
		name         string
		options      RegisterOptions
		project      string
		worktree     string
		wantKey      string
		wantProfile  string
		wantKind     string
		wantOwns     string
		wantWorktree string
	}{
		{
			name:    "coordinator",
			options: RegisterOptions{Kind: records.AgentKindCoordinator, Name: "coordinator", NameSet: true},
			wantKey: "coordinator-g1", wantProfile: "coordinator", wantKind: "coordinator",
		},
		{
			name:    "supervisor",
			options: RegisterOptions{Kind: records.AgentKindSupervisor, Name: "supervisor", NameSet: true},
			wantKey: "supervisor-g1", wantKind: "supervisor",
		},
		{
			name:    "planner",
			options: RegisterOptions{Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", Name: "manager-hive", NameSet: true},
			project: "hive", wantKey: "manager-hive-g1", wantProfile: "planner", wantOwns: "hive",
		},
		{
			name:    "builder",
			options: RegisterOptions{Profile: records.ProfileBuilder, Project: "hive", Task: testTaskID, Name: "builder-1", NameSet: true},
			project: "hive", worktree: "/tmp/worktrees/1-facts", wantKey: "builder-1", wantProfile: "builder", wantWorktree: "/tmp/worktrees/1-facts",
		},
		{
			name:    "verifier",
			options: RegisterOptions{Profile: records.ProfileVerifier, Project: "hive", Task: testTaskID, Name: "verifier-1", NameSet: true},
			project: "hive", worktree: "/tmp/worktrees/1-facts", wantKey: "verifier-1", wantProfile: "verifier", wantWorktree: "/tmp/worktrees/1-facts",
		},
		{
			name:    "general",
			options: RegisterOptions{Profile: records.ProfileGeneral, Name: "general-1", NameSet: true},
			wantKey: "general-1", wantProfile: "general",
		},
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
			options := c.options
			options.PaneID = "w1:p1"
			if options.Task == "" {
				options.Task = task
			}
			envelope := Register(home, options)
			if envelope.Outcome != "ok" {
				t.Fatalf("Register() = %#v; want ok", envelope)
			}
			record := readAgentFields(t, home, c.wantKey)
			for field, want := range map[string]string{
				"profile":  c.wantProfile,
				"kind":     c.wantKind,
				"owns":     c.wantOwns,
				"worktree": c.wantWorktree,
			} {
				if got, ok := record[field]; !ok || got != want {
					t.Errorf("%s = %#v, present %t; want %q", field, got, ok, want)
				}
			}
			if _, ok := record["role"]; ok {
				t.Errorf("record has a role field: %#v", record)
			}
		})
	}
}

// TestListAndShowExposeAgentFacts covers AC5.
func TestListAndShowExposeAgentFacts(t *testing.T) {
	home := testHome(t)
	writeProject(t, home, "hive")
	if envelope := Register(home, RegisterOptions{Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
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
	for _, want := range []string{`"profile":"planner"`, `"kind":""`, `"owns":"hive"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("List JSON = %s; missing %s", data, want)
		}
	}
	if strings.Contains(string(data), `"role"`) {
		t.Errorf("List JSON = %s; want no role field", data)
	}
	shown := Show(home, "manager-hive-g1")
	if shown.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", shown)
	}
	text := ShowText(shown)
	for _, want := range []string{"profile: planner", "kind: ", "owns: hive"} {
		if !strings.Contains(text, want) {
			t.Errorf("ShowText() = %q; missing %q", text, want)
		}
	}
	if strings.Contains(text, "role:") {
		t.Errorf("ShowText() = %q; want no role field", text)
	}
}
