package reopen

import (
	"testing"

	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// TestTaskReopenNeedsTheProjectOwnerOrCoordinator proves AC7 for reopen.
func TestTaskReopenNeedsTheProjectOwnerOrCoordinator(t *testing.T) {
	home := tasktest.SeedOwnerHome(t)
	task := tasktest.ReadOwnerRecord(t, home, "tasks", "1-step.json")
	task["status"] = "verified"
	task["verification"] = map[string]any{"agent": "verifier-1", "result": "pass", "reason": "", "at": "2026-01-01T00:00:00Z"}
	tasktest.WriteOwnerState(t, home, []string{"tasks", "1-step.json"}, task)

	refused := Run(home, tasklib.TransitionOptions{Task: "1-step", Agent: "builder-1", Reason: "repair"})
	if refused.Outcome != "failed" || refused.Error != "not the owner of project demo or the coordinator: builder-1" {
		t.Fatalf("Reopen(builder-1) = %#v; want the owner refusal", refused)
	}
	accepted := Run(home, tasklib.TransitionOptions{Task: "1-step", Agent: "owner-demo", Reason: "repair"})
	if accepted.Outcome != "ok" {
		t.Fatalf("Reopen(owner-demo) = %#v; want ok", accepted)
	}
}
