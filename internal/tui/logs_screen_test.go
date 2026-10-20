package tui

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

func TestLogsScreenRendersFilteredRows(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})

	view := ansi.Strip(logsScreen{model: m}.View(72, 14))
	for _, line := range []string{
		"12:00:01 INFO task queued 4ms",
		"12:00:03 WARN herdr network retry",
	} {
		if !strings.Contains(view, line) {
			t.Fatalf("logs view omitted row %q:\n%s", line, view)
		}
	}
	if strings.Contains(view, "task list 12ms") {
		t.Fatalf("default info level showed a debug command group:\n%s", view)
	}
}

func TestLogsScreenFoldsRequestsAndShowsSelectedJSON(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})
	for range 3 {
		m, _ = raise(t, m, key('l', "l"))
	}

	view := ansi.Strip(logsScreen{model: m}.View(72, 14))
	if !strings.Contains(view, "12:00:00 DEBUG cli task list 12ms") {
		t.Fatalf("command start and end were not folded into one row:\n%s", view)
	}
	if strings.Contains(view, "command end") {
		t.Fatalf("folded request still shows its end event:\n%s", view)
	}

	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	view = ansi.Strip(logsScreen{model: m}.View(120, 30))
	if !strings.Contains(view, `"msg": "command start"`) {
		t.Fatalf("enter did not show selected event JSON:\n%s", view)
	}
	m, _ = raise(t, m, key('r', "r"))
	view = ansi.Strip(logsScreen{model: m}.View(72, 14))
	if !strings.Contains(view, "request: r-11111111") || strings.Contains(view, "network retry") {
		t.Fatalf("request filter did not isolate the selected request:\n%s", view)
	}
	m, _ = raise(t, m, key('r', "r"))
	view = ansi.Strip(logsScreen{model: m}.View(72, 14))
	if strings.Contains(view, "request: r-11111111") || !strings.Contains(view, "network retry") {
		t.Fatalf("second r did not restore all requests:\n%s", view)
	}
}

func TestLogsScreenCyclesInfoLevelToWarn(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})
	m, _ = raise(t, m, key('l', "l"))

	view := ansi.Strip(logsScreen{model: m}.View(72, 14))
	if !strings.Contains(view, "level: warn") {
		t.Fatalf("logs screen did not show the cycled warn level:\n%s", view)
	}
}

func TestLogsScreenCyclesMetadataSearchAndDay(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})

	m, _ = raise(t, m, key('c', "c"))
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "component: cli") {
		t.Fatalf("component filter did not cycle:\n%s", view)
	}
	m, _ = raise(t, m, key('p', "p"))
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "project: demo") {
		t.Fatalf("project filter did not cycle:\n%s", view)
	}
	m, _ = raise(t, m, key('a', "a"))
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "agent: worker-1") {
		t.Fatalf("agent filter did not cycle:\n%s", view)
	}
	m, _ = raise(t, m, key('/', "/"))
	for _, character := range "task" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	view = ansi.Strip(m.View().Content)
	for _, filter := range []string{"component: cli", "project: demo", "agent: worker-1", "text: task"} {
		if !strings.Contains(view, filter) {
			t.Fatalf("filter line omitted %q:\n%s", filter, view)
		}
	}

	m, cmd := raise(t, m, key(']', "]"))
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "day: 2026-04-07") {
		t.Fatalf("right bracket did not select the next day:\n%s", view)
	}
	if cmd == nil {
		t.Fatal("day change did not request a background refresh")
	}
}

func TestLogsScreenKeepsSelectionAcrossRefresh(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	events := logsFixture(t, day)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: events})
	m, _ = raise(t, m, key('j', "j"))
	before := ansi.Strip(logsScreen{model: m}.View(72, 14))
	if !strings.Contains(before, "▶ 12:00:03 WARN herdr network retry") {
		t.Fatalf("down did not select the next event:\n%s", before)
	}
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: events})
	after := ansi.Strip(logsScreen{model: m}.View(72, 14))
	if !strings.Contains(after, "▶ 12:00:03 WARN herdr network retry") {
		t.Fatalf("refresh did not keep the selected event:\n%s", after)
	}
}

func TestLogsScreenDayNavigationLoadsSelectedFile(t *testing.T) {
	home := newTestHome(t)
	firstDay := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	secondDay := firstDay.AddDate(0, 0, 1)
	clock := fixedClock(t, secondDay)
	writeEvents(t, dailyPath(home, firstDay), eventJSON(t, firstDay, "INFO", "first day"))
	writeEvents(t, dailyPath(home, secondDay), eventJSON(t, secondDay, "INFO", "second day"))
	m := newTestLogsModel(t, home, clock)
	requireView(t, m, "second day")
	m.logger = nil
	m, cmd := raise(t, m, key('[', "["))
	if cmd == nil {
		t.Fatal("day change did not start a background read")
	}
	loaded, ok := cmd().(screenLoadMsg)
	if !ok {
		t.Fatalf("day read returned %T; want screenLoadMsg", loaded)
	}
	m, _ = raise(t, m, loaded)
	requireView(t, m, "first day")
	rejectView(t, m, "second day")
}

