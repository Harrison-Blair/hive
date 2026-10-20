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

func TestAgentsDefaultShowsCurrentGenerationsAndActiveTaskGhosts(t *testing.T) {
	home, clock := agentsGenerationFixture(t)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)

	want := []string{"coordinator-current", "supervisor-current", "manager-current", "live-worker", "active-ghost"}
	all := []string{
		"coordinator-old", "coordinator-current", "supervisor-old", "supervisor-current",
		"manager-old", "manager-current", "live-worker", "active-ghost", "landed-ghost",
	}
	for _, tab := range []agentsTab{agentsTree, agentsGraph, agentsTable} {
		m.agents.tab = tab
		lines, _ := m.agents.contentLines()
		for _, name := range want {
			if got := countAgentRows(lines, name); got != 1 {
				t.Errorf("tab %d has %d rows for %q, want one:\n%s", tab, got, name, ansi.Strip(strings.Join(lines, "\n")))
			}
		}
		for _, name := range all {
			if containsAgentName(want, name) {
				continue
			}
			if got := countAgentRows(lines, name); got != 0 {
				t.Errorf("tab %d has %d rows for hidden generation %q:\n%s", tab, got, name, ansi.Strip(strings.Join(lines, "\n")))
			}
		}
		var dimGhost string
		switch tab {
		case agentsTree:
			dimGhost = dimStyle.Render("× active-ghost · stopped · 3m")
		case agentsGraph:
			dimGhost = dimStyle.Render("× active-ghost")
		case agentsTable:
			dimGhost = dimStyle.Render("active-ghost · worker · demo · 11-active · — · — · stopped · 3m · manager-current")
		}
		if !strings.Contains(strings.Join(lines, "\n"), dimGhost) {
			t.Errorf("tab %d does not dim the active-task ghost: %s", tab, ansi.Strip(strings.Join(lines, "\n")))
		}
	}
}

func TestAgentsHistoryToggleShowsAllGenerations(t *testing.T) {
	home, clock := agentsGenerationFixture(t)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	if !strings.Contains(m.currentScreen().KeyHints(), "h hide stopped") {
		t.Fatalf("default footer does not show the stopped-agent hint: %s", m.currentScreen().KeyHints())
	}
	m, _ = raise(t, m, key('h', "h"))
	if !strings.Contains(m.currentScreen().KeyHints(), "h hide stopped") {
		t.Fatalf("footer does not keep the stopped-agent hint: %s", m.currentScreen().KeyHints())
	}

	want := []string{
		"coordinator-old", "coordinator-current", "supervisor-old", "supervisor-current",
		"manager-old", "manager-current", "live-worker", "active-ghost", "landed-ghost",
	}
	for _, tab := range []agentsTab{agentsTree, agentsGraph, agentsTable} {
		m.agents.tab = tab
		lines, _ := m.agents.contentLines()
		for _, name := range want {
			if got := countAgentRows(lines, name); got != 1 {
				t.Errorf("tab %d has %d history rows for %q, want one:\n%s", tab, got, name, ansi.Strip(strings.Join(lines, "\n")))
			}
		}
		var landedExample string
		switch tab {
		case agentsTree:
			landedExample = "✓ landed-ghost · stopped · 4m · 12 landed · abc1234"
		case agentsGraph:
			landedExample = "✓ landed-ghost"
		case agentsTable:
			landedExample = "✓ landed-ghost · worker · demo · 12-landed · — · — · stopped · 4m · manager-current · 12 landed · abc1234"
		}
		if !strings.Contains(ansi.Strip(strings.Join(lines, "\n")), landedExample) {
			t.Errorf("tab %d omits the landed ghost label %q: %s", tab, landedExample, ansi.Strip(strings.Join(lines, "\n")))
		}
		var dimExample string
		switch tab {
		case agentsTree:
			dimExample = dimStyle.Render("○ coordinator-old · idle · 1m")
		case agentsGraph:
			dimExample = dimStyle.Render("○ coordinator-old")
		case agentsTable:
			dimExample = dimStyle.Render("coordinator-old · coordinator · — · — · — · — · idle · 1m · none")
		}
		if !strings.Contains(strings.Join(lines, "\n"), dimExample) {
			t.Errorf("tab %d does not dim old generations: %s", tab, ansi.Strip(strings.Join(lines, "\n")))
		}
	}
}

