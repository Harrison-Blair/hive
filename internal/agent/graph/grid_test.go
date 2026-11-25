package graph

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/style"
	"github.com/charmbracelet/x/ansi"
)

// briefPod is the pod hive of the examples of the brief: the coordinator with
// three children, and one grandchild below graph-builder.
func briefPod() ([]list.Entry, map[string]string, map[string]string) {
	entries := []list.Entry{
		entry("hive-coordinator", "", "working", "claude", "claude-opus-5-5", 0),
		entry("docs-writer", "hive-coordinator", "idle", "pi", "sonnet-5-5", 0),
		entry("graph-builder", "hive-coordinator", "working", "claude", "claude-fable-5-2", 0),
		entry("graph-reviewer-1", "hive-coordinator", "done", "codex", "gpt-6", 0),
		entry("graph-tester", "graph-builder", "blocked", "claude", "claude-haiku-5-5", 0),
	}
	ages := map[string]string{
		"hive-coordinator": "≥2h04m00s", "docs-writer": "≥1m30s", "graph-builder": "≥14m05s",
		"graph-reviewer-1": "≥6m00s", "graph-tester": "≥30s",
	}
	checkouts := map[string]string{
		"hive-coordinator": "main", "docs-writer": "docs/graph", "graph-builder": "feat/graph-grid",
		"graph-reviewer-1": "feat/graph-grid", "graph-tester": "feat/graph-grid",
	}
	return entries, ages, checkouts
}

// TestTreeDrawsTheExampleOfTheBrief compares tree mode with the example of
// section 1.2 of the brief.
func TestTreeDrawsTheExampleOfTheBrief(t *testing.T) {
	want := strings.Join([]string{
		"╭─ hive ──────────────────────────────────────────────────────────────────────────────────────────────────────────────╮",
		"│  ╭───────────────────────────────╮                                                                                  │",
		"│  │ ⠋ coordinator       ≥2h04m00s │                                                                                  │",
		"│  │ main                          │                                                                                  │",
		"│  │ claude · opus-5-5             │                                                                                  │",
		"│  ╰───────────────────────────────╯                                                                                  │",
		"│                  ├───────────────────────────────────────┬───────────────────────────────────────╮                  │",
		"│  ╭───────────────────────────────╮       ╭───────────────────────────────╮       ╭───────────────────────────────╮  │",
		"│  │ ○ docs-writer          ≥1m30s │       │ ⠋ graph-builder       ≥14m05s │       │ ✓ graph-reviewer-1     ≥6m00s │  │",
		"│  │ docs/graph                    │       │ feat/graph-grid               │       │ feat/graph-grid               │  │",
		"│  │ pi · sonnet-5-5               │       │ claude · fable-5-2            │       │ codex · 6                     │  │",
		"│  ╰───────────────────────────────╯       ╰───────────────────────────────╯       ╰───────────────────────────────╯  │",
		"│                                                          │                                                          │",
		"│                                          ╭───────────────────────────────╮                                          │",
		"│                                          │ × graph-tester           ≥30s │                                          │",
		"│                                          │ feat/graph-grid               │                                          │",
		"│                                          │ claude · haiku-5-5            │                                          │",
		"│                                          ╰───────────────────────────────╯                                          │",
		"╰─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯",
	}, "\n")
	entries, ages, checkouts := briefPod()
	b := drawBoard(Pods(entries), 132, ages, "", 0, checkouts, true)
	if got := stripped(b.lines); got != want {
		t.Fatalf("tree =\n%s\nwant\n%s", got, want)
	}
}

// TestListDrawsTheExampleOfTheBrief compares list mode with the example of
// section 1.3 of the brief.
func TestListDrawsTheExampleOfTheBrief(t *testing.T) {
	want := strings.Join([]string{
		"╭─ hive ──────────────────────────────╮",
		"│  ╭───────────────────────────────╮  │",
		"│  │ ⠋ coordinator       ≥2h04m00s │  │",
		"│  │ main                          │  │",
		"│  │ claude · opus-5-5             │  │",
		"│  ╰───────────────────────────────╯  │",
		"│   │                                 │",
		"│   │ ╭────────────────────────────╮  │",
		"│   │ │ ○ docs-writer       ≥1m30s │  │",
		"│   ├─│ docs/graph                 │  │",
		"│   │ │ pi · sonnet-5-5            │  │",
		"│   │ ╰────────────────────────────╯  │",
		"│   │                                 │",
		"│   │ ╭────────────────────────────╮  │",
		"│   │ │ ⠋ graph-builder    ≥14m05s │  │",
		"│   ├─│ feat/graph-grid            │  │",
		"│   │ │ claude · fable-5-2         │  │",
		"│   │ ╰────────────────────────────╯  │",
		"│   │  │                              │",
		"│   │  │ ╭─────────────────────────╮  │",
		"│   │  │ │ × graph-tester     ≥30s │  │",
		"│   │  ╰─│ feat/graph-grid         │  │",
		"│   │    │ claude · haiku-5-5      │  │",
		"│   │    ╰─────────────────────────╯  │",
		"│   │                                 │",
		"│   │ ╭────────────────────────────╮  │",
		"│   │ │ ✓ graph-reviewer-1  ≥6m00s │  │",
		"│   ╰─│ feat/graph-grid            │  │",
		"│     │ codex · 6                  │  │",
		"│     ╰────────────────────────────╯  │",
		"╰─────────────────────────────────────╯",
	}, "\n")
	entries, ages, checkouts := briefPod()
	b := drawBoard(Pods(entries), 132, ages, "", 0, checkouts, false)
	if got := stripped(b.lines); got != want {
		t.Fatalf("list =\n%s\nwant\n%s", got, want)
	}
}

