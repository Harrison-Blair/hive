package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/logs"
)

// raiseRun sends msg and runs the command that it gives, so that an action
// log event reaches the log file.
func raiseRun(t *testing.T, m *Model, msg tea.Msg) *Model {
	t.Helper()
	m, cmd := raise(t, m, msg)
	runCmd(cmd)
	return m
}

// runCmd runs cmd and each command of a batch.
func runCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, next := range batch {
			runCmd(next)
		}
	}
}

// loggedActions gives the action of each tui action event in the log files
// of home.
func loggedActions(t *testing.T, home string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var actions []string
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
			var event map[string]any
			if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "tui action" {
				action, _ := event["action"].(string)
				actions = append(actions, action)
			}
		}
	}
	return actions
}

// viewGrid converts the view of m at its size.
func viewGrid(t *testing.T, m *Model) cellgrid.Grid {
	t.Helper()
	g, err := cellgrid.Convert(m.View().Content, m.width, m.height)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// findCell gives the row and the column of the first cell of text in the
// plain view, or -1.
func findCell(t *testing.T, m *Model, text string) (int, int) {
	t.Helper()
	for row, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if before, _, ok := strings.Cut(line, text); ok {
			return row, len([]rune(before))
		}
	}
	t.Fatalf("view does not contain %q:\n%s", text, ansi.Strip(m.View().Content))
	return -1, -1
}

// logs.md rule 5: j shows the raw JSON of the selected event with the title
// event HH:MM:SS · raw, and j again shows the fields.
func TestLogsJTogglesTheRawJSON(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	requireView(t, m, "event 12:00:03 · WARN", "msg           network retry", "j raw JSON · r filter by this request")
	m, _ = raise(t, m, key('j', "j"))
	requireView(t, m, "event 12:00:03 · raw", `"msg": "network retry",`, "j fields · r filter by this request")
	rejectView(t, m, "msg           network retry")
	m, _ = raise(t, m, key('j', "j"))
	requireView(t, m, "event 12:00:03 · WARN", "msg           network retry")
}

// logs.md rule 7: a field with a bad name shows the error on the notice
// line and keeps the filter line open.
func TestLogsFilterWithABadFieldKeepsTheLineOpen(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, key('/', "/"))
	for _, r := range " colour:red" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "error: unknown field colour", "/ level:info colour:red █")
	if m.focusState() != "filter" {
		t.Fatalf("focus = %q after a bad field; want the open filter line", m.focusState())
	}
	if m.textFilter != defaultLogFilter {
		t.Fatalf("applied filter = %q; want %q kept", m.textFilter, defaultLogFilter)
	}
	m, _ = raise(t, m, key(0, "esc"))
	rejectView(t, m, "unknown field")
}

// logs.md rule 7: with no match, the list shows no event matches TEXT in
// dim, the strip is empty, and the field panel is empty.
func TestLogsNoMatchEmptiesTheStrip(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m.textFilter = "nothing-here"
	m.refreshLogSelection()
	m.invalidateView()
	requireView(t, m, "no event matches nothing-here", " --:-- ")
	rejectView(t, m, "▁", "▇", "┴", "event 12:00")
	row, col := findCell(t, m, "no event matches")
	if tok := viewGrid(t, m).At(row, col).Token; tok != cellgrid.Dim {
		t.Errorf("no match text token %v, want dim", tok)
	}
}

// logs.md rule 8: a day with no log file shows the text in dim, an empty
// strip, and an empty field panel. The title row names the missing file.
func TestLogsDayWithNoFile(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, day: m.logDate})
	requireView(t, m, "logs · 0 events · no log file for 2026-04-05", "← 2026-04-05 → day", "no log file for 2026-04-05", "╭─ event ─", "─ events · 0 ─")
	rejectView(t, m, "▁", "┴", "TIME", "j raw JSON")
	row, col := findCell(t, m, "│ no log file for")
	if tok := viewGrid(t, m).At(row, col+2).Token; tok != cellgrid.Dim {
		t.Errorf("no log file text token %v, want dim", tok)
	}
}

// logs.md rule 8: f with follow on turns it off, and the filter line shows
// follow ○. f again selects the newest event.
func TestLogsFollowOff(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, key('f', "f"))
	requireView(t, m, "follow ●")
	m, _ = raise(t, m, key('f', "f"))
	requireView(t, m, "follow ○", "▶ 12:00:03")
	row, col := findCell(t, m, "○  f toggles")
	if tok := viewGrid(t, m).At(row, col).Token; tok != cellgrid.Dim {
		t.Errorf("○ token %v, want dim", tok)
	}
}

// logs.md rule 9: r applies the filter req: with the request of the
// selected event. g and G select the first and the last event.
func TestLogsRequestFilterAndEnds(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, key('r', "r"))
	if m.textFilter != "req:r-22222222" {
		t.Fatalf("filter after r = %q; want req:r-22222222", m.textFilter)
	}
	requireView(t, m, "/ req:r-22222222 ", "▶ 12:00:03 WARN")
	rejectView(t, m, "queued")

	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m, _ = raise(t, m, key('g', "g"))
	requireView(t, m, "▶ 12:00:00 DEBUG")
	m, _ = raise(t, m, key('G', "G"))
	requireView(t, m, "▶ 12:00:03 WARN")
}

