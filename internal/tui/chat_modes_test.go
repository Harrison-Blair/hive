package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

func TestChatModeBoardsMatch(t *testing.T) {
	for _, board := range []string{"ChatNormal", "ChatInput", "CompactChatNormal"} {
		t.Run(board, func(t *testing.T) { checkBoard(t, board) })
	}
}

// viewRow gives row of the plain view of m.
func viewRow(m *Model, row int) string {
	return strings.Split(ansi.Strip(m.View().Content), "\n")[row]
}

// requireCell checks the token of the cell at row and col.
func requireCell(t *testing.T, grid cellgrid.Grid, row, col int, char string, token cellgrid.Token) {
	t.Helper()
	cell := grid.At(row, col)
	if string(cell.Rune) != char || cell.Token != token {
		t.Errorf("cell %d,%d = %q %s, want %q %s", row, col, string(cell.Rune), cell.Token, char, token)
	}
}

func ctrlBracket() tea.KeyPressMsg { return tea.KeyPressMsg{Code: ']', Mod: tea.ModCtrl} }

func click(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func wheel(x, y int, button tea.MouseButton) tea.MouseWheelMsg {
	return tea.MouseWheelMsg{X: x, Y: y, Button: button}
}

// Golden-frames decision 19 and chat.md rule 9: Chat has no normal mode.
// The input line has the focus at start, the frame has the accent border,
// and the status row shows the chords.
func TestChatInputHasTheFocusAtStart(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	if !m.chatInputActive || m.focusState() != "input" {
		t.Fatalf("focus at start = %q, input %v, want the input line", m.focusState(), m.chatInputActive)
	}
	if got := viewRow(m, 37); !strings.HasPrefix(got, "│ > no coordinator · ^y health, r restarts it ") {
		t.Fatalf("input row = %q", got)
	}
	if got := viewRow(m, 39); !strings.Contains(got, " · ^g agents  ^t work  ^f friction  ^y health  ^l logs · ? help ") {
		t.Fatalf("status row = %q", got)
	}
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, 36, 0, "╭", cellgrid.Accent)
	requireCell(t, grid, 37, 2, ">", cellgrid.User)
	requireCell(t, grid, 37, 4, "n", cellgrid.Dim)
}

// Golden-frames decision 10: below 120 columns the status row cuts segments
// from the right and keeps the clock.
func TestChatCompactStatusRowKeepsTheClock(t *testing.T) {
	m := chatModel(t, 80, 24, 1)
	got := viewRow(m, 23)
	if !strings.HasPrefix(got, " ○ none · ○ supervisor · ? 0 asks · needs you 0 ") || !strings.HasSuffix(got, " 08:00 ") {
		t.Fatalf("compact status row = %q", got)
	}
	if strings.Contains(got, "^g agents") {
		t.Fatalf("compact status row = %q, want the chords cut", got)
	}
}

