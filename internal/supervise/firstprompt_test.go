package supervise

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const keptText = "⬢ You are the general agent helper. Your parent is boss."

// keptRecords writes the live parent boss and the general agent helper in
// pane w1:p5 with a kept first prompt and the given agent status.
func keptRecords(t *testing.T, home, agentStatus string) records.Agent {
	t.Helper()
	writeAgent(t, home, records.Agent{
		Key: "boss", Name: "boss", Profile: records.ProfileGeneral, PaneID: "w1:p9", WorkspaceID: "w1", TabID: "w1:t1",
		Status: records.AgentLive, AgentStatus: "working", Generation: 1, Spawned: true,
	})
	helper := records.Agent{
		Key: "helper", Name: "helper", Profile: records.ProfileGeneral, Parent: "boss", PaneID: "w1:p5",
		WorkspaceID: "w1", TabID: "w1:t3", TerminalID: "term-h", Status: records.AgentLive,
		AgentStatus: agentStatus, Generation: 1, Spawned: true, FirstPrompt: keptText,
	}
	writeAgent(t, home, helper)
	return helper
}

// keptOps gives the herdr status of the two panes of keptRecords.
func keptOps(helperStatus string) *fakeOps {
	ops := newFakeOps()
	ops.agent["w1:p9"] = herdr.AgentInfo{PaneID: "w1:p9", WorkspaceID: "w1", TabID: "w1:t1", AgentStatus: "working"}
	ops.agent["w1:p5"] = herdr.AgentInfo{PaneID: "w1:p5", TerminalID: "term-h", WorkspaceID: "w1", TabID: "w1:t3", AgentStatus: helperStatus}
	return ops
}

func unresolvedCount(t *testing.T, home, role string) int {
	t.Helper()
	count := 0
	for _, record := range readInbox(t, home, role) {
		if record.Kind == records.InboxUnresolved {
			count++
		}
	}
	return count
}

func TestKeptFirstPromptSentOneTimeAcrossTwoPasses(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	ops := keptOps("idle")
	s := handlerSupervisor(t, home, ops, newFakeClock())

	for pass := 1; pass <= 2; pass++ {
		if err := s.refresh(context.Background()); err != nil {
			t.Fatalf("pass %d: %v", pass, err)
		}
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	if prompt := ops.prompts[0]; prompt.Target != "w1:p5" || prompt.Text != keptText {
		t.Fatalf("prompt = %+v; want the kept first prompt to w1:p5", prompt)
	}
	stored := readAgent(t, home, "helper")
	if stored.FirstPromptAttemptAt == "" || stored.FirstPromptSentAt == "" || stored.Status != records.AgentLive {
		t.Fatalf("record = %+v; want live with both first prompt times", stored)
	}
}

// TestKeptFirstPromptSentWhenDone covers an agent in a tab that nobody saw.
// herdr gives done, the same ready status as idle.
func TestKeptFirstPromptSentWhenDone(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "blocked")
	ops := keptOps("done")
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 || ops.prompts[0].Text != keptText {
		t.Fatalf("prompts = %+v; want the kept first prompt", ops.prompts)
	}
}

func TestKeptFirstPromptWaitsWhileBlocked(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "blocked")
	ops := keptOps("blocked")
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none while blocked", ops.promptCount())
	}
	if stored := readAgent(t, home, "helper"); stored.FirstPromptAttemptAt != "" {
		t.Fatalf("first_prompt_attempt_at = %q; want empty", stored.FirstPromptAttemptAt)
	}
}

// TestKeptFirstPromptRestartBetweenAttemptAndSent covers a supervisor restart
// between step 2 and step 4 of decision 16.
func TestKeptFirstPromptRestartBetweenAttemptAndSent(t *testing.T) {
	home := testHome(t)
	helper := keptRecords(t, home, "idle")
	helper.FirstPromptAttemptAt = "2026-01-01T00:00:00Z"
	writeAgent(t, home, helper)
	ops := keptOps("idle")
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none after a restart between the steps", ops.promptCount())
	}
	if stored := readAgent(t, home, "helper"); stored.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", stored.Status)
	}
	if got := unresolvedCount(t, home, "boss"); got != 1 {
		t.Fatalf("unresolved records for the parent = %d; want 1", got)
	}
}

func TestKeptFirstPromptUncertainSendIsUnresolvedAndNotSentAgain(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	ops := keptOps("idle")
	ops.promptErr = errors.New("connection reset")
	s := handlerSupervisor(t, home, ops, newFakeClock())

	for pass := 1; pass <= 2; pass++ {
		if err := s.refresh(context.Background()); err != nil {
			t.Fatalf("pass %d: %v", pass, err)
		}
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	stored := readAgent(t, home, "helper")
	if stored.Status != records.AgentUnresolved || stored.FirstPromptAttemptAt == "" || stored.FirstPromptSentAt != "" {
		t.Fatalf("record = %+v; want unresolved with an attempt and no sent time", stored)
	}
	if got := unresolvedCount(t, home, "boss"); got != 1 {
		t.Fatalf("unresolved records for the parent = %d; want 1", got)
	}
}

func TestKeptFirstPromptFailedAttemptWriteDoesNotSend(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	ops := keptOps("idle")
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	agents := filepath.Join(home, ".hive", "state", "agents")
	if err := os.Chmod(agents, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(agents, 0o755) })

	_ = s.firstPromptRule(context.Background())
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want none after a failed attempt write", ops.promptCount())
	}
	if stored := readAgent(t, home, "helper"); stored.FirstPromptAttemptAt != "" {
		t.Fatalf("first_prompt_attempt_at = %q; want empty", stored.FirstPromptAttemptAt)
	}
}

func TestKeptFirstPromptFailedSentWriteIsUnresolvedAndNotSentAgain(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	ops := keptOps("idle")
	agents := filepath.Join(home, ".hive", "state", "agents")
	ops.promptFn = func(herdr.AgentPromptParams) error {
		return os.Chmod(agents, 0o555)
	}
	t.Cleanup(func() { _ = os.Chmod(agents, 0o755) })
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := s.firstPromptRule(context.Background()); err == nil {
		t.Fatal("firstPromptRule() = nil; want the failed sent write")
	}
	if err := os.Chmod(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	ops.promptFn = nil
	restarted := handlerSupervisor(t, home, ops, newFakeClock())
	if err := restarted.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	if stored := readAgent(t, home, "helper"); stored.Status != records.AgentUnresolved || stored.FirstPromptSentAt != "" {
		t.Fatalf("record = %+v; want unresolved with no sent time", stored)
	}
}

func TestKeptFirstPromptBlocksOtherNudges(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	appendInbox(t, home, "helper", records.InboxMessage, "", "boss", "more")
	ops := keptOps("idle")
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 0 {
		t.Fatalf("prompts = %d; want no nudge before the kept first prompt", ops.promptCount())
	}
}

func TestKeptFirstPromptGoesBeforeOtherNudges(t *testing.T) {
	home := testHome(t)
	keptRecords(t, home, "idle")
	writeAgent(t, home, managerRecord("aa-manager-g1", "aa", "w1:p1"))
	appendInbox(t, home, "manager-aa", records.InboxCompleted, "1-aa", "worker-1", "completed")
	ops := keptOps("idle")
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 2 || ops.prompts[0].Target != "w1:p5" || ops.prompts[1].Target != "w1:p1" {
		t.Fatalf("prompts = %+v; want the kept first prompt, then the nudge", ops.prompts)
	}
}
