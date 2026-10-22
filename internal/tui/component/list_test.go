package component

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

// sampleList is the sample of the List section, Components.dc.html lines
// 54-58.
func sampleList() List {
	done := lipgloss.NewStyle().Foreground(Done)
	dim := lipgloss.NewStyle().Foreground(Dim)
	row := func(id, status, wave string) []Cell {
		return []Cell{{Text: id}, {Text: status, Style: done}, {Text: wave, Style: dim}}
	}
	return List{
		Title:   "tasks · 4",
		Focused: true,
		Columns: []Column{{Title: "ID", Width: 26}, {Title: "STATUS", Width: 10}, {Title: "WAVE"}},
		Rows: [][]Cell{
			row("4-agents-equal-columns", "landed", "C"),
			row("3-tui-restart", "landed", "B"),
			row("2-stale-coordinator-startup", "landed", "A"),
		},
		Selected: 1,
	}
}

// TestListGolden renders the List section, Components.dc.html lines 54-58.
func TestListGolden(t *testing.T) {
	checkGolden(t, sampleList().View(60, 6), sheet, rect{33, 0, 60, 6})
}

// TestListKeepsSelectionVisible moves through 20 rows in a list with 3 row
// lines and checks that the selected row is on screen after each move.
func TestListKeepsSelectionVisible(t *testing.T) {
	l := List{Columns: []Column{{Title: "ID"}}}
	for i := range 20 {
		l.Rows = append(l.Rows, []Cell{{Text: fmt.Sprintf("row-%02d", i)}})
	}
	const height = 6 // 2 border rows, the header row, 3 row lines
	moves := []struct{ delta, selected int }{{1, 1}, {5, 6}, {30, 19}, {-2, 17}, {-17, 0}}
	for _, m := range moves {
		l.Move(m.delta, height)
		if l.Selected != m.selected {
			t.Fatalf("Move(%d): Selected %d, want %d", m.delta, l.Selected, m.selected)
		}
		want := fmt.Sprintf("│ ▶ row-%02d", m.selected)
		if !strings.Contains(ansi.Strip(l.View(20, height)), want) {
			t.Errorf("Move(%d): selected row %d is not visible:\n%s", m.delta, m.selected, ansi.Strip(l.View(20, height)))
		}
	}
}

// TestListScrollMark checks the ┃ mark on the right border when the rows do
// not fit, and no mark when they fit.
func TestListScrollMark(t *testing.T) {
	l := List{Columns: []Column{{Title: "ID"}}}
	for i := range 3 {
		l.Rows = append(l.Rows, []Cell{{Text: fmt.Sprintf("r%d", i)}})
	}
	want := "╭──────────╮\n│   ID     │\n│ ▶ r0     │\n│   r1     │\n│   r2     │\n╰──────────╯"
	if got := ansi.Strip(l.View(12, 6)); got != want {
		t.Errorf("rows that fit:\n%s\nwant:\n%s", got, want)
	}
	l.Selected, l.Offset = 2, 2
	want = "╭──────────╮\n│   ID     │\n│   r1     │\n│ ▶ r2     ┃\n╰──────────╯"
	if got := ansi.Strip(l.View(12, 5)); got != want {
		t.Errorf("rows that do not fit:\n%s\nwant:\n%s", got, want)
	}
}

// TestListFillGolden renders the tasks list of the Work board, rows 4-38 and
// columns 0-51. With Fill, the padding of each cell has the style of the
// cell.
func TestListFillGolden(t *testing.T) {
	l := sampleList()
	info := lipgloss.NewStyle().Foreground(Info)
	l.Rows = append(l.Rows, []Cell{{Text: "1-cp5-end-to-end"}, {Text: "assigned", Style: info}, {Text: "—", Style: lipgloss.NewStyle().Foreground(Dim)}})
	l.Fill = true
	checkGolden(t, l.View(52, 35), "Work", rect{4, 0, 52, 35})
}

// TestListRowCells checks the user decisions 4 and 7 of 2026-10-21 and the
// Friction decision of the manager: "▶" is accent bold, the space after it
// is fg and not bold, and the first cell is accent bold over its full column.
// Column 2 of each other row is fg and not bold.
func TestListRowCells(t *testing.T) {
	l := List{Columns: []Column{{Title: "ID", Width: 6}, {Title: "S"}}, Rows: [][]Cell{{{Text: "ab"}, {Text: "x"}}, {{Text: "cd"}, {Text: "y"}}}}
	g, err := cellgrid.Convert(strings.Join(strings.Split(l.View(14, 5), "\n")[2:4], "\n"), 14, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		row, col int
		rune     rune
		token    cellgrid.Token
		bold     bool
	}{
		{0, 2, '▶', cellgrid.Accent, true},
		{0, 3, ' ', cellgrid.FG, false},
		{0, 4, 'a', cellgrid.Accent, true},
		{0, 5, 'b', cellgrid.Accent, true},
		{0, 9, ' ', cellgrid.Accent, true},
		{0, 10, 'x', cellgrid.FG, false},
		{1, 2, ' ', cellgrid.FG, false},
		{1, 3, ' ', cellgrid.FG, false},
		{1, 4, 'c', cellgrid.FG, false},
	}
	for _, w := range want {
		c := g.At(w.row, w.col)
		if c.Rune != w.rune || c.Token != w.token || c.Bold != w.bold {
			t.Errorf("row %d col %d = %q %s bold %t; want %q %s bold %t", w.row, w.col, c.Rune, c.Token, c.Bold, w.rune, w.token, w.bold)
		}
	}
}