// gridEntries has five pods: a, b, c, d of one card each, and the home pod e
// with the root e over e1 and e2, which is two tracks.
func gridEntries() []list.Entry {
	entries := []list.Entry{}
	for _, p := range []string{"a", "b", "c", "d"} {
		entries = append(entries, entry(p+"-coordinator", "", "idle", "pi", "m", 0))
	}
	return append(entries,
		entry("e", "", "idle", "pi", "m", 0),
		entry("e1", "e", "idle", "pi", "m", 0),
		entry("e2", "e", "idle", "pi", "m", 0))
}

// TestTreeCardsAndPodsAreOnTheTrackGrid checks the columns of section 1.1:
// N = (W + 1) / 40 tracks in a row of pods, a pod of n tracks at a column 40k
// and 40n - 1 wide, and each card at a column 3 + 40k and 33 wide.
func TestTreeCardsAndPodsAreOnTheTrackGrid(t *testing.T) {
	type place struct{ left, top int }
	for _, tc := range []struct {
		width int
		cards map[string]place
		// pods holds the column, the first row, and the width of each pod
		// box, in pod order.
		pods [][3]int
	}{
		// At 200 columns, five tracks fit: e (2 tracks) does not fit beside
		// four pods of one track and starts the second row.
		{200, map[string]place{
			"a-coordinator": {3, 1}, "b-coordinator": {43, 1}, "c-coordinator": {83, 1}, "d-coordinator": {123, 1},
			"e": {3, 8}, "e1": {3, 14}, "e2": {43, 14},
		}, [][3]int{{0, 0, 39}, {40, 0, 39}, {80, 0, 39}, {120, 0, 39}, {0, 7, 79}}},
		// At 132, three tracks fit: d starts the second row, and e fits
		// beside it.
		{132, map[string]place{
			"a-coordinator": {3, 1}, "b-coordinator": {43, 1}, "c-coordinator": {83, 1}, "d-coordinator": {3, 8},
			"e": {43, 8}, "e1": {43, 14}, "e2": {83, 14},
		}, [][3]int{{0, 0, 39}, {40, 0, 39}, {80, 0, 39}, {0, 7, 39}, {40, 7, 79}}},
		// At 80, two tracks fit: two pods in a row, and e alone in the third.
		{80, map[string]place{
			"a-coordinator": {3, 1}, "b-coordinator": {43, 1}, "c-coordinator": {3, 8}, "d-coordinator": {43, 8},
			"e": {3, 15}, "e1": {3, 21}, "e2": {43, 21},
		}, [][3]int{{0, 0, 39}, {40, 0, 39}, {0, 7, 39}, {40, 7, 39}, {0, 14, 79}}},
	} {
		b := drawBoard(Pods(gridEntries()), tc.width, nil, "", 0, nil, true)
		for name, want := range tc.cards {
			if s := b.cards[name]; s.left != want.left || s.top != want.top || s.width != 33 {
				t.Errorf("width %d: span of %s = %+v; want left %d, top %d, width 33", tc.width, name, s, want.left, want.top)
			}
		}
		for i, pod := range tc.pods {
			row := []rune(ansi.Strip(b.lines[pod[1]]))
			title := []rune("╭─ " + Pods(gridEntries())[i].Title + " ")
			if len(row) < pod[0]+pod[2] || string(row[pod[0]:pod[0]+len(title)]) != string(title) || row[pod[0]+pod[2]-1] != '╮' {
				t.Errorf("width %d: pod %d row %d = %q; want a box from column %d, %d wide", tc.width, i, pod[1], string(row), pod[0], pod[2])
			}
		}
	}
}

