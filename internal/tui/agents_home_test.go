package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// generalAgent gives a live general agent with parent, started at started.
func generalAgent(key, parent, started string) records.Agent {
	return records.Agent{
		Key: key, Name: key, Role: "agent", Harness: "pi", Model: "gpt", PaneID: "w9:" + key, Parent: parent,
		Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-21T14:19:00Z",
		StartedAt: started, UpdatedAt: started,
	}
}

// homeSource gives one project with a manager and one worker, the
// coordinator, and the extra agents.
func homeSource(extra ...records.Agent) agentsSource {
	projects, agents := agentsProject("alpha", "pi", 1)
	agents = append([]records.Agent{{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Harness: "claude", Status: records.AgentLive, AgentStatus: "idle"}}, agents...)
	return agentsSource{projects: projects, agents: append(agents, extra...)}
}

// podOf gives the title of the pod that holds key and the level of its card.
func podOf(t *testing.T, m *Model, key string) (string, int) {
	t.Helper()
	for _, pod := range m.agentsPods() {
		if pod.manager != nil && pod.manager.record.Key == key {
			return pod.group.id, 0
		}
		for index, node := range pod.children {
			if node.record.Key == key && index < len(pod.levels) {
				return pod.group.id, pod.levels[index]
			}
		}
	}
	return "", -1
}

// cardColumn gives the column of the first cell of the name of key in the
// plain view.
func cardColumn(t *testing.T, view, name string) (int, int) {
	t.Helper()
	for row, line := range strings.Split(view, "\n") {
		if index := strings.Index(line, " "+name+" "); index >= 0 {
			return row, ansi.StringWidth(line[:index]) + 1
		}
	}
	t.Fatalf("the view omits %q:\n%s", name, view)
	return 0, 0
}

// AC1: a general agent with no parent card on the graph is in one last
// column with the title home. Home is not in the project count.
func TestAgentsHomeColumnHoldsAGeneralAgentWithoutAParentCard(t *testing.T) {
	m := agentsModel(t, homeSource(generalAgent("review-1", "someone-gone", "2026-10-21T14:00:00Z")))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if pod, _ := podOf(t, m, "review-1"); pod != "home" {
		t.Fatalf("review-1 is in pod %q; want home", pod)
	}
	pods := m.agentsPods()
	if last := pods[len(pods)-1]; last.group.id != "home" {
		t.Fatalf("the last column is %q; want home", last.group.id)
	}
	view := ansi.Strip(m.View().Content)
	requireView(t, m, "agents · 1 project", "─ home ─", "● review-1", "pi · working · 3m")
	_, alpha := cardColumn(t, view, "manager-alpha")
	_, home := cardColumn(t, view, "review-1")
	if home <= alpha {
		t.Fatalf("home column %d is not right of the project column %d:\n%s", home, alpha, view)
	}
}

// AC1: home is not shown when it is empty.
func TestAgentsHomeColumnIsNotShownWhenEmpty(t *testing.T) {
	m := agentsModel(t, homeSource(generalAgent("child-1", "manager-alpha", "2026-10-21T14:00:00Z")))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if strings.Contains(ansi.Strip(m.View().Content), " home ") {
		t.Fatal("the graph shows an empty home column")
	}
	if pod, _ := podOf(t, m, "child-1"); pod != "alpha" {
		t.Fatalf("child-1 is in pod %q; want alpha", pod)
	}
}

