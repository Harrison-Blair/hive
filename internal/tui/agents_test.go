package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestAgentsScreenRendersTopologyTabsAndDetail(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	requireAgentsView(t, m, "[ TREE ]", "home", "coordinator", "supervisor", "demo · Active demo goal", "manager-demo", "worker-1", "legacy-verifier")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	_ = m.currentScreen().View(120, 30)
	for range 4 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	requireAgentsView(t, m, "[ GRAPH ]", "parent manager-demo", "replaces old-worker", "model model-worker", "task 10-build", "pane w1:p4", "age 4m")
	view := ansi.Strip(m.currentScreen().View(120, 30))
	want := []string{
		"┌─ home ──────────────────────┐",
		"│  ● coordinator · claude · 1m│",
		"│  ○ supervisor · pi · 2m     │",
		"└─────────────────────────────┘",
		"┌─ demo · Active demo goal ────────────────────────────────┐",
		"│┄┄─ ○ manager-demo · claude · 3m ← coordinator            │",
		"│    ┄┄─ ● legacy-verifier · codex · 10-build              │",
		"│▶     └── ! worker-1 · pi · 10-build ⋯ replaces old-worker│",
		"└──────────────────────────────────────────────────────────┘",
		"coordinator ┄┄▶ manager-demo",
	}
	if !strings.Contains(view, strings.Join(want, "\n")) {
		t.Fatalf("graph lines are not the expected literal fixture:\n%s", view)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	requireAgentsView(t, m, "[ TABLE ]", "name", "role", "project", "task", "harness", "model", "status", "age", "parent", "worker-1 · worker · demo · 10-build · pi · model-worker · blocked · 4m · manager-demo")
}

func TestAgentsLoadsLongAgentKeysAsNodesAndParents(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	managerKey := "manager-stenographer-g7"
	writeAgentRecord(t, home, records.Agent{Key: managerKey, Name: "long-manager", Role: "manager", Project: "demo", Harness: "claude", Status: records.AgentLive, AgentStatus: "idle"})
	writeAgentRecord(t, home, records.Agent{Key: "long-key-worker", Name: "long-key-worker", Role: "worker", Project: "demo", Task: "11-build", Harness: "pi", Status: records.AgentLive, AgentStatus: "working", Parent: managerKey})

	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	requireAgentsView(t, m, "long-manager · manager · demo", "long-key-worker · worker · demo · 11-build · pi · — · working · ? · long-manager")
	if len(managerKey) <= 18 {
		t.Fatalf("fixture key %q must exceed the legacy ID limit", managerKey)
	}
}

func TestAgentsPanelFitsTerminalFrameAtStandardAndLargeSizes(t *testing.T) {
	for _, size := range []struct {
		width  int
		height int
	}{{width: 80, height: 24}, {width: 120, height: 40}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			home, clock := agentsFixture(t, 40)
			m := newTestModel(t, home, clock)
			m.width, m.height, m.screenIndex = size.width, size.height, 1
			m.layout()
			m, cmd := raise(t, m, key('2', "2"))
			m = deliverAgentCommands(t, m, cmd)

			headerRows := strings.Count(ansi.Strip(headerStyle.Width(size.width).Render(m.headerLine())), "\n") + 1
			bodyHeight := size.height - headerRows - statusRows - 5
			sidebarWidth := 24
			if size.width >= 100 {
				sidebarWidth = 30
			}
			sidebar := m.renderSidebar(sidebarWidth, bodyHeight)
			mainWidth := size.width - sidebarWidth
			panel := m.renderPanel("Agents", m.currentScreen().View(mainWidth, bodyHeight), mainWidth, bodyHeight)
			sidebarLines := strings.Count(ansi.Strip(sidebar), "\n") + 1
			panelLines := strings.Count(ansi.Strip(panel), "\n") + 1
			if panelLines != sidebarLines {
				t.Fatalf("Agents panel has %d rendered lines; sidebar has %d (body %d, screen %d)\n%s", panelLines, sidebarLines, bodyHeight, strings.Count(ansi.Strip(m.currentScreen().View(mainWidth, bodyHeight)), "\n")+1, ansi.Strip(panel))
			}
		})
	}
}

