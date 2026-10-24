package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

type testClock struct {
	nanos atomic.Int64
}

func (c *testClock) set(when time.Time) { c.nanos.Store(when.UnixNano()) }

func (c *testClock) now() time.Time { return time.Unix(0, c.nanos.Load()).In(time.Local) }

func fixedClock(t *testing.T, when time.Time) *testClock {
	t.Helper()
	clock := &testClock{}
	clock.set(when)
	return clock
}

func newTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func newTestModel(t *testing.T, home string, clock *testClock) *Model {
	t.Helper()
	model := New(home, clock.now)
	model.chatCallerCwd = home
	model.chatPaneID = ""
	model.chatCallerBlocked = false
	model.chatErr = ""
	// The test runs below projects/, so New blocked the caller and gave the
	// input no focus. Golden-frames decision 19 gives the input the focus.
	model.chatInputActive = true
	model.chatInput.Focus()
	t.Cleanup(model.Stop)
	return model
}

func newTestLogsModel(t *testing.T, home string, clock *testClock) *Model {
	t.Helper()
	m := newTestModel(t, home, clock)
	m.screenIndex = 7
	return loadLogs(t, m)
}

func loadHeader(t *testing.T, m *Model) *Model {
	t.Helper()
	cmd := m.startHeaderLoad()
	if cmd == nil {
		t.Fatal("header load did not start")
	}
	m, _ = raise(t, m, cmd())
	return m
}

func loadLogs(t *testing.T, m *Model) *Model {
	t.Helper()
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("screen data load did not start")
	}
	m, _ = raise(t, m, cmd())
	return m
}

func dailyPath(home string, day time.Time) string {
	return filepath.Join(home, ".hive", "logs", day.Format("2006-01-02")+".jsonl")
}

func eventJSON(t *testing.T, when time.Time, level, message string) string {
	t.Helper()
	event := map[string]any{
		"time":  when.Format(time.RFC3339Nano),
		"level": level,
		"msg":   message,
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeEvents(t *testing.T, path string, lines ...string) {
	t.Helper()
	data := strings.Join(lines, "\n")
	if len(lines) != 0 {
		data += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendEvent(t *testing.T, path, line string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(line + "\n"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func appendRaw(t *testing.T, path, text string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeRecord(t *testing.T, home string, collection []string, id string, value any) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn(collection, id, value); err != nil {
		t.Fatal(err)
	}
}

func raise(t *testing.T, m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	updated, ok := next.(*Model)
	if !ok {
		t.Fatalf("Update returned %T; want *Model", next)
	}
	return updated, cmd
}

func key(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func ctrlC() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
}

func requireView(t *testing.T, m *Model, wants ...string) {
	t.Helper()
	view := ansi.Strip(m.View().Content)
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Fatalf("view does not contain %q:\n%s", want, view)
		}
	}
}

func rejectView(t *testing.T, m *Model, unwanted ...string) {
	t.Helper()
	view := ansi.Strip(m.View().Content)
	for _, item := range unwanted {
		if strings.Contains(view, item) {
			t.Fatalf("view contains %q:\n%s", item, view)
		}
	}
}

func quitMessage(t *testing.T, cmd tea.Cmd) tea.QuitMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("got no command; want quit")
	}
	msg := cmd()
	quit, ok := msg.(tea.QuitMsg)
	if !ok {
		t.Fatalf("got %T; want tea.QuitMsg", msg)
	}
	return quit
}

func TestHeaderShowsProjectAgentAndNeedsUserCounts(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m = loadHeader(t, m)
	// Golden-frames decision 11: the status row shows the supervisor and
	// needs you. The header state keeps the counts.
	requireView(t, m, "○ supervisor", "needs you 0", "12:00")
	requireHeaderCounts(t, m.header, 0, 0, 0, 0)

	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 1, PID: int64(os.Getpid()), PIDStart: pidStart,
	})
	writeRecord(t, home, []string{"projects"}, "demo", records.Project{ID: "demo", Path: "/repo/demo", Base: "dev"})
	writeRecord(t, home, []string{"agents"}, "agent-one", records.Agent{
		Key: "agent-one", Name: "agent-one", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive, AgentStatus: "working",
	})
	writeRecord(t, home, []string{"agents"}, "agent-two", records.Agent{
		Key: "agent-two", Name: "agent-two", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive, AgentStatus: "done",
	})
	writeRecord(t, home, []string{"agents"}, "agent-starting", records.Agent{
		Key: "agent-starting", Name: "agent-starting", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentStarting,
	})
	writeRecord(t, home, []string{"tasks"}, "1-first", records.Task{ID: "1-first", Project: "demo", Title: "first", NeedsUser: "choose"})
	m = loadHeader(t, m)
	requireView(t, m, "● supervisor", "needs you 1")
	requireHeaderCounts(t, m.header, 1, 2, 1, 1)

	writeRecord(t, home, []string{"agents"}, "agent-two", records.Agent{
		Key: "agent-two", Name: "agent-two", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive, AgentStatus: "idle",
	})
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 2, PID: 0,
	})
	m = loadHeader(t, m)
	requireView(t, m, "○ supervisor", "needs you 1")
	requireHeaderCounts(t, m.header, 1, 2, 0, 0)
}

