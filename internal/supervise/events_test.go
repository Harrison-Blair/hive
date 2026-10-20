package supervise

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func workerRecord(key, project, task, pane string) records.Agent {
	return records.Agent{
		Key:         key,
		Name:        key,
		Role:        "worker",
		Project:     project,
		Task:        task,
		PaneID:      pane,
		WorkspaceID: "w1",
		TabID:       "w1:t2",
		Status:      records.AgentLive,
		AgentStatus: "working",
		Generation:  1,
		Spawned:     true,
	}
}

func managerRecord(key, project, pane string) records.Agent {
	return records.Agent{
		Key:         key,
		Name:        key,
		Role:        "manager",
		Project:     project,
		PaneID:      pane,
		WorkspaceID: "w1",
		TabID:       "w1:t1",
		Status:      records.AgentLive,
		AgentStatus: "idle",
		Generation:  1,
		Spawned:     true,
	}
}

func TestEventsBlockedWritesCounterRecordAndStatus(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"blocked"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxBlocked {
		t.Fatalf("inbox = %+v; want one blocked record", all)
	}
	if updated := readAgent(t, home, worker.Key); updated.AgentStatus != "blocked" {
		t.Fatalf("agent_status = %q; want blocked", updated.AgentStatus)
	}
	wantChangedAt := "2026-01-01T00:00:00Z"
	if got := readAgentStatusChangedAt(t, home, worker.Key); got != wantChangedAt {
		t.Fatalf("status_changed_at = %q; want %q", got, wantChangedAt)
	}

	// A repeated push with the stored status writes nothing.
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if all := readInbox(t, home, "manager-demo"); len(all) != 1 {
		t.Fatalf("inbox after repeat = %+v; want one blocked record", all)
	}
	if got := readAgentStatusChangedAt(t, home, worker.Key); got != wantChangedAt {
		t.Fatalf("status_changed_at after repeat = %q; want %q", got, wantChangedAt)
	}
}

