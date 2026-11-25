package graph

import (
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/canvas"
	"github.com/Harrison-Blair/hive/internal/lib/style"
	"github.com/charmbracelet/x/ansi"
)

// smallTree is c with the children a and b, and w below a, all in the home
// pod.
func smallTree() []list.Entry {
	return []list.Entry{
		entry("c", "", "idle", "pi", "m", 0),
		entry("a", "c", "idle", "pi", "m", 0),
		entry("b", "c", "idle", "pi", "m", 0),
		entry("w", "a", "idle", "pi", "m", 0),
	}
}

// listed is loaded in list mode.
func listed(t *testing.T, entries []list.Entry, width, height int) Model {
	t.Helper()
	return key(t, loaded(t, entries, width, height), 'g')
}

func stripped(lines []string) string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = ansi.Strip(line)
	}
	return strings.Join(out, "\n")
}

// TestTreeDrawsSiblingsInOneRowUnderTheFirstChild compares the tree of
// smallTree at 38 columns with a picture drawn by hand: a pod of two tracks,
// 79 columns, cards of 33 columns at columns 3 and 43, c over its first child
// a, b beside a, and w below a.
func TestTreeDrawsSiblingsInOneRowUnderTheFirstChild(t *testing.T) {
	card := func(right int, rows ...string) []string {
		out := make([]string, len(rows))
		for i, row := range rows {
			out[i] = "│  " + row + strings.Repeat(" ", right) + "│"
		}
		return out
	}
	top, bottom := "╭"+strings.Repeat("─", 31)+"╮", "╰"+strings.Repeat("─", 31)+"╯"
	blank := "│" + strings.Repeat(" ", 31) + "│"
	name := func(n string) string { return "│ ○ " + n + strings.Repeat(" ", 27) + "│" }
	sub := "│ pi · m" + strings.Repeat(" ", 24) + "│"
	lines := []string{"╭─ home " + strings.Repeat("─", 70) + "╮"}
	lines = append(lines, card(42, top, name("c"), blank, sub, bottom)...)
	lines = append(lines, "│"+strings.Repeat(" ", 18)+"├"+strings.Repeat("─", 39)+"╮"+strings.Repeat(" ", 18)+"│")
	gap := strings.Repeat(" ", 7)
	lines = append(lines, card(2, top+gap+top, name("a")+gap+name("b"), blank+gap+blank, sub+gap+sub, bottom+gap+bottom)...)
	lines = append(lines, "│"+strings.Repeat(" ", 18)+"│"+strings.Repeat(" ", 58)+"│")
	lines = append(lines, card(42, top, name("w"), blank, sub, bottom)...)
	lines = append(lines, "╰"+strings.Repeat("─", 77)+"╯")
	want := strings.Join(lines, "\n")
	b := drawBoard(Pods(smallTree()), 38, nil, "", 0, nil, true)
	if got := stripped(b.lines); got != want {
		t.Fatalf("tree =\n%s\nwant\n%s", got, want)
	}
	for name, at := range map[string][2]int{"c": {3, 1}, "a": {3, 7}, "b": {43, 7}, "w": {3, 13}} {
		if s := b.cards[name]; s.left != at[0] || s.top != at[1] || s.width != 33 {
			t.Errorf("span of %s = %+v; want left %d, top %d, width 33", name, s, at[0], at[1])
		}
	}
}

func TestTreeJoinsOfTheSelectedAndOfAWorkingParent(t *testing.T) {
	entries := smallTree()
	entries[1].Status = "working"
	b := drawBoard(Pods(entries), 38, nil, "c", 0, nil, true)
	if row := b.lines[5]; !strings.Contains(row, "\x1b["+style.Selected+"m┗"+strings.Repeat("━", 31)+"┛") {
		t.Errorf("selected parent bottom = %q; want a whole heavy edge in the selected style", row)
	}
	if row := b.lines[7]; !strings.Contains(row, "\x1b["+style.Working+"m╭"+strings.Repeat("─", 31)+"╮") {
		t.Errorf("working child top = %q; want a whole edge in the working style", row)
	}
	if row := b.lines[11]; !strings.Contains(row, "\x1b["+style.Working+"m╰"+strings.Repeat("─", 31)+"╯") {
		t.Errorf("working parent bottom = %q; want a whole edge in the working style", row)
	}
	b = drawBoard(Pods(entries), 38, nil, "b", 0, nil, true)
	if row := ansi.Strip(b.lines[7]); !strings.Contains(row, "┏"+strings.Repeat("━", 31)+"┓") {
		t.Errorf("selected child top = %q; want a whole heavy edge", row)
	}
}