func requireHeaderCounts(t *testing.T, state headerState, projects, agents, working, idle int) {
	t.Helper()
	if state.projects != projects || state.agents != agents || state.working != working || state.idle != idle {
		t.Fatalf("header counts = %d projects, %d agents, %d working, %d idle; want %d, %d, %d, %d",
			state.projects, state.agents, state.working, state.idle, projects, agents, working, idle)
	}
}

func TestFramesMatchTerminalSizesOnEveryScreen(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{width: 80, height: 24},
		{width: 120, height: 40},
		{width: 200, height: 57},
		{width: 207, height: 57},
		{width: 103, height: 57},
		{width: 57, height: 20},
		{width: 300, height: 80},
	}
	screens := []string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"}

	for screenIndex, screen := range screens {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%s/%dx%d", screen, size.width, size.height), func(t *testing.T) {
				home := newTestHome(t)
				clock := fixedClock(t, time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local))
				m := newTestModel(t, home, clock)
				m.screenIndex = screenIndex
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})

				lines := strings.Split(ansi.Strip(m.View().Content), "\n")
				if size.width < minimumWidth || size.height < minimumHeight {
					// Decision 7: below 80×24 the view is the too-small notice.
					if !strings.Contains(strings.Join(lines, "\n"), "terminal too small") {
						t.Fatalf("%s view below 80×24 is not the too-small notice", screen)
					}
					return
				}
				if len(lines) != size.height {
					t.Errorf("%s frame has %d lines, want %d", screen, len(lines), size.height)
				}
				for lineIndex, line := range lines {
					if got := ansi.StringWidth(line); got != size.width {
						t.Fatalf("%s frame line %d has display width %d, want %d: %q", screen, lineIndex+1, got, size.width, line)
					}
				}
				foundBorder := false
				for _, line := range lines {
					if !strings.Contains(line, "╭") || !strings.Contains(line, "╮") {
						continue
					}
					foundBorder = true
					if got := ansi.StringWidth(strings.TrimRight(line, " ")); got != size.width {
						t.Errorf("%s panel right border ends at display column %d, want %d: %q", screen, got, size.width, line)
					}
					break
				}
				if !foundBorder {
					t.Errorf("%s frame has no top border", screen)
				}
			})
		}
	}
}

