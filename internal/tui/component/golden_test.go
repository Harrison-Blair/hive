package component

import (
	"os"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// designDir holds the golden grid of each board.
const designDir = "../../../.hive/spec/tui-v2/design/"

// sheet is the component sheet, Components.dc.html.
const sheet = "Components"

// rect is a cell rectangle of a board: zero-based row and column of the
// top-left cell, then the width and the height.
type rect struct{ row, col, w, h int }

// boardRect gives the cells of r from the golden grid of board.
func boardRect(t *testing.T, board string, r rect) cellgrid.Grid {
	t.Helper()
	f, err := os.Open(designDir + board + ".grid")
	if err != nil {
		t.Fatalf("open board: %v", err)
	}
	defer func() { _ = f.Close() }()
	g, err := cellgrid.ReadGrid(f)
	if err != nil {
		t.Fatalf("read board %s: %v", board, err)
	}
	if r.row+r.h > g.Height || r.col+r.w > g.Width {
		t.Fatalf("rect %+v is outside the %dx%d board", r, g.Width, g.Height)
	}
	out := cellgrid.Grid{Width: r.w, Height: r.h}
	for row := r.row; row < r.row+r.h; row++ {
		for col := r.col; col < r.col+r.w; col++ {
			out.Cells = append(out.Cells, g.At(row, col))
		}
	}
	return out
}

// checkGolden converts the rendered ANSI text and compares it with the
// rectangle r of the golden grid of board. lipgloss renders each colour as a
// 24-bit SGR, which is the TrueColor profile: only a colorprofile writer
// downsamples. Each differing cell is printed with its board row and column.
func checkGolden(t *testing.T, got, board string, r rect) {
	t.Helper()
	if w := lipgloss.Width(got); w != r.w {
		t.Errorf("rendered width %d, want %d", w, r.w)
	}
	g, err := cellgrid.Convert(got, r.w, r.h)
	if err != nil {
		t.Fatalf("convert: %v\n%s", err, got)
	}
	diffs, err := cellgrid.Compare(boardRect(t, board, r), g, nil)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	for i := range diffs {
		diffs[i].Row += r.row
		diffs[i].Col += r.col
	}
	if len(diffs) > 0 {
		t.Errorf("%d cells differ from %s.grid (board row and column):\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(diffs), got)
	}
}
