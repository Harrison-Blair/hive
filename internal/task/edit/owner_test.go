package edit

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// TestTaskEditNeedsTheProjectOwnerOrCoordinator proves AC7 with the exact
// refusal text, for the owner, the coordinator, and another agent.
func TestTaskEditNeedsTheProjectOwnerOrCoordinator(t *testing.T) {
	home := tasktest.SeedOwnerHome(t)
	tasktest.WriteOwnerState(t, home, []string{"agents", "coordinator-g1.json"}, map[string]any{
		"key": "coordinator-g1", "name": "coordinator", "kind": "coordinator",
		"profile": "coordinator", "status": "live", "generation": 1, "pane_id": "w0:p1",
	})
	tasktest.WriteOwnerState(t, home, []string{"claims", "coordinator.json"}, map[string]any{
		"key": "coordinator", "agent": "coordinator-g1", "generation": 1,
	})

	refused := Run(home, Options{Task: "1-step", Agent: "builder-1", NeedsUserSet: true, NeedsUser: "wait"})
	if refused.Outcome != "failed" || refused.Error != "not the owner of project demo or the coordinator: builder-1" {
		t.Fatalf("Edit(builder-1) = %#v; want the owner refusal", refused)
	}
	if record := tasktest.ReadOwnerRecord(t, home, "tasks", "1-step.json"); record["needs_user"] != nil {
		t.Fatalf("task = %#v; want no needs_user write", record)
	}

	for _, agent := range []string{"owner-demo", "coordinator"} {
		envelope := Run(home, Options{Task: "1-step", Agent: agent, NeedsUserSet: true, NeedsUser: "wait for " + agent})
		if envelope.Outcome != "ok" {
			t.Fatalf("Edit(%s) = %#v; want ok", agent, envelope)
		}
	}
}
