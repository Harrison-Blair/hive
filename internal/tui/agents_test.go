package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

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
	var current []string
	for _, node := range data.nodes {
		if node.current {
			current = append(current, node.record.Key)
		}
	}
	if len(current) != 1 || current[0] != "legacy-c" {
		t.Fatalf("current legacy generations = %q, want only record legacy-c", current)
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

func TestAgentsPanelFitsTerminalFrameAtStandardAndLargeSizes(t *testing.T) {
	for _, size := range []struct {
		width  int
		height int
	}{{width: 80, height: 24}, {width: 120, height: 40}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			home, clock := agentsFixture(t, 40)
			m := newTestModel(t, home, clock)
			m.width, m.height = size.width, size.height
			m, cmd := raise(t, m, ctrl('g'))
			m = deliverAgentCommands(t, m, cmd)

			// Decisions 1 and 3: the frame has no sidebar. The Agents panel
			// fills the content area, and the view is exactly the pane.
			lines := strings.Split(ansi.Strip(m.View().Content), "\n")
			if len(lines) != size.height {
				t.Fatalf("view has %d lines; want %d", len(lines), size.height)
			}
			for index, line := range lines {
				if got := ansi.StringWidth(line); got != size.width {
					t.Fatalf("line %d is %d cells; want %d:\n%s", index, got, size.width, strings.Join(lines, "\n"))
				}
			}
			// Golden-frames decision 2: the title row and the rule are above
			// the body. Golden-frames agents.md rule 1: the body has no
			// outer box.
			if !strings.HasPrefix(lines[0], " agents · ") || strings.Contains(strings.Join(lines, "\n"), "project agents") {
				t.Fatalf("Agents body is not below the title row with no outer box:\n%s", strings.Join(lines, "\n"))
			}
		})
	}
}

func TestAgentsPeekReadsInCommandAndEscClosesIt(t *testing.T) {
	home, clock := agentsFixture(t, 0)
	m := newTestModel(t, home, clock)
	// At 120×40 the peek panel has room for each of the 20 lines.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, cmd := raise(t, m, ctrl('g'))
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
	// Decision 14: p peeks at the pane of the selected agent. The
	// coordinator is not on the graph, so the manager is selected.
	_, cmd = raise(t, m, key('p', "p"))
	if cmd == nil {
		t.Fatal("p did not start pane read")
	}
	m = deliverAgentCommands(t, m, cmd)
	requireView(t, m, "peek w1:p3 · esc closes", "line 06", "line 25")
	if view := ansi.Strip(m.View().Content); strings.Contains(view, "line 05") || strings.Contains(view, "line 01") {
		t.Fatalf("peek shows more than the last 20 lines:\n%s", view)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, argument := range []string{"pane", "read", "w1:p3", "--lines", "20", "--format", "text"} {
		if !strings.Contains(string(args), argument) {
			t.Fatalf("herdr pane read args omitted %q: %s", argument, args)
		}
	}
	m, cmd = raise(t, m, key(tea.KeyEscape, "esc"))
	m = deliverAgentCommands(t, m, cmd)
	if m.agents.peekOpen {
		t.Fatal("esc did not close the peek")
	}
	rejectView(t, m, "peek w1:p3 · esc closes")
	requireView(t, m, " agents · ", "manager-demo")
	logData, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent peek"`, `"action":"agent peek close"`, `"pane_id":"w1:p3"`} {
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
	// Golden-frames decision 11: the status row shows the state of the
	// supervisor and needs you. The header state keeps the agent counts.
	check := func(screen string) {
		frame := ansi.Strip(m.View().Content)
		for _, want := range []string{"● supervisor", "needs you 0"} {
			if !strings.Contains(frame, want) {
				t.Errorf("%s header omits %q:\n%s", screen, want, frame)
			}
		}
		if m.header.agents != 5 || m.header.working != 2 || m.header.idle != 3 {
			t.Errorf("header counts = %d agents, %d working, %d idle; want 5, 2, 3", m.header.agents, m.header.working, m.header.idle)
		}
		for _, unwanted := range []string{"supervisor fresh", "needs_user"} {
			if strings.Contains(frame, unwanted) {
				t.Errorf("%s header contains %q:\n%s", screen, unwanted, frame)
			}
		}
	}
	// Ask a27: the default row of Agents has no supervisor and no needs
	// you segment, so the check names them.
	cfg := home.DefaultTUI()
	cfg.StatusLines = [][]string{{"state", "supervisor", "needs_you", "clock"}}
	m.setTUIConfig(cfg)
	check("Agents")
	// Decision 34: only Chat has these segments in the default row.
	m.setTUIConfig(home.DefaultTUI())
	m.selectScreen(tabChat, 0)
	check("Chat")
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
	for _, node := range m.agents.data.nodes {
		if node.current && node.record.Key == "supervisor-g16" {
			supervisor = node
			break
		}
	}
	if supervisor.record.Key == "" {
		t.Fatal("live supervisor generation 16 is not current")
	}
	if supervisor.status != "working" || supervisor.age != "50m" {
		t.Fatalf("supervisor status/age = %q/%q, want working/50m", supervisor.status, supervisor.age)
	}
	// Health › Agents shows the pid in the details of the supervisor.
	if want := fmt.Sprint(os.Getpid()); supervisor.pid != want {
		t.Fatalf("supervisor pid = %q, want the claim pid %s", supervisor.pid, want)
	}
}

func TestAgentsBoardTaskShortcutOpensSelectedTask(t *testing.T) {
	m := agentsBoardModel(t)
	// agents.md rule 12: o opens the pod, and the arrows move in it.
	m, _ = raise(t, m, key('o', "o"))
	for range 3 {
		m, _ = raise(t, m, key('j', "j"))
	}
	m.screenLoading = true
	m, cmd := raise(t, m, key('t', "t"))
	if m.screenIndex != 2 {
		t.Fatalf("screen index = %d, want Work screen index 2", m.screenIndex)
	}
	if m.tasks.selectedID != "10-alpha" {
		t.Fatalf("Tasks selection = %q, want task 10-alpha", m.tasks.selectedID)
	}
	if cmd == nil {
		t.Fatal("t did not load the Tasks screen")
	}
	m = deliverAgentCommands(t, m, cmd)
	// Work rule 4: the selected card has the heavy border.
	if got := selectedCard(t, m); got != "10-alpha" {
		t.Fatalf("Work board selected %q, want task 10-alpha", got)
	}
}

func TestAgentsBoardKeyHintsMatchTheBoard(t *testing.T) {
	m := agentsBoardModel(t)
	want := "←↑↓→ move · tab view · enter peek · a attach · h stopped · t task · r restart coordinator/supervisor · ask in Chat"
	if got := m.currentScreen().KeyHints(); got != want {
		t.Fatalf("Agents key hints = %q, want %q", got, want)
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
	m.selectScreen(tabAgents, 0)
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
	// Decision 10: with no project, Agents shows the notice. Golden-frames
	// decision 16: ctrl+g, the byte 0x07, opens Agents.
	if err := RunHome(home, strings.NewReader("\x07\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	if !strings.Contains(out.String(), agentsNoneText) {
		t.Fatalf("program output did not render the Agents notice:\n%s", out.String())
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
