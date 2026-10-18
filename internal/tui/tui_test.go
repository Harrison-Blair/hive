package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
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
	t.Cleanup(model.Stop)
	return model
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
	view := m.View().Content
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Fatalf("view does not contain %q:\n%s", want, view)
		}
	}
}

func rejectView(t *testing.T, m *Model, unwanted ...string) {
	t.Helper()
	view := m.View().Content
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

func TestHeaderShowsHomeAndRefreshesStoreCounts(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)

	requireView(t, m, "hive "+home, "supervisor: down", "agents: 0", "inbox: 0 unread")

	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 1, PID: int64(os.Getpid()),
	})
	writeRecord(t, home, []string{"agents"}, "agent-one", records.Agent{
		Key: "agent-one", Name: "agent-one", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive,
	})
	writeRecord(t, home, []string{"agents"}, "agent-two", records.Agent{
		Key: "agent-two", Name: "agent-two", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive,
	})
	for sequence := uint64(1); sequence <= 3; sequence++ {
		writeRecord(t, home, []string{"inbox", "coordinator"}, fmt.Sprintf("%d", sequence), records.Inbox{
			Seq: sequence, Role: "coordinator", Kind: records.InboxMessage, Text: "waiting",
		})
	}
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "supervisor: live", "agents: 2", "inbox: 3 unread")

	writeRecord(t, home, []string{"agents"}, "agent-three", records.Agent{
		Key: "agent-three", Name: "agent-three", Role: "worker", Project: "demo", Task: "1-first", Status: records.AgentLive,
	})
	writeRecord(t, home, []string{"inbox", "coordinator"}, "4", records.Inbox{
		Seq: 4, Role: "coordinator", Kind: records.InboxMessage, Text: "waiting",
	})
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "agents: 3", "inbox: 4 unread")

	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 2, PID: 0,
	})
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "supervisor: down")
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

	requireView(t, m, "hive "+home, "error:")
	rejectView(t, m, "agents: 0", "inbox: 0 unread")
}

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
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 10})
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "alpha gq", "beta", "gamma gq", "epsilon gq", "zeta")

	// Editing must not let viewing-mode keys act.
	m, _ = raise(t, m, key('/', "/"))
	for _, code := range []rune{'f', 'l', 'G'} {
		m, _ = raise(t, m, key(code, string(code)))
	}
	requireView(t, m, "text: -", "level: debug", "follow: on")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "level: debug", "follow: on")

	// A q in editing mode is a printable character, not a quit.
	m, _ = raise(t, m, key('/', "/"))
	for _, code := range []rune{'g', 'q'} {
		var cmd tea.Cmd
		m, cmd = raise(t, m, key(code, string(code)))
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Fatal("q in editing mode quit the TUI")
			}
		}
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	requireView(t, m, "alpha gq", "gamma gq", "epsilon gq", "text: gq")
	rejectView(t, m, "beta", "delta", "zeta")

	// Escape clears the input and the applied filter.
	m, _ = raise(t, m, key('/', "/"))
	m, _ = raise(t, m, key('z', "z"))
	m, _ = raise(t, m, key('z', "z"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "beta", "delta", "zeta", "text: -", "level: debug")

	// l cycles debug, info, warn, error, and debug again.
	for _, want := range []string{"level: info", "level: warn", "level: error", "level: debug"} {
		m, _ = raise(t, m, key('l', "l"))
		requireView(t, m, want)
	}

	m, cmd := raise(t, m, key('q', "q"))
	quitMessage(t, cmd)
	m, _ = raise(t, m, tickMsg{})
	_, cmd = raise(t, m, ctrlC())
	quitMessage(t, cmd)
}

func TestFollowReadsAppendedLineOnTick(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 19, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "first"))
	m := newTestModel(t, home, clock)
	requireView(t, m, "first")

	appendEvent(t, path, eventJSON(t, day, "INFO", "second on tick"))
	time.Sleep(700 * time.Millisecond)
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "second on tick")
}

