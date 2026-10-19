package supervise

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

func TestSubscriptionEntriesPerPaneAndPlain(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	writeAgent(t, home, workerRecord("worker-1", "demo", "1-demo", "w1:p2"))
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(ops.subs) != 1 {
		t.Fatalf("subscribe calls = %d; want one", len(ops.subs))
	}
	subscriptions := ops.subs[0].Subscriptions
	status := 0
	plain := map[string]bool{}
	for _, entry := range subscriptions {
		if entry.Type == "pane.agent_status_changed" {
			if entry.PaneID == "" {
				t.Fatalf("status entry has no pane id")
			}
			status++
			continue
		}
		plain[entry.Type] = true
	}
	if status != 2 {
		t.Fatalf("status entries = %d; want two", status)
	}
	// The expected plain selectors are literal and independent of the
	// production list.
	wantPlain := []string{
		"pane.agent_detected",
		"pane.exited",
		"pane.closed",
		"tab.closed",
		"workspace.closed",
		"pane.moved",
	}
	for _, kind := range wantPlain {
		if !plain[kind] {
			t.Fatalf("plain entry %q is missing", kind)
		}
	}
	if len(plain) != len(wantPlain) {
		t.Fatalf("plain entries = %v; want six", plain)
	}
}

func TestSubscriptionRebuildOnNewRecord(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := ops.lastStream()

	writeAgent(t, home, workerRecord("worker-1", "demo", "1-demo", "w1:p2"))
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(ops.streams) != 2 {
		t.Fatalf("streams = %d; want two", len(ops.streams))
	}
	if !first.isClosed() {
		t.Fatalf("the old subscription was not closed")
	}
}

func TestSubscriptionNoRebuildWhenUnchanged(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, managerRecord("manager-demo-g1", "demo", "w1:p1"))
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(ops.streams) != 1 {
		t.Fatalf("streams = %d; want one", len(ops.streams))
	}
}

func TestSubscriptionAppliesPaneStatusAfterRebuild(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	record.AgentStatus = "working"
	record.TerminalID = "term-w"
	writeAgent(t, home, record)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-w", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "idle"}
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, record.Key); updated.AgentStatus != "idle" {
		t.Fatalf("agent_status = %q; want idle", updated.AgentStatus)
	}
}

func TestSubscriptionRebuildOnPaneMove(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, workerRecord("worker-1", "demo", "1-demo", "w1:p2"))
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.resubscribe(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := ops.lastStream()

	event := eventOf("pane_moved", `{"type":"pane_moved","previous_pane_id":"w1:p2","pane":{"pane_id":"w2:p5","workspace_id":"w2","tab_id":"w2:t1"}}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if len(ops.streams) != 2 {
		t.Fatalf("streams = %d; want two after the move", len(ops.streams))
	}
	if !first.isClosed() {
		t.Fatalf("the old subscription was not closed")
	}
	if updated := readAgent(t, home, "worker-1"); updated.PaneID != "w2:p5" {
		t.Fatalf("pane_id = %q; want w2:p5", updated.PaneID)
	}
}

func TestSubscriptionPlainFramesCarryType(t *testing.T) {
	event := eventOf("pane_exited", `{"type":"pane_exited","pane_id":"w1:p2","workspace_id":"w1"}`)
	var parsed eventData
	if err := json.Unmarshal(event.Data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.PaneID != "w1:p2" {
		t.Fatalf("pane_id = %q; want w1:p2", parsed.PaneID)
	}
}
