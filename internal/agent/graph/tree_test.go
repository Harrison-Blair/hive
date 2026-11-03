package graph

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
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

// TestTreeDrawsSiblingsInOneRowUnderACenteredParent compares the tree of
// smallTree at 38 columns with a picture drawn by hand: cards of 16 columns,
// a and b side by side, c centered over them, and w below a.
func TestTreeDrawsSiblingsInOneRowUnderACenteredParent(t *testing.T) {
	want := strings.Join([]string{
		"╭─ home ─────────────────────────────╮",
		"│          ╭──────────────╮          │",
		"│          │ ○ c          │          │",
		"│          │              │          │",
		"│          │ pi · m       │          │",
		"│          ╰───────┬──────╯          │",
		"│         ╭────────┴────────╮        │",
		"│ ╭───────┴──────╮  ╭───────┴──────╮ │",
		"│ │ ○ a          │  │ ○ b          │ │",
		"│ │              │  │              │ │",
		"│ │ pi · m       │  │ pi · m       │ │",
		"│ ╰───────┬──────╯  ╰──────────────╯ │",
		"│         │                          │",
		"│ ╭───────┴──────╮                   │",
		"│ │ ○ w          │                   │",
		"│ │              │                   │",
		"│ │ pi · m       │                   │",
		"│ ╰──────────────╯                   │",
		"╰────────────────────────────────────╯",
	}, "\n")
	b := drawBoard(Pods(smallTree()), 38, nil, "", 0, nil, true)
	if got := stripped(b.lines); got != want {
		t.Fatalf("tree =\n%s\nwant\n%s", got, want)
	}
	for name, at := range map[string][2]int{"c": {11, 1}, "a": {2, 7}, "b": {20, 7}, "w": {2, 13}} {
		if s := b.cards[name]; s.left != at[0] || s.top != at[1] || s.width != 16 {
			t.Errorf("span of %s = %+v; want left %d, top %d, width 16", name, s, at[0], at[1])
		}
	}
}

func TestTreeJoinsOfTheSelectedAndOfAWorkingParent(t *testing.T) {
	entries := smallTree()
	entries[1].Status = "working"
	b := drawBoard(Pods(entries), 38, nil, "c", 0, nil, true)
	if row := b.lines[5]; !strings.Contains(row, "\x1b["+styleSelected+"m┗━━━━━━━┳━━━━━━┛") {
		t.Errorf("selected parent bottom = %q; want a heavy ┳ in the selected style", row)
	}
	if row := b.lines[7]; !strings.Contains(row, "\x1b["+styleWorking+"m╭───────┴──────╮") {
		t.Errorf("working child top = %q; want ┴ in the working style", row)
	}
	if row := b.lines[11]; !strings.Contains(row, "\x1b["+styleWorking+"m╰───────┬──────╯") {
		t.Errorf("working parent bottom = %q; want ┬ in the working style", row)
	}
	b = drawBoard(Pods(entries), 38, nil, "b", 0, nil, true)
	if row := ansi.Strip(b.lines[7]); !strings.Contains(row, "┏━━━━━━━┻━━━━━━┓") {
		t.Errorf("selected child top = %q; want a heavy ┻", row)
	}
}

func TestTreeCardsShrinkToFitTheWidestPod(t *testing.T) {
	// Two cards of 33 and a gap of 2 need 68 columns, and 72 with the pod.
	for width, want := range map[int]int{72: 33, 71: 32, 38: 16, 20: 16} {
		b := drawBoard(Pods(smallTree()), width, nil, "", 0, nil, true)
		if got := b.cards["w"].width; got != want {
			t.Errorf("width %d: card width = %d; want %d", width, got, want)
		}
	}
}

func TestGraphOpensInTreeModeAndGSwitches(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	text := screen(m)
	if !strings.Contains(text, " [tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit") {
		t.Fatalf("the graph does not open in tree mode:\n%s", text)
	}
	if strings.Contains(text, "├┤ ✓ hive-reviewer") || !strings.Contains(text, "┴") {
		t.Fatalf("tree mode draws the list:\n%s", text)
	}
	m = key(t, m, 'g')
	text = screen(m)
	if !strings.Contains(text, " [list] ←↑↓→ move · enter details · esc hide · g tree · r refresh · q quit") || !strings.Contains(text, "├┤ ✓ hive-reviewer") {
		t.Fatalf("g does not switch to the list:\n%s", text)
	}
	m = key(t, m, 'g')
	if text := screen(m); !strings.Contains(text, " [tree] ") || strings.Contains(text, "├┤ ✓ hive-reviewer") {
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

func TestTreeDownTieGoesLeft(t *testing.T) {
	m := loaded(t, smallTree(), 38, 40)
	if m.selected != "c" {
		t.Fatalf("selected = %q; want c", m.selected)
	}
	if m = key(t, m, tea.KeyDown); m.selected != "a" {
		t.Fatalf("down from c, whose center is as far from a as from b: selected = %q; want a", m.selected)
	}
}

// TestWideTreeScrollsSidewaysToTheSelection draws one root with six children
// at 80 columns. The children are 16 wide with a gap of 2, so k6 spans
// columns 92 to 107 of the board.
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
	if len(row) != 80 || row[79] != '┃' || string(row[64:68]) != "┃ ○ " {
		t.Fatalf("k6 row 1 = %q; want the card in columns 64 to 79", string(row))
	}
	for range 5 {
		m = key(t, m, tea.KeyLeft)
	}
	if text := screen(m); m.selected != "k1" || !strings.HasPrefix(cardRow1(t, m, "k1"), "┃ ○ k1 ") {
		t.Fatalf("left back to k1 does not scroll back:\n%s", text)
	}
}

// TestHiddenTreeKeepsTheSideScrollAndAMoveKeyShowsWithoutMoving draws the
// root a beside the root r with six children at 80 columns. Cards are 16
// wide: a spans columns 2 to 17 and k6 columns 110 to 125.
func TestHiddenTreeKeepsTheSideScrollAndAMoveKeyShowsWithoutMoving(t *testing.T) {
	entries := []list.Entry{entry("a", "", "idle", "pi", "m", 0), entry("r", "", "idle", "pi", "m", 0)}
	for _, name := range []string{"k1", "k2", "k3", "k4", "k5", "k6"} {
		entries = append(entries, entry(name, "r", "idle", "pi", "m", 0))
	}
	m := key(t, loaded(t, entries, 80, 30), tea.KeyDown)
	for range 5 {
		m = key(t, m, tea.KeyRight)
	}
	if m.selected != "k6" || m.xscroll != 46 {
		t.Fatalf("selected = %q, xscroll = %d; want k6 with xscroll 46", m.selected, m.xscroll)
	}
	m = key(t, m, tea.KeyEscape)
	if text := screen(m); strings.ContainsAny(text, "┏┻") {
		t.Fatalf("esc did not hide the outline in tree mode:\n%s", text)
	}
	// The refresh replaces k6 with k7, so the selection goes to a, left of
	// the view, and the board keeps its width.
	m = update(t, m, fetchedMsg{entries: append(entries[:7:7], entry("k7", "r", "idle", "pi", "m", 0)), at: clock})
	if m.selected != "a" || m.xscroll != 46 {
		t.Fatalf("selected = %q, xscroll = %d; want a with xscroll 46 kept while hidden", m.selected, m.xscroll)
	}
	m = key(t, m, tea.KeyDown)
	if text := screen(m); m.selected != "a" || !strings.Contains(text, "┃ ○ a ") {
		t.Fatalf("down while hidden: selected = %q; want the outline back on a without a move:\n%s", m.selected, text)
	}
}
