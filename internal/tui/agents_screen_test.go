package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// agentsBoards are the boards of the Agents screen (AC2). The Help board
// has its own check, TestHelpMatchesItsBoard.
var agentsBoards = []string{"Main", "AgentsMany", "AgentDetails", "CompactAgents"}

// Decision 42: the demo view of each Agents board matches its golden grid
// cell for cell.
func TestAgentsBoardsMatchTheirGrids(t *testing.T) {
	for _, board := range agentsBoards {
		t.Run(board, func(t *testing.T) { checkBoardGolden(t, board) })
	}
}

// agentsProject gives one project with a manager and workers named from id.
func agentsProject(id, harness string, workers int) ([]records.Project, []records.Agent) {
	agents := []records.Agent{{
		Key: "manager-" + id, Name: "manager-" + id, Role: "manager", Project: id, Harness: harness,
		PaneID: "w1:" + id, Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-10-21T14:20:00Z",
	}}
	for index := range workers {
		name := fmt.Sprintf("t%d-%s-worker", index+1, id)
		agents = append(agents, records.Agent{
			Key: name, Name: name, Role: "worker", Project: id, Task: name, Harness: harness, Parent: "manager-" + id,
			PaneID: "w2:" + name, Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-21T14:19:00Z",
		})
	}
	return []records.Project{{ID: id, Path: "/repo/" + id, Base: "dev"}}, agents
}

// agentsModel gives the Main demo model with the agents of source.
func agentsModel(t *testing.T, source agentsSource) *Model {
	t.Helper()
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	source.supervisorLive = true
	m.agents.accept(buildAgents(source, m.now()))
	m.invalidateView()
	return m
}

// agentsDataSets are the data of the responsive checks of decision 46.
func agentsDataSets(t *testing.T) map[string]func(t *testing.T) (*Model, string) {
	return map[string]func(t *testing.T) (*Model, string){
		"Main": func(t *testing.T) (*Model, string) {
			return newTestDemoModel(t, "Main.fixture.json", demoSize{}), "manager-hive"
		},
		"AgentsMany": func(t *testing.T) (*Model, string) {
			return newTestDemoModel(t, "AgentsMany.fixture.json", demoSize{}), "t5-worker-1"
		},
		"nine projects": func(t *testing.T) (*Model, string) {
			var source agentsSource
			for index := range 9 {
				projects, agents := agentsProject(fmt.Sprintf("p%d", index+1), "pi", index%4)
				source.projects = append(source.projects, projects...)
				source.agents = append(source.agents, agents...)
			}
			m := agentsModel(t, source)
			m.graph.selected = "t2-p8-worker"
			return m, "t2-p8-worker"
		},
		"long text": func(t *testing.T) (*Model, string) {
			long := strings.Repeat("very-long-project-name-", 4)
			projects, agents := agentsProject(long, strings.Repeat("harness", 5), 3)
			goal := records.Goal{ID: "g1-" + long, Project: long, Status: "active", Waves: []records.GoalWave{{ID: "A", Status: "open"}}}
			m := agentsModel(t, agentsSource{projects: projects, agents: agents, goals: []records.Goal{goal}})
			m.graph.selected = agents[3].Key
			return m, "t3-very-long-proj…"
		},
		"forty workers": func(t *testing.T) (*Model, string) {
			projects, agents := agentsProject("p1", "pi", 40)
			m := agentsModel(t, agentsSource{projects: projects, agents: agents})
			m.graph.selected = "t40-p1-worker"
			return m, "t40-p1-worker"
		},
		"empty": func(t *testing.T) (*Model, string) {
			return agentsModel(t, agentsSource{}), agentsNoneText
		},
	}
}

// Decision 46: at 80 and 119 columns, at the minimum height and one row
// more, and at 120 columns, no line is wider than the pane, no border
// breaks, and the selected agent is on the screen. At 79 columns the screen
// shows the too-small notice.
func TestAgentsFitEachSize(t *testing.T) {
	sizes := []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}}
	for name, build := range agentsDataSets(t) {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%s/%dx%d", name, size.Width, size.Height), func(t *testing.T) {
				m, selected := build(t)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				view := ansi.Strip(m.View().Content)
				if size.Width < minimumWidth {
					if !strings.Contains(view, "terminal too small") {
						t.Fatalf("view at %d columns is not the too-small notice:\n%s", size.Width, view)
					}
					return
				}
				lines := strings.Split(view, "\n")
				if len(lines) != size.Height {
					t.Fatalf("view has %d lines, want %d:\n%s", len(lines), size.Height, view)
				}
				for index, line := range lines {
					if got := ansi.StringWidth(line); got != size.Width {
						t.Fatalf("line %d has width %d, want %d: %q", index+1, got, size.Width, line)
					}
				}
				requireWholeAgentsBorders(t, view)
				if !strings.Contains(view, selected) {
					t.Fatalf("view does not show %q:\n%s", selected, view)
				}
			})
		}
	}
}

