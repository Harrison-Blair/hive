package task

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

// writeOwnerState writes one JSON object below the state root.
func writeOwnerState(t *testing.T, home string, parts []string, value any) {
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

// readOwnerRecord reads one state file as a JSON object.
func readOwnerRecord(t *testing.T, home string, parts ...string) map[string]any {
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

// seedOwnerHome writes one project, one live owner, one builder, and one
// assigned task. The task file keeps the old controlling_role field so the
// base build can complete it.
func seedOwnerHome(t *testing.T) string {
	t.Helper()
	home := newHome(t)
	writeOwnerState(t, home, []string{"projects", "demo.json"}, map[string]any{
		"id": "demo", "path": filepath.Join(home, "repo-demo"), "base": "dev",
	})
	writeOwnerState(t, home, []string{"agents", "owner-demo-g1.json"}, map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager",
		"kind": "", "profile": "planner", "owns": "demo", "project": "demo",
		"status": "live", "generation": 1, "pane_id": "w1:p1",
	})
	writeOwnerState(t, home, []string{"claims", "owner-demo.json"}, map[string]any{
		"key": "owner-demo", "agent": "owner-demo-g1", "generation": 1,
	})
	writeOwnerState(t, home, []string{"agents", "builder-1.json"}, map[string]any{
		"key": "builder-1", "name": "builder-1", "role": "worker",
		"profile": "builder", "project": "demo", "task": "1-step",
		"status": "live", "generation": 1, "pane_id": "w1:p2",
	})
	writeOwnerState(t, home, []string{"tasks", "1-step.json"}, map[string]any{
		"id": "1-step", "project": "demo", "title": "Step", "controlling_role": "manager-demo",
		"acting_agent": "builder-1", "status": "assigned", "depends_on": []string{},
		"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	})
	return home
}

// TestTaskCompleteMailsTheProjectOwner proves AC2: task complete sends its
// record to the owner of the project of the task.
func TestTaskCompleteMailsTheProjectOwner(t *testing.T) {
	home := seedOwnerHome(t)
	envelope := Complete(home, TransitionOptions{Task: "1-step", Agent: "builder-1"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	found := readOwnerInbox(t, home, "owner-demo")
	if len(found) != 1 {
		t.Fatalf("owner inbox = %#v; want one completed record", found)
	}
	record := found[0]
	if record.Kind != records.InboxCompleted || record.Task != "1-step" || record.Agent != "builder-1" || record.Text != "completed" {
		t.Fatalf("record = %#v; want the completed record of 1-step", record)
	}
	if _, ok := readOwnerRecord(t, home, "tasks", "1-step.json")["controlling_role"]; ok {
		t.Fatal("the task record still has controlling_role")
	}
}

// TestTaskCompleteUsesCoordinatorWithoutOwner proves AC2 with no live owner.
func TestTaskCompleteUsesCoordinatorWithoutOwner(t *testing.T) {
	home := seedOwnerHome(t)
	stopped := readOwnerRecord(t, home, "agents", "owner-demo-g1.json")
	stopped["status"] = "stopped"
	writeOwnerState(t, home, []string{"agents", "owner-demo-g1.json"}, stopped)

	envelope := Complete(home, TransitionOptions{Task: "1-step", Agent: "builder-1"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	found := readOwnerInbox(t, home, "coordinator")
	if len(found) != 1 || found[0].Task != "1-step" {
		t.Fatalf("coordinator inbox = %#v; want the completed record", found)
	}
	if got := readOwnerInbox(t, home, "owner-demo"); len(got) != 0 {
		t.Fatalf("owner inbox = %#v; want no record for a stopped owner", got)
	}
}

func readOwnerInbox(t *testing.T, home, key string) []records.Inbox {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := registry.ListIn([]string{"inbox", key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := make([]records.Inbox, 0, len(raw))
	for _, data := range raw {
		var record records.Inbox
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		found = append(found, record)
	}
	return found
}

// TestCreateWritesNoControllingRole proves AC2 for a new task record.
func TestCreateWritesNoControllingRole(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "demo")
	envelope := Create(home, CreateOptions{Project: "demo", Title: "Step"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %#v; want ok", envelope)
	}
	record := envelope.Result.(records.Task)
	if _, ok := readOwnerRecord(t, home, "tasks", record.ID+".json")["controlling_role"]; ok {
		t.Fatal("the new task record has controlling_role")
	}
}

// TestTaskShowPrintsProjectMail proves AC8: task show prints mail: project:ID
// and no controlling_role line.
func TestTaskShowPrintsProjectMail(t *testing.T) {
	text := ShowText(okEnvelope(ShowResult{Task: records.Task{ID: "1-step", Project: "demo"}}))
	if !strings.Contains(text, "mail: project:demo") {
		t.Fatalf("ShowText() = %q; want mail: project:demo", text)
	}
	if strings.Contains(text, "controlling_role:") {
		t.Fatalf("ShowText() = %q; want no controlling_role line", text)
	}
}

// TestFailureReportPrintsProjectInbox proves AC8: the failure report names the
// project inbox and --inbox.
func TestFailureReportPrintsProjectInbox(t *testing.T) {
	envelope := output.Envelope{Outcome: "partial", Phase: "inbox", Result: records.Task{ID: "1-step", Project: "demo"}}
	text := FailureReport(envelope, "1-step", "builder-1")
	if !strings.Contains(text, "inspect: hive inbox --inbox project:demo --all") {
		t.Fatalf("FailureReport() = %q; want inspect: hive inbox --inbox project:demo --all", text)
	}
}

// TestTaskEditNeedsTheProjectOwnerOrCoordinator proves AC7 with the exact
// refusal text, for the owner, the coordinator, and another agent.
func TestTaskEditNeedsTheProjectOwnerOrCoordinator(t *testing.T) {
	home := seedOwnerHome(t)
	writeOwnerState(t, home, []string{"agents", "coordinator-g1.json"}, map[string]any{
		"key": "coordinator-g1", "name": "coordinator", "kind": "coordinator",
		"profile": "coordinator", "status": "live", "generation": 1, "pane_id": "w0:p1",
	})
	writeOwnerState(t, home, []string{"claims", "coordinator.json"}, map[string]any{
		"key": "coordinator", "agent": "coordinator-g1", "generation": 1,
	})

	refused := Edit(home, EditOptions{Task: "1-step", Agent: "builder-1", NeedsUserSet: true, NeedsUser: "wait"})
	if refused.Outcome != "failed" || refused.Error != "not the owner of project demo or the coordinator: builder-1" {
		t.Fatalf("Edit(builder-1) = %#v; want the owner refusal", refused)
	}
	if record := readOwnerRecord(t, home, "tasks", "1-step.json"); record["needs_user"] != nil {
		t.Fatalf("task = %#v; want no needs_user write", record)
	}

	for _, agent := range []string{"owner-demo", "coordinator"} {
		envelope := Edit(home, EditOptions{Task: "1-step", Agent: agent, NeedsUserSet: true, NeedsUser: "wait for " + agent})
		if envelope.Outcome != "ok" {
			t.Fatalf("Edit(%s) = %#v; want ok", agent, envelope)
		}
	}
}

// TestTaskReopenNeedsTheProjectOwnerOrCoordinator proves AC7 for reopen.
func TestTaskReopenNeedsTheProjectOwnerOrCoordinator(t *testing.T) {
	home := seedOwnerHome(t)
	task := readOwnerRecord(t, home, "tasks", "1-step.json")
	task["status"] = "verified"
	task["verification"] = map[string]any{"agent": "verifier-1", "result": "pass", "reason": "", "at": "2026-01-01T00:00:00Z"}
	writeOwnerState(t, home, []string{"tasks", "1-step.json"}, task)

	refused := Reopen(home, TransitionOptions{Task: "1-step", Agent: "builder-1", Reason: "repair"})
	if refused.Outcome != "failed" || refused.Error != "not the owner of project demo or the coordinator: builder-1" {
		t.Fatalf("Reopen(builder-1) = %#v; want the owner refusal", refused)
	}
	accepted := Reopen(home, TransitionOptions{Task: "1-step", Agent: "owner-demo", Reason: "repair"})
	if accepted.Outcome != "ok" {
		t.Fatalf("Reopen(owner-demo) = %#v; want ok", accepted)
	}
}
