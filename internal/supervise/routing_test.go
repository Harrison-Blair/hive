package supervise

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func coordinatorRecord(key, pane string) records.Agent {
	return records.Agent{
		Key: key, Name: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: pane,
		WorkspaceID: "w0", TabID: "w0:t1", Status: records.AgentLive,
		AgentStatus: "idle", Generation: 1,
	}
}

func blockedInboxes(t *testing.T, home string, agents []records.Agent, pane string) map[string]int {
	t.Helper()
	for _, record := range agents {
		writeAgent(t, home, record)
	}
	s := handlerSupervisor(t, home, newFakeOps(), newFakeClock())
	event := eventOf("pane.agent_status_changed", `{"pane_id":"`+pane+`","workspace_id":"w1","agent_status":"blocked"}`)
	if err := s.handleEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, key := range []string{"coordinator", "manager-demo", "manager-other"} {
		for _, record := range readInbox(t, home, key) {
			if record.Kind == records.InboxBlocked {
				counts[key]++
			}
		}
	}
	return counts
}

func TestRoutingUsualParentsMatchCurrentRule(t *testing.T) {
	coordinator := coordinatorRecord("coordinator-g1", "w0:p1")
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.Name = "manager-demo"
	manager.Parent = coordinator.Key
	worker := workerRecord("t1-worker-1", "demo", "1-demo", "w1:p2")
	worker.Parent = manager.Key
	verifier := workerRecord("t1-verifier-1", "demo", "1-demo", "w1:p3")
	verifier.Profile = records.ProfileVerifier
	verifier.Parent = manager.Key
	for _, tc := range []struct {
		name string
		pane string
		want string
	}{
		{"worker", worker.PaneID, "manager-demo"},
		{"verifier", verifier.PaneID, "manager-demo"},
		{"manager", manager.PaneID, "coordinator"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			counts := blockedInboxes(t, testHome(t), []records.Agent{coordinator, manager, worker, verifier}, tc.pane)
			if len(counts) != 1 || counts[tc.want] != 1 {
				t.Fatalf("blocked records = %v; want one in %s", counts, tc.want)
			}
		})
	}
}

func TestRoutingWorkerOfCoordinatorGoesToCoordinator(t *testing.T) {
	coordinator := coordinatorRecord("coordinator-g1", "w0:p1")
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	manager.Name = "manager-demo"
	worker := workerRecord("t1-worker-1", "demo", "1-demo", "w1:p2")
	worker.Parent = coordinator.Key
	counts := blockedInboxes(t, testHome(t), []records.Agent{coordinator, manager, worker}, worker.PaneID)
	if len(counts) != 1 || counts["coordinator"] != 1 {
		t.Fatalf("blocked records = %v; want one in coordinator", counts)
	}
}

func TestRoutingWorkerReachesRestartedManager(t *testing.T) {
	coordinator := coordinatorRecord("coordinator-g1", "w0:p1")
	old := managerRecord("manager-demo-g1", "demo", "w1:p1")
	old.Name = "manager-demo"
	old.Status = records.AgentStale
	middle := managerRecord("manager-demo-g2", "demo", "w1:p4")
	middle.Name = "manager-demo"
	middle.Status = records.AgentExited
	middle.Replaces = old.Key
	current := managerRecord("manager-demo-g3", "demo", "w1:p5")
	current.Name = "manager-demo"
	current.Replaces = middle.Key
	current.Generation = 3
	worker := workerRecord("t1-worker-1", "demo", "1-demo", "w1:p2")
	worker.Parent = old.Key
	counts := blockedInboxes(t, testHome(t), []records.Agent{coordinator, old, middle, current, worker}, worker.PaneID)
	if len(counts) != 1 || counts["manager-demo"] != 1 {
		t.Fatalf("blocked records = %v; want one in manager-demo", counts)
	}
}

func TestRoutingParentWithoutLiveSuccessorGoesToCoordinator(t *testing.T) {
	for _, tc := range []struct {
		name   string
		parent string
	}{
		{"stopped parent", "manager-other-g1"},
		{"missing parent", "manager-gone-g1"},
		{"no parent", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coordinator := coordinatorRecord("coordinator-g1", "w0:p1")
			stopped := managerRecord("manager-other-g1", "other", "w1:p1")
			stopped.Name = "manager-other"
			stopped.Status = records.AgentStopped
			worker := workerRecord("t1-worker-1", "other", "1-demo", "w1:p2")
			worker.Parent = tc.parent
			counts := blockedInboxes(t, testHome(t), []records.Agent{coordinator, stopped, worker}, worker.PaneID)
			if len(counts) != 1 || counts["coordinator"] != 1 {
				t.Fatalf("blocked records = %v; want one in coordinator", counts)
			}
		})
	}
}

func TestNudgeIdleWorkerGetsPromptForOwnInbox(t *testing.T) {
	home := testHome(t)
	worker := workerRecord("t1-worker-1", "demo", "1-demo", "w1:p2")
	worker.AgentStatus = "idle"
	writeAgent(t, home, worker)
	appendInbox(t, home, "t1-worker-1", records.InboxMessage, "", "manager-demo", "hello")
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	if err := s.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.nudgeRule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ops.promptCount() != 1 {
		t.Fatalf("prompts = %d; want one nudge to the worker", ops.promptCount())
	}
}
