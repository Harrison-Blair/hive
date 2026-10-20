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

func TestScreenKeysSelectAllEightScreens(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	for index, name := range []string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"} {
		number := rune('1' + index)
		m, _ = raise(t, m, key(number, string(number)))
		selected := "▶ " + string(number) + " " + name
		if name == "Agents" {
			selected = "AGENTS"
		}
		if !strings.Contains(m.View().Content, selected) {
			t.Fatalf("key %q did not select %s:\n%s", number, name, m.View().Content)
		}
		if name == "Chat" {
			if !strings.Contains(m.View().Content, "coordinator") {
				t.Fatalf("Chat screen does not show the coordinator:\n%s", m.View().Content)
			}
		} else if name == "Agents" {
			if !strings.Contains(m.View().Content, "[ TREE ]") {
				t.Fatalf("Agents screen does not show its tree tab:\n%s", m.View().Content)
			}
		} else if name == "Health" {
			if !strings.Contains(m.View().Content, "home") || !strings.Contains(m.View().Content, "agents") {
				t.Fatalf("Health screen does not show its panels:\n%s", m.View().Content)
			}
		} else if name == "Tasks" {
			if !strings.Contains(m.View().Content, "active 0") {
				t.Fatalf("Tasks screen does not show its task tabs:\n%s", m.View().Content)
			}
		} else if name == "Friction" {
			if !strings.Contains(m.View().Content, "open 0") {
				t.Fatalf("Friction screen does not show its entry tabs:\n%s", m.View().Content)
			}
		} else if name == "Goals" && !strings.Contains(m.View().Content, "unassigned") {
			t.Fatalf("Goals screen does not show its unassigned panel:\n%s", m.View().Content)
		} else if name == "Queue" && !strings.Contains(m.View().Content, "no open tasks") {
			t.Fatalf("Queue screen does not show its empty state:\n%s", m.View().Content)
		}
	}
}

func TestNarrowShellFitsAndShowsCompactSidebar(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View().Content
	if got := lipgloss.Height(view); got > 24 {
		t.Fatalf("rendered shell height %d exceeds 24:\n%s", got, view)
	}
	for _, text := range []string{"hive", "SCREENS", "▶ 1 Chat"} {
		if !strings.Contains(view, text) {
			t.Fatalf("narrow shell omitted %q:\n%s", text, view)
		}
	}
	for _, name := range []string{"1 Chat", "2 Agents", "3 Goals", "4 Queue", "5 Tasks", "6 Friction", "7 Health", "8 Logs"} {
		if !strings.Contains(view, name) {
			t.Fatalf("compact sidebar does not show %q:\n%s", name, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got > 80 {
			t.Fatalf("rendered line width %d exceeds 80: %q", got, line)
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
	m, keyCmd := raise(t, m, key('2', "2"))
	m = deliverAgentCommands(t, m, keyCmd)
	if !strings.Contains(m.View().Content, "AGENTS") {
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
	m, _ = raise(t, m, key('2', "2"))
	if !strings.Contains(m.View().Content, "AGENTS") {
		t.Fatalf("key did not update while screen load was blocked:\n%s", m.View().Content)
	}
	m, _ = raise(t, m, key('8', "8"))
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
	m, _ = raise(t, m, key('2', "2"))
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
	go func() { done <- RunHome(home, strings.NewReader("28\x03\x03"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("RunHome did not quit after the second ctrl+c")
	}
	if !strings.Contains(out.String(), "Agents") || !strings.Contains(out.String(), "Logs") {
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
			m.layout()
			output := m.View().Content
			assertNoPaintedBackgroundSGR(t, output)
			if !strings.Contains(output, "38;2;245;185;66") {
				t.Errorf("%s sidebar omits the honey foreground color:\n%s", name, output)
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
			if !strings.Contains(view.Content, "esc keep draft") {
				t.Error("active Chat footer omits its input controls")
			}
			if view.Cursor == nil {
				t.Fatal("active Chat input omits the real terminal cursor")
			}
			plain := ansi.Strip(view.Content)
			wantText := "Type a message"
			cursorOffset := 0
			if draft != "" {
				wantText = draft
				cursorOffset = len(draft)
			}
			textIndex := strings.Index(plain, wantText)
			if textIndex < 0 {
				t.Fatalf("active Chat input omits %q", wantText)
			}
			prefix := plain[:textIndex+cursorOffset]
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
				if row+4 >= len(lines) || !strings.Contains(lines[row+4], "╰") {
					t.Errorf("Chat composer omits its bottom border:\n%s", plain)
				}
			}
		})
	}
}

func fixtureBackgroundScreens(m *Model, now time.Time) {
	m.chatData = chatScreenData{
		entries: []chatlib.Entry{{ID: "c1", At: now.Format(time.RFC3339Nano), From: "coordinator", Text: "fixture chat"}},
		coordinator: records.Agent{
			Name: "coordinator", Harness: "fixture", Model: "fixture", PaneID: "fixture-pane",
			Status: records.AgentLive, AgentStatus: "working",
		},
		coordinatorFound: true,
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
	m.tasks.accept(tasksData{items: []taskItem{{record: task}}, projects: []string{"hive"}})
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
