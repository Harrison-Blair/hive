package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/home"
)

const helpTitle = "help · esc or ? closes · ↑↓ scroll"

// openHelpAt gives a frame model at width×height with the Help overlay open.
func openHelpAt(t *testing.T, width, height int) (*Model, string) {
	t.Helper()
	m, dir := newFrameModel(t, width, height)
	m, cmd := raise(t, m, key('?', "?"))
	drain(cmd)
	if !m.helpOpen {
		t.Fatal("? in an empty composer did not open Help")
	}
	return m, dir
}

// cellOf gives the row and the column of the first cell of text in rows, or
// -1, -1.
func cellOf(rows []string, text string) (int, int) {
	for index, row := range rows {
		if at := strings.Index(row, text); at >= 0 {
			return index, ansi.StringWidth(row[:at])
		}
	}
	return -1, -1
}

// helpBox gives the rows of the top and the bottom border of the Help box.
func helpBox(t *testing.T, rows []string) (int, int) {
	t.Helper()
	top, _ := cellOf(rows, "╭─ "+helpTitle)
	if top < 0 {
		t.Fatalf("no Help box with the title %q:\n%s", helpTitle, strings.Join(rows, "\n"))
	}
	for index := top + 1; index < len(rows); index++ {
		if strings.Contains(rows[index], "╰") {
			return top, index
		}
	}
	t.Fatalf("the Help box has no bottom border:\n%s", strings.Join(rows, "\n"))
	return 0, 0
}

// help.md rule 2: two columns. The left column lists SCREENS, CHAT, and ASK
// CARD. The right column lists AGENTS, WORK, FRICTION, HEALTH, and LOGS.
// STATUS ROW comes in the last rows, below both columns. Rule 3: SCREENS
// lists each chord of README decision 16, esc, ?, and ctrl+c ×2.
func TestHelpListsTheSectionsInTwoColumns(t *testing.T) {
	m, _ := openHelpAt(t, 120, 40)
	rows := viewRows(m)
	top, bottom := helpBox(t, rows)
	box := rows[top+1 : bottom]
	_, left := cellOf(box, "SCREENS")
	_, right := cellOf(box, "AGENTS")
	if left < 0 || right <= left {
		t.Fatalf("SCREENS at column %d, AGENTS at column %d; want two columns:\n%s", left, right, strings.Join(box, "\n"))
	}
	last := 0
	for _, column := range []struct {
		col      int
		sections []string
	}{{left, []string{"SCREENS", "CHAT", "ASK CARD"}}, {right, []string{"AGENTS", "WORK", "FRICTION", "HEALTH", "LOGS"}}} {
		before := -1
		for _, section := range column.sections {
			row, col := cellOf(box, section+" ")
			if col != column.col || row <= before {
				t.Fatalf("%s at row %d column %d; want column %d below row %d:\n%s", section, row, col, column.col, before, strings.Join(box, "\n"))
			}
			before = row
		}
		for row := before + 1; row < len(box) && strings.TrimSpace(ansi.Cut(box[row], column.col, column.col+40)) != ""; row++ {
			before = row
		}
		last = max(last, before)
	}
	status, col := cellOf(box, "STATUS ROW")
	if status <= last || col != left {
		t.Fatalf("STATUS ROW at row %d column %d; want column %d below row %d:\n%s", status, col, left, last, strings.Join(box, "\n"))
	}
	for _, want := range []string{"segments  state · supervisor", "config    [tui.status] lines"} {
		if row, _ := cellOf(box, want); row <= status {
			t.Errorf("%q at row %d; want it below STATUS ROW on row %d", want, row, status)
		}
	}
	screens, _ := cellOf(box, "SCREENS")
	for index, want := range []string{
		"ctrl+g      agents", "ctrl+t      work", "ctrl+f      friction", "ctrl+y      health", "ctrl+l      logs",
		"esc         back to chat", "?           this help", "ctrl+c ×2   quit",
	} {
		if got := ansi.Cut(box[screens+1+index], left, left+ansi.StringWidth(want)); got != want {
			t.Errorf("SCREENS row %d = %q; want %q", index+1, got, want)
		}
	}
	for _, want := range []string{"j           raw JSON", "ctrl+]      field panel"} {
		if row, col := cellOf(box, want); col != right || row < 0 {
			t.Errorf("LOGS row %q at column %d; want column %d", want, col, right)
		}
	}
}