func TestAppendAndStatusKeepsTimestampWhenAgentStatusDoesNotChange(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgentWithStatusChangedAt(t, home, worker, "2025-12-31T23:59:00Z")
	s := handlerSupervisor(t, home, newFakeOps(), newFakeClock())

	envelope := s.appendAndStatus(context.Background(), notification{
		role: "manager-demo", kind: records.InboxBlocked, task: worker.Task,
		agent: worker.Name, text: "blocked", agentKey: worker.Key, agentStatus: "working",
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("appendAndStatus() = %#v; want success", envelope)
	}
	if got := readAgentStatusChangedAt(t, home, worker.Key); got != "2025-12-31T23:59:00Z" {
		t.Fatalf("status_changed_at after unchanged write = %q; want the original timestamp", got)
	}
}

func TestStatusOnlyKeepsTimestampWhenAgentStatusDoesNotChange(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "working"
	writeAgentWithStatusChangedAt(t, home, worker, "2025-12-31T23:59:00Z")
	s := handlerSupervisor(t, home, newFakeOps(), newFakeClock())

	if envelope := s.statusOnly(context.Background(), worker, "idle"); envelope.Outcome != "ok" {
		t.Fatalf("statusOnly() = %#v; want success", envelope)
	}
	const changedAt = "2026-01-01T00:00:00Z"
	if got := readAgentStatusChangedAt(t, home, worker.Key); got != changedAt {
		t.Fatalf("status_changed_at = %q; want %q", got, changedAt)
	}
	if envelope := s.statusOnly(context.Background(), worker, "idle"); envelope.Outcome != "ok" {
		t.Fatalf("repeated statusOnly() = %#v; want success", envelope)
	}
	if got := readAgentStatusChangedAt(t, home, worker.Key); got != changedAt {
		t.Fatalf("status_changed_at after unchanged write = %q; want %q", got, changedAt)
	}
}

func writeAgentWithStatusChangedAt(t *testing.T, home string, record records.Agent, changedAt string) {
	t.Helper()
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	value["status_changed_at"] = changedAt
	data, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".hive", "state", "agents", record.Key+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readAgentStatusChangedAt(t *testing.T, home, key string) string {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "agents", key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	changedAt, _ := value["status_changed_at"].(string)
	return changedAt
}

func TestEventsWorkingStoresStatusWithoutRecord(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "idle"
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"working"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, worker.Key); updated.AgentStatus != "working" {
		t.Fatalf("agent_status = %q; want working", updated.AgentStatus)
	}
	if all := readInbox(t, home, "manager-demo"); len(all) != 0 {
		t.Fatalf("inbox = %+v; want no record", all)
	}
}

func TestEventsBlockedWorkingBlockedWritesTwoRecords(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "idle"
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	for _, status := range []string{"blocked", "working", "blocked"} {
		event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"`+status+`"}`)
		if err := s.handleEvent(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 2 {
		t.Fatalf("inbox = %+v; want two blocked records", all)
	}
}

func TestEventsFinishedAgentStops(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "idle"
	worker.Finished = true
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"done"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if ops.stopCount() != 1 {
		t.Fatalf("stops = %d; want one", ops.stopCount())
	}
}

func TestEventsFinishedMarkOnIdleStopsAtEvaluate(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "idle"
	worker.Finished = true
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.stopRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.stopCount() != 1 {
		t.Fatalf("stops = %d; want one", ops.stopCount())
	}
}

func TestEventsDoneIsIdleQuiet(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, worker)
	ops := newFakeOps()
	clock := newFakeClock()
	s := handlerSupervisor(t, home, ops, clock)

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"done"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	clock.fire()
	if err := s.handleQuiet(context.Background(), "worker-1"); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxQuiet {
		t.Fatalf("inbox = %+v; want one quiet record", all)
	}
	// A second quiet fire for the same idle period writes no second record.
	if err := s.handleQuiet(context.Background(), "worker-1"); err != nil {
		t.Fatal(err)
	}
	if all := readInbox(t, home, "manager-demo"); len(all) != 1 {
		t.Fatalf("inbox after second fire = %+v; want one record", all)
	}
}

func TestEventsAgentDetectedReleasedStartingIgnored(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.Status = records.AgentStarting
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	var log bytes.Buffer
	s.logger = slog.New(slog.NewTextHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))

	event := eventOf("pane_agent_detected", `{"pane_id":"w1:p2","workspace_id":"w1","released":true,"final_status":"unknown"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, worker.Key); updated.Status != records.AgentStarting {
		t.Fatalf("status = %q; want starting", updated.Status)
	}
	if all := readInbox(t, home, "manager-demo"); len(all) != 0 {
		t.Fatalf("inbox = %+v; want no records", all)
	}
	if !strings.Contains(log.String(), "ignored released pane while starting") {
		t.Fatalf("log = %q; want an ignored-event entry", log.String())
	}
}

func TestEventsAgentDetectedReleasedExited(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane_agent_detected", `{"pane_id":"w1:p2","workspace_id":"w1","released":true,"final_status":"unknown"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxExited {
		t.Fatalf("inbox = %+v; want one exited record", all)
	}
	if updated := readAgent(t, home, worker.Key); updated.Status != records.AgentExited {
		t.Fatalf("status = %q; want exited", updated.Status)
	}
}