func TestAgentsLegacyGenerationFallbackUsesStatusChangeThenRecordTime(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	for _, agent := range []records.Agent{
		{Key: "legacy-a", Name: "legacy-coordinator", Role: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:58:00Z", UpdatedAt: "2026-05-22T11:59:00Z"},
		{Key: "legacy-b", Name: "legacy-coordinator", Role: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:56:00Z"},
		{Key: "legacy-c", Name: "legacy-coordinator", Role: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:57:00Z"},
	} {
		writeAgentRecord(t, home, agent)
	}
	data, err := loadAgents(home, clock.now())
	if err != nil {
		t.Fatal(err)
	}
	state := agentsState{data: data}
	visible := state.visibleNodes()
	if len(visible) != 1 || visible[0].record.Key != "legacy-c" {
		t.Fatalf("visible legacy generations = %#v, want only record legacy-c", visible)
	}
}

func agentsGenerationFixture(t *testing.T) (string, *testClock) {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor-current", Generation: 2, PID: int64(os.Getpid()), PIDStart: pidStart})
	writeRecord(t, home, []string{"projects"}, "demo", records.Project{ID: "demo", Path: "/repo/demo", Base: "dev"})
	for _, record := range []records.Agent{
		{Key: "coordinator-old", Name: "coordinator-old", Role: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:55:00Z"},
		{Key: "coordinator-current", Name: "coordinator-current", Role: "coordinator", Status: records.AgentLive, AgentStatus: "working", Replaces: "coordinator-old", StatusChangedAt: "2026-05-22T11:58:00Z", UpdatedAt: "2026-05-22T11:59:00Z"},
		{Key: "supervisor-old", Name: "supervisor-old", Role: "supervisor", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:55:00Z"},
		{Key: "supervisor-current", Name: "supervisor-current", Role: "supervisor", Status: records.AgentLive, AgentStatus: "working", Parent: "coordinator-current", Replaces: "supervisor-old", StatusChangedAt: "2026-05-22T11:58:00Z", UpdatedAt: "2026-05-22T11:59:00Z"},
		{Key: "manager-old", Name: "manager-old", Role: "manager", Project: "demo", Status: records.AgentLive, AgentStatus: "idle", Parent: "coordinator-old", StatusChangedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:55:00Z"},
		{Key: "manager-current", Name: "manager-current", Role: "manager", Project: "demo", Status: records.AgentLive, AgentStatus: "working", Parent: "coordinator-current", Replaces: "manager-old", StatusChangedAt: "2026-05-22T11:58:00Z", UpdatedAt: "2026-05-22T11:59:00Z"},
		{Key: "live-worker", Name: "live-worker", Role: "worker", Project: "demo", Task: "10-live", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-current", StatusChangedAt: "2026-05-22T11:58:00Z"},
		{Key: "active-ghost", Name: "active-ghost", Role: "worker", Project: "demo", Task: "11-active", Status: records.AgentStopped, Parent: "manager-current", StatusChangedAt: "2026-05-22T11:57:00Z"},
		{Key: "landed-ghost", Name: "landed-ghost", Role: "worker", Project: "demo", Task: "12-landed", Status: records.AgentExited, Parent: "manager-current", StatusChangedAt: "2026-05-22T11:56:00Z"},
	} {
		writeAgentRecord(t, home, record)
	}
	writeRecord(t, home, []string{"tasks"}, "11-active", records.Task{ID: "11-active", Project: "demo", Status: "assigned"})
	writeRecord(t, home, []string{"tasks"}, "12-landed", records.Task{ID: "12-landed", Project: "demo", Status: "verified", Delivery: &records.Delivery{Commit: "abc1234def", Pushed: true}})
	return home, clock
}

func countAgentRows(lines []string, name string) int {
	count := 0
	for _, line := range lines {
		plain := ansi.Strip(line)
		index := strings.Index(plain, name+" ·")
		if strings.Contains(plain, "│ │") && strings.Contains(plain, name) {
			prefix := plain[:strings.Index(plain, name)]
			if hasAgentMarkAnywhere(prefix) {
				count++
			}
			continue
		}
		if index < 0 {
			continue
		}
		prefix := plain[:index]
		if index == 0 || strings.HasSuffix(prefix, "▶ ") || hasAgentMark(prefix) {
			count++
		}
	}
	return count
}

func hasAgentMarkAnywhere(prefix string) bool {
	for _, mark := range []string{"● ", "○ ", "! ", "× ", "? ", "✓ "} {
		if strings.Contains(prefix, mark) {
			return true
		}
	}
	return strings.Contains(prefix, "▶ ")
}

func hasAgentMark(prefix string) bool {
	for _, mark := range []string{"● ", "○ ", "! ", "× ", "? ", "✓ "} {
		if strings.HasSuffix(prefix, mark) {
			return true
		}
	}
	return false
}

func containsAgentName(names []string, name string) bool {
	for _, item := range names {
		if item == name {
			return true
		}
	}
	return false
}

func TestAgentsScreenRendersTopologyTabsAndDetail(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	requireAgentsView(t, m, "[ TREE ]", "home", "coordinator", "supervisor", "demo · Active demo goal", "manager-demo", "worker-1", "legacy-verifier")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	_ = m.currentScreen().View(120, 30)
	for range 4 {
		m, _ = raise(t, m, key('j', "j"))
	}
	requireAgentsView(t, m, "[ GRAPH ]", "parent manager-demo", "model model-worker", "task 10-build", "pane w1:p4")
	view := ansi.Strip(m.currentScreen().View(120, 30))
	for _, want := range []string{
		"≡ home", "≡ demo ·", "coordinator", "manager-demo", "legacy-verifier", "worker-1",
		"cc idle · 3m", "edges: ─ delegates · ┄ nudges · ⋯ replaced",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("graph omits board element %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "coordinator ──▶ manager-demo") {
		t.Fatalf("graph still lists edges as text below the pods:\n%s", view)
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
	for _, want := range []string{`"action":"agent tab"`, `"tab":"graph"`, `"action":"agent selection"`, `"agent_id":"supervisor"`, `"action":"agent history visibility"`, `"history":true`, `"action":"agent attach"`, `"pane_id":"w1:p2"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("TUI log omits %q:\n%s", want, data)
		}
	}
}

func TestAgentsScreenStatusHidesStoppedAndUsesUnknownWhenSupervisorStale(t *testing.T) {
	home, clock := agentsFixture(t, 4)
	writeAgentRecord(t, home, records.Agent{Key: "stale-worker", Name: "stale-worker", Role: "worker", Project: "demo", Task: "09-stale", Harness: "pi", PaneID: "w1:p9", Status: records.AgentStale, AgentStatus: "working", Parent: "manager-demo"})
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
	writeAgentRecord(t, home, records.Agent{Key: "idle-worker", Name: "idle-worker", Role: "worker", Project: "demo", Task: "11-idle", Harness: "pi", Model: "model-idle", PaneID: "w1:p5", Status: records.AgentLive, AgentStatus: "done", Parent: "manager-demo", StatusChangedAt: "2026-05-22T11:56:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "no-age", Name: "no-age", Role: "worker", Project: "demo", Task: "13-no-age", Harness: "pi", PaneID: "w1:p8", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-demo"})
	m = refreshAgents(t, m)
	requireAgentsView(t, m, "unknown", "idle-worker", "no-age", "?")
	for _, node := range m.agents.visibleNodes() {
		if node.record.Role == "supervisor" && node.status != "unknown" {
			t.Fatalf("stale supervisor status = %q, want unknown", node.status)
		}
	}
}

func TestAgentsGraphKeepsSelectedAgentInViewWithFortyAgents(t *testing.T) {
	home, clock := agentsFixture(t, 40)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	_ = m.currentScreen().View(70, 12)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	for range 44 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	view := ansi.Strip(m.currentScreen().View(70, 12))
	if !strings.Contains(view, "▶ ● worker-39") {
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
	requireAgentsView(t, m, "worker-1  parent manager-demo")
	m, _ = raise(t, m, key('k', "k"))
	requireAgentsView(t, m, "legacy-verifier  parent inferred")
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
	requireAgentsView(t, m, "worker-1  parent manager-demo")
	writeAgentRecord(t, home, records.Agent{Key: "aaa-worker", Name: "aaa-worker", Role: "worker", Project: "demo", Harness: "pi", PaneID: "w1:p8", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-demo"})
	m = refreshAgents(t, m)
	requireAgentsView(t, m, "worker-1  parent manager-demo")
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

func TestAgentsHeaderUsesBoardLabels(t *testing.T) {
	m := agentsBoardModel(t)
	frame := ansi.Strip(m.View().Content)
	for _, want := range []string{"5 agents", "2 working", "3 idle", "supervisor ●", "needs you 0"} {
		if !strings.Contains(frame, want) {
			t.Errorf("Agents header omits %q:\n%s", want, frame)
		}
	}
	for _, unwanted := range []string{"supervisor fresh", "needs_user"} {
		if strings.Contains(frame, unwanted) {
			t.Errorf("Agents header contains %q:\n%s", unwanted, frame)
		}
	}
}

func TestAgentsBoardGoldenFrames(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{{name: "120x40", width: 120, height: 40}, {name: "200x57", width: 200, height: 57}}
	tabs := []struct {
		name string
		tab  agentsTab
	}{{name: "tree", tab: agentsTree}, {name: "graph", tab: agentsGraph}, {name: "table", tab: agentsTable}}

	for _, size := range sizes {
		for _, tab := range tabs {
			t.Run(size.name+"/"+tab.name, func(t *testing.T) {
				m := agentsBoardModel(t)
				m.now = func() time.Time { return time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC) }
				for index := range m.agents.data.nodes {
					if m.agents.data.nodes[index].record.Role == "supervisor" {
						m.agents.data.nodes[index].pid = "551386"
					}
				}
				m.agents.tab = tab.tab
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
				frame := ansi.Strip(m.View().Content)
				if strings.Contains(frame, "SCREENS") || !strings.Contains(frame, "AGENTS") {
					t.Fatalf("Agents does not use the board explorer layout:\n%s", frame)
				}
				lines := strings.Split(frame, "\n")
				if len(lines) != size.height {
					t.Fatalf("frame has %d lines, want %d", len(lines), size.height)
				}
				for index, line := range lines {
					if got := ansi.StringWidth(line); got != size.width {
						t.Fatalf("line %d has width %d, want %d: %q", index+1, got, size.width, line)
					}
				}
				goldenPath := filepath.Join("testdata", "agents-board", tab.name+"-"+size.name+".golden")
				golden, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Fatal(err)
				}
				if got := strings.TrimSuffix(frame, "\n"); got != strings.TrimSuffix(string(golden), "\n") {
					t.Fatalf("Agents frame differs from %s:\n%s", goldenPath, got)
				}
			})
		}
	}
}

func TestAgentsBoardDetailShowsTaskStatusAndNextTask(t *testing.T) {
	m := agentsBoardModel(t)
	for range 3 {
		m, _ = raise(t, m, key('j', "j"))
	}
	if m.agents.selectedKey != "worker-alpha" {
		t.Fatalf("selection = %q, want worker-alpha", m.agents.selectedKey)
	}
	view := ansi.Strip(m.currentScreen().View(120, 36))
	for _, want := range []string{
		"worker-alpha  parent manager-alpha", "model model-alpha", "task 10-alpha  assigned",
		"pane w1:p4", "next 11-next", "╭", "╰",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("Agents detail omits %q:\n%s", want, view)
		}
	}
	unknown := agentsState{data: agentsData{nodes: []agentNode{{record: records.Agent{Name: "unknown"}, status: "unknown"}}}}
	if got := unknown.detailLine(unknown.data.nodes[0]); !strings.Contains(got, "parent ?  model ?  task ?  ?  pane ?  next ?") {
		t.Fatalf("unknown detail values are not marked with ?: %s", got)
	}
}

func TestAgentsGhostsRequireAnOpenTask(t *testing.T) {
	m := agentsBoardModel(t)
	writeRecord(t, m.home, []string{"goals"}, "g1-alpha", records.Goal{
		ID: "g1-alpha", Project: "alpha", Title: "Alpha goal", Status: "active",
		Waves: []records.GoalWave{{ID: "wave-a", Tasks: []string{"21-verified-active", "23-landed"}}},
	})
	for _, task := range []records.Task{
		{ID: "20-cancelled", Project: "alpha", Status: "cancelled"},
		{ID: "21-verified-active", Project: "alpha", Status: "verified"},
		{ID: "22-verified-unassigned", Project: "alpha", Status: "verified"},
		{ID: "23-landed", Project: "alpha", Status: "verified", Delivery: &records.Delivery{Commit: "abc1234", Pushed: true}},
		{ID: "24-completed", Project: "alpha", Status: "completed"},
	} {
		writeRecord(t, m.home, []string{"tasks"}, task.ID, task)
	}
	for _, record := range []records.Agent{
		{Key: "ghost-cancelled", Name: "ghost-cancelled", Role: "worker", Project: "alpha", Task: "20-cancelled", Parent: "manager-alpha", Status: records.AgentStopped},
		{Key: "ghost-verified-active", Name: "ghost-verified-active", Role: "worker", Project: "alpha", Task: "21-verified-active", Parent: "manager-alpha", Status: records.AgentStopped},
		{Key: "ghost-verified-unassigned", Name: "ghost-verified-unassigned", Role: "worker", Project: "alpha", Task: "22-verified-unassigned", Parent: "manager-alpha", Status: records.AgentStopped},
		{Key: "ghost-landed", Name: "ghost-landed", Role: "worker", Project: "alpha", Task: "23-landed", Parent: "manager-alpha", Status: records.AgentStopped},
		{Key: "ghost-completed", Name: "ghost-completed", Role: "worker", Project: "alpha", Task: "24-completed", Parent: "manager-alpha", Status: records.AgentStopped},
	} {
		writeAgentRecord(t, m.home, record)
	}
	m = refreshAgents(t, m)
	for _, name := range []string{"ghost-beta", "ghost-verified-active", "ghost-completed"} {
		if !containsAgentName(agentNames(m.agents.visibleNodes()), name) {
			t.Errorf("open task ghost %q is hidden", name)
		}
	}
	for _, name := range []string{"ghost-cancelled", "ghost-verified-unassigned", "ghost-landed"} {
		if containsAgentName(agentNames(m.agents.visibleNodes()), name) {
			t.Errorf("closed task ghost %q is visible without history", name)
		}
	}
	explorerLines, _ := m.agents.explorerLines()
	explorer := strings.Join(explorerLines, "\n")
	for _, name := range []string{"ghost-cancelled", "ghost-verified-unassigned", "ghost-landed"} {
		if strings.Contains(explorer, name) {
			t.Errorf("closed task ghost %q is visible in the explorer without history", name)
		}
	}
	m, _ = raise(t, m, key('h', "h"))
	for _, name := range []string{"ghost-cancelled", "ghost-verified-unassigned", "ghost-landed"} {
		if !containsAgentName(agentNames(m.agents.visibleNodes()), name) {
			t.Errorf("history hides stopped agent %q", name)
		}
	}
}

func TestAgentsSupervisorRowUsesLifecycleClaimPIDAndStartAge(t *testing.T) {
	m := agentsBoardModel(t)
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, m.home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor-g16", Generation: 16, PID: int64(os.Getpid()), PIDStart: pidStart,
	})
	writeAgentRecord(t, m.home, records.Agent{
		Key: "supervisor-g16", Name: "supervisor", Role: "supervisor", Status: records.AgentLive,
		Generation: 16, PaneID: "wBC:p19", StartedAt: "2026-10-20T11:10:00Z", UpdatedAt: "2026-10-20T11:10:00Z",
	})
	m = refreshAgents(t, m)
	var supervisor agentNode
	for _, node := range m.agents.visibleNodes() {
		if node.record.Key == "supervisor-g16" {
			supervisor = node
			break
		}
	}
	if supervisor.record.Key == "" {
		t.Fatal("live supervisor generation 16 is not visible")
	}
	if supervisor.status != "working" || supervisor.age != "50m" {
		t.Fatalf("supervisor status/age = %q/%q, want working/50m", supervisor.status, supervisor.age)
	}
	m.agents.tab = agentsGraph
	lines, _ := m.agents.graphLines()
	view := strings.Join(lines, "\n")
	if !strings.Contains(view, fmt.Sprintf("pid %d · 50m", os.Getpid())) {
		t.Fatalf("supervisor graph row does not show its pid and age:\n%s", view)
	}
}

func agentNames(nodes []agentNode) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.record.Name)
	}
	return names
}

func TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		readAt   string
		nudgedAt string
		wantEdge bool
	}{
		{name: "unread nudged manager", role: "manager-alpha", nudgedAt: "2026-10-20T11:59:00Z", wantEdge: true},
		{name: "read nudged manager", role: "manager-alpha", readAt: "2026-10-20T11:59:00Z", nudgedAt: "2026-10-20T11:58:00Z"},
		{name: "unread manager without nudge", role: "manager-alpha"},
		{name: "unread nudged coordinator", role: "coordinator", nudgedAt: "2026-10-20T11:59:00Z", wantEdge: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			m := agentsBoardModel(t)
			writeRecord(t, m.home, []string{"inbox", test.role}, "1", records.Inbox{
				Seq: 1, Role: test.role, Kind: records.InboxCompleted, Agent: "worker-alpha",
				Text: "completed", ReadAt: test.readAt, NudgedAt: test.nudgedAt,
			})
			m = refreshAgents(t, m)
			m.agents.tab = agentsGraph
			lines, _ := m.agents.graphLines()
			graph := strings.Join(lines, "\n")
			hasNudge := false
			for index, line := range lines {
				plain := ansi.Strip(line)
				if !strings.Contains(plain, "nudges") {
					continue
				}
				hasNudge = strings.Contains(plain, "┄") || index > 0 && strings.Contains(ansi.Strip(lines[index-1]), "┄")
			}
			if hasNudge != test.wantEdge {
				t.Fatalf("nudge edge present = %t, want %t:\n%s", hasNudge, test.wantEdge, graph)
			}
			if strings.Contains(graph, "supervisor ┄┄▶ "+test.role) {
				t.Fatalf("graph still lists the nudge as text:\n%s", graph)
			}
			view := ansi.Strip(m.currentScreen().View(90, 32))
			if !strings.Contains(view, "edges: ─ delegates · ┄ nudges · ⋯ replaced") {
				t.Fatalf("graph view omits the edge legend:\n%s", view)
			}
		})
	}
}

func TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers(t *testing.T) {
	m := agentsBoardModel(t)
	writeAgentRecord(t, m.home, records.Agent{
		Key: "verifier-alpha", Name: "verifier-alpha", Role: "verifier", Project: "alpha",
		Parent: "manager-alpha", Status: records.AgentLive, AgentStatus: "working",
	})
	for _, role := range []string{"worker-alpha", "verifier-alpha"} {
		writeRecord(t, m.home, []string{"inbox", role}, "1", records.Inbox{
			Seq: 1, Role: role, Kind: records.InboxCompleted, NudgedAt: "2026-10-20T11:59:00Z",
		})
	}
	m = refreshAgents(t, m)
	m.agents.tab = agentsGraph
	lines, _ := m.agents.graphLines()
	graph := strings.Join(lines, "\n")
	if strings.Contains(graph, "nudges") {
		t.Fatalf("drawn graph has a nudge route for a worker or verifier:\n%s", graph)
	}
}

func TestAgentsGraphConnectsSiblingWorkersToManagerStem(t *testing.T) {
	m := agentsBoardModel(t)
	writeAgentRecord(t, m.home, records.Agent{
		Key: "worker-alpha-sibling", Name: "worker-alpha-sibling", Role: "worker", Project: "alpha",
		Task: "11-next", Harness: "pi", Parent: "manager-alpha", Status: records.AgentLive, AgentStatus: "working",
	})
	m = refreshAgents(t, m)
	m.agents.tab = agentsGraph
	lines, positions := m.agents.graphLines()
	for _, key := range []string{"manager-alpha", "worker-alpha", "worker-alpha-sibling"} {
		if _, ok := positions[key]; !ok {
			t.Fatalf("graph position missing %q: %#v", key, positions)
		}
	}
	graph := strings.Join(lines, "\n")
	for _, connector := range []string{"┬", "┴"} {
		if !strings.Contains(graph, connector) {
			t.Fatalf("graph is missing manager stem connector %q:\n%s", connector, graph)
		}
	}
}

func TestAgentsGraphArrowSelectionMovesBetweenPodsAndScrolls(t *testing.T) {
	m := agentsBoardModel(t)
	m.agents.tab = agentsGraph
	_ = m.currentScreen().View(90, 32)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	if m.agents.selectedKey != "manager-alpha" {
		t.Fatalf("right selected %q, want manager-alpha", m.agents.selectedKey)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.agents.selectedKey != "worker-alpha" {
		t.Fatalf("down selected %q, want worker-alpha", m.agents.selectedKey)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	if m.agents.selectedKey != "ghost-beta" {
		t.Fatalf("right selected %q, want ghost-beta", m.agents.selectedKey)
	}
	view := ansi.Strip(m.currentScreen().View(90, 32))
	if !strings.Contains(view, "▶ × ghost-beta") || m.agents.graphOffset == 0 {
		t.Fatalf("graph did not scroll with the selected node (offset %d):\n%s", m.agents.graphOffset, view)
	}
}

func TestAgentsBoardTaskShortcutOpensSelectedTask(t *testing.T) {
	m := agentsBoardModel(t)
	for range 3 {
		m, _ = raise(t, m, key('j', "j"))
	}
	m.screenLoading = true
	m, cmd := raise(t, m, key('t', "t"))
	if m.screenIndex != 4 {
		t.Fatalf("screen index = %d, want Tasks screen index 4", m.screenIndex)
	}
	if m.tasks.selectedID != "10-alpha" || m.tasks.tab != tasksAll {
		t.Fatalf("Tasks selection = %q in tab %d, want task 10-alpha in all", m.tasks.selectedID, m.tasks.tab)
	}
	if cmd == nil {
		t.Fatal("t did not load the Tasks screen")
	}
	m = deliverAgentCommands(t, m, cmd)
	if view := ansi.Strip(m.currentScreen().View(120, 36)); !strings.Contains(view, "> 10-alpha") {
		t.Fatalf("Tasks screen did not select task 10-alpha:\n%s", view)
	}
}

func TestAgentsBoardKeyHintsMatchTheBoard(t *testing.T) {
	m := agentsBoardModel(t)
	want := "←↑↓→ move on graph · tab tree/graph/table · enter peek pane · a attach · h hide stopped · t open task · ask in Chat"
	if got := m.currentScreen().KeyHints(); got != want {
		t.Fatalf("Agents key hints = %q, want %q", got, want)
	}
}

func TestAgentsGraphPodTitleSharesTopBorder(t *testing.T) {
	m := agentsBoardModel(t)
	nodes := m.agents.nodeMap()
	homePod, _, _, _ := m.agents.graphPod(m.agents.data.groups[0], nodes)
	if !strings.HasPrefix(homePod[0], "╭─ ≡ home ─") || strings.Contains(homePod[1], "≡ home") {
		t.Fatalf("home pod title is not on the top border:\n%s", strings.Join(homePod, "\n"))
	}
	alphaPod, _, _, _ := m.agents.graphPod(m.agents.data.groups[1], nodes)
	borderText := strings.NewReplacer("╭", "", "╮", "", "─", "", "┬", "").Replace(alphaPod[0])
	if got := strings.Join(strings.Fields(borderText), " "); !strings.Contains(got, "≡ alpha · Alpha goal for long graph") {
		t.Fatalf("alpha pod title is incomplete on the top border: %s", alphaPod[0])
	}
}

func TestAgentsExplorerUsesIndentationWithoutConnectors(t *testing.T) {
	m := agentsBoardModel(t)
	lines, _ := m.agents.explorerLines()
	explorer := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"  ▶ ● coordinator", "  ○ supervisor", "  ○ manager-alpha", "    ● worker-alpha"} {
		if !strings.Contains(explorer, want) {
			t.Errorf("Agents explorer omits %q:\n%s", want, explorer)
		}
	}
	if got := strings.Count(explorer, "supervisor"); got != 1 {
		t.Fatalf("Agents explorer has %d supervisor rows, want one:\n%s", got, explorer)
	}
	if strings.Contains(explorer, "├─") || strings.Contains(explorer, "└─") || strings.Contains(explorer, "┄┄") {
		t.Fatalf("Agents explorer draws connectors:\n%s", explorer)
	}
}

func TestAgentsGraphHasNoStemForManagerWithOnlyHiddenChildren(t *testing.T) {
	m := agentsBoardModel(t)
	writeRecord(t, m.home, []string{"tasks"}, "20-ghost", records.Task{ID: "20-ghost", Project: "beta", Status: "cancelled"})
	data, err := loadAgents(m.home, m.now())
	if err != nil {
		t.Fatal(err)
	}
	m.agents.accept(data)
	m.agents.tab = agentsGraph
	lines, positions := m.agents.graphLayout()
	manager, ok := positions["manager-beta"]
	if !ok {
		t.Fatal("graph omits manager-beta")
	}
	cardBottom := ansi.Strip(ansi.Cut(lines[manager.row+2], manager.col, manager.col+manager.width))
	if strings.Contains(cardBottom, "┴") {
		t.Fatalf("manager-beta draws a stem with no visible child: %s", cardBottom)
	}
}

func TestAgentsGraphDelegationBusUsesOpenEndGlyphs(t *testing.T) {
	m := agentsBoardModel(t)
	lines, positions := m.agents.graphLayout()
	bus := []rune(ansi.Strip(lines[0]))
	for _, test := range []struct {
		key  string
		want rune
	}{{key: "coordinator", want: '┌'}, {key: "manager-alpha", want: '┬'}, {key: "manager-beta", want: '┐'}} {
		position := positions[test.key]
		column := position.col + position.width/2
		if got := bus[column]; got != test.want {
			t.Errorf("delegation bus at %s = %q, want %q: %s", test.key, got, test.want, string(bus))
		}
	}
}

func TestAgentsTreeLastVisibleChildUsesBottomConnector(t *testing.T) {
	m := agentsBoardModel(t)
	writeAgentRecord(t, m.home, records.Agent{
		Key: "z-hidden-worker", Name: "hidden-worker", Role: "worker", Project: "alpha",
		Parent: "manager-alpha", Status: records.AgentStopped,
	})
	m = refreshAgents(t, m)
	lines, positions := m.agents.treeLines()
	row, ok := positions["worker-alpha"]
	if !ok {
		t.Fatal("TREE omits worker-alpha")
	}
	if !strings.Contains(lines[row], "└─ ● worker-alpha") {
		t.Fatalf("last visible worker has the wrong connector: %s", lines[row])
	}
}

func TestAgentsGraphNodeSubtitlesHaveNoExtraSeparator(t *testing.T) {
	m := agentsBoardModel(t)
	m.agents.tab = agentsGraph
	lines, _ := m.agents.graphLines()
	graph := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"cc idle · 3m", "pi 10-alpha"} {
		if !strings.Contains(graph, want) {
			t.Errorf("graph omits subtitle %q:\n%s", want, graph)
		}
	}
	for _, unwanted := range []string{"cc · idle", "pi · 10-alpha"} {
		if strings.Contains(graph, unwanted) {
			t.Errorf("graph has extra subtitle separator %q:\n%s", unwanted, graph)
		}
	}
}

func TestAgentsGraphDelegationStemJoinsPodBorder(t *testing.T) {
	m := agentsBoardModel(t)
	writeAgentRecord(t, m.home, records.Agent{
		Key: "manager-alpha", Name: "manager-alpha", Role: "manager", Project: "alpha",
		Harness: "claude", Model: "model-manager-alpha", PaneID: "w1:p3",
		Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-10-20T11:57:00Z",
	})
	m = refreshAgents(t, m)
	lines, positions := m.agents.graphLayout()
	if !strings.Contains(ansi.Strip(lines[4]), "┄ inferred") {
		t.Fatalf("inferred manager label is missing from its row: %s", ansi.Strip(lines[4]))
	}
	manager := positions["manager-alpha"]
	column := manager.col + manager.width/2
	topOfNode := manager.row - 1
	for row := 0; row <= topOfNode; row++ {
		want := '│'
		if row == 0 || row == 2 || row == topOfNode {
			want = '┬'
		}
		line := []rune(ansi.Strip(lines[row]))
		if got := line[column]; got != want {
			t.Fatalf("delegation stem at row %d = %q, want %q: %s", row, got, want, string(line))
		}
	}
}

func TestAgentsGraphNodeCardsStayInsetInLongTitlePods(t *testing.T) {
	m := agentsBoardModel(t)
	_, _, podWidth, cardWidth := m.agents.graphPod(m.agents.data.groups[1], m.agents.nodeMap())
	if gap := podWidth - cardWidth; gap < 6 {
		t.Fatalf("alpha node cards have only %d cells of total pod margin, want at least 6", gap)
	}
}

func TestAgentsGraphPodsHaveRoundedCorners(t *testing.T) {
	m := agentsBoardModel(t)
	pod, _, _, _ := m.agents.graphPod(m.agents.data.groups[0], m.agents.nodeMap())
	for _, test := range []struct {
		line int
		col  int
		want rune
	}{{line: 0, col: 0, want: '╭'}, {line: 0, col: -1, want: '╮'}, {line: -1, col: 0, want: '╰'}, {line: -1, col: -1, want: '╯'}} {
		line := test.line
		if line < 0 {
			line = len(pod) + line
		}
		column := test.col
		if column < 0 {
			column = len([]rune(pod[line])) + column
		}
		if got := []rune(pod[line])[column]; got != test.want {
			t.Errorf("pod corner at line %d column %d = %q, want %q: %s", line, column, got, test.want, pod[line])
		}
	}
}

func TestAgentsGraphNodeHasOneCellOfPadding(t *testing.T) {
	m := agentsBoardModel(t)
	m.agents.tab = agentsGraph
	lines, _ := m.agents.graphLines()
	for _, line := range lines {
		plain := ansi.Strip(line)
		name := "▶ ● coordinator"
		start := strings.Index(plain, name)
		if start < 0 {
			continue
		}
		runes := []rune(plain)
		runeStart := len([]rune(plain[:start]))
		if runeStart < 2 || runes[runeStart-1] != ' ' || runes[runeStart-2] != '│' {
			t.Fatalf("coordinator node has no left padding: %s", plain)
		}
		end := runeStart + len([]rune(name))
		rightBorder := end
		for rightBorder < len(runes) && runes[rightBorder] != '│' {
			rightBorder++
		}
		if rightBorder == len(runes) || rightBorder == end || runes[rightBorder-1] != ' ' {
			t.Fatalf("coordinator node has no right padding: %s", plain)
		}
		return
	}
	t.Fatal("graph has no coordinator node")
}

func TestAgentsHeaderUsesSpacesAndRightAlignsTime(t *testing.T) {
	m := agentsBoardModel(t)
	m.width = 120
	m.now = func() time.Time { return time.Date(2026, 10, 20, 12, 0, 0, 0, time.Local) }
	line := ansi.Strip(m.agentsHeaderLine())
	if !strings.HasPrefix(line, "⬢ hive  2 projects  5 agents  2 working  3 idle  supervisor ●  needs you 0") {
		t.Fatalf("Agents header groups are not separated by spaces: %s", line)
	}
	if ansi.StringWidth(line) != 120 || !strings.HasSuffix(line, "12:00") {
		t.Fatalf("Agents header time is not at the right edge: width=%d line=%q", ansi.StringWidth(line), line)
	}
}

func TestAgentsDetailFieldsUseSpaces(t *testing.T) {
	m := agentsBoardModel(t)
	var worker agentNode
	for _, node := range m.agents.data.nodes {
		if node.record.Key == "worker-alpha" {
			worker = node
			break
		}
	}
	want := "worker-alpha  parent manager-alpha  model model-alpha  task 10-alpha  assigned  pane w1:p4  next 11-next"
	if got := m.agents.detailLine(worker); got != want {
		t.Fatalf("Agents detail = %q, want %q", got, want)
	}
	strip := strings.Join(agentDetailStrip(want, 120), "\n")
	if !strings.Contains(strip, "worker-alpha  parent manager-alpha  model model-alpha") {
		t.Fatalf("Agents detail strip does not keep space separators:\n%s", strip)
	}
}

func agentsBoardModel(t *testing.T) *Model {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", PID: int64(os.Getpid()), PIDStart: pidStart})
	for _, project := range []records.Project{
		{ID: "alpha", Path: "/repo/alpha", Base: "dev"},
		{ID: "beta", Path: "/repo/beta", Base: "dev"},
	} {
		writeRecord(t, home, []string{"projects"}, project.ID, project)
	}
	for _, goal := range []records.Goal{
		{ID: "g1-alpha", Project: "alpha", Title: "Alpha goal for long graph", Status: "active"},
		{ID: "g2-beta", Project: "beta", Title: "Beta goal for long graph", Status: "active"},
	} {
		writeRecord(t, home, []string{"goals"}, goal.ID, goal)
	}
	for _, task := range []records.Task{
		{ID: "10-alpha", Project: "alpha", Title: "Alpha task", Status: "assigned"},
		{ID: "11-next", Project: "alpha", Title: "Next task", Status: "created", DependsOn: []string{"10-alpha"}},
		{ID: "20-ghost", Project: "beta", Title: "Ghost task", Status: "assigned"},
	} {
		writeRecord(t, home, []string{"tasks"}, task.ID, task)
	}
	for _, record := range []records.Agent{
		{Key: "coordinator", Name: "coordinator", Role: "coordinator", Harness: "claude", Model: "model-coordinator", PaneID: "w1:p1", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-20T11:59:00Z"},
		{Key: "supervisor", Name: "supervisor", Role: "supervisor", Harness: "pi", Model: "model-supervisor", PaneID: "w1:p2", Parent: "coordinator", Status: records.AgentLive, AgentStatus: "done", StatusChangedAt: "2026-10-20T11:58:00Z"},
		{Key: "manager-alpha", Name: "manager-alpha", Role: "manager", Project: "alpha", Harness: "claude", Model: "model-manager-alpha", PaneID: "w1:p3", Parent: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-10-20T11:57:00Z"},
		{Key: "worker-alpha", Name: "worker-alpha", Role: "worker", Project: "alpha", Task: "10-alpha", Harness: "pi", Model: "model-alpha", PaneID: "w1:p4", Parent: "manager-alpha", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: "2026-10-20T11:56:00Z"},
		{Key: "manager-beta", Name: "manager-beta", Role: "manager", Project: "beta", Harness: "claude", Model: "model-manager-beta", PaneID: "w1:p5", Parent: "coordinator", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: "2026-10-20T11:55:00Z"},
		{Key: "ghost-beta", Name: "ghost-beta", Role: "worker", Project: "beta", Task: "20-ghost", Harness: "pi", Model: "model-ghost", PaneID: "w1:p6", Parent: "manager-beta", Status: records.AgentStopped, StatusChangedAt: "2026-10-20T11:54:00Z"},
	} {
		writeAgentRecord(t, home, record)
	}
	m := newTestModel(t, home, clock)
	m = loadHeader(t, m)
	m.screenIndex = 1
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Agents screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	return m
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