func TestTreeCardsKeepFullWidth(t *testing.T) {
	// smallTree has two leaves, so its pod is two tracks: 79 columns.
	for _, width := range []int{120, 79, 78, 38, 20, 1} {
		b := drawBoard(Pods(smallTree()), width, nil, "", 0, nil, true)
		if got := b.cards["w"].width; got != 33 {
			t.Errorf("width %d: card width = %d; want 33", width, got)
		}
		if b.width != 79 {
			t.Errorf("width %d: board width = %d; want 79", width, b.width)
		}
	}
}

// TestListCardWidthFollowsTheScreenAndTheDepth checks the root c, of
// min(33, max(16, W - 6)), and w at depth 2, 6 narrower down to 16.
func TestListCardWidthFollowsTheScreenAndTheDepth(t *testing.T) {
	for width, want := range map[int][2]int{72: {33, 27}, 38: {32, 26}, 20: {16, 16}, 1: {16, 16}} {
		b := drawBoard(Pods(smallTree()), width, nil, "", 0, nil, false)
		if got := [2]int{b.cards["c"].width, b.cards["w"].width}; got != want {
			t.Errorf("width %d: list card widths of c and w = %v; want %v", width, got, want)
		}
	}
}

func TestNarrowTreeNavigationFollowsSelection(t *testing.T) {
	m := loaded(t, smallTree(), 20, 40)
	for _, step := range []struct {
		key      rune
		selected string
		xscroll  int
		row      int
	}{
		{tea.KeyDown, "a", 16, 10},
		{tea.KeyRight, "b", 56, 10},
		{tea.KeyLeft, "a", 16, 10},
		{tea.KeyUp, "c", 16, 4},
	} {
		m = key(t, m, step.key)
		if m.selected != step.selected || m.xscroll != step.xscroll {
			t.Fatalf("key %q: selected = %q, xscroll = %d; want %q, %d", step.key, m.selected, m.xscroll, step.selected, step.xscroll)
		}
		row := []rune(strings.Split(screen(m), "\n")[step.row])
		if len(row) != 20 || row[19] != '┃' {
			t.Fatalf("key %q: selected card row = %q; want 20 columns ending at selected border", step.key, string(row))
		}
	}
}

func TestGraphOpensInTreeModeAndGSwitches(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	text := screen(m)
	if !strings.Contains(text, " [tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit") {
		t.Fatalf("the graph does not open in tree mode:\n%s", text)
	}
	if strings.Contains(text, "├─│") || !strings.Contains(text, "├───") {
		t.Fatalf("tree mode draws the list:\n%s", text)
	}
	m = key(t, m, 'g')
	text = screen(m)
	if !strings.Contains(text, " [list] ←↑↓→ move · enter details · esc hide · g tree · r refresh · q quit") || !strings.Contains(text, "├─│") || strings.Contains(text, "├───") {
		t.Fatalf("g does not switch to the list:\n%s", text)
	}
	m = key(t, m, 'g')
	if text := screen(m); !strings.Contains(text, " [tree] ") || strings.Contains(text, "├─│") {
		t.Fatalf("a second g does not switch back to the tree:\n%s", text)
	}
}