// help.md rule 3: when the config rebinds a chord, SCREENS shows the bound
// key. README decision 18 also rebinds ?.
func TestHelpShowsTheBoundKeys(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	cfg := home.DefaultTUI()
	cfg.Keys["agents"] = "ctrl+a"
	cfg.Keys["help"] = "ctrl+k"
	m.setTUIConfig(cfg)
	m, _ = raise(t, m, ctrl('k'))
	if !m.helpOpen {
		t.Fatal("the bound key ctrl+k did not open Help")
	}
	requireView(t, m, "ctrl+a      agents", "ctrl+k      this help")
	rejectView(t, m, "ctrl+g      agents", "?           this help")
}

// help.md rule 1: the box border is accent, and the screen behind it is
// border. Rule 2: each section name is accent bold, each key is fg bold, and
// each action is dim.
func TestHelpColours(t *testing.T) {
	m, _ := openHelpAt(t, 120, 40)
	rows := viewRows(m)
	top, bottom := helpBox(t, rows)
	grid := renderGrid(t, m, 120, 40)
	_, x := cellOf(rows, "╭─ "+helpTitle)
	cell := func(row, col int) cellgrid.Cell { return grid.At(row, col) }
	if c := cell(top, x); c.Token != cellgrid.Accent {
		t.Errorf("box corner token %s; want accent", c.Token)
	}
	for row := range 40 {
		if row >= top && row <= bottom {
			continue
		}
		for col := range 120 {
			if c := cell(row, col); c.Rune != ' ' && (c.Token != cellgrid.Border || c.Bold) {
				t.Fatalf("row %d col %d %q is %s bold %v behind Help; want border, not bold", row, col, string(c.Rune), c.Token, c.Bold)
			}
		}
	}
	for _, tt := range []struct {
		text  string
		token cellgrid.Token
		bold  bool
	}{
		{"SCREENS", cellgrid.Accent, true},
		{"LOGS", cellgrid.Accent, true},
		{"ctrl+g", cellgrid.FG, true},
		{"shift+enter", cellgrid.FG, true},
		{"open or fold all events", cellgrid.Dim, false},
		{"details popup", cellgrid.Dim, false},
	} {
		row, col := cellOf(rows, tt.text)
		if row < 0 {
			t.Fatalf("no %q in the view", tt.text)
		}
		for i := range len([]rune(tt.text)) {
			if c := cell(row, col+i); c.Rune != ' ' && (c.Token != tt.token || c.Bold != tt.bold) {
				t.Fatalf("%q cell %d is %s bold %v; want %s bold %v", tt.text, i, c.Token, c.Bold, tt.token, tt.bold)
			}
		}
	}
}