func TestRunHomeLogsKeysOnStringReader(t *testing.T) {
	home := newTestHome(t)
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	day := time.Now().In(time.Local)
	writeEvents(t, dailyPath(home, day), logsFixtureEvent(t, day, "INFO", "task", "queued", map[string]any{"project_id": "demo"}))
	input, sendInput := io.Pipe()
	output := newSynchronizedOutput()
	done := make(chan error, 1)
	go func() { done <- RunHome(home, input, output) }()
	t.Cleanup(func() {
		_ = input.Close()
		_ = sendInput.Close()
	})

	if _, err := sendInput.Write([]byte("8")); err != nil {
		t.Fatalf("send screen key: %v", err)
	}
	waitForOutput(t, output, "▶ 8 Logs", "task queued")
	if _, err := sendInput.Write([]byte("l")); err != nil {
		t.Fatalf("send level key: %v", err)
	}
	view := waitForOutput(t, output, "▶ 8 Logs", "warn", "no events today")
	if err := sendInput.Close(); err != nil {
		t.Fatalf("close input: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	eventAt := strings.Index(view, "task queued")
	if !strings.Contains(view, "▶ 8 Logs") || eventAt < 0 || !strings.Contains(view[eventAt:], "warnno events today") {
		t.Fatalf("program output did not show Logs with the warn filter applied:\n%s", view)
	}
	screenView := ansi.Strip(output.String())
	if !strings.Contains(screenView, "▶ 8 Logs") {
		t.Fatalf("program output did not show Logs:\n%s", screenView)
	}
	events, _, err := logs.NewReader(home).Read(t.Context(), day)
	if err != nil {
		t.Fatalf("read TUI action log: %v", err)
	}
	cycledLevel := false
	for _, event := range events {
		if event.Fields["msg"] == "tui action" && event.Fields["action"] == "level" {
			cycledLevel = true
			break
		}
	}
	if !cycledLevel {
		t.Fatalf("program did not apply the level key; events: %#v", events)
	}
}

type synchronizedOutput struct {
	mu      sync.Mutex
	output  strings.Builder
	changed chan struct{}
}

func newSynchronizedOutput() *synchronizedOutput {
	return &synchronizedOutput{changed: make(chan struct{}, 1)}
}

func (w *synchronizedOutput) Write(data []byte) (int, error) {
	w.mu.Lock()
	n, err := w.output.Write(data)
	w.mu.Unlock()
	select {
	case w.changed <- struct{}{}:
	default:
	}
	return n, err
}

func (w *synchronizedOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.output.String()
}

func waitForOutput(t *testing.T, output *synchronizedOutput, wants ...string) string {
	t.Helper()
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	for {
		view := ansi.Strip(output.String())
		found := true
		for _, want := range wants {
			if !strings.Contains(view, want) {
				found = false
				break
			}
		}
		if found {
			return view
		}
		select {
		case <-output.changed:
		case <-timeout.C:
			t.Fatalf("program output did not reach state containing %q:\n%s", strings.Join(wants, ", "), view)
			return view
		}
	}
}

func logsFixture(t *testing.T, day time.Time) []logs.Event {
	t.Helper()
	home := t.TempDir()
	path := dailyPath(home, day)
	writeEvents(t, path,
		logsFixtureEvent(t, day, "DEBUG", "cli", "command start", map[string]any{"command": "task list", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(time.Second), "INFO", "task", "queued", map[string]any{"duration_ms": 4, "project_id": "demo", "agent_id": "worker-1", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(2*time.Second), "DEBUG", "cli", "command end", map[string]any{"command": "task list", "duration_ms": 12, "outcome": "ok", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(3*time.Second), "WARN", "herdr", "network retry", map[string]any{"project_id": "other", "agent_id": "worker-2", "request_id": "r-22222222"}),
	)
	reader := logs.NewReaderWithClock(home, func() time.Time { return day })
	events, _, err := reader.Read(t.Context(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func logsFixtureEvent(t *testing.T, at time.Time, level, component, message string, extra map[string]any) string {
	t.Helper()
	event := map[string]any{"time": at.Format(time.RFC3339Nano), "level": level, "component": component, "msg": message}
	for key, value := range extra {
		event[key] = value
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