// Decisions 10 and 14: the graph shows the project agents only, and the
// selected card has a heavy accent border (user decision 2026-10-21,
// option A).
func TestAgentsGraphShowsProjectAgentsOnly(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	view := ansi.Strip(m.View().Content)
	for _, name := range []string{"coordinator ●", "supervisor ●"} {
		if strings.Count(view, strings.Fields(name)[0]) != 1 {
			t.Fatalf("the graph shows %q outside the header:\n%s", name, view)
		}
	}
	for _, name := range []string{"◐ manager-hive", "○ t4-worker-1", "○ t4-verifier-1"} {
		if !strings.Contains(view, name) {
			t.Fatalf("the graph omits %q:\n%s", name, view)
		}
	}
	grid := renderGrid(t, m, 120, 40)
	row, col := findText(t, grid, "manager-hive")
	if got := grid.At(row, col-4); got.Rune != '┃' || got.Token != cellgrid.Accent {
		t.Fatalf("the selected card border is %+v, want an accent ┃", got)
	}
	if got := grid.At(row, col-2); got.Rune != '◐' || got.Token != cellgrid.Accent {
		t.Fatalf("the idle dot is %+v, want an accent ◐", got)
	}
}

// findText gives the row and the column of the first cell of text in g.
func findText(t *testing.T, g cellgrid.Grid, text string) (int, int) {
	t.Helper()
	for row := range g.Height {
		var line []rune
		for col := range g.Width {
			line = append(line, g.At(row, col).Rune)
		}
		if index := strings.Index(string(line), text); index >= 0 {
			return row, len([]rune(string(line)[:index]))
		}
	}
	t.Fatalf("grid omits %q", text)
	return 0, 0
}

// Decision 10: → at the edge of the screen scrolls the pods and keeps the
// selected card on the screen. ← scrolls back.
func TestAgentsArrowsScrollThePods(t *testing.T) {
	m := newTestDemoModel(t, "AgentsMany.fixture.json", demoSize{})
	requireView(t, m, "projects 1–4 of 5", "example-ml →")
	for range 4 {
		m, _ = raise(t, m, key(tea.KeyRight, ""))
	}
	if m.graph.selected != "manager-ml-g1" {
		t.Fatalf("selection %q after 4 × →; want manager-ml-g1", m.graph.selected)
	}
	requireView(t, m, "projects 2–5 of 5", "← hive", "◐ manager-ml")
	for range 4 {
		m, _ = raise(t, m, key(tea.KeyLeft, ""))
	}
	requireView(t, m, "projects 1–4 of 5")
	// Each step keeps the index in the pod, cut to the pod.
	if m.graph.selected != "manager-hive-g2" {
		t.Fatalf("selection %q after 4 × ←; want manager-hive-g2", m.graph.selected)
	}
}

// A light connector meets the heavy border of the selected card with a
// mixed-weight joint (manager decision): ┯ below a tree manager, ┷ above a
// tree child, ┡ on the corner of a rail manager, and ┨ beside a rail child.
func TestAgentsSelectedCardJoints(t *testing.T) {
	for _, tc := range []struct {
		fixture, name string
		keys          []tea.KeyPressMsg
		row, col      int // the joint cell relative to the first cell of the name
		want          rune
	}{
		{"Main.fixture.json", "manager-hive", nil, 2, 8, '┯'},
		{"Main.fixture.json", "t4-worker-1", []tea.KeyPressMsg{key(tea.KeyDown, "")}, -1, 8, '┷'},
		{"AgentsMany.fixture.json", "t5-worker-1", nil, 0, -4, '┨'},
		{"AgentsMany.fixture.json", "manager-hive", []tea.KeyPressMsg{key('k', "k"), key('k', "k")}, 2, -4, '┡'},
	} {
		m := newTestDemoModel(t, tc.fixture, demoSize{})
		for _, k := range tc.keys {
			m, _ = raise(t, m, k)
		}
		grid := renderGrid(t, m, 120, 40)
		row, col := findText(t, grid, tc.name)
		if got := grid.At(row+tc.row, col+tc.col); got.Rune != tc.want || got.Token != cellgrid.Accent {
			t.Errorf("%s %s: joint is %+v, want an accent %q", tc.fixture, tc.name, got, tc.want)
		}
	}
}

