package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

// logsModelAt gives a normal-mode Logs model at width×height with the
// events of logsFixture.
func logsModelAt(t *testing.T, width, height int) *Model {
	t.Helper()
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m := newTestLogsModel(t, newTestHome(t), fixedClock(t, day))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})
	return m
}

// AC1: a command start and its command end of one request fold to one row
// with the command. The row shows the msg and no duration.
func TestLogsFoldsCommandRequests(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	requireView(t, m, "12:00:00 DEBUG cli       task list ", "12:00:01 INFO  task      queued ")
	rejectView(t, m, "command end", "4ms")
}

// AC1, decision 25: each typed field and free text filters the rows.
func TestLogsTypedFilterFields(t *testing.T) {
	for _, tt := range []struct {
		filter string
		shows  []string
		hides  []string
	}{
		{"level:debug comp:cli", []string{"task list"}, []string{"queued", "network retry"}},
		{"level:debug proj:demo", []string{"queued"}, []string{"task list", "network retry"}},
		{"agent:worker-1", []string{"queued"}, []string{"network retry"}},
		{"agent:worker-2", []string{"network retry"}, []string{"queued"}},
		{"level:debug req:r-22222222", []string{"network retry"}, []string{"queued", "task list"}},
		{"task:t9", []string{"no events today"}, []string{"queued", "network retry"}},
		{"level:warn", []string{"network retry"}, []string{"queued"}},
		{"retry", []string{"network retry"}, []string{"queued"}},
	} {
		t.Run(tt.filter, func(t *testing.T) {
			m := logsModelAt(t, 120, 40)
			m.textFilter = tt.filter
			m.refreshLogSelection()
			m.invalidateView()
			requireView(t, m, tt.shows...)
			rejectView(t, m, tt.hides...)
		})
	}
}

// Decision 25: agent: matches the agent ID, or the agent name before its
// generation.
func TestLogsAgentFieldMatchesTheAgentName(t *testing.T) {
	row := logRow{event: logs.Event{Fields: map[string]any{"agent_id": "manager-hive-g2", "level": "INFO"}}}
	for filter, want := range map[string]bool{
		"agent:manager-hive-g2": true,
		"agent:manager-hive":    true,
		"agent:manager":         false,
		"agent:manager-hive-g":  false,
	} {
		if got := parseLogFilter(filter).matches(row); got != want {
			t.Errorf("%s matches %v; want %v", filter, got, want)
		}
	}
}

// AC1: the filter bar takes a paste while it is open, and only then.
func TestLogsFilterBarTakesPaste(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.PasteMsg{Content: " retry"})
	requireView(t, m, "/ level:info ")
	rejectView(t, m, "level:info retry")
	m, _ = raise(t, m, key('/', "/"))
	m, _ = raise(t, m, tea.PasteMsg{Content: " retry"})
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "/ level:info retry ", "network retry")
	rejectView(t, m, "queued")
}

// AC1: < and > change the day, and f turns follow off and on. The filter
// bar shows the day, and follow as ● or ○.
func TestLogsDayAndFollowKeys(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	requireView(t, m, "follow ●  < 2026-04-06 >")
	m, _ = raise(t, m, key('f', "f"))
	requireView(t, m, "follow ○  < 2026-04-06 >")
	m, cmd := raise(t, m, key('>', ">"))
	requireView(t, m, "< 2026-04-07 >")
	if cmd == nil {
		t.Fatal("> did not start a load of the next day")
	}
	m, _ = raise(t, m, key('<', "<"))
	m, _ = raise(t, m, key('<', "<"))
	requireView(t, m, "< 2026-04-05 >")
}

// AC4, decisions 21 and 22: from 120 columns, enter and tab give the focus
// to the JSON panel, and ↑↓ g G scroll it. esc gives the focus back to the
// events.
func TestLogsWideJSONKeys(t *testing.T) {
	m := logsModelAt(t, 120, 24)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m, _ = raise(t, m, key(0, "enter"))
	if m.focusState() != "json" {
		t.Fatalf("enter gave focus %q; want json", m.focusState())
	}
	requireView(t, m, "[json] ↑↓ scroll")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.detailsOffset[tabLogs] != 1 {
		t.Fatalf("down gave JSON offset %d; want 1", m.detailsOffset[tabLogs])
	}
	m, _ = raise(t, m, key('g', "g"))
	if m.detailsOffset[tabLogs] != 0 {
		t.Fatalf("g gave JSON offset %d; want 0", m.detailsOffset[tabLogs])
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.focusState() != "events" {
		t.Fatalf("esc gave focus %q; want events", m.focusState())
	}
	m, _ = raise(t, m, key(0, "tab"))
	if m.focusState() != "json" {
		t.Fatalf("tab gave focus %q; want json", m.focusState())
	}
}

// AC4, decision 9: below 120 columns, enter opens the JSON overlay of the
// selected event, < and > still change the day, and esc closes it.
func TestLogsCompactJSONOverlay(t *testing.T) {
	m := logsModelAt(t, 80, 24)
	requireView(t, m, "[events] ↑↓ move · enter details")
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "event 12:00:03 · esc back", `"msg": "network retry"`, "[json] ↑↓ scroll · esc back to events")
	rejectView(t, m, "─ events · ")
	m, _ = raise(t, m, key('>', ">"))
	requireView(t, m, "< 2026-04-07 >")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "─ events · ")
}