// Decision 19 and the key table of chat.md: printable keys type, also i, j,
// k, g, and G. esc, tab, and ctrl+] do nothing. enter keeps the focus.
func TestChatInputModeKeys(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	for _, press := range []tea.KeyPressMsg{
		key('x', "x"), {Code: tea.KeyEscape}, {Code: tea.KeyTab}, ctrlBracket(),
		key('i', "i"), key('j', "j"), key('k', "k"), key('g', "g"), key('G', "G"),
	} {
		m, _ = raise(t, m, press)
		if !m.chatInputActive {
			t.Fatalf("%s took the focus from the input line", press)
		}
	}
	if got := m.chatInput.Value(); got != "xijkgG" {
		t.Fatalf("draft = %q, want xijkgG", got)
	}
	if got := viewRow(m, 37); !strings.HasPrefix(got, "│ > xijkgG█ ") {
		t.Fatalf("input row = %q", got)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.chatInputActive {
		t.Fatal("enter took the focus from the input line")
	}
}

// Rule 13: ? on an empty input line opens Help; on a line with text it
// types. A chord opens its screen while text is in the line.
func TestChatQuestionMarkAndChordsFromTheInputLine(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	m, _ = raise(t, m, key('?', "?"))
	if !m.helpOpen {
		t.Fatal("? on the empty line did not open help")
	}
	m, _ = raise(t, m, key('?', "?"))
	m, _ = raise(t, m, key('a', "a"))
	m, _ = raise(t, m, key('?', "?"))
	if m.helpOpen || m.chatInput.Value() != "a?" {
		t.Fatalf("help %v, draft %q, want no help and a?", m.helpOpen, m.chatInput.Value())
	}
	m, _ = raise(t, m, ctrl('t'))
	if m.activeTab() != tabWork {
		t.Fatalf("^t gave tab %d", m.activeTab())
	}
	if got := m.chatInput.Value(); got != "a?" {
		t.Fatalf("draft = %q, want a? kept", got)
	}
}

// Decision 19: a click on the transcript does not take the focus from the
// input line.
func TestChatClicksKeepTheInputFocus(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	m, _ = raise(t, m, click(30, 10))
	if !m.chatInputActive {
		t.Fatal("a click on the transcript took the focus")
	}
	m, _ = raise(t, m, click(30, 37))
	if !m.chatInputActive {
		t.Fatal("a click on the input line took the focus")
	}
}

// Rule 7: the wheel over the transcript scrolls it and keeps the focus.
func TestChatWheelScrollsTheFeed(t *testing.T) {
	m := chatModel(t, 120, 40, 40)
	requireView(t, m, "line-40")
	for range 10 {
		m, _ = raise(t, m, wheel(30, 10, tea.MouseWheelUp))
	}
	rejectView(t, m, "line-40")
	for range 10 {
		m, _ = raise(t, m, wheel(30, 10, tea.MouseWheelDown))
	}
	requireView(t, m, "line-40")

	for range 10 {
		m, _ = raise(t, m, wheel(30, 10, tea.MouseWheelUp))
	}
	rejectView(t, m, "line-40")
	if !m.chatInputActive {
		t.Fatal("the wheel took the focus from the input line")
	}
}

// Rule 7 and C10: a transcript taller than its rows follows its end. The
// last message is on the last transcript row, row 0 has the earlier row
// from column 52, and the ┃ thumb on column 119 ends on the last row.
func TestChatTallTranscriptFollowsItsEnd(t *testing.T) {
	m := chatModel(t, 120, 40, 40)
	rows := viewRows(m)
	if got := rows[35]; !strings.HasPrefix(got, " line-40 ") {
		t.Fatalf("row 35 = %q, want the last message line", got)
	}
	if !strings.HasPrefix(rows[36], "╭") {
		t.Fatalf("row 36 = %q, want the composer frame", rows[36])
	}
	// 35 rows below row 0 hold 12 messages of 2 rows with a blank row
	// between each two, so 28 messages are earlier.
	if got := rows[1]; !strings.HasPrefix(got, " ● coordinator ") {
		t.Fatalf("row 1 = %q, want the first line of message 29", got)
	}
	head := rows[0]
	if col := strings.Index(head, "·"); col != 52 || !strings.Contains(head, "· · · 28 earlier · · ·") {
		t.Fatalf("row 0 = %q (· at %d), want the earlier row from column 52", head, col)
	}
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, 35, 119, "┃", cellgrid.Dim)
	requireCell(t, grid, 0, 52, "·", cellgrid.Dim)
	for row := 1; row < 36; row++ {
		if strings.HasSuffix(rows[row], "┃") {
			return
		}
	}
	t.Fatal("no thumb on column 119")
}

// selectedRow gives the row of the view of m with the marker ▶ at column 2.
func selectedRow(m *Model) int {
	for row, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if strings.HasPrefix(line, "│ ▶") {
			return row
		}
	}
	return -1
}

// The wheel over a list on another tab moves its selection. The wheel over
// the details scrolls them and keeps the focus. Golden-frames decision 2:
// the list starts one row higher than in tui-v2.
func TestMouseWheelScrollsTheListUnderThePointer(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	if got := selectedRow(m); got != 9 {
		t.Fatalf("selected row = %d, want 9", got)
	}
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelDown))
	if got := selectedRow(m); got != 10 {
		t.Fatalf("selected row after wheel down = %d, want 10", got)
	}
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelUp))
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelUp))
	if got := selectedRow(m); got != 8 {
		t.Fatalf("selected row after wheel up = %d, want 8", got)
	}
	// The details of the first selected entry have more lines than the
	// region, so they can scroll.
	m = newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	before := m.detailsOffset[tabFriction]
	m, _ = raise(t, m, wheel(80, 12, tea.MouseWheelDown))
	if m.detailsOffset[tabFriction] <= before {
		t.Fatalf("details offset = %d, want more than %d", m.detailsOffset[tabFriction], before)
	}
	if m.regionFocus[tabFriction] != 0 {
		t.Fatal("the wheel moved the focus")
	}
}

// The wheel over the Logs event list moves its selection.
func TestMouseWheelScrollsTheLogsEvents(t *testing.T) {
	m := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	before := selectedRow(m)
	m, _ = raise(t, m, wheel(10, 20, tea.MouseWheelUp))
	if got := selectedRow(m); got != before-1 {
		t.Fatalf("selected row = %d, want %d", got, before-1)
	}
}