// ↑ and ↓ move between the manager and the children of a pod. The
// children are in order of creation time (user decision 2 of
// general-agent): t4-verifier-1, t5-worker-1, t6-worker-1.
func TestAgentsUpDownMoveInThePod(t *testing.T) {
	m := newTestDemoModel(t, "AgentsMany.fixture.json", demoSize{})
	for _, step := range []struct {
		key  tea.KeyPressMsg
		want string
	}{
		{key('j', "j"), "t6-worker-1"},
		{key(tea.KeyDown, ""), "t6-worker-1"},
		{key('k', "k"), "t5-worker-1"},
		{key(tea.KeyUp, ""), "t4-verifier-1"},
		{key(tea.KeyDown, ""), "t5-worker-1"},
		{key(tea.KeyUp, ""), "t4-verifier-1"},
		{key(tea.KeyUp, ""), "manager-hive-g2"},
	} {
		m, _ = raise(t, m, step.key)
		if m.graph.selected != step.want {
			t.Fatalf("selection %q after %q; want %q", m.graph.selected, step.key.String(), step.want)
		}
	}
	m = newTestDemoModel(t, "Main.fixture.json", demoSize{})
	for _, step := range []struct {
		key  tea.KeyPressMsg
		want string
	}{
		{key(tea.KeyDown, ""), "t4-worker-1"},
		{key(tea.KeyRight, ""), "t4-verifier-1"},
		{key(tea.KeyUp, ""), "manager-hive-g2"},
	} {
		m, _ = raise(t, m, step.key)
		if m.graph.selected != step.want {
			t.Fatalf("Main: selection %q after %q; want %q", m.graph.selected, step.key.String(), step.want)
		}
	}
}

// Decision 14: h hides the stopped agents and shows them again.
func TestAgentsHHidesStoppedAgents(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	requireView(t, m, "t4-worker-1", "h hides 2 stopped")
	m, _ = raise(t, m, key('h', "h"))
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, "t4-worker-1") || !strings.Contains(view, "h shows 2 stopped") {
		t.Fatalf("h did not hide the stopped agents:\n%s", view)
	}
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "t4-worker-1", "h hides 2 stopped")
}

// Decisions 14 and 19: enter opens the details popup, which keeps a, p, and
// t, and esc closes it. r does nothing on the graph and in the popup.
func TestAgentsDetailsPopupKeys(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	before := m.View().Content
	m, cmd := raise(t, m, key('r', "r"))
	if m.View().Content != before || cmd != nil || m.agents.restartConfirm {
		t.Fatal("r changed the Agents graph")
	}
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	if m.focusState() != "details" {
		t.Fatalf("focus %q after enter; want details", m.focusState())
	}
	requireView(t, m, "manager-hive · esc closes", "ACTIONS", "[details] a attach · p peek · t task · ↑↓ scroll · esc close")
	popup := m.View().Content
	m, cmd = raise(t, m, key('r', "r"))
	if m.View().Content != popup || cmd != nil {
		t.Fatal("r changed the details popup")
	}
	for _, action := range []string{"a", "p"} {
		m, _ = raise(t, m, key(rune(action[0]), action))
		requireView(t, m, demoNotice)
	}
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	if !m.popupOpen {
		m, _ = raise(t, m, key(tea.KeyEnter, ""))
	}
	m, _ = raise(t, m, key(tea.KeyEscape, "esc"))
	if m.popupOpen || m.focusState() != "graph" {
		t.Fatalf("esc did not close the popup: focus %q", m.focusState())
	}
	requireView(t, m, "[graph] ←↑↓→ move · enter details · a attach · p peek · t task · h stopped · ? help")
}

// Decision 35: a and p on the graph show the demo notice.
func TestAgentsDemoActionsShowTheNotice(t *testing.T) {
	for _, action := range []string{"a", "p"} {
		m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
		m, _ = raise(t, m, key(rune(action[0]), action))
		requireView(t, m, demoNotice)
	}
}