// Decision 22 and the Logs board: the JSON panel shows each line that fits.
// When the lines do not fit, the last body row shows the count of the lines
// below it, and the scroll mark shows.
func TestLogsJSONOverflowShowsMoreLines(t *testing.T) {
	m := logsModelAt(t, 80, 24)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m.moveLogSelection(-10)
	m, _ = raise(t, m, key(0, "enter"))
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, "more lines") || strings.Contains(view, "┃") {
		t.Fatalf("JSON that fits shows an overflow mark:\n%s", view)
	}
	event := map[string]any{"time": "2026-04-06T12:00:00Z", "level": "INFO", "msg": "big"}
	for index := range 30 {
		event[fmt.Sprintf("field_%02d", index)] = index
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: []logs.Event{{Raw: raw, Fields: fields}}})
	// 35 JSON lines in 15 body rows: 14 lines, then the count of the 21
	// lines below them.
	requireView(t, m, "↓ 21 more lines", "┃")
	// The overlay takes the scroll keys of decision 21, not G.
	for range 2 {
		m, _ = raise(t, m, key(0, "pgdown"))
	}
	requireView(t, m, `"time": "2026-04-06T12:00:00Z"`, "}")
	rejectView(t, m, "more lines")
}

// AC2, decision 46: at each size, each line has the pane width, no border
// breaks, and the selected row shows. The overlay shows after enter below
// 120 columns. Empty data and text longer than its column keep the frame.
func TestLogsFitsEachSize(t *testing.T) {
	long := strings.Repeat("long-", 40)
	for _, data := range []string{"fixture", "empty", "long"} {
		for _, size := range []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}} {
			t.Run(fmt.Sprintf("%s/%dx%d", data, size.Width, size.Height), func(t *testing.T) {
				m := logsModelAt(t, size.Width, size.Height)
				switch data {
				case "empty":
					m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest})
				case "long":
					day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
					line := logsFixtureEvent(t, day, "WARN", long, long, map[string]any{"reason": long})
					var fields map[string]any
					if err := json.Unmarshal([]byte(line), &fields); err != nil {
						t.Fatal(err)
					}
					m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: []logs.Event{{Raw: []byte(line), Fields: fields}}})
					m.textFilter = "level:info " + long
					m.refreshLogSelection()
					m.invalidateView()
				}
				requireLogsFit(t, m, size, data != "empty")
				if size.Width >= minimumWidth && size.Width < 120 {
					m, _ = raise(t, m, key(0, "enter"))
					requireLogsFit(t, m, size, false)
					if data != "empty" {
						requireView(t, m, "· esc back")
					}
				}
			})
		}
	}
}

func requireLogsFit(t *testing.T, m *Model, size demoSize, selected bool) {
	t.Helper()
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
	requireWholeBorders(t, view)
	if selected && !strings.Contains(view, "▶") {
		t.Fatalf("the selected row is not visible:\n%s", view)
	}
}

func TestLogsFollowAndSelectionKeepLargeLogRowsOnScreen(t *testing.T) {
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	home := newTestHome(t)
	lines := make([]string, 5000)
	for index := range lines {
		lines[index] = logsFixtureEvent(t, day.Add(time.Duration(index)*time.Second), "INFO", "tui", fmt.Sprintf("event-%04d", index), nil)
	}
	writeEvents(t, dailyPath(home, day), lines...)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m = loadLogs(t, m)

	assertSelectedRowVisible := func(want string) {
		t.Helper()
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "▶ "+want) {
			t.Fatalf("selected log row %q is off screen:\n%s", want, view)
		}
	}
	assertSelectedRowVisible("13:23:19 INFO  tui       event-4999")
	if m.logSelection != 4999 {
		t.Fatalf("follow selected row %d, want newest row 4999", m.logSelection)
	}

	for range 60 {
		m, _ = raise(t, m, key('k', "k"))
	}
	assertSelectedRowVisible("13:22:19 INFO  tui       event-4939")
	m, _ = raise(t, m, key('j', "j"))
	assertSelectedRowVisible("13:22:20 INFO  tui       event-4940")
	m, _ = raise(t, m, key(0, "pgup"))
	assertSelectedRowVisible("13:21:50 INFO  tui       event-4910")
}

func TestLogsScreenKeepsSelectionAcrossRefresh(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, key('k', "k"))
	requireView(t, m, "▶ 12:00:01 INFO  task      queued")
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})
	requireView(t, m, "▶ 12:00:01 INFO  task      queued")
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
	m, cmd := raise(t, m, key('<', "<"))
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

