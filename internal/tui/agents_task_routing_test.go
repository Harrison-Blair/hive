package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// agentsTaskRoutingData gives two goals and one task with no goal. Each task
// has a live worker, so the Agents screen selects one card of each goal case.
func agentsTaskRoutingData() demoData {
	return demoData{
		Supervisor: demoSupervisor{Live: true, Agent: "supervisor-g1", PID: 4242},
		Projects:   []records.Project{{ID: "hive", Path: "/home/hive"}},
		Goals: []records.Goal{
			{ID: "g1-alpha", Project: "hive", Title: "Alpha goal", Status: "active", CreatedAt: "2026-10-01T10:00:00Z", Waves: []records.GoalWave{
				{ID: "A", DoneWhen: "dev", Tasks: []string{"10-alpha"}},
			}},
			{ID: "g2-beta", Project: "hive", Title: "Beta goal", Status: "active", CreatedAt: "2026-10-02T10:00:00Z", Waves: []records.GoalWave{
				{ID: "B", DoneWhen: "dev", Tasks: []string{"20-beta"}},
			}},
		},
		Tasks: []records.Task{
			{ID: "10-alpha", Project: "hive", Title: "Alpha task", Feature: "routing", Status: "assigned"},
			{ID: "20-beta", Project: "hive", Title: "Beta task", Feature: "routing", Status: "created"},
			{ID: "30-loose", Project: "hive", Title: "Loose task", Feature: "routing", Status: "created"},
		},
		Agents: []records.Agent{
			{Key: "coordinator-g1", Name: "coordinator-g1", Role: "coordinator", PaneID: "w1:p1", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T14:20:00Z"},
			{Key: "manager-hive", Name: "manager-hive", Role: "manager", Project: "hive", PaneID: "w1:p2", Parent: "coordinator-g1", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T14:19:00Z"},
			{Key: "worker-alpha", Name: "worker-alpha", Role: "worker", Project: "hive", Task: "10-alpha", PaneID: "w1:p3", Parent: "manager-hive", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T14:18:00Z"},
			{Key: "worker-beta", Name: "worker-beta", Role: "worker", Project: "hive", Task: "20-beta", PaneID: "w1:p4", Parent: "manager-hive", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T14:17:00Z"},
			{Key: "worker-loose", Name: "worker-loose", Role: "worker", Project: "hive", Task: "30-loose", PaneID: "w1:p5", Parent: "manager-hive", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T14:16:00Z"},
		},
	}
}

// newAgentsTaskRoutingModel gives a demo model on the Agents screen with
// selection agent. The Work data is cold until the test loads it.
func newAgentsTaskRoutingModel(t *testing.T, agent string, popup string) *Model {
	t.Helper()
	fixture := &demoFixture{
		Terminal: demoSize{Width: 120, Height: 40},
		Clock:    demoClock{Time: "2026-10-23T14:22:00Z", Zone: "UTC"},
		Home:     "~/hive",
		Data:     agentsTaskRoutingData(),
		UI:       demoUIState{Screen: "agents", Selection: agent, Popup: popup},
		now:      workTestNow,
	}
	m := newDemoModel(t.TempDir(), &demoState{fixture: fixture})
	t.Cleanup(m.Stop)
	return m
}

// loadAgentsTaskRoutingWork puts the Work loader data in the model before the
// key, so the check covers the already-loaded Work state.
func loadAgentsTaskRoutingWork(t *testing.T, m *Model) {
	t.Helper()
	data, err := m.tasksLoader(m.home)
	if err != nil {
		t.Fatal(err)
	}
	m.tasks.accept(data)
}

// requireAgentTaskPopup checks the exact task of the open Work task popup.
func requireAgentTaskPopup(t *testing.T, m *Model, id, goal string) {
	t.Helper()
	if m.activeTab() != tabWork {
		t.Fatalf("active tab = %d, want Work", m.activeTab())
	}
	if m.tasks.selectedID != id {
		t.Fatalf("selected task = %q, want the literal ID %q", m.tasks.selectedID, id)
	}
	if m.tasks.goal != goal {
		t.Fatalf("goal = %q, want %q", m.tasks.goal, goal)
	}
	if m.tasks.popup != workPopupTask {
		t.Fatalf("popup = %q, want the Work task popup", m.tasks.popup)
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, id+" · esc closes") {
		t.Fatalf("Work view has no task popup titled %q:\n%s", id+" · esc closes", view)
	}
}

func TestAgentsTaskKeyOpensTheExactWorkTaskPopup(t *testing.T) {
	cases := []struct {
		name  string
		agent string
		task  string
		goal  string
	}{
		{"goal alpha", "worker-alpha", "10-alpha", "g1-alpha"},
		{"goal beta", "worker-beta", "20-beta", "g2-beta"},
		{"pseudo goal", "worker-loose", "30-loose", workNoGoal},
	}
	for _, test := range cases {
		for _, loaded := range []bool{false, true} {
			name := fmt.Sprintf("%s/work loaded %t", test.name, loaded)
			t.Run(name, func(t *testing.T) {
				m := newAgentsTaskRoutingModel(t, test.agent, "")
				if loaded {
					loadAgentsTaskRoutingWork(t, m)
					// A different task was open before the key.
					m.tasks.selectedID = "10-alpha"
					m.tasks.popup = workPopupTask
					m.selectScreen(tabAgents, 0)
					m.graph.selected = test.agent
				}
				m, cmd := raise(t, m, key('t', "t"))
				if m.activeTab() != tabWork {
					t.Fatalf("t left the Agents screen on tab %d", m.activeTab())
				}
				m = deliverAgentCommands(t, m, cmd)
				requireAgentTaskPopup(t, m, test.task, test.goal)
			})
		}
	}
}

func TestAgentsDetailsTaskKeyOpensTheExactWorkTaskPopup(t *testing.T) {
	for _, test := range []struct {
		name  string
		agent string
		task  string
		goal  string
	}{
		{"goal", "worker-beta", "20-beta", "g2-beta"},
		{"pseudo goal", "worker-loose", "30-loose", workNoGoal},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := newAgentsTaskRoutingModel(t, test.agent, "agent-details")
			if !m.popupOpen || m.focusState() != "details" {
				t.Fatalf("setup: popup open %v, focus %q; want the open details popup", m.popupOpen, m.focusState())
			}
			m, cmd := raise(t, m, key('t', "t"))
			if m.popupOpen {
				t.Fatal("t left the Agents details popup open")
			}
			m = deliverAgentCommands(t, m, cmd)
			requireAgentTaskPopup(t, m, test.task, test.goal)
		})
	}
}

// TestAgentsTaskKeyKeepsTheTaskAcrossAStaleLoad gives a Work load that does
// not hold the agent task before the fresh load. The stale load must not
// select another task or open its popup.
func TestAgentsTaskKeyKeepsTheTaskAcrossAStaleLoad(t *testing.T) {
	m := newAgentsTaskRoutingModel(t, "worker-beta", "")
	m, _ = raise(t, m, key('t', "t"))
	if m.tasks.openTask != "20-beta" {
		t.Fatalf("pending task = %q, want 20-beta before the load", m.tasks.openTask)
	}
	data := agentsTaskRoutingData()
	data.Tasks = data.Tasks[:1]
	stale, err := buildTasks(data.Tasks, data.Projects, data.Goals, data.Agents, taskFiles{
		evidence: func(taskItem) ([]string, error) { return nil, nil },
		read:     func(taskItem, string) (string, error) { return "", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	m, _ = raise(t, m, screenLoadMsg{request: tasksDataRequest, tasks: stale})
	if m.tasks.popup != "" {
		t.Fatalf("stale load opened popup %q, want no popup", m.tasks.popup)
	}
	cmd := m.startCurrentScreenLoad()
	m = deliverAgentCommands(t, m, cmd)
	requireAgentTaskPopup(t, m, "20-beta", "g2-beta")
}