// t opens the task of the selected agent in Work › Tasks.
func TestAgentsTOpensTheTask(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	m, _ = raise(t, m, key('t', "t"))
	if m.activeTab() != tabWork || m.viewName() != "Tasks" || m.tasks.selectedID != "4-agents-equal-columns" {
		t.Fatalf("t gave tab %d view %q task %q", m.activeTab(), m.viewName(), m.tasks.selectedID)
	}
}

// Decision 21: below 120 columns the footer shows the compact hints of the
// graph.
func TestAgentsCompactFooter(t *testing.T) {
	m := newTestDemoModel(t, "CompactAgents.fixture.json", demoSize{})
	requireView(t, m, "[graph] ←↑↓→ move · enter details · a attach · p peek · ?")
}

// Decisions 37 and 38: normal mode reads the records of a scratch home with
// the same renderer, and R reloads them and keeps the selection.
func TestAgentsNormalModeLoadsAndRefreshes(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 14, 22, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", PID: int64(os.Getpid()), PIDStart: pidStart})
	writeRecord(t, home, []string{"projects"}, "alpha", records.Project{ID: "alpha", Path: "/repo/alpha", Base: "dev"})
	writeRecord(t, home, []string{"goals"}, "g1-alpha", records.Goal{ID: "g1-alpha", Project: "alpha", Status: "active", Waves: []records.GoalWave{{ID: "A", Status: "open"}}})
	_, agents := agentsProject("alpha", "pi", 2)
	for _, record := range agents {
		writeAgentRecord(t, home, record)
	}
	logDir := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The popup shows the event time in the zone of the clock.
	at := time.Date(2026, 10, 21, 14, 10, 0, 0, clock.now().Location()).Format(time.RFC3339)
	event := `{"time":"` + at + `","level":"INFO","msg":"task land 1-alpha","agent_id":"manager-alpha"}` + "\n"
	if err := os.WriteFile(dailyPath(home, clock.now()), []byte(event), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, cmd := raise(t, m, key('1', "1"))
	m = deliverAgentCommands(t, m, cmd)
	requireView(t, m, "project agents · 1 project", "g1-alpha · wave A", "◐ manager-alpha", "● t1-alpha-worker", "pi · working · 3m")
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	writeAgentRecord(t, home, records.Agent{Key: "t3-alpha-worker", Name: "t3-alpha-worker", Role: "worker", Project: "alpha", Task: "t3-alpha-worker", Harness: "pi", Parent: "manager-alpha", Status: records.AgentStopped, AgentStatus: "done"})
	m, cmd = raise(t, m, key('R', "R"))
	m = deliverAgentCommands(t, m, cmd)
	requireView(t, m, "h shows 1 stopped")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "○ t3-alpha-worker", "h hides 1 stopped")
	if m.graph.selected != "t1-alpha-worker" {
		t.Fatalf("selection %q after R; want t1-alpha-worker", m.graph.selected)
	}
	m, _ = raise(t, m, key(tea.KeyUp, ""))
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	requireView(t, m, "RECENT", "14:10 task land 1-alpha")
}

// openAgents opens the Agents tab of a normal-mode model at 120×40 and
// loads its records.
func openAgents(t *testing.T, home string, clock *testClock) *Model {
	t.Helper()
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, cmd := raise(t, m, key('1', "1"))
	return deliverAgentCommands(t, m, cmd)
}

// The graph shows the current generation of each agent. A replaced
// generation is not on the graph. A stopped current agent is on it after h.
func TestAgentsGraphShowsTheCurrentGenerationOnly(t *testing.T) {
	home, clock := agentsGenerationFixture(t)
	writeAgentRecord(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Role: "worker", Project: "demo", Task: "13-replaced", Status: records.AgentLive, AgentStatus: "idle", Parent: "manager-current"})
	writeAgentRecord(t, home, records.Agent{Key: "worker-new", Name: "worker-new", Role: "worker", Project: "demo", Task: "13-replaced", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-current", Replaces: "worker-old"})
	m := openAgents(t, home, clock)
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "manager-current", "● live-worker", "● worker-new", "○ active-ghost", "○ landed-ghost", "h hides 2 stopped")
	view := ansi.Strip(m.View().Content)
	for _, old := range []string{"manager-old", "worker-old"} {
		if strings.Contains(view, old) {
			t.Fatalf("the graph shows the replaced generation %s:\n%s", old, view)
		}
	}
}

