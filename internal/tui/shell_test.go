package tui

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

// TestScreenKeysSelectAllSixScreens presses the chords of golden-frames
// decision 16, esc, and [ ] (decisions 3, 22). Each tab shows the current
// screen of its view until wave 6. The title row of decision 2 names the
// screen, and Chat has no title row.
func TestScreenKeysSelectAllSixScreens(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	for _, step := range []struct {
		key    tea.KeyPressMsg
		tab    int
		screen string
		shows  string
	}{
		{ctrl('g'), tabAgents, "Agents", "no project agents"},
		{escKey(), tabChat, "Chat", "type a message to the coordinator."},
		{ctrl('t'), tabWork, "Goals", "goals · 0"},
		{key(']', "]"), tabWork, "Queue", "no tasks"},
		{key(']', "]"), tabWork, "Tasks", "tasks · 0"},
		{ctrl('f'), tabFriction, "Friction", "open · 0"},
		{ctrl('y'), tabHealth, "Health", "↑↓ ←→ tile"},
		{ctrl(']'), tabHealth, "Health", "agents ·"},
		{ctrl('l'), tabLogs, "Logs", "events · 0"},
	} {
		m, _ = raise(t, m, step.key)
		view := ansi.Strip(m.View().Content)
		if m.activeTab() != step.tab || m.legacyName() != step.screen {
			t.Fatalf("key %q gave tab %d screen %s; want tab %d screen %s", step.key, m.activeTab(), m.legacyName(), step.tab, step.screen)
		}
		title := " " + strings.ToLower(screenRegistry[step.tab].label)
		titleRow := strings.Split(view, "\n")[0]
		if step.tab == tabChat {
			// Golden-frames chat.md rule 8: the welcome block from row 1.
			title, titleRow = " hive · ", strings.Split(view, "\n")[1]
		}
		if !strings.HasPrefix(titleRow, title) || !strings.Contains(view, step.shows) {
			t.Fatalf("key %q view omits %q or %q:\n%s", step.key, title, step.shows, view)
		}
	}
}

// TestNarrowShellFitsAndShowsTheTabBar checks the frame at 80×24
// (decisions 1, 3) and no sidebar. Golden-frames decisions 2, 3, and 10: no
// header and no tab bar, and the status row drops whole segments from the
// right to fit.
func TestNarrowShellFitsAndShowsTheTabBar(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View().Content
	if got := lipgloss.Height(view); got != 24 {
		t.Fatalf("rendered shell height %d; want 24:\n%s", got, view)
	}
	plain := ansi.Strip(view)
	// The welcome block of golden-frames chat.md rule 8 has hive · HOME on
	// row 1; the header was on row 0.
	if strings.Contains(strings.Split(plain, "\n")[0], " hive ") || strings.Contains(plain, "1 Agents") {
		t.Fatalf("narrow shell shows the header or the tab bar:\n%s", plain)
	}
	if rows := strings.Split(plain, "\n"); !strings.HasPrefix(rows[23], " ○ none · ○ supervisor · ? 0 asks · needs you 0 ") {
		t.Fatalf("narrow status row = %q", rows[23])
	}
	if strings.Contains(plain, "SCREENS") {
		t.Fatalf("narrow shell shows the old sidebar:\n%s", plain)
	}
	for _, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != 80 {
			t.Fatalf("rendered line width %d; want 80: %q", got, line)
		}
	}
}

func TestSlowHeaderLoadDoesNotBlockKeysOrStartTwice(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseLoad := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseLoad()
	var loads atomic.Int32
	m.headerLoader = func(string) (headerState, error) {
		started <- struct{}{}
		if loads.Add(1) == 1 {
			<-release
		}
		return headerState{}, nil
	}
	m.logLoader = func(*logs.Reader, time.Time) ([]logs.Event, error) { return nil, nil }
	cmd := m.startHeaderLoad()
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("header load did not start")
	}
	m, keyCmd := raise(t, m, ctrl('g'))
	m = deliverAgentCommands(t, m, keyCmd)
	if !strings.Contains(ansi.Strip(m.View().Content), " agents · ") {
		t.Fatalf("key did not update while header load was blocked:\n%s", m.View().Content)
	}
	m.screenIndex = 7
	_, refresh := raise(t, m, tickMsg{})
	runRefreshLoads(t, refresh)
	select {
	case <-started:
		t.Fatal("refresh started a second header load before the first settled")
	default:
	}
	releaseLoad()
	msg := <-done
	m, _ = raise(t, m, msg)
	if m.headerLoading {
		t.Fatal("header load did not settle")
	}
}