func TestFollowControlsViewport(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 7, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	lines := make([]string, 0, 10)
	for index := 1; index <= 10; index++ {
		lines = append(lines, eventJSON(t, day, "INFO", fmt.Sprintf("line %02d", index)))
	}
	path := dailyPath(home, day)
	writeEvents(t, path, lines...)
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 5})
	m, _ = raise(t, m, tickMsg{})
	if !m.viewport.AtBottom() {
		t.Fatalf("follow on left the viewport away from the end")
	}

	m, _ = raise(t, m, key('f', "f"))
	requireView(t, m, "follow: off")
	before := m.viewport.YOffset()
	if before == 0 {
		t.Fatal("test needs a scrolled viewport")
	}
	appendEvent(t, path, eventJSON(t, day, "INFO", "line eleven"))
	m = tickUntilCount(t, m, 11)
	if got := m.viewport.YOffset(); got != before {
		t.Fatalf("follow off moved the viewport from %d to %d", before, got)
	}

	m, _ = raise(t, m, key('g', "g"))
	if !m.viewport.AtTop() {
		t.Fatal("g did not move the viewport to the first line")
	}
	m, _ = raise(t, m, key('G', "G"))
	if !m.viewport.AtBottom() {
		t.Fatal("G did not move the viewport to the last line")
	}
	requireView(t, m, "follow: on")
}

func TestDailyFileRolloverKeepsEarlierLines(t *testing.T) {
	home := newTestHome(t)
	firstDay := time.Date(2026, 4, 8, 12, 0, 0, 0, time.Local)
	secondDay := firstDay.AddDate(0, 0, 1)
	clock := fixedClock(t, firstDay)
	firstPath := dailyPath(home, firstDay)
	writeEvents(t, firstPath,
		eventJSON(t, firstDay, "INFO", "before midnight one"),
		eventJSON(t, firstDay, "INFO", "before midnight two"),
	)
	m := newTestModel(t, home, clock)
	requireView(t, m, "before midnight one", "before midnight two")

	clock.set(secondDay)
	writeEvents(t, dailyPath(home, secondDay), eventJSON(t, secondDay, "INFO", "after midnight"))
	m = tickUntilCount(t, m, 3)
	requireView(t, m, "before midnight one", "before midnight two", "after midnight")
}

func TestFollowKeepsIncompleteLineUntilNewline(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 18, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "complete line"))
	m := newTestModel(t, home, clock)
	requireView(t, m, "complete line")

	partial := `{"time":"2026-04-18T12:00:01-04:00","level":"INFO","msg":"half`
	appendRaw(t, path, partial)
	time.Sleep(700 * time.Millisecond)
	rejectView(t, m, "half")
	appendRaw(t, path, " line\"}\n")
	m = tickUntilCount(t, m, 2)
	requireView(t, m, "half line")
}

func TestDailyFileWaitsForFirstFile(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 9, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestModel(t, home, clock)
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
	m := newTestModel(t, home, clock)
	requireView(t, m, "error:", "not a regular file")
	rejectView(t, m, "no events today")
}

func TestFollowFailureShowsErrorAndStops(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 11, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "one"))
	m := newTestModel(t, home, clock)
	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- m.waitFollow()() }()
	select {
	case msg := <-done:
		m, _ = raise(t, m, msg)
	case <-time.After(4 * time.Second):
		t.Fatal("follower failure did not arrive")
	}
	requireView(t, m, "error:", "truncated")
}

func TestStopSettlesFollower(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 12, 12, 0, 0, 0, time.Local))
	m := New(home, clock.now)
	m.Stop()
	select {
	case _, ok := <-m.eventCh:
		if ok {
			t.Fatal("event stream stayed open after Stop")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not settle the follower")
	}
}

func TestQuitKeysQuit(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 13, 12, 0, 0, 0, time.Local))
	for _, press := range []tea.KeyPressMsg{key('q', "q"), ctrlC()} {
		m := newTestModel(t, home, clock)
		_, cmd := raise(t, m, press)
		quitMessage(t, cmd)
	}
}

func TestOtherKeysDoNotQuit(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 4, 14, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
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
	go func() { done <- RunHome(home, strings.NewReader("q"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("RunHome did not exit after q")
	}
	if !strings.Contains(out.String(), home) {
		t.Fatalf("output %q does not hold the home path", out.String())
	}
}

func TestRunResolvesFixtureHome(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	writeEvents(t, dailyPath(home, day), eventJSON(t, day, "INFO", "wrapper event"))
	t.Chdir(home)
	var out bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- Run(strings.NewReader("q"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Run did not exit after q")
	}
	if !strings.Contains(out.String(), home) {
		t.Fatalf("output %q does not hold the resolved home", out.String())
	}
}