func TestHeaderInvalidStoreShowsBoundedError(t *testing.T) {
	home := newTestHome(t)
	if err := os.MkdirAll(filepath.Join(home, ".hive", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "state", "agents"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	clock := fixedClock(t, time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m = loadHeader(t, m)
	requireView(t, m, " hive ", "error:")
	rejectView(t, m, "0 agents")
}

// The filter bar takes each printable key while it is open (decision 21):
// enter applies the text, and esc closes the bar and keeps the applied
// filter. The typed field level: sets the lowest level (decision 25).
func TestKeysFilterLevelFollowAndQuit(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	writeEvents(t, dailyPath(home, day),
		eventJSON(t, day, "INFO", "alpha gq"),
		eventJSON(t, day, "INFO", "beta"),
		eventJSON(t, day, "WARN", "gamma gq"),
		eventJSON(t, day, "INFO", "delta"),
		eventJSON(t, day, "ERROR", "epsilon gq"),
		eventJSON(t, day, "DEBUG", "zeta"),
	)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	requireView(t, m, "/ level:info ", "alpha gq", "beta", "gamma gq", "epsilon gq")
	rejectView(t, m, "zeta")

	// Editing must not let viewing-mode keys act.
	m, _ = raise(t, m, key('/', "/"))
	requireKeys(t, m, "filter", "enter apply · esc close")
	for _, code := range []rune{'f', 'G'} {
		m, _ = raise(t, m, key(code, string(code)))
	}
	requireView(t, m, "/ level:infofG█")
	if !m.follow {
		t.Fatal("f in the filter bar turned follow off")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "/ level:info ", "beta")
	requireKeys(t, m, "events", "")

	// A q in editing mode is a printable character, not a quit.
	m, _ = raise(t, m, key('/', "/"))
	for _, code := range " gq" {
		var cmd tea.Cmd
		m, cmd = raise(t, m, key(code, string(code)))
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Fatal("q in editing mode quit the TUI")
			}
		}
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	requireView(t, m, "/ level:info gq ", "alpha gq", "gamma gq", "epsilon gq")
	rejectView(t, m, "beta", "delta", "zeta")

	// Escape in the event list clears the applied filter.
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "/ level:info ", "beta", "delta")
	rejectView(t, m, "zeta")

	// Backspace and a typed level field.
	m, _ = raise(t, m, key('/', "/"))
	for range "info" {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	for _, code := range "warn" {
		m, _ = raise(t, m, key(code, string(code)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	requireView(t, m, "/ level:warn ", "gamma gq", "epsilon gq")
	rejectView(t, m, "alpha gq", "beta", "delta", "zeta")

	m, cmd := raise(t, m, key('q', "q"))
	if cmd != nil {
		t.Fatal("q returned a command")
	}
	m, _ = raise(t, m, ctrlC())
	requireView(t, m, "press ctrl+c again to quit")
	_, cmd = raise(t, m, ctrlC())
	quitMessage(t, cmd)
}

func TestFollowReadsAppendedLineOnTick(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 19, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "first"))
	m := newTestLogsModel(t, home, clock)
	requireView(t, m, "first")

	appendEvent(t, path, eventJSON(t, day, "INFO", "second on tick"))
	m = loadLogs(t, m)
	requireView(t, m, "second on tick")
}

// Follow selects the newest event on each load. With follow off, a load
// keeps the selected event. g selects the first event, and G selects the
// newest event and turns follow on.
func TestFollowControlsSelection(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 7, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	lines := make([]string, 0, 40)
	for index := 1; index <= 40; index++ {
		lines = append(lines, eventJSON(t, day, "INFO", fmt.Sprintf("line %02d", index)))
	}
	path := dailyPath(home, day)
	writeEvents(t, path, lines...)
	m := newTestLogsModel(t, home, clock)
	// Decision 7: 80×24 is the smallest frame.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	requireView(t, m, "▶ 12:00:00 INFO  -         line 40")

	m, _ = raise(t, m, key('f', "f"))
	if m.follow {
		t.Fatal("f did not turn follow off")
	}
	appendEvent(t, path, eventJSON(t, day, "INFO", "line 41"))
	m = tickUntilCount(t, m, 41)
	requireView(t, m, "▶ 12:00:00 INFO  -         line 40")

	m, _ = raise(t, m, key('g', "g"))
	requireView(t, m, "▶ 12:00:00 INFO  -         line 01")
	rejectView(t, m, "line 40")
	m, _ = raise(t, m, key('G', "G"))
	requireView(t, m, "▶ 12:00:00 INFO  -         line 41")
	if !m.follow {
		t.Fatal("G did not turn follow on")
	}
	appendEvent(t, path, eventJSON(t, day, "INFO", "line 42"))
	m = tickUntilCount(t, m, 42)
	requireView(t, m, "▶ 12:00:00 INFO  -         line 42")
}

func TestDailyFileRolloverSelectsCurrentDay(t *testing.T) {
	home := newTestHome(t)
	firstDay := time.Date(2026, 4, 8, 12, 0, 0, 0, time.Local)
	secondDay := firstDay.AddDate(0, 0, 1)
	clock := fixedClock(t, firstDay)
	firstPath := dailyPath(home, firstDay)
	writeEvents(t, firstPath,
		eventJSON(t, firstDay, "INFO", "before midnight one"),
		eventJSON(t, firstDay, "INFO", "before midnight two"),
	)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 160, Height: 24})
	requireView(t, m, "before midnight one", "before midnight two")

	clock.set(secondDay)
	writeEvents(t, dailyPath(home, secondDay), eventJSON(t, secondDay, "INFO", "after midnight"))
	m, _ = raise(t, m, tickMsg{})
	m.screenLoading = false
	m = loadLogs(t, m)
	if len(m.events) != 1 {
		t.Fatalf("new day loaded %d events; want only its daily file", len(m.events))
	}
	requireView(t, m, "after midnight")
	rejectView(t, m, "before midnight one", "before midnight two")
}

func TestFollowKeepsIncompleteLineUntilNewline(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 18, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "complete line"))
	m := newTestLogsModel(t, home, clock)
	requireView(t, m, "complete line")

	partial := `{"time":"2026-04-18T12:00:01-04:00","level":"INFO","msg":"half`
	appendRaw(t, path, partial)
	m = loadLogs(t, m)
	rejectView(t, m, "half")
	appendRaw(t, path, " line\"}\n")
	m = tickUntilCount(t, m, 2)
	requireView(t, m, "half line")
}