func TestSlowScreenLoadDoesNotBlockKeysOrStartTwice(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m.screenIndex = 7
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseLoad := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseLoad()
	var loads atomic.Int32
	m.logLoader = func(*logs.Reader, time.Time) ([]logs.Event, error) {
		started <- struct{}{}
		if loads.Add(1) == 1 {
			<-release
		}
		return nil, nil
	}
	cmd := m.startCurrentScreenLoad()
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("screen load did not start")
	}
	m, _ = raise(t, m, ctrl('g'))
	if !strings.Contains(ansi.Strip(m.View().Content), " agents · ") {
		t.Fatalf("key did not update while screen load was blocked:\n%s", m.View().Content)
	}
	m, _ = raise(t, m, ctrl('l'))
	_, refresh := raise(t, m, tickMsg{})
	runRefreshLoads(t, refresh)
	select {
	case <-started:
		t.Fatal("refresh started a second screen load before the first settled")
	default:
	}
	releaseLoad()
	msg := <-done
	m, _ = raise(t, m, msg)
	if m.screenLoading {
		t.Fatal("screen load did not settle")
	}
}

func runRefreshLoads(t *testing.T, refresh tea.Cmd) {
	t.Helper()
	msg := refresh()
	commands, ok := msg.(tea.BatchMsg)
	if !ok || len(commands) < 2 {
		t.Fatalf("refresh command returned %T with no load command", msg)
	}
	// The first command schedules the next tick. Run the data commands now.
	for _, command := range commands[1:] {
		command()
	}
}

func TestQuitNeedsTwoCtrlCWithinThreeSeconds(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	var cmd tea.Cmd
	m, cmd = raise(t, m, ctrlC())
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("first ctrl+c quit the TUI")
		}
	}
	requireView(t, m, "press ctrl+c again to quit")
	clock.set(clock.now().Add(2 * time.Second))
	_, cmd = raise(t, m, ctrlC())
	quitMessage(t, cmd)
}

func TestQuitNoticeExpiresAndOtherKeyClearsIt(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	firstPress := clock.now()
	m, cmd := raise(t, m, ctrlC())
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("first ctrl+c quit the TUI")
		}
	}
	clock.set(firstPress.Add(4 * time.Second))
	m, _ = raise(t, m, quitNoticeExpiredMsg{at: firstPress})
	rejectView(t, m, "press ctrl+c again to quit")
	m, _ = raise(t, m, ctrlC())
	requireView(t, m, "press ctrl+c again to quit")
	m, _ = raise(t, m, ctrl('g'))
	rejectView(t, m, "press ctrl+c again to quit")
}

func TestQDoesNotQuitOnAnyScreen(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	for _, screen := range []rune{'1', '8'} {
		m, _ = raise(t, m, key(screen, string(screen)))
		_, cmd := raise(t, m, key('q', "q"))
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Fatalf("q quit screen %q", screen)
			}
		}
	}
}

func TestAttachUsesBackgroundFocusAndReportsError(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	cmd := m.attachCmd("w1:pA")
	if cmd == nil {
		t.Fatal("attach returned no command")
	}
	m, _ = raise(t, m, cmd())
	requireView(t, m, "error:")
}

func TestRunHomeHandlesScreenAndQuitKeys(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader("\x07\x0c\x03\x03"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("RunHome did not quit after the second ctrl+c")
	}
	// Golden-frames decision 2: the title row names the screen. The
	// renderer draws only the last frame of the fast input in full.
	if !strings.Contains(ansi.Strip(out.String()), " logs") {
		t.Fatalf("program output omitted selected screens:\n%s", out.String())
	}
}

func TestEightScreensDoNotPaintTerminalBackground(t *testing.T) {
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, newTestHome(t), clock)
	fixtureBackgroundScreens(m, clock.now())
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})

	for index, name := range []string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"} {
		t.Run(name, func(t *testing.T) {
			m.screenIndex = index
			output := m.View().Content
			assertNoPaintedBackgroundSGR(t, output)
			if !strings.Contains(output, "38;2;245;185;66") {
				t.Errorf("%s tab bar omits the honey foreground color:\n%s", name, output)
			}
		})
	}

	m.screenIndex = 0
	m.chatInputActive = true
	for _, draft := range []string{"", "selection marker draft"} {
		t.Run("Chat input "+draft, func(t *testing.T) {
			m.chatInput.Reset()
			m.chatInput.InsertString(draft)
			m.chatInput.Focus()
			m.invalidateView()
			view := m.View()
			assertNoPaintedBackgroundSGR(t, view.Content)
			// Golden-frames decision 19: the status row shows the chords.
			if !strings.Contains(ansi.Strip(view.Content), "^g agents  ^t work") {
				t.Error("active Chat status row omits the chords")
			}
			if view.Cursor == nil {
				t.Fatal("active Chat input omits the real terminal cursor")
			}
			// The composer draws the cursor cell █ after the text, and the
			// real terminal cursor sits on it. On the empty line the
			// placeholder follows "> " and the cursor sits on its first cell
			// (golden-frames chat.md rule 10).
			plain := ansi.Strip(view.Content)
			wantText, mark := "> "+draft+"█", "█"
			if draft == "" {
				wantText, mark = "> no coordinator", "no coordinator"
			}
			textIndex := strings.Index(plain, wantText)
			if textIndex < 0 {
				t.Fatalf("active Chat input omits %q:\n%s", wantText, plain)
			}
			prefix := plain[:strings.Index(plain[textIndex:], mark)+textIndex]
			lineStart := strings.LastIndex(prefix, "\n") + 1
			want := tea.Position{X: ansi.StringWidth(prefix[lineStart:]), Y: strings.Count(prefix, "\n")}
			if got := view.Cursor.Position; got != want {
				t.Errorf("active Chat cursor position = %+v, want text position %+v", got, want)
			}
			if draft != "" {
				lines := strings.Split(plain, "\n")
				row := strings.Count(plain[:strings.Index(plain, draft)], "\n")
				if row < 1 || !strings.Contains(lines[row-1], "╭") {
					t.Errorf("Chat composer omits its top border:\n%s", plain)
				}
				if row+1 >= len(lines) || !strings.Contains(lines[row+1], "╰") {
					t.Errorf("Chat composer omits its bottom border:\n%s", plain)
				}
			}
		})
	}
}

