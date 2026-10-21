package supervise

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestVerifyInitialIdleStartsQuietWindow covers B2. Startup reconcile must
// start a quiet timer for an unfinished live agent that is already idle.
func TestVerifyInitialIdleStartsQuietWindow(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.TerminalID = "term-w"
	worker.AgentStatus = "idle"
	writeAgent(t, home, worker)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-w", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "idle"}
	clock := newFakeClock()
	s := handlerSupervisor(t, home, ops, clock)

	if err := s.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := clock.timerCount(); got != 1 {
		t.Fatalf("quiet timers after startup refresh = %d; want 1", got)
	}
	clock.fire()
	if err := s.handleQuiet(context.Background(), worker.Key); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "coordinator")
	if len(all) != 1 || all[0].Kind != records.InboxQuiet {
		t.Fatalf("inbox = %+v; want one quiet record", all)
	}
}

// TestVerifyStoreRefreshKeepsBlockedPush covers B3. A store refresh must not
// discard the blocked notification of a queued status push.
func TestVerifyStoreRefreshKeepsBlockedPush(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.TerminalID = "term-w"
	worker.AgentStatus = "working"
	writeAgent(t, home, worker)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-w", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "blocked"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	// The store refresh runs before the queued blocked push.
	if err := s.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"blocked"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "coordinator")
	if len(all) != 1 || all[0].Kind != records.InboxBlocked {
		t.Fatalf("blocked notifications = %+v; want one blocked record", all)
	}
	if updated := readAgent(t, home, worker.Key); updated.AgentStatus != "blocked" {
		t.Fatalf("agent_status = %q; want blocked", updated.AgentStatus)
	}
}

// TestVerifyDoneThenIdleIsOneQuietPeriod covers B4. A done then idle change
// with no working between must not start a second quiet window.
func TestVerifyDoneThenIdleIsOneQuietPeriod(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "working"
	writeAgent(t, home, worker)
	ops := newFakeOps()
	clock := newFakeClock()
	s := handlerSupervisor(t, home, ops, clock)

	done := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"done"}`)
	if err := s.handleEvent(context.Background(), done); err != nil {
		t.Fatal(err)
	}
	clock.fire()
	if err := s.handleQuiet(context.Background(), worker.Key); err != nil {
		t.Fatal(err)
	}

	// herdr reports done to idle when a tab receives focus. No working
	// transition starts a new period.
	idle := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"idle"}`)
	if err := s.handleEvent(context.Background(), idle); err != nil {
		t.Fatal(err)
	}
	clock.fire()
	if err := s.handleQuiet(context.Background(), worker.Key); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "coordinator")
	if len(all) != 1 {
		t.Fatalf("quiet records without working transition = %d; want 1", len(all))
	}
}

// TestVerifyRunKeepsPromptEffectOnMarkFailure covers B5. The exit envelope
// keeps the committed prompt effect after the mark write fails.
func TestVerifyRunKeepsPromptEffectOnMarkFailure(t *testing.T) {
	home := testHome(t)
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.TerminalID = "term-m"
	writeAgent(t, home, manager)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	all := readInbox(t, home, "manager-demo")
	markPath := filepath.Join(home, ".hive", "state", "inbox", "manager-demo", strconv.FormatUint(all[0].Seq, 10)+".json")

	ops := newFakeOps()
	ops.agent["w1:p1"] = herdr.AgentInfo{PaneID: "w1:p1", TerminalID: "term-m", WorkspaceID: "w1", TabID: "w1:t1", AgentStatus: "idle"}
	ops.promptFn = func(herdr.AgentPromptParams) error {
		// The prompt succeeds, then the mark path becomes unwritable.
		_ = os.Remove(markPath)
		_ = os.Mkdir(markPath, 0o755)
		return nil
	}
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	_, envelope := Run(context.Background(), Options{
		Home:        home,
		Out:         io.Discard,
		PaneID:      "w1:p0",
		WorkspaceID: "w1",
		TabID:       "w1:t0",
		PID:         pid,
		PIDStart:    start,
		Operations:  ops,
		Clock:       newFakeClock(),
	})
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one", ops.promptCount())
	}
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q error = %v; want partial", envelope.Outcome, envelope.Error)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Action != "nudged" {
		t.Fatalf("effects = %+v; want the nudged prompt effect", envelope.Effects)
	}
}

// TestVerifySignalDuringInitialRebuildExitsCleanly covers B6. A signal during
// the initial rebuild must exit cleanly and close the subscription.
func TestVerifySignalDuringInitialRebuildExitsCleanly(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.TerminalID = "term-w"
	worker.AgentStatus = "idle"
	writeAgent(t, home, worker)

	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-w", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "idle"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gets := 0
	ops.getAgentFn = func(string) {
		gets++
		// The first get is the reconcile before the subscription opens.
		// Cancel on the first post-subscribe get of the rebuild.
		if gets == 2 {
			cancel()
		}
	}

	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	_, envelope := Run(ctx, Options{
		Home:        home,
		Out:         io.Discard,
		PaneID:      "w1:p0",
		WorkspaceID: "w1",
		TabID:       "w1:t0",
		PID:         pid,
		PIDStart:    start,
		Operations:  ops,
		Clock:       newFakeClock(),
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome = %q error = %v; want a clean signal exit", envelope.Outcome, envelope.Error)
	}
	stream := ops.lastStream()
	if stream == nil || !stream.isClosed() {
		t.Fatalf("stream_closed = false; want the subscription closed")
	}
}

// TestVerifyMalformedInboxReadStopsEvent covers B7. A completion lookup that
// fails to read the store is a storage failure, not a missing completion.
func TestVerifyMalformedInboxReadStopsEvent(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "working"
	writeAgent(t, home, worker)
	inboxDir := filepath.Join(home, ".hive", "state", "inbox", "manager-demo")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, "1.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"idle"}`)
	err := s.handleEvent(context.Background(), event)
	if err == nil {
		t.Fatalf("handleEvent returned nil; want a storage read failure")
	}
	if updated := readAgent(t, home, worker.Key); updated.AgentStatus != "working" {
		t.Fatalf("agent_status = %q; want an unchanged working", updated.AgentStatus)
	}
}

// TestVerifyWorkspaceClosedMarksEveryLiveRecord covers B8. One topology close
// resolves every matching live record.
func TestVerifyWorkspaceClosedMarksEveryLiveRecord(t *testing.T) {
	home := testHome(t)
	first := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	second := workerRecord("worker-2", "demo", "2-demo", "w1:p3")
	writeAgent(t, home, first)
	writeAgent(t, home, second)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("workspace_closed", `{"type":"workspace_closed","workspace_id":"w1"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "coordinator")
	if len(all) != 2 {
		t.Fatalf("workspace close notifications = %d; want 2", len(all))
	}
	for _, key := range []string{first.Key, second.Key} {
		if updated := readAgent(t, home, key); updated.Status != records.AgentUnresolved {
			t.Fatalf("%s status = %q; want unresolved", key, updated.Status)
		}
	}
}
