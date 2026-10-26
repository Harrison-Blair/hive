package supervise

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// ownerRecord returns one live owner record with the facts of D1 and D7.
func ownerRecord(key, name, project, pane string) records.Agent {
	return records.Agent{
		Key: key, Name: name, Profile: records.ProfilePlanner, Owns: project, Project: project, PaneID: pane,
		WorkspaceID: "w1", TabID: "w1:t1", Status: records.AgentLive,
		AgentStatus: "idle", Generation: 1, Spawned: true,
	}
}

// TestNudgeProjectOwner proves AC4: the supervisor nudges the owner for a
// record in the inbox of the owner.
func TestNudgeProjectOwner(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, ownerRecord("owner-demo-g1", "owner-demo", "demo", "w1:p1"))
	appendInbox(t, home, "owner-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one nudge of the owner", ops.promptCount())
	}
	if target := ops.prompts[0].Target; target != "w1:p1" {
		t.Fatalf("target = %q; want the owner pane w1:p1", target)
	}
	all := readInbox(t, home, "owner-demo")
	if len(all) != 1 || all[0].NudgedAt == "" || all[0].NudgedGeneration != 1 {
		t.Fatalf("record = %+v; want nudged_at and generation 1", all)
	}
}

// TestWorkerSeesProjectCompletedRecord proves that the completed record of a
// task reaches the owner inbox that the task uses.
func TestWorkerSeesProjectCompletedRecord(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, ownerRecord("owner-demo-g1", "owner-demo", "demo", "w1:p1"))
	writeAgent(t, home, workerRecord("builder-1", "demo", "1-demo", "w1:p2"))
	appendInbox(t, home, "owner-demo", records.InboxCompleted, "1-demo", "builder-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	completed, err := s.hasCompletedRecord(context.Background(), s.agents["builder-1"])
	if err != nil {
		t.Fatal(err)
	}
	if !completed {
		t.Fatal("hasCompletedRecord() = false; want the completed record in the owner inbox")
	}
}