// Escape in the event list clears the applied filter and writes one action
// event. It keeps follow and the viewing mode.
func TestLogsEscapeClearsAppliedFilter(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 20, 12, 0, 0, 0, time.Local)
	writeEvents(t, dailyPath(home, day),
		eventJSON(t, day, "INFO", "project alpha"),
		eventJSON(t, day, "INFO", "other beta"),
		eventJSON(t, day, "DEBUG", "project debug"),
	)
	m := newTestLogsModel(t, home, fixedClock(t, day))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = raise(t, m, key('/', "/"))
	for _, r := range " project" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, key(0, "enter"))
	rejectView(t, m, "other beta")

	before := countLogLines(t, home)
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		m, _ = raise(t, m, cmd())
	}
	if after := countLogLines(t, home); after != before+1 {
		t.Fatalf("log lines went from %d to %d after Escape; want one event", before, after)
	}
	data, err := os.ReadFile(dailyPath(home, day))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatal(err)
	}
	if event["msg"] != "tui action" || event["action"] != "filter clear" || event["component"] != "tui" {
		t.Fatalf("Escape log event = %v", event)
	}
	requireView(t, m, "/ level:info ", "project alpha", "other beta")
	rejectView(t, m, "project debug")
	if m.mode != viewing || !m.follow {
		t.Fatalf("mode %v, follow %v after Escape; want viewing and follow on", m.mode, m.follow)
	}
}

// AC3: the program in normal mode shows Logs on 6 and applies a typed
// filter, from a string reader.
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

	if _, err := sendInput.Write([]byte("6")); err != nil {
		t.Fatalf("send screen key: %v", err)
	}
	waitForOutput(t, output, "─ events · ", "queued")
	if _, err := sendInput.Write([]byte("/ proj:other\r")); err != nil {
		t.Fatalf("send filter keys: %v", err)
	}
	waitForOutput(t, output, "proj:other", "no events today")
	if err := sendInput.Close(); err != nil {
		t.Fatalf("close input: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	events, _, err := logs.NewReader(home).Read(t.Context(), day)
	if err != nil {
		t.Fatalf("read TUI action log: %v", err)
	}
	applied := false
	for _, event := range events {
		if event.Fields["msg"] == "tui action" && event.Fields["action"] == "filter apply" {
			applied = true
			break
		}
	}
	if !applied {
		t.Fatalf("program did not apply the filter; events: %#v", events)
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

// AC3, decision 37: normal mode and demo mode draw Logs with the same
// renderer. A normal-mode model that reads the Logs fixture events from a
// log file in a scratch home draws the same content rows as the demo model.
func TestLogsNormalModeDrawsTheDemoContent(t *testing.T) {
	demo := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	fixture := demo.demo.fixture
	home := newTestHome(t)
	lines := make([]string, len(fixture.Data.Logs))
	for index, raw := range fixture.Data.Logs {
		lines[index] = string(raw)
	}
	writeEvents(t, dailyPath(home, fixture.now), lines...)
	// The fixture clock is in UTC; testClock gives the local zone.
	m := New(home, func() time.Time { return fixture.now })
	t.Cleanup(m.Stop)
	m.width, m.height = fixture.Terminal.Width, fixture.Terminal.Height
	m.applyDemoUI(fixture.UI)
	m = loadLogs(t, m)

	want := strings.Split(demo.View().Content, "\n")
	got := strings.Split(m.View().Content, "\n")
	if len(got) != len(want) {
		t.Fatalf("normal view has %d rows; want %d", len(got), len(want))
	}
	for row := 3; row < len(want)-1; row++ {
		if got[row] != want[row] {
			t.Fatalf("content row %d differs:\nnormal %q\ndemo   %q", row, ansi.Strip(got[row]), ansi.Strip(want[row]))
		}
	}
}

// Manager decision: an event with a reason shows its msg, " · ", and the
// reason text before the first ";". The JSON panel keeps the full reason.
func TestLogsRowShowsTheReasonBeforeItsFirstSemicolon(t *testing.T) {
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m := logsModelAt(t, 120, 40)
	events := []string{
		logsFixtureEvent(t, day, "WARN", "task", "verify fail", map[string]any{"reason": "V4 diff scope; Design section missing"}),
		logsFixtureEvent(t, day.Add(time.Second), "WARN", "task", "verify fail", map[string]any{"reason": "no evidence"}),
	}
	loaded := make([]logs.Event, len(events))
	for index, line := range events {
		var fields map[string]any
		if err := json.Unmarshal([]byte(line), &fields); err != nil {
			t.Fatal(err)
		}
		loaded[index] = logs.Event{Raw: []byte(line), Fields: fields}
	}
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: loaded})
	requireView(t, m, "12:00:00 WARN  task      verify fail · V4 diff scope ", "verify fail · no evidence ")
	rejectView(t, m, "verify fail · V4 diff scope;")
	m, _ = raise(t, m, key('k', "k"))
	requireView(t, m, `"reason": "V4 diff scope; Design section missin…`)
}
