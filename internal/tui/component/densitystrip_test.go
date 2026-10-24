package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// convertStrip converts the three rows of a strip at width cells.
func convertStrip(t *testing.T, s DensityStrip, width int) cellgrid.Grid {
	t.Helper()
	got := s.View(width)
	if rows := strings.Split(got, "\n"); len(rows) != 3 {
		t.Fatalf("View gives %d rows, want 3:\n%s", len(rows), got)
	}
	g, err := cellgrid.Convert(got, width, 3)
	if err != nil {
		t.Fatalf("convert: %v\n%s", err, got)
	}
	return g
}

// logs.md rule 1: a cell is ▁ to ▇ by its count relative to the largest
// count, and an empty bucket is blank. The start and the end time sit at the
// ends. A WARN cell is accent, an ERROR cell is error, each other cell dim.
func TestDensityStripCells(t *testing.T) {
	s := DensityStrip{
		Start: "13:00", End: "14:22",
		Cells: []DensityCell{{Count: 7}, {Count: 1}, {Count: 0}, {Count: 4, Level: "WARN"}, {Count: 2, Level: "ERROR"}, {Count: 3}},
		Mark:  -1,
	}
	g := convertStrip(t, s, 22)
	rows := strings.Split(ansi.Strip(s.View(22)), "\n")
	if want := " 13:00 ▇▁ ▄▂▃ 14:22   "; rows[0] != want {
		t.Errorf("strip row = %q, want %q", rows[0], want)
	}
	if want := "       ──────         "; rows[1] != want {
		t.Errorf("rule row = %q, want %q", rows[1], want)
	}
	if strings.TrimSpace(rows[2]) != "" {
		t.Errorf("marker row = %q, want blank with no mark", rows[2])
	}
	for col, want := range map[int]cellgrid.Token{1: cellgrid.Dim, 7: cellgrid.Dim, 8: cellgrid.Dim, 10: cellgrid.Accent, 11: cellgrid.Error, 12: cellgrid.Dim, 14: cellgrid.Dim} {
		if got := g.At(0, col).Token; got != want {
			t.Errorf("strip col %d token %v, want %v", col, got, want)
		}
	}
	if got := g.At(1, 7).Token; got != cellgrid.Border {
		t.Errorf("rule token %v, want border", got)
	}
}

// logs.md rule 2: the rule shows ┴ under the cell of the selected event.
// The row under it shows the marker text in the colour of the level, with
// the level word under the ┴.
func TestDensityStripMark(t *testing.T) {
	cells := make([]DensityCell, 20)
	for i := range cells {
		cells[i].Count = 1
	}
	s := DensityStrip{Start: "13:00", End: "14:22", Cells: cells, Mark: 10, MarkText: "13:38 warn · verify fail", MarkLevel: "WARN"}
	g := convertStrip(t, s, 36)
	rows := strings.Split(ansi.Strip(s.View(36)), "\n")
	if got := []rune(rows[1])[17]; got != '┴' {
		t.Errorf("rule col 17 = %q, want ┴: %q", got, rows[1])
	}
	if g.At(1, 17).Token != cellgrid.Accent {
		t.Errorf("┴ token %v, want accent", g.At(1, 17).Token)
	}
	if want := "           13:38 warn · verify fail"; strings.TrimRight(rows[2], " ") != want {
		t.Errorf("marker row = %q, want %q", rows[2], want)
	}
	if g.At(2, 17).Rune != 'w' || g.At(2, 17).Token != cellgrid.Accent {
		t.Errorf("marker col 17 = %q %v, want w in accent", g.At(2, 17).Rune, g.At(2, 17).Token)
	}
	// A mark near the end keeps its text inside the row. A mark at the
	// start keeps the text from the first cell.
	s.Mark, s.MarkLevel = 19, "ERROR"
	end := strings.Split(ansi.Strip(s.View(36)), "\n")[2]
	if !strings.HasSuffix(strings.TrimRight(end, " "), "verify fail") || len([]rune(end)) != 36 {
		t.Errorf("marker row at the end = %q, want the full text inside 36 cells", end)
	}
	if g := convertStrip(t, s, 36); g.At(1, 26).Token != cellgrid.Error {
		t.Errorf("ERROR ┴ token %v, want error", g.At(1, 26).Token)
	}
	s.Mark = 0
	if start := strings.Split(ansi.Strip(s.View(36)), "\n")[2]; !strings.HasPrefix(start, "       13:38") {
		t.Errorf("marker row at the start = %q, want the text from column 7", start)
	}
}