// TestTreeParentIsOverItsFirstChild checks that each parent card is in the
// column of its first child, and that the joins are as in section 1.2.
func TestTreeParentIsOverItsFirstChild(t *testing.T) {
	entries, ages, checkouts := briefPod()
	b := drawBoard(Pods(entries), 200, ages, "", 0, checkouts, true)
	for parent, child := range map[string]string{"hive-coordinator": "docs-writer", "graph-builder": "graph-tester"} {
		if p, c := b.cards[parent], b.cards[child]; p.left != c.left {
			t.Errorf("%s at column %d; want the column %d of its first child %s", parent, p.left, c.left, child)
		}
	}
	joins := []rune(ansi.Strip(b.lines[6]))
	for col, want := range map[int]rune{19: '├', 20: '─', 58: '─', 59: '┬', 60: '─', 98: '─', 99: '╮', 100: ' ', 18: ' '} {
		if joins[col] != want {
			t.Errorf("join row column %d = %q; want %q", col, joins[col], want)
		}
	}
	if one := []rune(ansi.Strip(b.lines[12])); one[59] != '│' || strings.Count(string(one), "│") != 3 {
		t.Errorf("join row of one child = %q; want │ in column 59", string(one))
	}
}

// TestListChildrenStepRightAndEndInOneColumn checks section 1.3 at several
// widths: a root of min(33, max(16, W - 6)), each depth 3 narrower down to 16,
// and each card ending in the column of the root.
func TestListChildrenStepRightAndEndInOneColumn(t *testing.T) {
	// A chain d0 → d1 → … → d7 is depth 0 to 7.
	entries := []list.Entry{entry("d0", "", "idle", "pi", "m", 0)}
	for i := 1; i < 8; i++ {
		entries = append(entries, entry("d"+string(rune('0'+i)), "d"+string(rune('0'+i-1)), "idle", "pi", "m", 0))
	}
	for width, widths := range map[int][]int{
		200: {33, 30, 27, 24, 21, 18, 18, 18},
		30:  {24, 21, 18, 16, 16, 16, 16, 16},
		10:  {16, 16, 16, 16, 16, 16, 16, 16},
	} {
		b := drawBoard(Pods(entries), width, nil, "", 0, nil, false)
		root := b.cards["d0"]
		if root.left != 3 {
			t.Errorf("width %d: root at column %d; want 3", width, root.left)
		}
		for d, w := range widths {
			s := b.cards["d"+string(rune('0'+d))]
			if s.width != w || s.left+s.width != root.left+root.width {
				t.Errorf("width %d: depth %d = %+v; want width %d ending at column %d", width, d, s, w, root.left+root.width-1)
			}
		}
	}
}

// TestListRailSkipsAChildLessThanThreeRight draws the chain at 30 columns:
// d3 starts 2 cells right of d2, and d4 in the column of d3, so neither has a
// rail or a tick. d1 and d2 have them.
func TestListRailSkipsAChildLessThanThreeRight(t *testing.T) {
	entries := []list.Entry{entry("d0", "", "idle", "pi", "m", 0)}
	for i := 1; i < 5; i++ {
		entries = append(entries, entry("d"+string(rune('0'+i)), "d"+string(rune('0'+i-1)), "idle", "pi", "m", 0))
	}
	b := drawBoard(Pods(entries), 30, nil, "", 0, nil, false)
	rows := strings.Split(stripped(b.lines), "\n")
	for name, tick := range map[string]bool{"d1": true, "d2": true, "d3": false, "d4": false} {
		s := b.cards[name]
		row := []rune(rows[s.top+2])
		has := row[s.left-2] == '╰' && row[s.left-1] == '─'
		if has != tick {
			t.Errorf("%s middle row %q: tick %v; want %v", name, string(row), has, tick)
		}
	}
	d2 := b.cards["d2"]
	for y := d2.top + cardHeight; y < len(rows)-1; y++ {
		if r := []rune(rows[y])[d2.left+1]; r != ' ' {
			t.Errorf("row %d column %d = %q; want no rail below d2", y, d2.left+1, r)
		}
	}
}

func viewLines(m Model) []string {
	return strings.Split(m.View().Content, "\n")
}