// help.md rule 4: when the body is taller than the box, the box scrolls with
// the ┃ thumb on its right border. ↑ and ↓ scroll it, and the scroll stops at
// the end of the body.
func TestHelpScrollsWithTheThumb(t *testing.T) {
	m, _ := openHelpAt(t, 80, 24)
	thumb := func() []int {
		rows := viewRows(m)
		top, bottom := helpBox(t, rows)
		var at []int
		for row := top + 1; row < bottom; row++ {
			if strings.HasSuffix(strings.TrimRight(rows[row], " "), "┃") {
				at = append(at, row)
			}
		}
		return at
	}
	first := thumb()
	if len(first) == 0 {
		t.Fatalf("the body is taller than the box but the right border has no ┃ thumb:\n%s", strings.Join(viewRows(m), "\n"))
	}
	requireView(t, m, "SCREENS")
	for range 60 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	end := m.helpOffset
	requireView(t, m, "config    [tui.status]")
	rejectView(t, m, "SCREENS")
	if moved := thumb(); len(moved) == 0 || moved[0] <= first[0] {
		t.Fatalf("thumb rows %v after the scroll; want below %v", moved, first)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.helpOffset != end {
		t.Fatalf("↓ at the end moved the offset from %d to %d", end, m.helpOffset)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	if m.helpOffset != end-1 {
		t.Fatalf("↑ at the end gave offset %d; want %d", m.helpOffset, end-1)
	}
}

// help.md: at 120×40 the body fits the box, so no thumb shows and ↓ does
// not scroll.
func TestHelpFitsAt120x40(t *testing.T) {
	m, _ := openHelpAt(t, 120, 40)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.helpOffset != 0 {
		t.Fatalf("↓ moved the offset to %d; the body fits the box", m.helpOffset)
	}
	rows := viewRows(m)
	top, bottom := helpBox(t, rows)
	for row := top + 1; row < bottom; row++ {
		if strings.Contains(rows[row], "┃") {
			t.Fatalf("row %d has a thumb, but the body fits: %q", row, rows[row])
		}
	}
}

// help.md: ? or esc closes Help. Each other key does nothing: no chord, no
// typed text, no scroll.
func TestHelpOtherKeysDoNothing(t *testing.T) {
	m, _ := openHelpAt(t, 80, 24)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	before := strings.Join(viewRows(m), "\n")
	for _, press := range []tea.KeyPressMsg{
		key('x', "x"), key('j', "j"), key('k', "k"), key('g', "g"), key('G', "G"), key(0, "enter"),
		{Code: tea.KeyPgDown}, {Code: tea.KeyPgUp}, {Code: tea.KeyLeft}, {Code: tea.KeyRight}, {Code: tea.KeyTab},
		ctrl('t'), ctrl('g'), ctrl('o'), ctrl('r'), ctrl(']'),
	} {
		m, _ = raise(t, m, press)
		if after := strings.Join(viewRows(m), "\n"); !m.helpOpen || m.activeTab() != tabChat || after != before {
			t.Fatalf("key %q changed the Help overlay (open %v, screen %d):\n%s", press.String(), m.helpOpen, m.activeTab(), after)
		}
	}
	if got := m.chatInput.Value(); got != "" {
		t.Fatalf("keys behind Help typed %q into the composer", got)
	}
	m, _ = raise(t, m, escKey())
	if m.helpOpen {
		t.Fatal("esc did not close Help")
	}
	m, _ = raise(t, m, key('?', "?"))
	m, _ = raise(t, m, key('?', "?"))
	if m.helpOpen {
		t.Fatal("? did not close Help")
	}
}

// help.md: ? opens Help from each screen.
func TestHelpOpensFromEachScreen(t *testing.T) {
	for _, chord := range chordTabs {
		m, _ := newFrameModel(t, 120, 40)
		m, _ = raise(t, m, tea.KeyPressMsg{Code: rune(m.tuiConfig.Keys[chord.action][len("ctrl+")]), Mod: tea.ModCtrl})
		m, _ = raise(t, m, key('?', "?"))
		if !m.helpOpen || m.activeTab() != chord.tab {
			t.Fatalf("? on %s: help %v, screen %d", chord.action, m.helpOpen, m.activeTab())
		}
		requireView(t, m, helpTitle)
	}
}

// AGENTS.md logging: each user action on Help is logged: open, scroll, and
// close.
func TestHelpLogsEachAction(t *testing.T) {
	m, dir := openHelpAt(t, 80, 24)
	var cmds []tea.Cmd
	_, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	cmds = append(cmds, cmd)
	_, cmd = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	cmds = append(cmds, cmd)
	_, cmd = raise(t, m, escKey())
	cmds = append(cmds, cmd)
	for _, cmd := range cmds {
		drain(cmd)
	}
	var got []string
	for _, event := range tuiActions(t, dir) {
		action, _ := event["action"].(string)
		if strings.HasPrefix(action, "help") {
			got = append(got, action)
		}
	}
	want := []string{"help open", "help scroll", "help scroll", "help close"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("help actions %v; want %v", got, want)
	}
}

// help.md rule 4 and README decision 14: while a notice exists, the notice
// line shows on the row above the status rows, over the dimmed screen. The
// box at the center stays clear of it, and the body scrolls with the thumb.
func TestHelpKeepsTheNoticeLineClear(t *testing.T) {
	for _, chord := range []tea.KeyPressMsg{{}, ctrl('t')} {
		m, _ := newFrameModel(t, 120, 40)
		if chord.Code != 0 {
			m, _ = raise(t, m, chord)
		}
		m, _ = raise(t, m, key('?', "?"))
		m, _ = raise(t, m, ctrlC())
		rows := viewRows(m)
		notice := len(rows) - len(m.statusRows(120, 40)) - 1
		if !strings.HasPrefix(rows[notice], " press ctrl+c again to quit") {
			t.Fatalf("screen %d: row %d = %q; want the quit notice above the status rows:\n%s", m.activeTab(), notice, rows[notice], strings.Join(rows, "\n"))
		}
		if c := renderGrid(t, m, 120, 40).At(notice, 1); c.Token != cellgrid.Accent {
			t.Errorf("notice token %s; want accent", c.Token)
		}
		top, bottom := helpBox(t, rows)
		if bottom >= notice || top != 40-1-bottom {
			t.Fatalf("box rows %d to %d; want it at the center and above row %d", top, bottom, notice)
		}
		_, left := cellOf(rows, "╭─ "+helpTitle)
		edge := left + helpBoxWidth - 1
		thumb := false
		for row := top + 1; row < bottom; row++ {
			thumb = thumb || ansi.Cut(rows[row], edge, edge+1) == "┃"
		}
		if !thumb {
			t.Errorf("screen %d: the body does not fit the smaller box but has no thumb", m.activeTab())
		}
	}
}