// logs.md rules 2 and 3: the colours of the list rows and the marker text
// of the selected event under the strip.
func TestLogsRowColours(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	m.textFilter = "level:debug"
	m.refreshLogSelection()
	m.invalidateView()
	g := viewGrid(t, m)
	check := func(text string, offset int, token cellgrid.Token, bold bool) {
		t.Helper()
		row, col := findCell(t, m, text)
		if cell := g.At(row, col+offset); cell.Token != token || cell.Bold != bold {
			t.Errorf("%q +%d = %v bold %v, want %v bold %v", text, offset, cell.Token, cell.Bold, token, bold)
		}
	}
	check("▶ 12:00:03", 0, cellgrid.Accent, true)
	check("▶ 12:00:03", 2, cellgrid.Accent, true)
	check("WARN   herdr", 0, cellgrid.Accent, true)
	check("WARN   herdr", 7, cellgrid.Dim, false)
	check("herdr      network retry", 11, cellgrid.Accent, true)
	check("12:00:00 DEBUG", 0, cellgrid.Dim, false)
	check("12:00:00 DEBUG", 9, cellgrid.Dim, false)
	check("task list", 0, cellgrid.Dim, false)
	check("INFO   task", 0, cellgrid.FG, false)
	check("queued", 0, cellgrid.FG, false)
	check("12:00 warn · network retry", 0, cellgrid.Accent, false)
	row, col := findCell(t, m, "┴")
	if cell := g.At(row, col); cell.Token != cellgrid.Accent {
		t.Errorf("┴ token %v, want accent", cell.Token)
	}
	if g.At(row+1, col).Rune != 'w' {
		t.Errorf("the level word is not under the ┴: %q", g.At(row+1, col).Rune)
	}
}

// logs.md layout: the title row gives the counts and the path of the day
// file. The state table: with the panel focus, the field panel has the
// accent border.
func TestLogsTitleRowAndPanelFocus(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	requireView(t, m, " logs · 3 events · 1 warn · ")
	rejectView(t, m, "error ·")
	row, col := findCell(t, m, "╭─ event 12:00:03")
	if tok := viewGrid(t, m).At(row, col).Token; tok != cellgrid.Border {
		t.Errorf("panel border %v without focus, want border", tok)
	}
	m, _ = raise(t, m, ctrl(']'))
	g := viewGrid(t, m)
	if tok := g.At(row, col).Token; tok != cellgrid.Accent {
		t.Errorf("panel border %v with focus, want accent", tok)
	}
	if tok := g.At(row, 0).Token; tok != cellgrid.Border {
		t.Errorf("list border %v without focus, want border", tok)
	}
}

// AGENTS.md logging: each user action on the Logs screen writes a tui
// action event.
func TestLogsLogsEachUserAction(t *testing.T) {
	m := logsModelAt(t, 120, 40)
	for _, msg := range []tea.Msg{
		key('j', "j"), key('f', "f"), key('r', "r"), key('g', "g"), key('G', "G"),
		key('/', "/"), key(0, "enter"), tea.KeyPressMsg{Code: tea.KeyRight},
	} {
		m = raiseRun(t, m, msg)
	}
	got := loggedActions(t, m.home)
	for _, want := range []string{"raw json", "follow", "filter request", "top", "bottom", "edit start", "filter apply", "log day"} {
		if !slices.Contains(got, want) {
			t.Errorf("actions %q do not include %q", got, want)
		}
	}
}

// logs.md rule 3 (case L11): the MESSAGE of an event with an error is the
// msg, then · and the error.
func TestLogsRowShowsTheError(t *testing.T) {
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	m := logsModelAt(t, 120, 40)
	line := logsFixtureEvent(t, day, "ERROR", "agent", "spawn worker", map[string]any{"error": "workspace refused"})
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		t.Fatal(err)
	}
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: []logs.Event{{Raw: []byte(line), Fields: fields}}})
	requireView(t, m, "ERROR  agent      spawn worker · workspace refused ")
}

// logs.md rule 4 and README decision 7: a value with wide runes wraps at the
// cell width and keeps each rune. The field panel does not panic and loses
// no text.
func TestLogsFieldPanelWrapsAWideValue(t *testing.T) {
	day := time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local)
	value := strings.Repeat("界", 20)
	line := logsFixtureEvent(t, day, "WARN", "task", "verify fail", map[string]any{"reason": value})
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		t.Fatal(err)
	}
	m := logsModelAt(t, 120, 40)
	m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: []logs.Event{{Raw: []byte(line), Fields: fields}}})
	view := ansi.Strip(m.View().Content)
	index := strings.Index(view, "reason        ")
	if index < 0 {
		t.Fatalf("view has no reason row:\n%s", view)
	}
	if got := strings.Count(view[index:], "界"); got != 20 {
		t.Fatalf("panel has %d wide runes; want 20:\n%s", got, view)
	}
	requireView(t, m, "reason        界界界")
}
