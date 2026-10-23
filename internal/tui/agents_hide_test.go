package tui

import (
	"os"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Goal g9 (decisions 105 to 108): a stopped, exited, or stale manager hides
// with the other stopped cards. A pod with no visible card disappears. A
// hidden manager with a live child keeps the pod, and the child hangs from
// the pod top border as in the home pod.

// managerStates are the states of a manager that h hides.
var managerStates = []string{records.AgentStopped, records.AgentExited, records.AgentStale}

// hideSizes are the two widths of the checks: from 120 columns and below.
var hideSizes = []tea.WindowSizeMsg{{Width: 120, Height: 40}, {Width: 80, Height: 24}}

// hideHome gives a home with the projects alpha and beta, a live
// coordinator, a live supervisor claim, and the agent records. alpha has a
// live manager and a live worker.
func hideHome(t *testing.T, agents ...records.Agent) (string, *testClock) {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", PID: int64(os.Getpid()), PIDStart: pidStart})
	for _, id := range []string{"alpha", "beta"} {
		writeRecord(t, home, []string{"projects"}, id, records.Project{ID: id, Path: "/repo/" + id, Base: "dev"})
	}
	base := []records.Agent{
		{Key: "coordinator", Name: "coordinator", Role: "coordinator", Harness: "claude", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-23T11:58:00Z", StartedAt: "2026-10-23T10:00:00Z"},
		{Key: "manager-alpha", Name: "manager-alpha", Role: "manager", Project: "alpha", Harness: "claude", Status: records.AgentLive, AgentStatus: "idle", Parent: "coordinator", StatusChangedAt: "2026-10-23T11:58:00Z", StartedAt: "2026-10-23T10:01:00Z"},
		{Key: "a1-worker", Name: "a1-worker", Role: "worker", Project: "alpha", Task: "1-a", Harness: "pi", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-alpha", StatusChangedAt: "2026-10-23T11:58:00Z", StartedAt: "2026-10-23T10:02:00Z"},
	}
	for _, record := range append(base, agents...) {
		writeAgentRecord(t, home, record)
	}
	return home, clock
}

// betaManager gives the beta manager record in state.
func betaManager(state string) records.Agent {
	return records.Agent{Key: "manager-beta", Name: "manager-beta", Role: "manager", Project: "beta", Harness: "codex", Status: state, AgentStatus: "done", Parent: "coordinator", StatusChangedAt: "2026-10-23T11:00:00Z", StartedAt: "2026-10-23T10:03:00Z"}
}

// betaWorker gives a beta worker record with the parent manager-beta.
func betaWorker(name, state, agentStatus, started string) records.Agent {
	return records.Agent{Key: name, Name: name, Role: "worker", Project: "beta", Task: name, Harness: "pi", Status: state, AgentStatus: agentStatus, Parent: "manager-beta", StatusChangedAt: "2026-10-23T11:50:00Z", StartedAt: started}
}

// openHide opens Agents on the home at size.
func openHide(t *testing.T, home string, clock *testClock, size tea.WindowSizeMsg) *Model {
	t.Helper()
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, size)
	m, cmd := raise(t, m, key('1', "1"))
	return deliverAgentCommands(t, m, cmd)
}

// A manager in each hidden state with no live child: at start the pod of
// beta is gone and the title hint counts the 2 hidden cards. h shows the
// pod with the manager card, and its note counts the manager. h hides the
// pod again.
func TestAgentsHHidesAManagerWithNoLiveChild(t *testing.T) {
	for _, state := range managerStates {
		for _, size := range hideSizes {
			home, clock := hideHome(t, betaManager(state), betaWorker("b2-worker", records.AgentStopped, "done", "2026-10-23T10:05:00Z"))
			m := openHide(t, home, clock, size)
			// Only alpha is left, so the one-pod view opens (decision 64).
			requireView(t, m, "project agents · alpha · h shows 2 stopped", "● a1-worker")
			rejectView(t, m, "manager-beta", "b2-worker", "beta")
			m, _ = raise(t, m, key('h', "h"))
			requireView(t, m, "project agents · 2 projects ─", "─ beta ─", "○ manager-beta")
			rejectView(t, m, "h shows")
			if size.Width >= 120 {
				requireView(t, m, "h hides 2 stopped")
			}
			m, _ = raise(t, m, key('h', "h"))
			requireView(t, m, "project agents · alpha · h shows 2 stopped")
			rejectView(t, m, "manager-beta", "─ beta ─")
		}
	}
}

// A manager in each hidden state with a live child: the pod of beta stays
// with its border and title, the manager card hides, and the note counts
// the manager and the stopped worker. The overview draws the live child as
// the home pod draws its cards, with no "no manager" text and no rail joint.
func TestAgentsHiddenManagerKeepsThePodOfALiveChild(t *testing.T) {
	for _, state := range managerStates {
		for _, size := range hideSizes {
			home, clock := hideHome(t, betaManager(state),
				betaWorker("b1-worker", records.AgentLive, "working", "2026-10-23T10:04:00Z"),
				betaWorker("b2-worker", records.AgentStopped, "done", "2026-10-23T10:05:00Z"))
			m := openHide(t, home, clock, size)
			requireView(t, m, "project agents · 2 projects · h shows 2 stopped", "─ beta ─", "● b1-worker")
			rejectView(t, m, "manager-beta", "b2-worker", "no manager")
			if size.Width >= 120 {
				requireView(t, m, "h shows 2 stopped  ")
			}
			grid := renderGrid(t, m, size.Width, size.Height)
			row, col := findText(t, grid, "● b1-worker")
			// The card is on content column 0 of the pod, as a home card:
			// its left border has no rail joint, and no rail is left of it.
			if got := grid.At(row, col-2).Rune; got != '│' {
				t.Errorf("%s %dx%d: left border of b1-worker = %q; want │", state, size.Width, size.Height, got)
			}
			if got := grid.At(row, col-3).Rune; got != ' ' {
				t.Errorf("%s %dx%d: cell left of b1-worker = %q; want a space", state, size.Width, size.Height, got)
			}
			m, _ = raise(t, m, key('h', "h"))
			requireView(t, m, "○ manager-beta", "● b1-worker")
			rejectView(t, m, "h shows")
			if size.Width >= 120 {
				requireView(t, m, "○ b2-worker", "h hides 2 stopped")
			}
		}
	}
}

// In the one-pod view of a pod with a hidden manager, the coordinator line
// goes down through the pod top border (┼) to the top joint (┴) of each top
// child, as rule A draws home. A nested child keeps its parent.
func TestAgentsPodViewHangsTheChildrenFromTheTopBorder(t *testing.T) {
	for _, state := range managerStates {
		for _, size := range hideSizes {
			nested := records.Agent{Key: "b1-helper", Name: "b1-helper", Role: "agent", Harness: "claude", Status: records.AgentLive, AgentStatus: "idle", Parent: "b1-worker", StatusChangedAt: "2026-10-23T11:50:00Z", StartedAt: "2026-10-23T10:06:00Z"}
			home, clock := hideHome(t, betaManager(state),
				betaWorker("b1-worker", records.AgentLive, "working", "2026-10-23T10:04:00Z"), nested)
			m := openHide(t, home, clock, size)
			m, _ = raise(t, m, key(tea.KeyRight, ""))
			m, _ = raise(t, m, key(tea.KeyEnter, ""))
			requireView(t, m, "project agents · beta · h shows 1 stopped", "b1-worker", "b1-helper")
			rejectView(t, m, "manager-beta", "no manager")
			grid := renderGrid(t, m, size.Width, size.Height)
			_, coordCol := findText(t, grid, "● coordinator")
			joint := coordCol - 2 + 12
			podRow, _ := findText(t, grid, "─ beta ─")
			workerRow, workerCol := findText(t, grid, "b1-worker")
			if got := grid.At(podRow, joint).Rune; got != '┼' {
				t.Errorf("%s %dx%d: pod top border at the joint = %q; want ┼", state, size.Width, size.Height, got)
			}
			if got := grid.At(workerRow-1, workerCol-4+12).Rune; got != '┷' {
				t.Errorf("%s %dx%d: top joint of the selected b1-worker = %q; want ┷", state, size.Width, size.Height, got)
			}
			for row := podRow + 1; row < workerRow-1; row++ {
				if got := grid.At(row, joint).Rune; got != '│' {
					t.Errorf("%s %dx%d: line row %d = %q; want │", state, size.Width, size.Height, row, got)
				}
			}
			helperRow, helperCol := findText(t, grid, "b1-helper")
			if got := grid.At(helperRow-1, helperCol-4+12).Rune; got != '┴' {
				t.Errorf("%s %dx%d: top joint of b1-helper = %q; want ┴", state, size.Width, size.Height, got)
			}
			if got := grid.At(workerRow+2, workerCol-4+12).Rune; got != '┯' {
				t.Errorf("%s %dx%d: bottom joint of b1-worker = %q; want ┯", state, size.Width, size.Height, got)
			}
			if t.Failed() {
				t.Logf("view:\n%s", m.View().Content)
				return
			}
		}
	}
}

// When every project pod disappears, the hint stays on the title with the
// count of each hidden card, at both widths. h shows the pod again.
func TestAgentsHintStaysWhenEveryProjectPodHides(t *testing.T) {
	for _, size := range hideSizes {
		home, clock := hideHome(t, betaManager(records.AgentStopped), betaWorker("b2-worker", records.AgentStopped, "done", "2026-10-23T10:05:00Z"))
		// alpha also stops: the manager is exited and the worker is stale.
		writeAgentRecord(t, home, records.Agent{Key: "manager-alpha", Name: "manager-alpha", Role: "manager", Project: "alpha", Harness: "claude", Status: records.AgentExited, Parent: "coordinator", StartedAt: "2026-10-23T10:01:00Z"})
		writeAgentRecord(t, home, records.Agent{Key: "a1-worker", Name: "a1-worker", Role: "worker", Project: "alpha", Task: "1-a", Harness: "pi", Status: records.AgentStale, AgentStatus: "working", Parent: "manager-alpha", StartedAt: "2026-10-23T10:02:00Z"})
		m := openHide(t, home, clock, size)
		requireView(t, m, "project agents · 0 projects · h shows 4 stopped", agentsNoneText)
		rejectView(t, m, "alpha", "beta")
		m, _ = raise(t, m, key('h', "h"))
		requireView(t, m, "project agents · 2 projects ─", "○ manager-alpha", "○ manager-beta")
		rejectView(t, m, agentsNoneText)
	}
}

// h with nothing to hide adds no hint.
func TestAgentsNoHintWithNothingHidden(t *testing.T) {
	home, clock := hideHome(t, betaManager(records.AgentLive))
	m := openHide(t, home, clock, hideSizes[0])
	requireView(t, m, "project agents · 2 projects ─", "manager-beta")
	rejectView(t, m, "h shows", "h hides")
}

// The selected pod disappears with h, and the selection moves to a pod on
// the screen. The second h shows the pod again with the selection on it.
func TestAgentsSelectionFollowsThePodsOfH(t *testing.T) {
	for _, size := range hideSizes {
		home, clock := hideHome(t, betaManager(records.AgentStopped))
		writeAgentRecord(t, home, records.Agent{Key: "helper", Name: "helper", Role: "agent", Harness: "pi", Status: records.AgentLive, AgentStatus: "idle", Parent: "coordinator", StartedAt: "2026-10-23T10:07:00Z"})
		m := openHide(t, home, clock, size)
		requireView(t, m, "project agents · 1 project · h shows 1 stopped", "━ alpha ━", "─ home ─")
		m, _ = raise(t, m, key('h', "h"))
		m, _ = raise(t, m, key(tea.KeyRight, ""))
		requireView(t, m, "━ beta ━", "─ alpha ─")
		m, _ = raise(t, m, key('h', "h"))
		requireView(t, m, "━ alpha ━", "─ home ─")
		rejectView(t, m, "beta")
		m, _ = raise(t, m, key('h', "h"))
		requireView(t, m, "━ beta ━", "─ alpha ─")
		// In the one-pod view the selected stopped card hides, and the
		// first visible card of the pod is selected.
		home, clock = hideHome(t, betaManager(records.AgentStopped), betaWorker("b1-worker", records.AgentLive, "working", "2026-10-23T10:04:00Z"))
		m = openHide(t, home, clock, size)
		m, _ = raise(t, m, key('h', "h"))
		m, _ = raise(t, m, key(tea.KeyRight, ""))
		m, _ = raise(t, m, key(tea.KeyEnter, ""))
		if m.graph.selected != "manager-beta" {
			t.Fatalf("%dx%d: selection %q in the pod; want manager-beta", size.Width, size.Height, m.graph.selected)
		}
		m, _ = raise(t, m, key('h', "h"))
		requireView(t, m, "project agents · beta · h shows 1 stopped", "┃ ● b1-worker")
		if node, _, _ := m.selectedAgent(); node.record.Key != "b1-worker" {
			t.Fatalf("%dx%d: selection %q after h; want b1-worker", size.Width, size.Height, node.record.Key)
		}
	}
}

// At 80 columns two pods fit. The last pod is selected, and h hides the pod
// before it: the first pod column stays valid and the selected pod stays on
// the screen.
func TestAgentsScrollStaysValidWhenPodsHide(t *testing.T) {
	home, clock := hideHome(t, betaManager(records.AgentStopped))
	writeRecord(t, home, []string{"projects"}, "gamma", records.Project{ID: "gamma", Path: "/repo/gamma", Base: "dev"})
	writeAgentRecord(t, home, records.Agent{Key: "manager-gamma", Name: "manager-gamma", Role: "manager", Project: "gamma", Harness: "pi", Status: records.AgentLive, AgentStatus: "working", Parent: "coordinator", StartedAt: "2026-10-23T10:08:00Z"})
	m := openHide(t, home, clock, hideSizes[1])
	m, _ = raise(t, m, key('h', "h"))
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	requireView(t, m, "━ gamma ━", "projects 2–3 of 3")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "━ gamma ━", "─ alpha ─", "project agents · 2 projects · h shows 1 stopped")
	rejectView(t, m, "beta", "at the edge scrolls")
	if m.graph.left != 0 {
		t.Fatalf("first pod column = %d; want 0", m.graph.left)
	}
}