// TestTreeMovesFollowTheScreen walks the fixture at 140 columns. Cards are
// 33 wide. The api pod (api-coordinator over api-worker) and the hive pod
// (hive-coordinator over hive-reviewer and hive-worker, hive-worker-kid
// below hive-worker) share the first row of pods; home (coordinator and
// loose) wraps to the second.
func TestTreeMovesFollowTheScreen(t *testing.T) {
	m := loaded(t, fixture(), 140, 60)
	for i, step := range []struct {
		key  rune
		want string
	}{
		{tea.KeyUp, "api-coordinator"},
		{tea.KeyLeft, "api-coordinator"},
		{tea.KeyRight, "hive-coordinator"},
		{tea.KeyDown, "hive-reviewer"},
		{tea.KeyLeft, "api-worker"},
		{tea.KeyRight, "hive-reviewer"},
		{tea.KeyRight, "hive-worker"},
		{tea.KeyRight, "hive-worker"},
		{tea.KeyDown, "hive-worker-kid"},
		{tea.KeyDown, "loose"},
		{tea.KeyLeft, "coordinator"},
		{tea.KeyDown, "coordinator"},
		{tea.KeyUp, "hive-worker-kid"},
		{'k', "hive-worker"},
		{'h', "hive-reviewer"},
		{'j', "hive-worker-kid"},
		{'l', "hive-worker-kid"},
	} {
		m = key(t, m, step.key)
		if m.selected != step.want {
			t.Fatalf("step %d: selected = %q; want %q", i, m.selected, step.want)
		}
	}
}

// TestTreeDownTieGoesLeft draws the roots a, b, and c on the tracks 0, 1,
// and 2, with a1 below a and c1 below c. From b, a1 and c1 are one track
// away.
func TestTreeDownTieGoesLeft(t *testing.T) {
	m := loaded(t, []list.Entry{
		entry("a", "", "idle", "pi", "m", 0), entry("a1", "a", "idle", "pi", "m", 0),
		entry("b", "", "idle", "pi", "m", 0),
		entry("c", "", "idle", "pi", "m", 0), entry("c1", "c", "idle", "pi", "m", 0),
	}, 200, 40)
	if m = key(t, m, tea.KeyRight); m.selected != "b" {
		t.Fatalf("selected = %q; want b", m.selected)
	}
	if m = key(t, m, tea.KeyDown); m.selected != "a1" {
		t.Fatalf("down from b, whose center is as far from a1 as from c1: selected = %q; want a1", m.selected)
	}
}

// TestWideTreeScrollsSidewaysToTheSelection draws one root with six children
// at 80 columns. The children are on the tracks 0 to 5, so k6 spans
// columns 203 to 235 of the board.
func TestWideTreeScrollsSidewaysToTheSelection(t *testing.T) {
	entries := []list.Entry{entry("r", "", "idle", "pi", "m", 0)}
	for _, name := range []string{"k1", "k2", "k3", "k4", "k5", "k6"} {
		entries = append(entries, entry(name, "r", "idle", "pi", "m", 0))
	}
	m := key(t, loaded(t, entries, 80, 30), tea.KeyDown)
	m = key(t, m, tea.KeyLeft)
	m = key(t, m, tea.KeyLeft)
	if text := screen(m); m.selected != "k1" || !strings.Contains(text, "○ k1 ") {
		t.Fatalf("selected = %q; want k1 in view:\n%s", m.selected, text)
	}
	for range 5 {
		m = key(t, m, tea.KeyRight)
	}
	text := screen(m)
	if m.selected != "k6" || !strings.Contains(text, "○ k6 ") || strings.Contains(text, "○ k1 ") {
		t.Fatalf("selected = %q; want k6 in view and k1 off it:\n%s", m.selected, text)
	}
	for _, line := range strings.Split(m.View().Content, "\n") {
		if w := ansi.StringWidth(line); w > 80 {
			t.Errorf("line %q is %d cells; want at most 80", ansi.Strip(line), w)
		}
	}
	row := []rune(cardRow1(t, m, "k6"))
	if len(row) != 80 || row[79] != '┃' || string(row[47:51]) != "┃ ○ " {
		t.Fatalf("k6 row 1 = %q; want the card in columns 47 to 79", string(row))
	}
	for range 5 {
		m = key(t, m, tea.KeyLeft)
	}
	if text := screen(m); m.selected != "k1" || !strings.HasPrefix(cardRow1(t, m, "k1"), "┃ ○ k1 ") {
		t.Fatalf("left back to k1 does not scroll back:\n%s", text)
	}
}

