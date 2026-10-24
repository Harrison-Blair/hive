package tui

import (
	"encoding/json"
	"fmt"
	"io"
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
	requireView(t, m, "12:00:00 DEBUG  cli        task list ", "12:00:01 INFO   task       queued ")
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
		{"task:t9", []string{"no event matches task:t9"}, []string{"queued", "network retry"}},
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

// logs.md rule 8: ← and → change the day, and f turns follow off and on.
// The title row shows the day, and the filter line shows follow as ● or ○.
func TestLogsDayAndFollowKeys(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	requireView(t, m, "follow ●  f toggles follow", "← 2026-04-06 → day")
	m, _ = raise(t, m, key('f', "f"))
	requireView(t, m, "follow ○  f toggles follow")
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	requireView(t, m, "← 2026-04-07 → day")
	if cmd == nil {
		t.Fatal("→ did not start a load of the next day")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
	requireView(t, m, "← 2026-04-05 → day")
}

// logs.md rule 6 and the key table: from 120 columns, ctrl+] moves the
// focus between the list and the field panel, the panel takes ↑↓ g G, and
// esc returns to Chat from the panel.
func TestLogsWideFieldPanelKeys(t *testing.T) {
	m := logsModelAt(t, 120, 24)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m, _ = raise(t, m, ctrl(']'))
	if m.focusState() != "json" {
		t.Fatalf("ctrl+] gave focus %q; want json", m.focusState())
	}
	requireKeys(t, m, "json", "↑↓ scroll")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.detailsOffset[tabLogs] != 1 {
		t.Fatalf("down gave panel offset %d; want 1", m.detailsOffset[tabLogs])
	}
	m, _ = raise(t, m, key('g', "g"))
	if m.detailsOffset[tabLogs] != 0 {
		t.Fatalf("g gave panel offset %d; want 0", m.detailsOffset[tabLogs])
	}
	m, _ = raise(t, m, ctrl(']'))
	if m.focusState() != "events" {
		t.Fatalf("ctrl+] gave focus %q; want events", m.focusState())
	}
	m, _ = raise(t, m, ctrl(']'))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.activeTab() != tabChat {
		t.Fatalf("esc on the panel gave tab %d; want Chat", m.activeTab())
	}
}

// AC4, decision 9: below 120 columns, enter opens the field panel overlay
// of the selected event, → still changes the day, and esc closes it.
func TestLogsCompactFieldOverlay(t *testing.T) {
	m := logsModelAt(t, 80, 24)
	requireKeys(t, m, "events", "↑↓ move · enter details")
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "event 12:00:03 · WARN · esc back", "msg           network retry")
	requireKeys(t, m, "json", "↑↓ scroll · esc back to events")
	rejectView(t, m, "─ events · ")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	requireView(t, m, "← 2026-04-07 → day")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "─ events · ")
}

// Decision 22 and the Logs boards: the field panel shows each line that
// fits, with no scroll mark. When the lines do not fit, the scroll mark
// shows, and pgdown scrolls to the last line.
func TestLogsFieldPanelOverflowShowsTheThumb(t *testing.T) {
	m := logsModelAt(t, 80, 24)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m.moveLogSelection(-10)
	m, _ = raise(t, m, key(0, "enter"))
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, "┃") {
		t.Fatalf("fields that fit show a scroll mark:\n%s", view)
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
	// 33 field lines do not fit in the overlay at 80×24.
	requireView(t, m, "time          2026-04-06T12:00:00Z", "┃")
	rejectView(t, m, "field_29")
	for range 3 {
		m, _ = raise(t, m, key(0, "pgdown"))
	}
	requireView(t, m, "field_29      29")
	rejectView(t, m, "time          2026-04-06T12:00:00Z")
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
	assertSelectedRowVisible("13:23:19 INFO   tui        event-4999")
	if m.logSelection != 4999 {
		t.Fatalf("follow selected row %d, want newest row 4999", m.logSelection)
	}

	for range 60 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	}
	assertSelectedRowVisible("13:22:19 INFO   tui        event-4939")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	assertSelectedRowVisible("13:22:20 INFO   tui        event-4940")
	m, _ = raise(t, m, key(0, "pgup"))
	assertSelectedRowVisible(fmt.Sprintf("%s INFO   tui        event-%04d", day.Add(time.Duration(4940-m.logPage())*time.Second).Format("15:04:05"), 4940-m.logPage()))
}

