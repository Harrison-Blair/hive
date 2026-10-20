package tui

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
)

func TestScreenKeysSelectAllEightScreens(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	for index, name := range []string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"} {
		number := rune('1' + index)
		m, _ = raise(t, m, key(number, string(number)))
		if !strings.Contains(m.View().Content, "▶ "+string(number)+" "+name) {
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
		} else if name == "Goals" && !strings.Contains(m.View().Content, "unassigned") {
			t.Fatalf("Goals screen does not show its unassigned panel:\n%s", m.View().Content)
		} else if name == "Queue" && !strings.Contains(m.View().Content, "no open tasks") {
			t.Fatalf("Queue screen does not show its empty state:\n%s", m.View().Content)
		} else if name != "Goals" && name != "Queue" && name != "Logs" && !strings.Contains(m.View().Content, "next wave") {
			t.Fatalf("%s screen does not show its placeholder:\n%s", name, m.View().Content)
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
	if !strings.Contains(m.View().Content, "▶ 2 Agents") {
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
	if !strings.Contains(m.View().Content, "▶ 2 Agents") {
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
