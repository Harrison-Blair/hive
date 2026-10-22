package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
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
	got := renderGrid(t, m, want.Width, want.Height)
	rows := []int{0, 1, 2, want.Height - 1}
	diffs, err := cellgrid.Compare(gridRows(want, rows...), gridRows(got, rows...), nil)
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