func fixtureBackgroundScreens(m *Model, now time.Time) {
	m.chatData = chatScreenData{
		entries: []chatlib.Entry{{ID: "c1", At: now.Format(time.RFC3339Nano), From: "coordinator", Text: "fixture chat"}},
	}
	m.rebuildChat()

	node := agentNode{
		record: records.Agent{Key: "coordinator", Name: "coordinator", Role: "coordinator", Status: records.AgentLive},
		status: "working", age: "now",
	}
	m.agents.accept(agentsData{
		nodes: []agentNode{node}, groups: []agentGroup{{id: "home", title: "home", keys: []string{"coordinator"}}},
		supervisorLive: true,
	})

	goal := records.Goal{
		ID: "g1-fixture", Project: "hive", Title: "Fixture goal", Grant: "fixture grant", Status: "active",
		Waves: []records.GoalWave{{ID: "wave-a", Title: "Fixture wave", DoneWhen: "dev", Tasks: []string{"100-fixture"}, Status: "active"}},
	}
	task := records.Task{ID: "100-fixture", Project: "hive", Title: "Fixture task", Status: "assigned", DependsOn: []string{}}
	m.goalsQueue.accept(goalsQueueData{goals: []records.Goal{goal}, tasks: []records.Task{task}})
	m.tasks.accept(tasksData{items: []taskItem{{record: task}}})
	m.friction.accept(frictionData{entries: []frictionEntry{{
		file: "fixture.md", path: "entries/fixture.md", title: "Fixture friction", status: "Open",
		opened: "2026-10-20", body: "Fixture summary.", valid: true,
		sections: []frictionSection{{title: "Summary", body: []string{"Fixture summary."}}},
	}}})
	m.health.accept(doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Required: true, Detail: "fixture home"},
		{Name: "harness fixture", Status: doctor.StatusOK, Detail: "fixture agent"},
	}})

	m.acceptLogEvents([]logs.Event{{
		Raw:    []byte(`{"time":"2026-10-20T12:00:00Z","level":"INFO","msg":"fixture log"}`),
		Fields: map[string]any{"time": "2026-10-20T12:00:00Z", "level": "INFO", "msg": "fixture log"},
	}})
}

func assertNoPaintedBackgroundSGR(t *testing.T, output string) {
	t.Helper()
	var background string
	for index := 0; index < len(output); {
		if output[index] != '\x1b' || index+1 >= len(output) || output[index+1] != '[' {
			index++
			continue
		}
		end := index + 2
		for end < len(output) && (output[end] < 0x40 || output[end] > 0x7e) {
			end++
		}
		if end == len(output) {
			t.Errorf("unterminated ANSI sequence at byte %d", index)
			return
		}
		if background == "" && output[end] == 'm' && sgrSetsBackground(output[index+2:end]) {
			background = output[index : end+1]
		}
		index = end + 1
	}
	if background != "" {
		t.Errorf("SGR sets a terminal background: %q", background)
	}
}

func sgrSetsBackground(parameters string) bool {
	parts := strings.Split(parameters, ";")
	for index := 0; index < len(parts); index++ {
		parameter := parts[index]
		if parameter == "" {
			continue
		}
		code, err := strconv.Atoi(parameter)
		if err != nil {
			continue
		}
		if code == 48 || code >= 40 && code <= 47 || code >= 100 && code <= 107 || code == 7 {
			return true
		}
		if code != 38 || index+1 >= len(parts) {
			continue
		}
		switch parts[index+1] {
		case "2":
			index += 4
		case "5":
			index += 2
		}
	}
	return false
}
