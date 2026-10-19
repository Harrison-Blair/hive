package supervise

import (
	"context"
	"fmt"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func appendInbox(t *testing.T, home, role, kind, task, agent, text string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope := inbox.AppendHeld(locked, inbox.AppendOptions{Role: role, Kind: kind, Task: task, Agent: agent, Text: text})
		if envelope.Outcome != "ok" {
			return fmt.Errorf("append: %v", envelope.Error)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNudgeIdleManagerGetsPrompt(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	prompt := ops.prompts[0]
	if prompt.Target != "w1:p1" {
		t.Fatalf("target = %q; want w1:p1", prompt.Target)
	}
	if prompt.Text == "" {
		t.Fatalf("prompt text is empty")
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].NudgedAt == "" || all[0].NudgedGeneration != 1 {
		t.Fatalf("record = %+v; want nudged_at and generation 1", all)
	}
}

func TestNudgeDoneManagerGetsPrompt(t *testing.T) {
	home := testHome(t)
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.AgentStatus = "done"
	writeAgent(t, home, manager)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
}

func TestNudgeWorkingManagerWaitsForIdle(t *testing.T) {
	home := testHome(t)
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.AgentStatus = "working"
	writeAgent(t, home, manager)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none while working", ops.promptCount())
	}

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p1","workspace_id":"w1","agent_status":"idle"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one after idle", ops.promptCount())
	}
}

func TestNudgeBlockedManagerNoPrompt(t *testing.T) {
	home := testHome(t)
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.AgentStatus = "blocked"
	writeAgent(t, home, manager)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none while blocked", ops.promptCount())
	}
}

func TestNudgeNewGenerationPromptsAgain(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	// A restarted manager has generation two.
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.Generation = 2
	writeAgent(t, home, manager)
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 2 {
		t.Fatalf("prompts = %d; want two after a new generation", ops.promptCount())
	}
}

func TestNudgePromptFailureRetries(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	ops.promptErr = &herdr.Error{Code: herdr.CodeAgentPromptStalled, Message: "stalled"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "manager-demo")
	if all[0].NudgedAt != "" {
		t.Fatalf("nudged_at = %q; want empty after a failed prompt", all[0].NudgedAt)
	}
	ops.promptErr = nil
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 2 {
		t.Fatalf("prompts = %d; want a retry", ops.promptCount())
	}
}

func TestNudgeNoLiveAgentNoPrompt(t *testing.T) {
	home := testHome(t)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none without a live agent", ops.promptCount())
	}
}