func TestAgentsScreenLogsTabAndSelectionActions(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if cmd == nil {
		t.Fatal("tab action did not log")
	}
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd == nil {
		t.Fatal("selection action did not log")
	}
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, key('h', "h"))
	if cmd == nil {
		t.Fatal("stopped visibility action did not log")
	}
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, key('a', "a"))
	if cmd == nil {
		t.Fatal("attach action did not return a command")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("attach action did not batch its log with attach: %#v", batch)
	}
	_, _ = raise(t, m, batch[0]())
	data, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent tab"`, `"tab":"graph"`, `"action":"agent selection"`, `"agent_id":"supervisor"`, `"action":"agent stopped visibility"`, `"hide_stopped":false`, `"action":"agent attach"`, `"pane_id":"w1:p2"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("TUI log omits %q:\n%s", want, data)
		}
	}
}

func TestAgentsScreenStatusHidesStoppedAndUsesUnknownWhenSupervisorStale(t *testing.T) {
	home, clock := agentsFixture(t, 4)
	writeAgentRecord(t, home, records.Agent{Key: "stale-worker", Name: "stale-worker", Role: "worker", Project: "demo", Harness: "pi", PaneID: "w1:p9", Status: records.AgentStale, AgentStatus: "working", Parent: "manager-demo"})
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	requireAgentsView(t, m, "working", "idle", "blocked", "supervisor · idle", "stale-worker · unknown")
	rejectAgentsView(t, m, "× old-worker", "stopped")
	m, _ = raise(t, m, key('h', "h"))
	requireAgentsView(t, m, "× old-worker", "× old-verifier", "stopped")
	view := m.currentScreen().View(120, 30)
	for _, stopped := range []string{"× old-worker · stopped · 6m", "× old-verifier · stopped · 7m"} {
		if !strings.Contains(view, dimStyle.Render(stopped)) {
			t.Fatalf("stopped agent is not dim: %s", stopped)
		}
	}
	m, _ = raise(t, m, key('h', "h"))
	rejectAgentsView(t, m, "× old-worker")

	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", Generation: 1, PID: 0})
	writeAgentRecord(t, home, records.Agent{Key: "idle-worker", Name: "idle-worker", Role: "worker", Project: "demo", Task: "10-build", Harness: "pi", Model: "model-idle", PaneID: "w1:p5", Status: records.AgentLive, AgentStatus: "done", Parent: "manager-demo", StatusChangedAt: "2026-05-22T11:56:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "no-age", Name: "no-age", Role: "worker", Project: "demo", Harness: "pi", PaneID: "w1:p8", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-demo"})
	m = refreshAgents(t, m)
	requireAgentsView(t, m, "unknown", "idle-worker", "no-age", "?")
}

func TestAgentsGraphKeepsSelectedAgentInViewWithFortyAgents(t *testing.T) {
	home, clock := agentsFixture(t, 40)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	_ = m.currentScreen().View(70, 12)
	for range 44 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	view := ansi.Strip(m.currentScreen().View(70, 12))
	if !strings.Contains(view, "▶     └── ● worker-39") {
		t.Fatalf("selected graph node is not in view:\n%s", view)
	}
	if strings.Contains(view, "worker-00") {
		t.Fatalf("graph did not scroll to the selected node:\n%s", view)
	}
}

