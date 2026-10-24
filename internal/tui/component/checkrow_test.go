package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var checkRowTiles = []CheckTile{{Text: "starting 0", Passed: true}, {Text: "unresolved 1", Selected: true}}

// Golden-frames health.md rule 1: the area name in dim bold in 11 cells, a
// passed tile with ✓ in done, and the selected tile with ▶, ! in accent, and
// its text in accent bold.
func TestCheckRowMarksAndSelection(t *testing.T) {
	want := "hive-grid 1 46 1\n" +
		"t| agents       ✓ starting 0  ▶ ! unresolved 1  \n" +
		"k|fdddddddddddffofffffffffffffafafaaaaaaaaaaaaff\n" +
		"b|0111111111110010000000000000101011111111111100\n"
	checkHandGrid(t, CheckRow{Area: "agents", Tiles: checkRowTiles}.View(46), want)
}

// Rule 7: while hive doctor runs, each mark and text is dim. The ▶ and the
// text of the selected tile stay in accent bold.
func TestCheckRowDimWhileChecking(t *testing.T) {
	tiles := []CheckTile{checkRowTiles[0], checkRowTiles[1]}
	for index := range tiles {
		tiles[index].Dim = true
	}
	want := "hive-grid 1 46 1\n" +
		"t| agents       ✓ starting 0  ▶ ! unresolved 1  \n" +
		"k|fdddddddddddffdfddddddddddffafdfaaaaaaaaaaaaff\n" +
		"b|0111111111110010000000000000101011111111111100\n"
	checkHandGrid(t, CheckRow{Area: "agents", Tiles: tiles}.View(46), want)
}

// A row wider than the width is cut at the width, and a long area name is
// cut with … in its 11 cells.
func TestCheckRowCutsAtTheWidth(t *testing.T) {
	cases := []struct {
		row   CheckRow
		width int
		want  string
	}{
		{CheckRow{Area: "agents", Tiles: checkRowTiles}, 24, " agents       ✓ starting"},
		{CheckRow{Area: "a very long area"}, 14, " a very lon…  "},
	}
	for _, tc := range cases {
		if got := ansi.Strip(tc.row.View(tc.width)); got != tc.want {
			t.Errorf("View(%d) = %q; want %q", tc.width, got, tc.want)
		}
	}
}