func TestDailyFileWaitsForFirstFile(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 9, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	requireView(t, m, "no events today")

	writeEvents(t, dailyPath(home, day), eventJSON(t, day, "INFO", "first arrives"))
	m = tickUntilCount(t, m, 1)
	requireView(t, m, "first arrives")
}

func TestDailyFileInvalidTypeShowsError(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 10, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	if err := os.MkdirAll(dailyPath(home, day), 0o755); err != nil {
		t.Fatal(err)
	}
	m := newTestLogsModel(t, home, clock)
	requireView(t, m, "error:", "not a regular file")
	rejectView(t, m, "no events today")
}

func TestScreenLoadErrorShowsInFooter(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 11, 12, 0, 0, 0, time.Local))
	m := newTestLogsModel(t, home, clock)
	m.logLoader = func(*logs.Reader, time.Time) ([]logs.Event, error) {
		return nil, errors.New("fixture load failure")
	}
	m.screenLoading = false
	cmd := m.startCurrentScreenLoad()
	m, _ = raise(t, m, cmd())
	requireView(t, m, "error:", "fixture load failure")
}

func TestStopClosesWriterMoreThanOnce(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 12, 12, 0, 0, 0, time.Local))
	m := New(home, clock.now)
	m.Stop()
	m.Stop()
}

func TestQuitKeysRequireTwoCtrlCPresses(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 13, 12, 0, 0, 0, time.Local))
	m := newTestLogsModel(t, home, clock)
	_, cmd := raise(t, m, key('q', "q"))
	if cmd != nil {
		t.Fatal("q returned a command")
	}
	m, cmd = raise(t, m, ctrlC())
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("first ctrl+c quit the TUI")
		}
	}
	requireView(t, m, "press ctrl+c again to quit")
	_, cmd = raise(t, m, ctrlC())
	quitMessage(t, cmd)
}

func TestOtherKeysDoNotQuit(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 14, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	// Decision 24: a printable key in the Chat feed types in the input, so
	// the check uses the Agents screen.
	m.screenIndex = 1
	_, cmd := raise(t, m, key('x', "x"))
	if cmd != nil {
		t.Fatalf("got command %T; want none", cmd())
	}
}

func TestViewNamesHive(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 15, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	if got := m.View().Content; !strings.Contains(got, "hive") {
		t.Fatalf("view %q does not name hive", got)
	}
}

func TestRunQuits(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 16, 12, 0, 0, 0, time.Local)
	writeEvents(t, dailyPath(home, day), eventJSON(t, day, "INFO", "run event"))
	var out bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader("\x03\x03"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("RunHome did not exit after q")
	}
	if !strings.Contains(out.String(), "hive") {
		t.Fatalf("output %q does not show the shell", out.String())
	}
}

func TestRunResolvesFixtureHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HIVE_HOME", "")
	home := newTestHome(t)
	day := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	writeEvents(t, dailyPath(home, day), eventJSON(t, day, "INFO", "wrapper event"))
	t.Chdir(home)
	var out bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- Run(strings.NewReader("\x03\x03"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Run did not exit after q")
	}
	if !strings.Contains(out.String(), "hive") {
		t.Fatalf("output %q does not show the shell", out.String())
	}
}