func TestAgentsSelectionMovesWithViKeysAndDetailShowsInferredParent(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, key('j', "j"))
	m, _ = raise(t, m, key('j', "j"))
	requireAgentsView(t, m, "manager-demo", "parent inferred")
	m, _ = raise(t, m, key('j', "j"))
	requireAgentsView(t, m, "legacy-verifier", "parent inferred")
	m, _ = raise(t, m, key('j', "j"))
	requireAgentsView(t, m, "DETAIL worker-1", "parent manager-demo")
	m, _ = raise(t, m, key('k', "k"))
	requireAgentsView(t, m, "DETAIL legacy-verifier")
}

func TestAgentsRefreshKeepsSelectionByAgentKey(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	for range 4 {
		m, _ = raise(t, m, key('j', "j"))
	}
	requireAgentsView(t, m, "DETAIL worker-1")
	writeAgentRecord(t, home, records.Agent{Key: "aaa-worker", Name: "aaa-worker", Role: "worker", Project: "demo", Harness: "pi", PaneID: "w1:p8", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-demo"})
	m = refreshAgents(t, m)
	requireAgentsView(t, m, "DETAIL worker-1")
}

func TestAgentsProjectsFollowProjectRecordOrder(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	writeRecord(t, home, []string{"projects"}, "alpha", records.Project{ID: "alpha", Path: "/repo/alpha", Base: "dev"})
	writeRecord(t, home, []string{"projects"}, "zeta", records.Project{ID: "zeta", Path: "/repo/zeta", Base: "dev"})
	m := newTestModel(t, home, clock)
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	view := ansi.Strip(m.currentScreen().View(120, 30))
	alpha := strings.Index(view, "▾ alpha")
	demo := strings.Index(view, "▾ demo · Active demo goal")
	zeta := strings.Index(view, "▾ zeta")
	if alpha < 0 || demo <= alpha || zeta <= demo {
		t.Fatalf("project groups are missing or out of order:\n%s", view)
	}
}

func TestAgentsPeekReadsInCommandAndEscClosesIt(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	bin := t.TempDir()
	argsPath := filepath.Join(bin, "args")
	herdrPath := filepath.Join(bin, "herdr")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$HERDR_ARGS\"\n"
	for index := 1; index <= 25; index++ {
		script += fmt.Sprintf("printf 'line %%02d\\n' %d\n", index)
	}
	if err := os.WriteFile(herdrPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_ARGS", argsPath)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	_, cmd = raise(t, m, key(tea.KeyEnter, "enter"))
	if cmd == nil {
		t.Fatal("enter did not start pane read")
	}
	m = deliverAgentCommands(t, m, cmd)
	requireAgentsView(t, m, "PEEK · w1:p1", "line 06", "line 25")
	rejectAgentsView(t, m, "line 05", "line 01")
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, argument := range []string{"pane", "read", "w1:p1", "--lines", "20", "--format", "text"} {
		if !strings.Contains(string(args), argument) {
			t.Fatalf("herdr pane read args omitted %q: %s", argument, args)
		}
	}
	m, cmd = raise(t, m, key(tea.KeyEscape, "esc"))
	m = deliverAgentCommands(t, m, cmd)
	requireAgentsView(t, m, "[ TREE ]")
	logData, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent peek"`, `"action":"agent peek close"`, `"pane_id":"w1:p1"`} {
		if !strings.Contains(string(logData), want) {
			t.Fatalf("Agents log omits %q:\n%s", want, logData)
		}
	}
	_, cmd = raise(t, m, key('a', "a"))
	if cmd == nil {
		t.Fatal("a did not start attach")
	}
}

func TestAgentsProgramUsesStringReader(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	if err := RunHome(home, strings.NewReader("2\t\t\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	if !strings.Contains(out.String(), "[ TABLE ]") {
		t.Fatalf("program output did not render the Agents table:\n%s", out.String())
	}
}

func agentsFixture(t *testing.T, workers int) (string, *testClock) {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", Generation: 1, PID: int64(os.Getpid()), PIDStart: pidStart})
	writeRecord(t, home, []string{"projects"}, "demo", records.Project{ID: "demo", Path: "/repo/demo", Base: "dev"})
	writeRecord(t, home, []string{"goals"}, "g1-demo", records.Goal{ID: "g1-demo", Project: "demo", Title: "Active demo goal", Status: "active"})
	writeRecord(t, home, []string{"agents"}, "coordinator", records.Agent{Key: "coordinator", Name: "coordinator", Role: "coordinator", Harness: "claude", Model: "model-coordinator", PaneID: "w1:p1", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-05-22T11:59:00Z"})
	writeRecord(t, home, []string{"agents"}, "supervisor", records.Agent{Key: "supervisor", Name: "supervisor", Role: "supervisor", Harness: "pi", Model: "model-supervisor", PaneID: "w1:p2", Status: records.AgentLive, AgentStatus: "done", StatusChangedAt: "2026-05-22T11:58:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "manager-demo", Name: "manager-demo", Role: "manager", Project: "demo", Harness: "claude", Model: "model-manager", PaneID: "w1:p3", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:57:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Role: "worker", Project: "demo", Task: "10-build", Harness: "pi", Model: "model-worker", PaneID: "w1:p4", Status: records.AgentLive, AgentStatus: "blocked", Parent: "manager-demo", Replaces: "old-worker", StatusChangedAt: "2026-05-22T11:56:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "legacy-verifier", Name: "legacy-verifier", Role: "verifier", Project: "demo", Task: "10-build", Harness: "codex", Model: "model-verifier", PaneID: "w1:p6", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-05-22T11:55:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "old-worker", Name: "old-worker", Role: "worker", Project: "demo", Task: "9-old", Harness: "pi", PaneID: "w1:p7", Status: records.AgentExited, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:54:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "old-verifier", Name: "old-verifier", Role: "verifier", Project: "demo", Task: "8-old", Harness: "codex", PaneID: "w1:p8", Status: records.AgentStopped, AgentStatus: "working", StatusChangedAt: "2026-05-22T11:53:00Z"})
	for i := 0; i < workers; i++ {
		key := fmt.Sprintf("worker-%02d", i)
		writeAgentRecord(t, home, records.Agent{Key: key, Name: key, Role: "worker", Project: "demo", Task: fmt.Sprintf("task-%02d", i), Harness: "pi", Model: "model", PaneID: fmt.Sprintf("w1:p%d", i+10), Status: records.AgentLive, AgentStatus: "working", Parent: "manager-demo", StatusChangedAt: "2026-05-22T11:56:00Z"})
	}
	return home, clock
}

func writeAgentRecord(t *testing.T, home string, agent records.Agent) {
	t.Helper()
	writeRecord(t, home, []string{"agents"}, agent.Key, agent)
}

func refreshAgents(t *testing.T, model *Model) *Model {
	t.Helper()
	model, cmd := raise(t, model, tickMsg{})
	if cmd == nil {
		return model
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return deliverAgentCommands(t, model, func() tea.Msg { return msg })
	}
	for _, command := range batch[1:] {
		model = deliverAgentCommands(t, model, command)
	}
	return model
}

func deliverAgentCommands(t *testing.T, model *Model, cmd tea.Cmd) *Model {
	t.Helper()
	if cmd == nil {
		return model
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, command := range batch {
			model = deliverAgentCommands(t, model, command)
		}
		return model
	}
	model, next := raise(t, model, msg)
	return deliverAgentCommands(t, model, next)
}

func requireAgentsView(t *testing.T, model *Model, wants ...string) {
	t.Helper()
	view := ansi.Strip(model.currentScreen().View(120, 30))
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Fatalf("Agents view does not contain %q:\n%s", want, view)
		}
	}
}

func rejectAgentsView(t *testing.T, model *Model, unwanted ...string) {
	t.Helper()
	view := ansi.Strip(model.currentScreen().View(120, 30))
	for _, item := range unwanted {
		if strings.Contains(view, item) {
			t.Fatalf("Agents view contains %q:\n%s", item, view)
		}
	}
}