func TestLogsScreenKeepsSelectionAcrossRefresh(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	requireView(t, m, "▶ 12:00:01 INFO   task       queued")
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, day)})
	requireView(t, m, "▶ 12:00:01 INFO   task       queued")
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
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
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

// logs.md keys: esc in the event list returns to Chat when no filter line
// is open. The applied filter stays, so Logs shows it again on ctrl+l.
func TestLogsEscapeReturnsToChatAndKeepsTheFilter(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 20, 12, 0, 0, 0, time.Local)
	writeEvents(t, dailyPath(home, day),
		eventJSON(t, day, "INFO", "project alpha"),
		eventJSON(t, day, "INFO", "other beta"),
	)
	m := newTestLogsModel(t, home, fixedClock(t, day))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = raise(t, m, key('/', "/"))
	for _, r := range " project" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, key(0, "enter"))
	rejectView(t, m, "other beta")

	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.activeTab() != tabChat {
		t.Fatalf("esc gave tab %d; want Chat", m.activeTab())
	}
	m, _ = raise(t, m, ctrl('l'))
	if m.activeTab() != tabLogs {
		t.Fatalf("ctrl+l gave tab %d; want Logs", m.activeTab())
	}
	requireView(t, m, "/ level:info project ", "project alpha")
	rejectView(t, m, "other beta")
}

// AC3: the program in normal mode shows Logs on ctrl+l (golden-frames
// decision 16) and applies a typed filter, from a string reader.
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

	if _, err := sendInput.Write([]byte("\x0c")); err != nil {
		t.Fatalf("send screen key: %v", err)
	}
	waitForOutput(t, output, "─ events · ", "queued")
	if _, err := sendInput.Write([]byte("/ proj:other\r")); err != nil {
		t.Fatalf("send filter keys: %v", err)
	}
	waitForOutput(t, output, "proj:other", "no event matches level:info proj:other")
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

// logs.md rule 3: an event with a reason shows its msg, " · ", and the
// reason. The row cuts the text at the column width only. The JSON panel
// keeps the full reason.
func TestLogsRowShowsTheWholeReasonWhenItFits(t *testing.T) {
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m := logsModelAt(t, 120, 40)
	events := []string{
		logsFixtureEvent(t, day, "WARN", "task", "fault", map[string]any{"reason": "first; suffix"}),
		logsFixtureEvent(t, day.Add(time.Second), "WARN", "task", "verify fail", map[string]any{"reason": "V4 diff scope; Design section missing"}),
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
	requireView(t, m, "12:00:00 WARN   task       fault · first; suffix", "verify fail · V4 diff scope;")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	requireView(t, m, "reason        first; suffix")
}

// Follow selects the newest event when a load brings a new newest event. A
// load with the same newest event keeps the selection, so a refresh of the
// Logs fixture keeps its selected event.
func TestLogsFollowKeepsTheSelectionWithoutNewEvents(t *testing.T) {
	demo := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	requireView(t, demo, "event 13:38:02")
	events := append([]logs.Event(nil), demo.events...)
	demo, _ = raise(t, demo, screenLoadMsg{request: logsDataRequest, events: events})
	requireView(t, demo, "event 13:38:02")

	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, key('f', "f"))
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	more := logsFixture(t, day)
	line := logsFixtureEvent(t, day.Add(time.Minute), "INFO", "task", "landed", nil)
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		t.Fatal(err)
	}
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: append(more, logs.Event{Raw: []byte(line), Fields: fields})})
	requireView(t, m, "▶ 12:01:00 INFO   task       landed")
}

// A filter that hides each row keeps the selected event. When the rows show
// again, the same event is selected.
func TestLogsSelectionSurvivesAnEmptyFilter(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	requireView(t, m, "▶ 12:00:03 WARN   herdr      network retry", "follow ○")
	m, _ = raise(t, m, key('/', "/"))
	for _, r := range " nothing-matches" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "no event matches level:info nothing-matches")
	m, _ = raise(t, m, key('/', "/"))
	for range " nothing-matches" {
		m, _ = raise(t, m, key(0, "backspace"))
	}
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "▶ 12:00:03 WARN   herdr      network retry")
}
