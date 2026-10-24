package component

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// checkTestdataGrid compares the rendered text with the grid file name in
// testdata. Each grid file is a cut of a Work board.
func checkTestdataGrid(t *testing.T, got, name string) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	checkHandGrid(t, got, string(want))
}

// TestWaveColumnSelectedWithRails checks work.md rules 3 and 5 on column B of
// GfWork (rows 3 to 36, columns 31 to 58): the heavy accent border of the
// column of the selected card, the title, the cards at a stride of 5 rows,
// the rail joint ┠ and ▶ on the left border, the joint ┨ on the right
// border, and the foot in dim.
func TestWaveColumnSelectedWithRails(t *testing.T) {
	c := WaveColumn{
		Title:    "B · running · 2/4",
		Selected: true,
		Cards: []TaskCard{
			{ID: "10-board-export", Line: "landed · 2adaf0", Dot: "●", DotColor: Done},
			{ID: "11-demo-mode", Line: "verified · V5 pass", Dot: "◐", DotColor: Working},
			{ID: "12-board-fixtures", Line: "t12-worker-1 · 12m", Dot: "○", DotColor: Dim, Selected: true, Working: true},
			{ID: "21-components-1", Line: "needs you · mask ok?", Dot: "!", DotColor: Accent},
		},
		Foot:     "done when dev",
		Height:   34,
		RailsIn:  []int{3},
		RailsOut: []int{18},
	}
	checkTestdataGrid(t, c.View(), "wavecolumn-selected-rails.grid")
}

// TestWaveColumnDone checks column A of GfWork (rows 3 to 36, columns 1 to
// 28): the rounded border in border, the title in accent bold, the foot with
// ✓ done in done, and the joint ┤ of a rail that leaves the right border.
func TestWaveColumnDone(t *testing.T) {
	c := WaveColumn{
		Title: "A · done · 3/3",
		Cards: []TaskCard{
			{ID: "5-tui-v2-spec", Line: "landed · 478dc8", Dot: "●", DotColor: Done},
			{ID: "8-no-startup", Line: "landed · 61afa0", Dot: "●", DotColor: Done},
			{ID: "9-cellgrid", Line: "landed · 0026af", Dot: "●", DotColor: Done},
		},
		Foot:     "done when dev",
		Done:     true,
		Height:   34,
		RailsOut: []int{13},
	}
	checkTestdataGrid(t, c.View(), "wavecolumn-done-rail-out.grid")
}

// TestWaveColumnPlanned checks column E of GfWorkScroll (rows 3 to 35,
// columns 91 to 118), a column of 33 rows with two cards.
func TestWaveColumnPlanned(t *testing.T) {
	c := WaveColumn{
		Title: "E · planned · 0/2",
		Cards: []TaskCard{
			{ID: "47-screenshots", Line: "waits on wave D", Dot: "○", DotColor: Dim},
			{ID: "48-pr-update", Line: "waits on 47", Dot: "○", DotColor: Dim},
		},
		Foot:   "done when main",
		Height: 33,
	}
	checkTestdataGrid(t, c.View(), "wavecolumn-planned.grid")
}

// TestWaveColumnSlots checks the cards that fit: a card needs 4 rows and one
// blank row, from row 2, above the foot row. GfWork (34 rows) and
// GfWorkScroll (33 rows) show 6 cards.
func TestWaveColumnSlots(t *testing.T) {
	for _, tc := range []struct{ height, want int }{{34, 6}, {33, 6}, {32, 5}, {9, 1}, {8, 1}, {7, 0}} {
		if got := WaveColumnSlots(tc.height); got != tc.want {
			t.Errorf("WaveColumnSlots(%d) = %d, want %d", tc.height, got, tc.want)
		}
	}
}

// TestWaveColumnCutsCardsAndDrawsThumb checks rule 7: the column shows only
// the cards that fit, and the ┃ thumb is on the right border in dim. Rule 12:
// the row above the foot shows the hidden cancelled cards in dim, and the
// cards keep that row free.
func TestWaveColumnCutsCardsAndDrawsThumb(t *testing.T) {
	card := TaskCard{ID: "1-a", Line: "created", Dot: "○", DotColor: Dim}
	c := WaveColumn{Title: "A · planned · 0/3", Cards: []TaskCard{card, card}, Foot: "done when dev", Hint: "h shows 1 cancelled", Height: 10, Thumb: ScrollMark{Start: 0, Size: 3}}
	got := c.View()
	want := strings.Join([]string{
		"╭─ A · planned · 0/3 ──────╮",
		"│                          ┃",
		"│ ╭──────────────────────╮ ┃",
		"│ │ ○ 1-a                │ ┃",
		"│ │ created              │ │",
		"│ ╰──────────────────────╯ │",
		"│                          │",
		"│ h shows 1 cancelled      │",
		"│ done when dev            │",
		"╰──────────────────────────╯",
	}, "\n")
	if plain := ansi.Strip(got); plain != want {
		t.Fatalf("View:\n%s\nwant:\n%s", plain, want)
	}
	grid, err := cellgrid.Convert(got, 28, 10)
	if err != nil {
		t.Fatal(err)
	}
	if cell := grid.At(1, 27); cell.Token != cellgrid.Dim {
		t.Errorf("thumb token %v, want dim", cell.Token)
	}
	if cell := grid.At(7, 2); cell.Token != cellgrid.Dim {
		t.Errorf("hint token %v, want dim", cell.Token)
	}
}
