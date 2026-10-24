package component

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// designDir holds the golden grid of each tui-v2 board. gfDesignDir holds
// the golden grid of each golden-frames board, whose name starts with Gf.
const (
	designDir   = "../../../.hive/spec/tui-v2/design/"
	gfDesignDir = "../../../.hive/spec/golden-frames/design/"
)

// sheet is the component sheet, Components.dc.html. gfSheet is the
// golden-frames component sheet, GfComponents.dc.html.
const (
	sheet   = "Components"
	gfSheet = "GfComponents"
)

// boardPath gives the golden grid file of board.
func boardPath(board string) string {
	if strings.HasPrefix(board, "Gf") {
		return gfDesignDir + board + ".grid"
	}
	return designDir + board + ".grid"
}

// rect is a cell rectangle of a board: zero-based row and column of the
// top-left cell, then the width and the height.
type rect struct{ row, col, w, h int }

// boardRect gives the cells of r from the golden grid of board.
func boardRect(t *testing.T, board string, r rect) cellgrid.Grid {
	t.Helper()
	requireSpecFile(t, boardPath(board))
	f, err := os.Open(boardPath(board))
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

// checkHandGrid compares the rendered ANSI text with a hand-written grid in
// the grid file format of cellgrid. Each differing cell is printed.
func checkHandGrid(t *testing.T, got, want string) {
	t.Helper()
	w, err := cellgrid.ReadGrid(strings.NewReader(want))
	if err != nil {
		t.Fatalf("read the expected grid: %v", err)
	}
	g, err := cellgrid.Convert(got, w.Width, w.Height)
	if err != nil {
		t.Fatalf("convert: %v\n%s", err, got)
	}
	diffs, err := cellgrid.Compare(w, g, nil)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if len(diffs) > 0 {
		t.Errorf("%d cells differ from the expected grid:\n%s\nrendered:\n%s", len(diffs), cellgrid.FormatDiffs(diffs), got)
	}
}

// requireSpecFile skips the test when a file or a folder of the local spec is
// missing. Git does not track .hive/spec/, so a clean clone has no grid and
// no fixture. The test runs as before when the file exists.
func requireSpecFile(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("local spec file missing: %s: Git does not track .hive/spec/", path)
	}
}