// The details popup shows the nudge of the supervisor only for a current
// coordinator or manager with an unread nudged inbox item.
func TestAgentsDetailsShowAnUnreadNudge(t *testing.T) {
	for _, tc := range []struct {
		name, inbox, selected, readAt, nudgedAt string
		want                                    bool
	}{
		{name: "unread nudged manager", inbox: "manager-alpha", selected: "manager-alpha", nudgedAt: "2026-10-20T11:59:00Z", want: true},
		{name: "read nudged manager", inbox: "manager-alpha", selected: "manager-alpha", readAt: "2026-10-20T11:59:00Z", nudgedAt: "2026-10-20T11:58:00Z"},
		{name: "unread manager without nudge", inbox: "manager-alpha", selected: "manager-alpha"},
		{name: "nudged worker", inbox: "worker-alpha", selected: "worker-alpha", nudgedAt: "2026-10-20T11:59:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := agentsBoardModel(t)
			writeRecord(t, m.home, []string{"inbox", tc.inbox}, "1", records.Inbox{
				Seq: 1, Role: tc.inbox, Kind: records.InboxCompleted, Agent: "worker-alpha",
				Text: "completed", ReadAt: tc.readAt, NudgedAt: tc.nudgedAt,
			})
			m = refreshAgents(t, m)
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
			m.graph.selected = tc.selected
			m, _ = raise(t, m, key(tea.KeyEnter, ""))
			view := ansi.Strip(m.View().Content)
			if !strings.Contains(view, tc.selected+" · esc closes") {
				t.Fatalf("the popup of %s is not open:\n%s", tc.selected, view)
			}
			if got := strings.Contains(view, "1m ago by supervisor"); got != tc.want {
				t.Fatalf("popup shows the nudge = %t, want %t:\n%s", got, tc.want, view)
			}
		})
	}
}

// A worker or verifier without a parent record, or with a long agent key,
// sits in the pod of its project.
func TestAgentsPodHoldsAgentsWithoutAParentRecord(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	writeAgentRecord(t, home, records.Agent{Key: "worker-stenographer-long-key-g7", Name: "long-key-worker", Role: "worker", Project: "demo", Task: "11-build", Harness: "pi", Status: records.AgentLive, AgentStatus: "working"})
	m := openAgents(t, home, clock)
	requireView(t, m, "project agents · 1 project", "◐ manager-demo", "● legacy-verifier", "● long-key-worker")
}

// The pods follow the order of the project records.
func TestAgentsPodsFollowProjectRecordOrder(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	writeRecord(t, home, []string{"projects"}, "alpha", records.Project{ID: "alpha", Path: "/repo/alpha", Base: "dev"})
	writeRecord(t, home, []string{"projects"}, "zeta", records.Project{ID: "zeta", Path: "/repo/zeta", Base: "dev"})
	m := openAgents(t, home, clock)
	view := ansi.Strip(m.View().Content)
	alpha, demo, zeta := strings.Index(view, " alpha "), strings.Index(view, " demo "), strings.Index(view, " zeta ")
	if alpha < 0 || demo <= alpha || zeta <= demo {
		t.Fatalf("pods are missing or out of order: alpha %d, demo %d, zeta %d:\n%s", alpha, demo, zeta, view)
	}
}

// A stale agent is unknown. When the supervisor claim is stale, each card
// is unknown.
func TestAgentsCardsShowUnknownWhenTheSupervisorIsStale(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	writeAgentRecord(t, home, records.Agent{Key: "stale-worker", Name: "stale-worker", Role: "worker", Project: "demo", Task: "09-stale", Harness: "pi", PaneID: "w1:p9", Status: records.AgentStale, AgentStatus: "working", Parent: "manager-demo"})
	m := openAgents(t, home, clock)
	requireView(t, m, "claude · idle · 3m", "pi · unknown")
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", Generation: 1, PID: 0})
	m = refreshAgents(t, m)
	requireView(t, m, "claude · unknown", "codex · unknown · 5m")
}

// Each key of the graph logs its user action.
func TestAgentsKeysLogTheirActions(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := openAgents(t, home, clock)
	m, cmd := raise(t, m, key(tea.KeyDown, ""))
	if cmd == nil {
		t.Fatal("the selection did not log")
	}
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, key('h', "h"))
	if cmd == nil {
		t.Fatal("h did not log")
	}
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, key('a', "a"))
	if cmd == nil {
		t.Fatal("a did not return a command")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("a did not batch its log with the attach: %#v", batch)
	}
	_, _ = raise(t, m, batch[0]())
	data, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent selection"`, `"action":"agent hide stopped"`, `"hide_stopped":false`, `"action":"agent attach"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("TUI log omits %q:\n%s", want, data)
		}
	}
}