// TestHeaderStartsAtColumnThreeAndTheClockEndsFiveFromTheRight checks
// section 1.4.
func TestHeaderStartsAtColumnThreeAndTheClockEndsFiveFromTheRight(t *testing.T) {
	for _, width := range []int{200, 80} {
		lines := viewLines(loaded(t, fixture(), width, 40))
		title := ansi.Strip(lines[0])
		if !strings.HasPrefix(title, "   hive agents · 3 pods · 8 agents ") {
			t.Errorf("width %d: title %q does not start at column 3", width, title)
		}
		if !strings.HasSuffix(title, " 14:22:05    ") || ansi.StringWidth(title) != width {
			t.Errorf("width %d: title %q; want the clock to end at column %d", width, title, width-5)
		}
		if status := ansi.Strip(lines[1]); status != "   updated 14:22:05" {
			t.Errorf("width %d: status %q; want it at column 3", width, status)
		}
	}
	m := New(Options{Fetch: fetchOf(nil), Now: func() time.Time { return clock }})
	m = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if status := ansi.Strip(viewLines(m)[1]); status != "   loading…" {
		t.Errorf("status %q; want loading at column 3", status)
	}
	if body := ansi.Strip(viewLines(loaded(t, nil, 80, 24))[2]); body != "   no agents" {
		t.Errorf("no agents line %q; want it at column 3", body)
	}
}

// footerCells returns the runes of the last rows of the view and checks
// that each cell has the footer background.
func footerCells(t *testing.T, m Model, rows int) [][]rune {
	t.Helper()
	lines := viewLines(m)
	out := make([][]rune, rows)
	width := m.screenWidth()
	for i, line := range lines[len(lines)-rows:] {
		cells := parseCells(line)
		if len(cells) != width {
			t.Fatalf("footer row %q has %d cells; want %d", ansi.Strip(line), len(cells), width)
		}
		for x, c := range cells {
			if !strings.HasSuffix(c.Style, style.Footer) {
				t.Errorf("footer row %d cell %d %q has style %q; want the footer background", i, x, c.R, c.Style)
			}
			out[i] = append(out[i], c.R)
		}
	}
	return out
}

const legendText = "⠋ working   ○ idle   × blocked   ✓ done   · unknown   – gone"

// TestFooterHasOneRowWhenTheLegendAndTheKeysFit checks section 1.5 at 200
// columns: one row with the legend at column 3 and the keys ending at column
// 195, and each cell in the footer background.
func TestFooterHasOneRowWhenTheLegendAndTheKeysFit(t *testing.T) {
	keys := "[tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit"
	m := loaded(t, fixture(), 200, 40)
	row := string(footerCells(t, m, 1)[0])
	want := "   " + legendText + strings.Repeat(" ", 200-3-60-73-4) + keys + "    "
	if row != want {
		t.Errorf("footer = %q\nwant     %q", row, want)
	}
	if above := ansi.Strip(viewLines(m)[len(viewLines(m))-2]); strings.Contains(above, "working") {
		t.Errorf("the row above the footer holds the legend: %q", above)
	}
	if lines := viewLines(m); len(lines) != 40 {
		t.Errorf("view has %d lines; want 40 with a pod area of 37 rows", len(lines))
	}
	if m.area() != 37 {
		t.Errorf("pod area = %d; want 37", m.area())
	}
	// The hidden and the details key lines are shorter. They end in the same
	// column, and the footer keeps one row.
	for _, m := range []Model{key(t, m, tea.KeyEscape), key(t, m, tea.KeyEnter)} {
		row := strings.TrimRight(string(footerCells(t, m, 1)[0]), " ")
		if len([]rune(row)) != 196 || !strings.HasPrefix(row, "   "+legendText) {
			t.Errorf("footer = %q; want the legend and keys ending at column 195", row)
		}
	}
}

// TestFooterHasTwoRowsWhenTheyDoNotFit checks the two-row footer at 132 and
// 80 columns, and at 142, one column below the limit of the brief's rule
// 3 + L + 3 + K + 4 <= W with L = 60 and K = 73.
func TestFooterHasTwoRowsWhenTheyDoNotFit(t *testing.T) {
	keys := "[tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit"
	for _, width := range []int{142, 132, 80} {
		m := loaded(t, fixture(), width, 40)
		rows := footerCells(t, m, 2)
		if got := strings.TrimRight(string(rows[0]), " "); got != "   "+legendText {
			t.Errorf("width %d: legend row = %q", width, got)
		}
		wantKeys := ansi.Truncate("   "+keys, width, "")
		if got := strings.TrimRight(string(rows[1]), " "); got != wantKeys {
			t.Errorf("width %d: key row = %q; want %q", width, got, wantKeys)
		}
		if m.area() != 36 {
			t.Errorf("width %d: pod area = %d; want 36", width, m.area())
		}
	}
	if m := loaded(t, fixture(), 143, 40); m.area() != 37 {
		t.Errorf("width 143: pod area = %d; want 37 with a one-row footer", m.area())
	}
}
