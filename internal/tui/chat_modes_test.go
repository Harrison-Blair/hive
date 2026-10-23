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

// Normal mode is the default. The feed has the accent border, and the input
// has the border token and the dim placeholder.
func TestChatNormalModeIsTheDefault(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	if m.chatInputActive {
		t.Fatal("chat starts in input mode")
	}
	if got := viewRow(m, 39); got != " [normal] i type · ↑↓ j/k scroll · wheel scroll · pgup/pgdn page · g/G top/end · ctrl+o events · ? help                 " {
		t.Fatalf("footer = %q", got)
	}
	if got := viewRow(m, 37); !strings.HasPrefix(got, "│ i or click here to type   ") {
		t.Fatalf("input row = %q", got)
	}
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, 3, 0, "╭", cellgrid.Accent)
	requireCell(t, grid, 36, 0, "╭", cellgrid.Border)
	requireCell(t, grid, 37, 2, "i", cellgrid.Dim)
	requireCell(t, grid, 37, 24, "e", cellgrid.Dim)
	requireCell(t, grid, 39, 1, "[", cellgrid.Accent)
}

func TestChatCompactFooterShowsTheModes(t *testing.T) {
	m := chatModel(t, 80, 24, 1)
	if got := viewRow(m, 23); got != " [normal] i type · ↑↓ wheel scroll · pgup/pgdn · ctrl+o events · ? help         " {
		t.Fatalf("compact normal footer = %q", got)
	}
	m, _ = raise(t, m, key('i', "i"))
	if got := viewRow(m, 23); !strings.HasPrefix(got, " [input] enter send · ctrl+] leave · wheel or pgup/pgdn scroll") {
		t.Fatalf("compact input footer = %q", got)
	}
}

// i starts input mode. esc and tab keep it. enter sends and keeps it.
// ctrl+] stops it.
func TestChatInputModeKeys(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	m, _ = raise(t, m, key('i', "i"))
	if !m.chatInputActive {
		t.Fatal("i did not start input mode")
	}
	if got := viewRow(m, 39); got != " [input] enter send · shift+enter newline · ctrl+] leave input · ctrl+o events · wheel or pgup/pgdn scroll              " {
		t.Fatalf("footer = %q", got)
	}
	if got := viewRow(m, 37); !strings.HasPrefix(got, "│ > █ ") {
		t.Fatalf("input row = %q", got)
	}
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, 3, 0, "╭", cellgrid.Border)
	requireCell(t, grid, 36, 0, "╭", cellgrid.Accent)
	if cell := grid.At(36, 3); string(cell.Rune) != "m" || !cell.Bold || cell.Token != cellgrid.Accent {
		t.Errorf("input title cell = %q %s bold %v, want accent bold m", string(cell.Rune), cell.Token, cell.Bold)
	}

	for _, press := range []tea.KeyPressMsg{key('x', "x"), {Code: tea.KeyEscape}, {Code: tea.KeyTab}, key('j', "j")} {
		m, _ = raise(t, m, press)
		if !m.chatInputActive {
			t.Fatalf("%s stopped input mode", press)
		}
	}
	if got := m.chatInput.Value(); got != "xj" {
		t.Fatalf("draft = %q, want xj", got)
	}
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || !m.chatSending {
		t.Fatal("enter did not send")
	}
	if !m.chatInputActive {
		t.Fatal("enter stopped input mode")
	}
	m, _ = raise(t, m, ctrlBracket())
	if m.chatInputActive {
		t.Fatal("ctrl+] did not stop input mode")
	}
	if got := viewRow(m, 39); !strings.HasPrefix(got, " [normal] ") {
		t.Fatalf("footer = %q", got)
	}
}

// In normal mode a letter keeps its shortcut and does not go into the input.
// esc, tab, and enter do not start input mode.
func TestChatNormalModeLettersStayShortcuts(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	for _, press := range []tea.KeyPressMsg{key('x', "x"), key('a', "a"), {Code: tea.KeyEscape}, {Code: tea.KeyTab}, {Code: tea.KeyEnter}} {
		m, _ = raise(t, m, press)
		if m.chatInputActive {
			t.Fatalf("%s started input mode", press)
		}
	}
	if got := m.chatInput.Value(); got != "" {
		t.Fatalf("draft = %q, want empty", got)
	}
	m, _ = raise(t, m, key('?', "?"))
	if !m.helpOpen {
		t.Fatal("? did not open help")
	}
	m, _ = raise(t, m, key('?', "?"))
	m, _ = raise(t, m, key('3', "3"))
	if m.activeTab() != tabWork {
		t.Fatalf("3 gave tab %d", m.activeTab())
	}
	if got := m.chatInput.Value(); got != "" {
		t.Fatalf("draft = %q, want empty", got)
	}
}