// requireWholeAgentsBorders is requireWholeBorders for the graph. On the rail,
// "├─" is the corner of a manager card and "├┤" and "╰┤" join a child card.
// "┡━", "├┨", and "╰┨" are the same joints on the heavy border of the
// selected card. Each ┏ needs a ┓, and each ┗ needs a ┛, in the same line.
func requireWholeAgentsBorders(t *testing.T, view string) {
	t.Helper()
	view = strings.NewReplacer("├┤", "││", "╰┤", "││", "├─", "╰─", "├┨", "│┃", "╰┨", "│┃", "┡━", "┗━").Replace(view)
	requireWholeBorders(t, view)
	for index, line := range strings.Split(view, "\n") {
		if strings.Count(line, "┏") != strings.Count(line, "┓") || strings.Count(line, "┗") != strings.Count(line, "┛") {
			t.Fatalf("line %d has a broken heavy border: %q\n%s", index+1, line, view)
		}
	}
}

// Decision 58: in normal mode the graph hides each stopped agent at start.
// h shows them and h hides them again. An unresolved agent stays visible.
func TestAgentsHideStoppedAtStart(t *testing.T) {
	home, clock := agentsGenerationFixture(t)
	writeAgentRecord(t, home, records.Agent{Key: "lost-worker", Name: "lost-worker", Role: "worker", Project: "demo", Task: "14-lost", Status: records.AgentUnresolved, Parent: "manager-current", StatusChangedAt: "2026-05-22T11:58:30Z"})
	m := openAgents(t, home, clock)
	requireView(t, m, "manager-current", "● live-worker", "lost-worker", "h shows 2 stopped")
	rejectView(t, m, "active-ghost", "landed-ghost")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "lost-worker", "○ active-ghost", "○ landed-ghost", "h hides 2 stopped")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "lost-worker", "h shows 2 stopped")
	rejectView(t, m, "active-ghost", "landed-ghost")
}

// Decision 60: a working card and each joint on its border are in working.
// An idle and a stopped card, their joints, and the rail outside a card keep
// border. A selected working card keeps the heavy accent border. The cells
// are those of AgentsHome.grid, column hive.
func TestAgentsWorkingCardBorder(t *testing.T) {
	m := newTestDemoModel(t, "AgentsHome.fixture.json", demoSize{})
	type cell struct {
		row, col int
		r        rune
		token    cellgrid.Token
	}
	check := func(m *Model, cells []cell) {
		t.Helper()
		got := renderGrid(t, m, 120, 40)
		for _, want := range cells {
			if c := got.At(want.row, want.col); c.Rune != want.r || c.Token != want.token {
				t.Errorf("row %d col %d: got %q %s, want %q %s", want.row, want.col, c.Rune, c.Token, want.r, want.token)
			}
		}
	}
	check(m, []cell{
		{7, 4, '╭', cellgrid.Working},  // manager-hive, working
		{10, 4, '├', cellgrid.Working}, // joint on the working manager card
		{12, 5, '╭', cellgrid.Border},  // t4-verifier-1, stopped
		{13, 4, '├', cellgrid.Border},  // rail outside the card
		{13, 5, '┤', cellgrid.Border},  // joint on the stopped card
		{17, 5, '╭', cellgrid.Working}, // t5-worker-1, working
		{18, 4, '├', cellgrid.Border},  // rail outside the working card
		{18, 5, '┤', cellgrid.Working}, // joint on the working card
		{20, 28, '╯', cellgrid.Working},
		{23, 4, '╰', cellgrid.Border},  // rail outside t6-worker-1
		{23, 5, '┤', cellgrid.Working}, // joint on t6-worker-1, working
		{17, 44, '╭', cellgrid.Border}, // t3-verifier-1, idle
		{18, 44, '┤', cellgrid.Border}, // joint on the idle card
	})
	m.graph.selected = "t5-worker-1"
	m.invalidateView()
	check(m, []cell{
		{17, 5, '┏', cellgrid.Accent},
		{18, 5, '┨', cellgrid.Accent},
		{20, 28, '┛', cellgrid.Accent},
		{18, 4, '├', cellgrid.Border},
	})
}