// TestHiddenTreeKeepsTheSideScrollAndAMoveKeyShowsWithoutMoving draws the
// root a beside the root r with six children at 80 columns. Cards are 33
// wide: a spans columns 3 to 35 and k6, on track 6, columns 243 to 275.
func TestHiddenTreeKeepsTheSideScrollAndAMoveKeyShowsWithoutMoving(t *testing.T) {
	entries := []list.Entry{entry("a", "", "idle", "pi", "m", 0), entry("r", "", "idle", "pi", "m", 0)}
	for _, name := range []string{"k1", "k2", "k3", "k4", "k5", "k6"} {
		entries = append(entries, entry(name, "r", "idle", "pi", "m", 0))
	}
	m := key(t, loaded(t, entries, 80, 30), tea.KeyDown)
	for range 5 {
		m = key(t, m, tea.KeyRight)
	}
	if m.selected != "k6" || m.xscroll != 196 {
		t.Fatalf("selected = %q, xscroll = %d; want k6 with xscroll 196", m.selected, m.xscroll)
	}
	m = key(t, m, tea.KeyEscape)
	if text := screen(m); strings.ContainsAny(text, "┏┃") {
		t.Fatalf("esc did not hide the outline in tree mode:\n%s", text)
	}
	// The refresh replaces k6 with k7, so the selection goes to a, left of
	// the view, and the board keeps its width.
	m = update(t, m, fetchedMsg{entries: append(entries[:7:7], entry("k7", "r", "idle", "pi", "m", 0)), at: clock})
	if m.selected != "a" || m.xscroll != 196 {
		t.Fatalf("selected = %q, xscroll = %d; want a with xscroll 196 kept while hidden", m.selected, m.xscroll)
	}
	m = key(t, m, tea.KeyDown)
	if text := screen(m); m.selected != "a" || !strings.Contains(text, "┃ ○ a ") {
		t.Fatalf("down while hidden: selected = %q; want the outline back on a without a move:\n%s", m.selected, text)
	}
}

// parseCells reads a rendered line back into one cell per rune with its SGR
// style.
func parseCells(line string) []canvas.Cell {
	var out []canvas.Cell
	style := ""
	for i := 0; i < len(line); {
		if strings.HasPrefix(line[i:], "\x1b[") {
			end := strings.IndexByte(line[i:], 'm')
			style = line[i+2 : i+end]
			if style == "0" {
				style = ""
			}
			i += end + 1
			continue
		}
		r, n := utf8.DecodeRuneInString(line[i:])
		out = append(out, canvas.Cell{R: r, Style: style})
		i += n
	}
	return out
}

// TestJoinsLeaveTheCardBordersWhole checks, in both modes, that each cell of
// the border of a selected, a working, and an idle card with joins is the
// plain box rune of that card in the style of that card.
func TestJoinsLeaveTheCardBordersWhole(t *testing.T) {
	entries := smallTree()
	entries[1].Status = "working"
	for _, tree := range []bool{true, false} {
		for _, selected := range []string{"c", "a", "b", "w"} {
			b := drawBoard(Pods(entries), 38, nil, selected, 0, nil, tree)
			rows := make([][]canvas.Cell, len(b.lines))
			for i, line := range b.lines {
				rows[i] = parseCells(line)
			}
			for name, s := range b.cards {
				box, edge := []rune("╭╮╰╯─│"), style.Border
				switch {
				case name == selected:
					box, edge = []rune("┏┓┗┛━┃"), style.Selected
				case s.status == "working":
					edge = style.Working
				}
				want := canvas.New(s.left+s.width, s.top+cardHeight)
				want.Box(s.left, s.top, s.width, cardHeight, box, edge)
				for y := s.top; y < s.top+cardHeight; y++ {
					for x := s.left; x < s.left+s.width; x++ {
						w := want.Cells[y][x]
						if w.R == 0 {
							continue
						}
						if got := rows[y][x]; got != w {
							t.Errorf("tree %v, selected %s: card %s cell %d,%d = %q in %q; want %q in %q",
								tree, selected, name, x, y, got.R, got.Style, w.R, w.Style)
						}
					}
				}
			}
		}
	}
}