func TestEventsStartingTopologyClosedUnresolved(t *testing.T) {
	for _, test := range []struct {
		name  string
		event string
		data  string
	}{
		{name: "pane", event: "pane_closed", data: `{"pane_id":"w1:p2"}`},
		{name: "tab", event: "tab_closed", data: `{"tab_id":"w1:t2"}`},
		{name: "workspace", event: "workspace_closed", data: `{"workspace_id":"w1"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
			worker.Status = records.AgentStarting
			writeAgent(t, home, worker)
			s := handlerSupervisor(t, home, newFakeOps(), newFakeClock())

			if err := s.handleEvent(context.Background(), eventOf(test.event, test.data)); err != nil {
				t.Fatal(err)
			}
			if updated := readAgent(t, home, worker.Key); updated.Status != records.AgentUnresolved {
				t.Fatalf("status = %q; want unresolved", updated.Status)
			}
			all := readInbox(t, home, "manager-demo")
			if len(all) != 1 || all[0].Kind != records.InboxUnresolved {
				t.Fatalf("inbox = %+v; want one unresolved record", all)
			}
		})
	}
}

func TestEventsStartingTopologyClosedWithoutControllerUnresolved(t *testing.T) {
	home := testHome(t)
	coordinator := records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator",
		PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1",
		Status: records.AgentStarting, Generation: 1, Spawned: true,
	}
	writeAgent(t, home, coordinator)
	s := handlerSupervisor(t, home, newFakeOps(), newFakeClock())

	if err := s.handleEvent(context.Background(), eventOf("workspace_closed", `{"workspace_id":"w1"}`)); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, coordinator.Key); updated.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", updated.Status)
	}
	if all := readInbox(t, home, "coordinator"); len(all) != 0 {
		t.Fatalf("inbox = %+v; want no record without a controlling role", all)
	}
}

func TestEventsWorkspaceClosedUnresolved(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("workspace_closed", `{"type":"workspace_closed","workspace_id":"w1"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	all := readInbox(t, home, "manager-demo")
	if len(all) != 1 || all[0].Kind != records.InboxUnresolved {
		t.Fatalf("inbox = %+v; want one unresolved record", all)
	}
	if updated := readAgent(t, home, worker.Key); updated.Status != records.AgentUnresolved {
		t.Fatalf("status = %q; want unresolved", updated.Status)
	}
}

func TestEventsPaneMovedUpdatesRecord(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	writeAgent(t, home, worker)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane_moved", `{"type":"pane_moved","previous_pane_id":"w1:p2","previous_workspace_id":"w1","previous_tab_id":"w1:t2","pane":{"pane_id":"w2:p5","workspace_id":"w2","tab_id":"w2:t1"}}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	updated := readAgent(t, home, worker.Key)
	if updated.PaneID != "w2:p5" || updated.WorkspaceID != "w2" || updated.TabID != "w2:t1" {
		t.Fatalf("moved record = %+v; want the new pane fields", updated)
	}
	if len(updated.PaneHistory) != 1 || updated.PaneHistory[0] != "w1:p2" {
		t.Fatalf("pane_history = %v; want [w1:p2]", updated.PaneHistory)
	}
}

func TestEventsNoControllingRoleLogsOnly(t *testing.T) {
	home := testHome(t)
	coordinator := records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator",
		PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1",
		Status: records.AgentLive, AgentStatus: "working", Generation: 1, Spawned: true,
	}
	writeAgent(t, home, coordinator)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	event := eventOf("pane.agent_status_changed", `{"pane_id":"w1:p1","workspace_id":"w1","agent_status":"blocked"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if updated := readAgent(t, home, coordinator.Key); updated.AgentStatus != "working" {
		t.Fatalf("agent_status = %q; want an unchanged working", updated.AgentStatus)
	}
}

func TestEventsStatusGroupFailures(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, workerRecord("worker-1", "demo", "1-demo", "w1:p2"))
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	// The agent key of the final status is absent, so the status write fails.
	envelope := s.applyNotification(context.Background(), notification{
		role:     "manager-demo",
		kind:     records.InboxBlocked,
		task:     "1-demo",
		agent:    "worker-1",
		text:     "blocked",
		agentKey: "missing-worker",
	})
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if len(readInbox(t, home, "manager-demo")) != 1 {
		t.Fatalf("want one appended record")
	}
}

func TestEventsStatusOnlyFailure(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	envelope := s.statusOnly(context.Background(), records.Agent{Key: "missing", Generation: 1}, "working")
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if len(envelope.Effects) != 0 {
		t.Fatalf("effects = %v; want none", envelope.Effects)
	}
}

func TestEventsCounterFailureHasNoEffects(t *testing.T) {
	home := testHome(t)
	writeAgent(t, home, workerRecord("worker-1", "demo", "1-demo", "w1:p2"))
	// A directory at the counter path makes the counter read fail.
	state := home + "/.hive/state"
	if err := os.MkdirAll(state+"/counters/inbox:manager-demo.json", 0o755); err != nil {
		t.Fatal(err)
	}
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())

	envelope := s.applyNotification(context.Background(), notification{
		role: "manager-demo", kind: records.InboxBlocked, agent: "worker-1", text: "blocked",
	})
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("envelope = %+v; want failed with no effects", envelope)
	}
}
