package component

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
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

// TestListSelectedRowFillsItsColumn checks user decisions 4 and 7 of
// 2026-10-21: the marker ▶ is accent and bold, the space after it is fg and
// not bold, and the selected name is accent and bold over the full column,
// its gap included.
func TestListSelectedRowFillsItsColumn(t *testing.T) {
	l := List{Columns: []Column{{Title: "ID", Width: 6}, {Title: "S"}}, Rows: [][]Cell{{{Text: "ab"}, {Text: "x"}}}}
	g, err := cellgrid.Convert(l.View(16, 4), 16, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]cellgrid.Cell{2: {Rune: '▶', Token: cellgrid.Accent, Bold: true}, 3: {Rune: ' ', Token: cellgrid.FG}, 10: {Rune: 'x', Token: cellgrid.FG}}
	for col, r := range "ab    " {
		want[4+col] = cellgrid.Cell{Rune: r, Token: cellgrid.Accent, Bold: true}
	}
	for col, cell := range want {
		if got := g.At(2, col); got != cell {
			t.Errorf("col %d: got %+v, want %+v", col, got, cell)
		}
	}
}
