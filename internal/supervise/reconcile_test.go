package supervise

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestReconcileMatchingTerminalKeepsLive(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	record.TerminalID = "term-a"
	writeAgent(t, home, record)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-a", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "working"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	updated := readAgent(t, home, record.Key)
	if updated.Status != records.AgentLive || updated.AgentStatus != "working" {
		t.Fatalf("record = %+v; want live with the stored status", updated)
	}
	if len(readInbox(t, home, "manager-demo")) != 0 {
		t.Fatalf("want no inbox record")
	}
}

func TestReconcileAnotherTerminalUnresolved(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	record.TerminalID = "term-a"
	writeAgent(t, home, record)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-b", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "working"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	updated := readAgent(t, home, record.Key)
	if updated.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", updated.Status)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxUnresolved {
		t.Fatalf("inbox = %+v; want one unresolved record", all)
	}
}

func TestReconcileAgentNotFoundShellPaneExited(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, record)
	ops := newFakeOps()
	ops.pane["w1:p2"] = herdr.PaneInfo{PaneID: "w1:p2", TerminalID: "term-a", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "unknown"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	updated := readAgent(t, home, record.Key)
	if updated.Status != records.AgentExited {
		t.Fatalf("status = %q; want exited", updated.Status)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxExited {
		t.Fatalf("inbox = %+v; want one exited record", all)
	}
}

func TestReconcileAgentNotFoundMissingPaneUnresolved(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, record)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, record.Key); updated.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", updated.Status)
	}
}

func TestReconcileStartingWithMissingPaneUnresolved(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	record.Status = records.AgentStarting
	writeAgent(t, home, record)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, record.Key); updated.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", updated.Status)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxUnresolved {
		t.Fatalf("inbox = %+v; want one unresolved record", all)
	}
}

func TestReconcileStartingShellWaitsForSpawn(t *testing.T) {
	home := testHome(t)
	record := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	record.Status = records.AgentStarting
	writeAgent(t, home, record)
	ops := newFakeOps()
	ops.pane["w1:p2"] = herdr.PaneInfo{PaneID: "w1:p2", TerminalID: "term-a", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "unknown"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, record.Key); updated.Status != records.AgentStarting {
		t.Fatalf("status = %q; want starting", updated.Status)
	}
	if all := readInbox(t, home, "manager-demo"); len(all) != 0 {
		t.Fatalf("inbox = %+v; want no records", all)
	}
}

func TestReconcileAdoptsTerminalID(t *testing.T) {
	home := testHome(t)
	coordinator := records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator",
		PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1",
		Status: records.AgentLive, AgentStatus: "idle", Generation: 1, Spawned: true,
	}
	writeAgent(t, home, coordinator)
	ops := newFakeOps()
	ops.agent["w1:p1"] = herdr.AgentInfo{PaneID: "w1:p1", TerminalID: "term-c", WorkspaceID: "w1", TabID: "w1:t1", AgentStatus: "idle"}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, coordinator.Key); updated.TerminalID != "term-c" {
		t.Fatalf("terminal_id = %q; want term-c", updated.TerminalID)
	}
}

func TestReconcileOldSupervisorStoppedWithoutHerdr(t *testing.T) {
	home := testHome(t)
	old := records.Agent{
		Key: "supervisor-g0", Name: "supervisor", Role: "supervisor",
		PaneID: "w1:p9", WorkspaceID: "w1", TabID: "w1:t9",
		Status: records.AgentLive, AgentStatus: "idle", Generation: 0, Spawned: false,
	}
	writeAgent(t, home, old)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, old.Key); updated.Status != records.AgentStopped {
		t.Fatalf("status = %q; want stopped", updated.Status)
	}
}

func TestReconcileUnknownAgentLogsOnly(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	ops.list = []herdr.AgentInfo{{PaneID: "w1:p7", TerminalID: "term-z", WorkspaceID: "w1", TabID: "w1:t7", AgentStatus: "idle"}}
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
}