// A click on the input starts input mode. A click on the feed stops it.
func TestChatClicksChangeTheMode(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	m, _ = raise(t, m, click(30, 37))
	if !m.chatInputActive {
		t.Fatal("a click on the input did not start input mode")
	}
	m, _ = raise(t, m, click(30, 10))
	if m.chatInputActive {
		t.Fatal("a click on the feed did not stop input mode")
	}
	m, _ = raise(t, m, click(30, 10))
	if m.chatInputActive {
		t.Fatal("a second click on the feed started input mode")
	}
}

// The wheel over the feed scrolls it, in normal mode and in input mode.
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

	m, _ = raise(t, m, key('i', "i"))
	for range 10 {
		m, _ = raise(t, m, wheel(30, 10, tea.MouseWheelUp))
	}
	rejectView(t, m, "line-40")
	if !m.chatInputActive {
		t.Fatal("the wheel stopped input mode")
	}
}

// The feed title shows the count of earlier messages also below 120
// columns.
func TestCompactChatFeedShowsTheEarlierNote(t *testing.T) {
	m := newTestDemoModel(t, "CompactChatNormal.fixture.json", demoSize{})
	if got := viewRow(m, 3); !strings.Contains(got, " ↑ 42 earlier ") || got[strings.Index(got, " ↑"):] == "" {
		t.Fatalf("feed title = %q", got)
	}
	if col := ansi.StringWidth(viewRow(m, 3)[:strings.Index(viewRow(m, 3), " ↑")]); col != 40 {
		t.Fatalf("note starts at %d, want 40", col)
	}
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
// the details scrolls them and keeps the focus.
func TestMouseWheelScrollsTheListUnderThePointer(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	if got := selectedRow(m); got != 10 {
		t.Fatalf("selected row = %d, want 10", got)
	}
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelDown))
	if got := selectedRow(m); got != 11 {
		t.Fatalf("selected row after wheel down = %d, want 11", got)
	}
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelUp))
	m, _ = raise(t, m, wheel(10, 12, tea.MouseWheelUp))
	if got := selectedRow(m); got != 9 {
		t.Fatalf("selected row after wheel up = %d, want 9", got)
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

// A click on a tab name opens that tab while the Logs filter is open.
func TestMouseClickOnATabNameWorksWhileTheLogsFilterIsOpen(t *testing.T) {
	m := newTestDemoModel(t, "Logs.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = raise(t, m, click(22, 1))
	if m.activeTab() != tabWork {
		t.Fatalf("a click on 3 Work gave tab %d", m.activeTab())
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
// focus.
func TestMouseClickSelectsAListRow(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	m, _ = raise(t, m, click(10, 7))
	if got := selectedRow(m); got != 7 {
		t.Fatalf("selected row = %d, want 7", got)
	}
	requireView(t, m, "╭─ agent-can-send-user-chat ")
	m, _ = raise(t, m, click(10, 16))
	if got := selectedRow(m); got != 16 {
		t.Fatalf("selected row = %d, want 16", got)
	}
	m, _ = raise(t, m, click(10, 30))
	if got := selectedRow(m); got != 20 {
		t.Fatalf("a click below the last row moved the selection to %d, want 20", got)
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

// A click on a tab name opens the tab, at both widths.
func TestMouseClickOnATabNameOpensTheTab(t *testing.T) {
	for _, width := range []int{120, 80} {
		m := chatModel(t, width, 40, 1)
		row := viewRow(m, 1)
		for tab, label := range []string{"1 Agents", "2 Chat", "3 Work", "4 Friction", "5 Health", "6 Logs"} {
			x := ansi.StringWidth(row[:strings.Index(row, label)]) + 2
			m, _ = raise(t, m, click(x, 1))
			if m.activeTab() != tab {
				t.Fatalf("width %d: a click on %q gave tab %d", width, label, m.activeTab())
			}
		}
		m, _ = raise(t, m, click(width-2, 1))
		if m.activeTab() != tabLogs {
			t.Fatalf("width %d: a click after the last label gave tab %d", width, m.activeTab())
		}
	}
}

// The header shows the ask count of the header state, at both widths.
func TestHeaderShowsTheAskCountOfTheState(t *testing.T) {
	for _, tc := range []struct {
		asks, width int
		want        string
	}{
		{0, 120, "   ? 0 asks   needs you 0   08:00 "},
		{1, 120, "   ? 1 ask   needs you 0   08:00 "},
		{0, 80, " · ? 0 · needs you 0 · 08:00 "},
		{1, 80, " · ? 1 · needs you 0 · 08:00 "},
	} {
		m := chatModel(t, tc.width, 40, 1)
		m.now = func() time.Time { return time.Date(2026, 10, 20, 8, 0, 0, 0, time.Local) }
		m.header.asks = tc.asks
		m.invalidateView()
		if got := viewRow(m, 0); !strings.HasSuffix(got, tc.want) {
			t.Errorf("asks %d at %d: header = %q, want suffix %q", tc.asks, tc.width, got, tc.want)
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