// AC1 and agents.md rule 13: with no project, the graph shows the home pod
// and the notice, as with each project pod hidden (goal g9).
func TestAgentsHomeAgentsWithNoProjectKeepTheNotice(t *testing.T) {
	m := agentsModel(t, agentsSource{agents: []records.Agent{generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z")}})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, agentsNoneText) || !strings.Contains(view, "━ home ━") || !strings.Contains(view, "● review-1") {
		t.Fatalf("with no project the graph does not show the home pod and the notice:\n%s", view)
	}
	m.health.agents = m.agents.data
	m.selectScreen(tabHealth, 1)
	m.invalidateView()
	requireView(t, m, "review-1")
}

// AC2: a general agent is under the card of its parent, also a general agent
// without a project. In the one-pod view a child of a child is one more tree
// level below its parent: in the same column, with the joint ┬ on the
// bottom of the parent card and one │ row (decision 74).
func TestAgentsNestGeneralAgentsUnderTheirParentCard(t *testing.T) {
	m := agentsModel(t, homeSource(
		generalAgent("helper-1", "t1-alpha-worker", "2026-10-21T14:00:00Z"),
		generalAgent("helper-2", "helper-1", "2026-10-21T14:01:00Z"),
	))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	for key, want := range map[string]int{"t1-alpha-worker": 1, "helper-1": 2, "helper-2": 3} {
		if pod, level := podOf(t, m, key); pod != "alpha" || level != want {
			t.Fatalf("%s is in pod %q at level %d; want alpha at level %d", key, pod, level, want)
		}
	}
	view := ansi.Strip(m.View().Content)
	workerRow, worker := cardColumn(t, view, "t1-alpha-worker")
	helperRow, helper := cardColumn(t, view, "helper-1")
	nestedRow, nested := cardColumn(t, view, "helper-2")
	if helper != worker || nested != helper || helperRow-workerRow != 5 || nestedRow-helperRow != 5 {
		t.Fatalf("cards at %d,%d %d,%d %d,%d; want each child 5 rows below its parent in its column:\n%s",
			workerRow, worker, helperRow, helper, nestedRow, nested, view)
	}
	_, manager := cardColumn(t, view, "manager-alpha")
	if worker != manager {
		t.Fatalf("the worker is not below the manager in the tree:\n%s", view)
	}
	lines := strings.Split(view, "\n")
	// The line to helper-1 leaves the bottom border of the worker card at
	// its center, and goes down one │ row.
	if got := []rune(lines[workerRow+2])[worker+8]; got != '┬' {
		t.Fatalf("the worker card bottom has %q above the line; want ┬:\n%s", got, view)
	}
	if got := []rune(lines[workerRow+3])[worker+8]; got != '│' {
		t.Fatalf("the row below the worker card has %q; want │:\n%s", got, view)
	}
	requireWholeAgentsBorders(t, strings.NewReplacer("─┬─", "───", "─┴─", "───").Replace(view))
}

// AC2: in the overview, a pod with nesting uses the rail form, also when its
// children fit side by side in the tree form. The one-pod view always uses
// the tree (decision 74).
func TestAgentsNestingUsesTheRailFormInTheOverview(t *testing.T) {
	source := homeSource(generalAgent("helper-1", "t1-alpha-worker", "2026-10-21T14:00:00Z"))
	projects, agents := agentsProject("beta", "pi", 1)
	source.projects, source.agents = append(source.projects, projects...), append(source.agents, agents...)
	m := agentsModel(t, source)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	view := ansi.Strip(m.View().Content)
	_, manager := cardColumn(t, view, "manager-alpha")
	_, worker := cardColumn(t, view, "t1-alpha-worker")
	if worker-manager != 1 {
		t.Fatalf("the worker is at column %d, the manager at %d; want the rail form:\n%s", worker, manager, view)
	}
}

// AC2: the children of a card are in order of creation time, for each role
// together.
func TestAgentsChildrenAreInOrderOfCreationTime(t *testing.T) {
	source := homeSource(
		generalAgent("late-helper", "manager-alpha", "2026-10-21T14:10:00Z"),
		generalAgent("early-helper", "manager-alpha", "2026-10-21T13:50:00Z"),
	)
	source.agents[2].StartedAt = "2026-10-21T14:00:00Z" // t1-alpha-worker
	m := agentsModel(t, source)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	var order []string
	for _, pod := range m.agentsPods() {
		if pod.group.id == "alpha" {
			for _, node := range pod.children {
				order = append(order, node.record.Key)
			}
		}
	}
	if got := strings.Join(order, " "); got != "early-helper t1-alpha-worker late-helper" {
		t.Fatalf("children %q; want early-helper t1-alpha-worker late-helper", got)
	}
}

// AC2: when h hides a stopped parent, its live child moves to home. It moves
// back when h shows the parent.
func TestAgentsHMovesTheChildOfAHiddenParent(t *testing.T) {
	source := homeSource(generalAgent("helper-1", "t1-alpha-worker", "2026-10-21T14:00:00Z"))
	source.agents[2].Status = records.AgentStopped
	m := agentsModel(t, source)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if pod, _ := podOf(t, m, "helper-1"); pod != "alpha" {
		t.Fatalf("helper-1 is in pod %q; want alpha", pod)
	}
	m, _ = raise(t, m, key('h', "h"))
	if pod, _ := podOf(t, m, "helper-1"); pod != "home" {
		t.Fatalf("after h helper-1 is in pod %q; want home", pod)
	}
	requireView(t, m, "─ home ─", "● helper-1", "h shows 1 stopped")
	m, _ = raise(t, m, key('h', "h"))
	if pod, _ := podOf(t, m, "helper-1"); pod != "alpha" {
		t.Fatalf("after h h helper-1 is in pod %q; want alpha", pod)
	}
}

// AC2: a child of the coordinator is in home, and two general agents with
// the same parent are both on the graph.
func TestAgentsCoordinatorChildrenAreInHome(t *testing.T) {
	m := agentsModel(t, homeSource(
		generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z"),
		generalAgent("review-2", "coordinator-g1", "2026-10-21T14:01:00Z"),
	))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	for _, key := range []string{"review-1", "review-2"} {
		if pod, level := podOf(t, m, key); pod != "home" || level != 0 {
			t.Fatalf("%s is in pod %q at level %d; want home at level 0", key, pod, level)
		}
	}
}

// AC2: a manager with no child shows "no child agents".
func TestAgentsManagerWithoutChildrenShowsNoChildAgents(t *testing.T) {
	projects, agents := agentsProject("alpha", "pi", 0)
	m := agentsModel(t, agentsSource{projects: projects, agents: agents})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	requireView(t, m, "no child agents")
}

// AC1 and AC2: the arrows reach the nested child and the home column.
// Decision 62 replaces → from a card to the next pod: → selects the home pod
// in the overview, and enter opens it.
func TestAgentsArrowsReachNestedAndHomeAgents(t *testing.T) {
	m := agentsModel(t, homeSource(
		generalAgent("helper-1", "t1-alpha-worker", "2026-10-21T14:00:00Z"),
		generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z"),
	))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.graph.selected = "t1-alpha-worker"
	m, _ = raise(t, m, key('o', "o"))
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	if m.graph.selected != "helper-1" {
		t.Fatalf("↓ from the worker selects %q; want helper-1", m.graph.selected)
	}
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	if m.graph.selected != "helper-1" {
		t.Fatalf("→ in the one-pod view selects %q; want helper-1", m.graph.selected)
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	m, _ = raise(t, m, key('o', "o"))
	if m.graph.selected != "review-1" {
		t.Fatalf("→ and o from the alpha pod select %q; want review-1", m.graph.selected)
	}
}

// AC1: from 80 to 119 columns, home is the last page of ←→.
func TestAgentsCompactHomeIsTheLastPage(t *testing.T) {
	source := homeSource(generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z"))
	for _, id := range []string{"beta", "gamma"} {
		projects, agents := agentsProject(id, "pi", 0)
		source.projects = append(source.projects, projects...)
		source.agents = append(source.agents, agents...)
	}
	m := agentsModel(t, source)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m.graph.selected = "manager-alpha"
	requireView(t, m, "agents · 3 projects", "projects 1–2 of 3")
	for range 3 {
		m, _ = raise(t, m, key(tea.KeyRight, ""))
	}
	if m.graph.pod != "home" {
		t.Fatalf("3 × → selects the pod %q; want home", m.graph.pod)
	}
	requireView(t, m, "● review-1", "━ home ━", "projects 3–3 of 3 · home", "← beta")
}

// AC3: the details popup of a general agent has the fields of another agent
// that have a value.
func TestAgentsDetailsOfAGeneralAgent(t *testing.T) {
	m := agentsModel(t, homeSource(generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z")))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.graph.selected = "review-1"
	m, _ = raise(t, m, key('o', "o"))
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	// agents.md rule 14: the model is on the harness row, and the parent
	// row shows the name of the parent record coordinator-g1.
	requireView(t, m, "review-1 · esc closes", "role", "agent", "harness", "pi · gpt", "parent    coordinator")
	_, lines, _ := m.agentDetailsPopup()
	for _, line := range lines {
		fields := strings.Fields(ansi.Strip(line))
		if len(fields) == 1 && (fields[0] == "project" || fields[0] == "nudged" || fields[0] == "goal" || fields[0] == "replaces") {
			t.Fatalf("the popup shows the empty field %q:\n%s", fields[0], strings.Join(lines, "\n"))
		}
	}
}

// AC3: Health › Agents lists a general agent with the role agent, and r on it
// shows the notice and does nothing.
func TestHealthAgentsRestartRefusesAGeneralAgent(t *testing.T) {
	m := newTestDemoModel(t, "HealthAgents.fixture.json", demoSize{})
	m.health.agents = buildAgents(homeSource(generalAgent("review-1", "coordinator-g1", "2026-10-21T14:00:00Z")), m.now())
	m.health.agents.supervisorLive = true
	m.health.selectedAgent = "review-1"
	// The board opens with the restart prompt of the supervisor. Close it.
	m.agents.restartConfirm, m.agents.restartNotice = false, ""
	m.invalidateView()
	requireView(t, m, "review-1")
	m, _ = raise(t, m, key('r', "r"))
	if m.agents.restartConfirm || m.agents.restartRunning {
		t.Fatal("r started a restart of a general agent")
	}
	requireView(t, m, healthRestartOnly)
}

// AC4: each new board fixture loads at its size with the demo loader, each
// time in it is at or before its clock, and its view shows the rule of the
// board.
func TestGeneralAgentBoardFixtures(t *testing.T) {
	for _, tc := range []struct {
		board         string
		width, height int
		wants         []string
		home          map[string]bool // key: in home
	}{
		{"AgentsHome", 120, 40, []string{"agents · 2 projects", "━ home ━", "plan-review-1", "spec-check-1"}, map[string]bool{"plan-review-1": true, "spec-check-1": true}},
		{"AgentsNested", 120, 40, []string{"t5-research-1", "t5-research-2"}, map[string]bool{"t5-research-1": false, "t5-research-2": false}},
		{"AgentsHiddenParent", 120, 40, []string{"━ home ━", "t4-review-1", "h shows 1 stopped"}, map[string]bool{"t4-review-1": true}},
		{"AgentsCoordinatorChild", 120, 40, []string{"agents · home", "plan-review-2"}, map[string]bool{"plan-review-1": true, "plan-review-2": true}},
		{"CompactAgentsHome", 80, 24, []string{"━ home ━", "plan-review-1", "projects 2–2 of 2 · home", "← hive"}, map[string]bool{"plan-review-1": true}},
		// agents.md rule 14 (ask a29): the parent row shows the name of the
		// parent, and the key row is in the place of RECENT.
		{"AgentDetailsGeneral", 120, 40, []string{"plan-review-1 · esc closes", "agent", "parent    coordinator", "a attach · p peek · esc closes"}, nil},
		{"HealthAgentsGeneral", 120, 40, []string{"▶ ● plan-review-1", "agent"}, nil},
	} {
		t.Run(tc.board, func(t *testing.T) {
			path := filepath.Join(designDir, tc.board+".fixture.json")
			requireSpecFile(t, path)
			fixture, err := loadDemoFixture(path)
			if err != nil {
				t.Fatal(err)
			}
			if fixture.Board != tc.board+".dc.html" || fixture.Terminal.Width != tc.width || fixture.Terminal.Height != tc.height {
				t.Fatalf("board %q at %dx%d; want %s.dc.html at %dx%d", fixture.Board, fixture.Terminal.Width, fixture.Terminal.Height, tc.board, tc.width, tc.height)
			}
			requireTimesAtOrBefore(t, path, fixture.now)
			m := newTestDemoModel(t, tc.board+".fixture.json", demoSize{})
			requireView(t, m, tc.wants...)
			for key, home := range tc.home {
				if pod, _ := podOf(t, m, key); (pod == "home") != home {
					t.Errorf("%s is in pod %q; want home %t", key, pod, home)
				}
			}
		})
	}
}

// requireTimesAtOrBefore fails when a time in the JSON file at path is after
// clock.
func requireTimesAtOrBefore(t *testing.T, path string, clock time.Time) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			for _, item := range v {
				walk(item)
			}
		case []any:
			for _, item := range v {
				walk(item)
			}
		case string:
			if at, err := time.Parse(time.RFC3339Nano, v); err == nil && at.After(clock) {
				t.Errorf("time %s is after the clock %s", v, clock.Format(time.RFC3339))
			}
		}
	}
	walk(value)
}

// AC5, decision 42: the demo view of each general agent board matches its
// golden grid cell for cell.
func TestGeneralAgentBoardsMatchTheirGrids(t *testing.T) {
	for _, board := range []string{"AgentsHome", "AgentsNested", "AgentsHiddenParent", "AgentsCoordinatorChild", "CompactAgentsHome", "AgentDetailsGeneral", "HealthAgentsGeneral"} {
		t.Run(board, func(t *testing.T) { checkBoardGolden(t, board) })
	}
}
