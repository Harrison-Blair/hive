package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

// readBoardGrid reads the golden grid of one board.
func readBoardGrid(t *testing.T, board string) cellgrid.Grid {
	t.Helper()
	file, err := os.Open(filepath.Join(designDir, board+".grid"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	grid, err := cellgrid.ReadGrid(file)
	if err != nil {
		t.Fatal(err)
	}
	return grid
}

// renderGrid converts the current view of m to a grid of width×height cells.
func renderGrid(t *testing.T, m *Model, width, height int) cellgrid.Grid {
	t.Helper()
	grid, err := cellgrid.Convert(m.View().Content, width, height)
	if err != nil {
		t.Fatalf("convert: %v\n%s", err, m.View().Content)
	}
	return grid
}

// renderRows converts only the rows of the view of m in rows, in order. The
// content area of a current screen can use colors outside the palette until
// wave 6, so the frame check converts only the frame rows.
func renderRows(t *testing.T, m *Model, width int, rows ...int) cellgrid.Grid {
	t.Helper()
	lines := strings.Split(m.View().Content, "\n")
	out := cellgrid.Grid{Width: width, Height: len(rows)}
	for _, row := range rows {
		grid, err := cellgrid.Convert(lines[row], width, 1)
		if err != nil {
			t.Fatalf("convert row %d: %v\n%q", row, err, lines[row])
		}
		out.Cells = append(out.Cells, grid.Cells...)
	}
	return out
}

// gridRows gives the rows of g in rows, in order.
func gridRows(g cellgrid.Grid, rows ...int) cellgrid.Grid {
	out := cellgrid.Grid{Width: g.Width, Height: len(rows)}
	for _, row := range rows {
		out.Cells = append(out.Cells, g.Cells[row*g.Width:(row+1)*g.Width]...)
	}
	return out
}

// checkFrameRows compares the header, the tab bar, and the footer rows of the
// demo view of board with the same rows of its golden grid.
func checkFrameRows(t *testing.T, board string) {
	t.Helper()
	want := readBoardGrid(t, board)
	m := newTestDemoModel(t, board+".fixture.json", demoSize{})
	rows := []int{0, 1, 2, want.Height - 1}
	got := renderRows(t, m, want.Width, rows...)
	diffs, err := cellgrid.Compare(gridRows(want, rows...), got, nil)
	if err != nil {
		t.Fatal(err)
	}
	for index := range diffs {
		diffs[index].Row = rows[diffs[index].Row]
	}
	if len(diffs) != 0 {
		t.Fatalf("%d frame cells differ from %s.grid:\n%s", len(diffs), board, cellgrid.FormatDiffs(diffs))
	}
}

func TestFrameRowsMatchMain(t *testing.T) {
	checkFrameRows(t, "Main")
}

// The startup notice takes the footer before an error, as the old footer
// did. The error then shows when the notice is empty.
func TestFooterShowsStartupNoticeBeforeError(t *testing.T) {
	m := newTestModel(t, newTestHome(t), fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.startupNotice = "not in herdr: no coordinator or supervisor started"
	m.chatErr = "chat send is for the user only"
	m.invalidateView()
	requireView(t, m, "[feed] not in herdr: no coordinator or supervisor started")
	m.startupNotice = ""
	m.invalidateView()
	requireView(t, m, "[feed] error: chat send is for the user only")
}

// screenBoards are the boards of one screen each. Their header, tab bar, and
// footer rows are the expected frame (AC5).
var screenBoards = []string{
	"AgentDetails", "AgentsMany", "Chat", "Compact", "CompactAgents", "CompactChat",
	"CompactFriction", "CompactHealth", "CompactLogs", "Friction", "FrictionResolved",
	"Health", "HealthAgents", "Logs", "Main", "Work", "WorkGoals", "WorkQueue",
}

func TestFrameRowsMatchEachScreenBoard(t *testing.T) {
	for _, board := range screenBoards {
		t.Run(board, func(t *testing.T) { checkFrameRows(t, board) })
	}
}

// The fixture UI state sets the tab, the view, the focus, the overlay, the
// popup, and the restart prompt of the frame.
func TestDemoUIStateSetsTheFrameState(t *testing.T) {
	for _, tt := range []struct {
		board, focus string
		tab, view    int
	}{
		{"AgentDetails", "details", tabAgents, 0},
		{"Chat", "input", tabChat, 0},
		{"Compact", "details", tabWork, 2},
		{"FrictionResolved", "list", tabFriction, 1},
		{"HealthAgents", "restart", tabHealth, 1},
		{"Help", "help", tabAgents, 0},
		{"Logs", "json", tabLogs, 0},
		{"WorkQueue", "list", tabWork, 1},
	} {
		t.Run(tt.board, func(t *testing.T) {
			m := newTestDemoModel(t, tt.board+".fixture.json", demoSize{})
			if m.activeTab() != tt.tab || m.activeView() != tt.view || m.focusState() != tt.focus {
				t.Fatalf("tab %d, view %d, focus %q; want %d, %d, %q", m.activeTab(), m.activeView(), m.focusState(), tt.tab, tt.view, tt.focus)
			}
		})
	}
}

// y in the restart prompt of demo mode shows the demo notice and restarts
// no agent.
func TestDemoRestartPromptYesDoesNotRestart(t *testing.T) {
	m := newTestDemoModel(t, "HealthAgents.fixture.json", demoSize{})
	m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if m.agents.restartRunning {
		t.Fatal("y in demo mode started a restart")
	}
	requireView(t, m, demoNotice)
}

// The Help view matches its board (AC5). The header, the tab bar, the
// footer, and the help box match cell for cell. The box of the board is in
// rows 6-34 and columns 18-101. Each other cell of the content area,
// columns 17 and 102 included, matches only in its token, because the
// Agents screen behind the box is wave 6 work.
func TestHelpMatchesItsBoard(t *testing.T) {
	want := readBoardGrid(t, "Help")
	m := newTestDemoModel(t, "Help.fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	var diffs []cellgrid.Diff
	for row := range want.Height {
		for col := range want.Width {
			w, g := want.At(row, col), got.At(row, col)
			frame := row <= 2 || row == want.Height-1
			box := row >= 6 && row <= 34 && col >= 18 && col <= 101
			if frame || box {
				if w != g {
					diffs = append(diffs, cellgrid.Diff{Row: row, Col: col, Want: w, Got: g})
				}
			} else if w.Token != g.Token {
				diffs = append(diffs, cellgrid.Diff{Row: row, Col: col, Want: w, Got: g})
			}
		}
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from Help.grid:\n%s", len(diffs), cellgrid.FormatDiffs(diffs))
	}
}

// frameScreens gives each tab and view of the registry.
func frameScreens() [][2]int {
	var out [][2]int
	for tab := range tabCount {
		views := max(len(screenRegistry[tab].views), 1)
		for view := range views {
			out = append(out, [2]int{tab, view})
		}
	}
	return out
}

// requireWholeBorders fails when a line of view has a border corner without
// its pair: each ╭ needs a ╮, and each ╰ needs a ╯, in the same line.
func requireWholeBorders(t *testing.T, view string) {
	t.Helper()
	for index, line := range strings.Split(view, "\n") {
		if strings.Count(line, "╭") != strings.Count(line, "╮") || strings.Count(line, "╰") != strings.Count(line, "╯") {
			t.Fatalf("line %d has a broken border: %q\n%s", index+1, line, view)
		}
	}
}

// AC3 and AC5: each screen at 79×24 shows the too-small notice. At 80×24,
// 119×40, and 120×40, each screen fills the pane, and no border breaks.
func TestEachScreenKeepsTheFrameAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, screen := range frameScreens() {
			name := fmt.Sprintf("%dx%d/%s/%d", size.Width, size.Height, screenRegistry[screen[0]].label, screen[1])
			t.Run(name, func(t *testing.T) {
				m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
				m.selectScreen(screen[0], screen[1])
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
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
			})
		}
	}
}

// focusTokens gives the token of the top-left corner of the list panel and
// of the details panel in the first content row of a 120×40 Health view.
func focusTokens(t *testing.T, m *Model) (cellgrid.Token, cellgrid.Token) {
	t.Helper()
	row := renderRows(t, m, 120, 4)
	listWidth := 120 * 13 / 30
	return row.At(0, 0).Token, row.At(0, listWidth).Token
}

// AC2: the focused region has the accent border, and tab moves the focus.
func TestTabMovesTheAccentBorder(t *testing.T) {
	m := newTestDemoModel(t, "Health.fixture.json", demoSize{})
	if list, details := focusTokens(t, m); list != cellgrid.Accent || details != cellgrid.Border {
		t.Fatalf("list focus: list border %s, details border %s; want accent, border", list, details)
	}
	requireView(t, m, "[list]")
	m, _ = raise(t, m, key(0, "tab"))
	if list, details := focusTokens(t, m); list != cellgrid.Border || details != cellgrid.Accent {
		t.Fatalf("details focus: list border %s, details border %s; want border, accent", list, details)
	}
	requireView(t, m, "[details]")
	m, _ = raise(t, m, key(0, "tab"))
	if m.focusState() != "list" {
		t.Fatalf("second tab gave focus %q; want list", m.focusState())
	}
}

// AC4: ? opens the help overlay, ↑ and ↓ scroll it, and ? or esc closes it.
// Help does not list --demo.
func TestHelpKeys(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	m, _ = raise(t, m, key('?', "?"))
	requireView(t, m, "keys · ? or esc closes", "[help]")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	if m.helpOffset != 1 {
		t.Fatalf("help offset = %d after down down up; want 1", m.helpOffset)
	}
	m, _ = raise(t, m, key('2', "2"))
	if m.activeTab() != tabAgents || !m.helpOpen {
		t.Fatal("a screen key changed the screen behind the help overlay")
	}
	m, _ = raise(t, m, key('?', "?"))
	rejectView(t, m, "keys · ? or esc closes")
	m, _ = raise(t, m, key('?', "?"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	rejectView(t, m, "keys · ? or esc closes")
	if text := strings.Join(helpLines(), "\n"); strings.Contains(text, "--demo") {
		t.Fatalf("help lists --demo:\n%s", text)
	}
}

// AC3: below 120 columns enter opens the details as a full-width overlay,
// and esc closes it. A resize keeps the overlay and its scroll position.
func TestDetailsOverlayBelow120Columns(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	rejectView(t, m, "esc back")
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, " · esc back", "[details]")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.focusState() != "details" || m.detailsOffset[tabFriction] != 1 {
		t.Fatalf("resize gave focus %q, offset %d; want details, 1", m.focusState(), m.detailsOffset[tabFriction])
	}
	requireView(t, m, " · esc back")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	rejectView(t, m, "esc back")
	requireView(t, m, "[list]")
}

// AC2: in each focus state of decision 21 that does not keep the global
// keys, a screen key, a view key, or tab does not change the screen or the
// view. An overlay, a popup, help, and the filter bar also keep the focus.
func TestOverlayKeepsScreenAndViewKeys(t *testing.T) {
	compact := tea.WindowSizeMsg{Width: 80, Height: 24}
	enter := key(0, "enter")
	for _, tt := range []struct {
		name, board, focus string
		keepFocus          bool
		keys               []tea.Msg
	}{
		{"help", "Help", "help", true, nil},
		{"popup", "AgentDetails", "details", true, nil},
		{"work overlay", "Compact", "details", true, nil},
		{"friction overlay", "Friction", "details", true, []tea.Msg{compact, enter}},
		{"health overlay", "Health", "details", true, []tea.Msg{compact, enter}},
		{"logs overlay", "CompactLogs", "json", true, nil},
		{"filter", "Logs", "filter", true, []tea.Msg{tea.KeyPressMsg{Code: tea.KeyEscape}, key('/', "/")}},
		{"input", "Chat", "input", false, nil},
		{"restart", "HealthAgents", "restart", false, nil},
	} {
		for _, press := range []tea.KeyPressMsg{key('1', "1"), key('2', "2"), key('3', "3"), key('4', "4"), key('5', "5"), key('6', "6"), key('[', "["), key(']', "]"), {Code: tea.KeyTab}} {
			t.Run(tt.name+"/"+press.String(), func(t *testing.T) {
				m := newTestDemoModel(t, tt.board+".fixture.json", demoSize{})
				for _, msg := range tt.keys {
					m, _ = raise(t, m, msg)
				}
				if m.focusState() != tt.focus {
					t.Fatalf("focus = %q; want %q", m.focusState(), tt.focus)
				}
				tab, view := m.activeTab(), m.activeView()
				m, _ = raise(t, m, press)
				if m.activeTab() != tab || m.activeView() != view {
					t.Fatalf("key %q changed screen %d view %d to screen %d view %d", press.String(), tab, view, m.activeTab(), m.activeView())
				}
				if tt.keepFocus && m.focusState() != tt.focus {
					t.Fatalf("key %q changed focus %q to %q", press.String(), tt.focus, m.focusState())
				}
			})
		}
	}
}

// AC2: an action key of the overlay still reaches its screen. In the Work
// Tasks overlay, d opens the dependency graph.
func TestOverlayGivesActionKeysToTheScreen(t *testing.T) {
	m := newTestDemoModel(t, "Compact.fixture.json", demoSize{})
	m, _ = raise(t, m, key('d', "d"))
	if !m.tasks.fileOpen || !strings.HasPrefix(m.tasks.fileTitle, "dependency graph") {
		t.Fatalf("d in the Work overlay gave open %t, title %q; want the dependency graph", m.tasks.fileOpen, m.tasks.fileTitle)
	}
	if m.focusState() != "details" {
		t.Fatalf("focus = %q after b; want details", m.focusState())
	}
}

// AC2: from 120 columns the details region sits beside the list, so the
// screen keys stay global (decision 21).
func TestWideDetailsKeepsScreenKeys(t *testing.T) {
	m := newTestDemoModel(t, "Work.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.focusState() != "details" {
		t.Fatalf("focus = %q; want details", m.focusState())
	}
	m, _ = raise(t, m, key('1', "1"))
	if m.activeTab() != tabAgents {
		t.Fatalf("1 in the wide details gave screen %d; want Agents", m.activeTab())
	}
}

// AC2: ctrl+c two times quits in each focus state. The first ctrl+c shows
// the quit notice.
func TestCtrlCTwiceQuitsInEachFocusState(t *testing.T) {
	for _, tt := range []struct {
		name, board, focus string
		keys               []tea.KeyPressMsg
	}{
		{"graph", "Main", "graph", nil},
		{"feed", "CompactChat", "feed", []tea.KeyPressMsg{{Code: tea.KeyEscape}}},
		{"input", "Chat", "input", nil},
		{"filter", "Logs", "filter", []tea.KeyPressMsg{{Code: tea.KeyEscape}, {Code: '/', Text: "/"}}},
		{"help", "Help", "help", nil},
		{"overlay", "Compact", "details", nil},
		{"popup", "AgentDetails", "details", nil},
		{"restart", "HealthAgents", "restart", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestDemoModel(t, tt.board+".fixture.json", demoSize{})
			for _, press := range tt.keys {
				m, _ = raise(t, m, press)
			}
			if m.focusState() != tt.focus {
				t.Fatalf("focus = %q; want %q", m.focusState(), tt.focus)
			}
			m, cmd := raise(t, m, ctrlC())
			if cmd != nil {
				if _, ok := cmd().(tea.QuitMsg); ok {
					t.Fatal("first ctrl+c quit")
				}
			}
			requireView(t, m, "press ctrl+c again to quit")
			_, cmd = raise(t, m, ctrlC())
			quitMessage(t, cmd)
		})
	}
}

// AC6: normal mode uses the frame of demo mode, and below 80×24 it shows the
// too-small notice.
func TestNormalModeUsesTheFrame(t *testing.T) {
	m := newTestModel(t, newTestHome(t), fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if !strings.HasPrefix(lines[0], " hive ") || !strings.Contains(lines[1], "1 Agents   2 Chat") || !strings.HasPrefix(lines[39], " [feed] ") {
		t.Fatalf("normal mode does not draw the frame:\n%s", strings.Join(lines, "\n"))
	}
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 23})
	requireView(t, m, "terminal too small", "hive needs at least 80×24")
}
