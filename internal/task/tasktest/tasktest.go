// Package tasktest holds the fixture helpers that the tests of several task
// operation packages share. Only tests import it.
package tasktest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// NewHome makes one fixture home with its marker.
func NewHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

// RegisterProject writes one project record.
func RegisterProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: filepath.Join(home, "repo-"+id), Base: "main"}
	if err := registry.Write("projects", id, record); err != nil {
		t.Fatal(err)
	}
}

// SeedTask writes one task record.
func SeedTask(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
}

// SeedAgent writes one live agent record for the task fixture.
func SeedAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

// NewSeed makes a home with one task and its worker and verifier records.
func NewSeed(t *testing.T, status string) (string, records.Task) {
	t.Helper()
	home := NewHome(t)
	RegisterProject(t, home, "one")
	task := records.Task{
		ID:          "1-seed",
		Project:     "one",
		Title:       "Seed",
		ActingAgent: "worker-1",
		Status:      status,
		DependsOn:   []string{},
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-01T00:00:00Z",
	}
	SeedTask(t, home, task)
	for i, name := range []string{"worker-1", "worker-2"} {
		SeedAgent(t, home, records.Agent{
			Key: name, Name: name, Profile: records.ProfileBuilder, Project: "one", Task: task.ID,
			Status: records.AgentLive, PaneID: "w1:p" + string(rune('1'+i)), Generation: 1,
			StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		})
	}
	for i, name := range []string{"verifier-1", "verifier-2"} {
		SeedAgent(t, home, records.Agent{
			Key: name, Name: name, Profile: records.ProfileVerifier, Project: "one", Task: task.ID,
			Status: records.AgentLive, PaneID: "w1:p" + string(rune('3'+i)), Generation: 1,
			StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		})
	}
	return home, task
}

// TaskRecordBytes returns the stored bytes of one task record.
func TaskRecordBytes(t *testing.T, home, id string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "tasks", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// ReadTask reads one task record.
func ReadTask(t *testing.T, home, id string) records.Task {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, id, &task); err != nil {
		t.Fatal(err)
	}
	return task
}

// WriteFile writes one fixture file and its parent folders.
func WriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ReadFileOrEmpty returns the body of one file and whether it exists.
func ReadFileOrEmpty(t *testing.T, path string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data), true
}

// HasTaskEffect reports whether an envelope has one effect.
func HasTaskEffect(envelope output.Envelope, resource, action, path string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action && effect.Path == path {
			return true
		}
	}
	return false
}

// WriteOwnerState writes one JSON object below the state root.
func WriteOwnerState(t *testing.T, home string, parts []string, value any) {
	t.Helper()
	path := filepath.Join(append([]string{home, ".hive", "state"}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// ReadOwnerRecord reads one state file as a JSON object.
func ReadOwnerRecord(t *testing.T, home string, parts ...string) map[string]any {
	t.Helper()
	path := filepath.Join(append([]string{home, ".hive", "state"}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

// SeedOwnerHome writes one project, one live owner, one builder, and one
// assigned task. The task file keeps the old controlling_role field so the
// base build can complete it.
func SeedOwnerHome(t *testing.T) string {
	t.Helper()
	home := NewHome(t)
	WriteOwnerState(t, home, []string{"projects", "demo.json"}, map[string]any{
		"id": "demo", "path": filepath.Join(home, "repo-demo"), "base": "dev",
	})
	WriteOwnerState(t, home, []string{"agents", "owner-demo-g1.json"}, map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager",
		"kind": "", "profile": "planner", "owns": "demo", "project": "demo",
		"status": "live", "generation": 1, "pane_id": "w1:p1",
	})
	WriteOwnerState(t, home, []string{"claims", "owner-demo.json"}, map[string]any{
		"key": "owner-demo", "agent": "owner-demo-g1", "generation": 1,
	})
	WriteOwnerState(t, home, []string{"agents", "builder-1.json"}, map[string]any{
		"key": "builder-1", "name": "builder-1", "role": "worker",
		"profile": "builder", "project": "demo", "task": "1-step",
		"status": "live", "generation": 1, "pane_id": "w1:p2",
	})
	WriteOwnerState(t, home, []string{"tasks", "1-step.json"}, map[string]any{
		"id": "1-step", "project": "demo", "title": "Step", "controlling_role": "manager-demo",
		"acting_agent": "builder-1", "status": "assigned", "depends_on": []string{},
		"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	})
	return home
}

// AssertOneLineError checks that an envelope error is one plain line.
func AssertOneLineError(t *testing.T, label string, envelopeError any) string {
	t.Helper()
	message, ok := envelopeError.(string)
	if !ok {
		t.Fatalf("%s error = %#v; want a string", label, envelopeError)
	}
	if strings.ContainsAny(message, "\n\r\t") {
		t.Fatalf("%s error = %q; want one plain line", label, message)
	}
	return message
}

// BriefAncestorPath returns the computed brief directory of a home.
func BriefAncestorPath(home string) string {
	return filepath.Join(home, ".hive", "state", "briefs")
}