// The wheel over the event list moves its selection while the Logs filter is
// open. The filter keeps the focus and its draft.
func TestMouseWheelWorksWhileTheLogsFilterIsOpen(t *testing.T) {
	m := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: 'x', Text: "x"})
	if got := m.focusState(); got != "filter" {
		t.Fatalf("focus = %q, want filter", got)
	}
	draft := m.filterDraft
	before := selectedRow(m)
	m, _ = raise(t, m, wheel(10, 20, tea.MouseWheelDown))
	if got := selectedRow(m); got != before+1 {
		t.Fatalf("selected row = %d, want %d", got, before+1)
	}
	if got := m.focusState(); got != "filter" || m.filterDraft != draft {
		t.Fatalf("after the wheel: focus %q, draft %q, want filter, %q", got, m.filterDraft, draft)
	}
}

// Golden-frames decision 2: no tab bar exists. A click on the rule, the row
// of the tui-v2 tab names, does not change the screen while the Logs filter
// is open.
func TestMouseClickOnATabNameWorksWhileTheLogsFilterIsOpen(t *testing.T) {
	m := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = raise(t, m, click(22, 1))
	if m.activeTab() != tabLogs || m.focusState() != "filter" {
		t.Fatalf("a click on the rule gave tab %d, focus %q; want Logs, filter", m.activeTab(), m.focusState())
	}
}

// The wheel over a region that cannot scroll does nothing.
func TestMouseWheelOverTheAgentsGraphDoesNothing(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	before := m.View().Content
	m, _ = raise(t, m, wheel(60, 20, tea.MouseWheelDown))
	if m.View().Content != before {
		t.Fatal("the wheel changed the Agents graph")
	}
}

// A click on a list row selects it. A click on the details gives them the
// focus. Golden-frames decision 2: the list starts one row higher than in
// tui-v2, so each row is one less.
func TestMouseClickSelectsAListRow(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	m, _ = raise(t, m, click(10, 6))
	if got := selectedRow(m); got != 6 {
		t.Fatalf("selected row = %d, want 6", got)
	}
	requireView(t, m, "╭─ agent-can-send-user-chat ")
	m, _ = raise(t, m, click(10, 15))
	if got := selectedRow(m); got != 15 {
		t.Fatalf("selected row = %d, want 15", got)
	}
	m, _ = raise(t, m, click(10, 30))
	if got := selectedRow(m); got != 19 {
		t.Fatalf("a click below the last row moved the selection to %d, want 19", got)
	}
	m, _ = raise(t, m, click(80, 12))
	if m.regionFocus[tabFriction] != 1 {
		t.Fatal("a click on the details did not give them the focus")
	}
	m, _ = raise(t, m, click(10, 8))
	if m.regionFocus[tabFriction] != 0 || selectedRow(m) != 8 {
		t.Fatalf("a click on the list: focus %d, row %d", m.regionFocus[tabFriction], selectedRow(m))
	}
}

// Golden-frames decision 2: no tab bar exists, so a click on row 1 of a
// screen, the rule, does not change the screen, at both widths.
func TestMouseClickOnATabNameOpensTheTab(t *testing.T) {
	for _, width := range []int{120, 80} {
		m := chatModel(t, width, 40, 1)
		m, _ = raise(t, m, ctrl('t'))
		for x := 0; x < width; x += 7 {
			m, _ = raise(t, m, click(x, 1))
			if m.activeTab() != tabWork {
				t.Fatalf("width %d: a click at %d,1 gave screen %d; want Work", width, x, m.activeTab())
			}
		}
	}
}

// Golden-frames decision 11: the status row shows the ask count of the
// header state, at both widths.
func TestHeaderShowsTheAskCountOfTheState(t *testing.T) {
	for _, tc := range []struct {
		asks, width int
		want        string
	}{
		{0, 120, " · ? 0 asks · needs you 0 · "},
		{1, 120, " · ? 1 ask · needs you 0 · "},
		{0, 80, " · ? 0 asks · needs you 0 "},
		{1, 80, " · ? 1 ask · needs you 0 "},
	} {
		m := chatModel(t, tc.width, 40, 1)
		m.now = func() time.Time { return time.Date(2026, 10, 20, 8, 0, 0, 0, time.Local) }
		m.header.asks = tc.asks
		m.invalidateView()
		rows := viewRows(m)
		if got := rows[len(rows)-1]; !strings.Contains(got, tc.want) || !strings.HasSuffix(got, " 08:00 ") {
			t.Errorf("asks %d at %d: status row = %q, want %q and the clock 08:00", tc.asks, tc.width, got, tc.want)
		}
	}
}

// The program asks the terminal for click and wheel reports. Cell motion
// mode leaves shift and drag to the terminal for copy.
func TestViewTurnsOnMouseReports(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	if got := m.View().MouseMode; got != tea.MouseModeCellMotion {
		t.Fatalf("mouse mode = %v, want cell motion", got)
	}
}
