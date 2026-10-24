package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestTaskCardSelected checks work.md rule 4 on the selected card of GfWork
// (rows 15 to 18, columns 33 to 56): the heavy accent border, the dot, the
// ID in accent bold, and line 2 in dim.
func TestTaskCardSelected(t *testing.T) {
	c := TaskCard{ID: "12-board-fixtures", Line: "t12-worker-1 · 12m", Dot: "○", DotColor: Dim, Selected: true, Working: true}
	want := "hive-grid 1 24 4\n" +
		"t|┏━━━━━━━━━━━━━━━━━━━━━━┓\n" + "k|aaaaaaaaaaaaaaaaaaaaaaaa\n" + "b|000000000000000000000000\n" +
		"t|┃ ○ 12-board-fixtures  ┃\n" + "k|afdfaaaaaaaaaaaaaaaaaafa\n" + "b|000011111111111111111100\n" +
		"t|┃ t12-worker-1 · 12m   ┃\n" + "k|afddddddddddddddddddddfa\n" + "b|000000000000000000000000\n" +
		"t|┗━━━━━━━━━━━━━━━━━━━━━━┛\n" + "k|aaaaaaaaaaaaaaaaaaaaaaaa\n" + "b|000000000000000000000000\n"
	checkHandGrid(t, c.View(), want)
}

// TestTaskCardPlain checks a card that is not selected, GfWork rows 5 to 8,
// columns 63 to 86: the border in border, the ID in fg.
func TestTaskCardPlain(t *testing.T) {
	c := TaskCard{ID: "29-frame-layout", Line: "waits on 21", Dot: "○", DotColor: Dim}
	want := "hive-grid 1 24 4\n" +
		"t|╭──────────────────────╮\n" + "k|rrrrrrrrrrrrrrrrrrrrrrrr\n" + "b|000000000000000000000000\n" +
		"t|│ ○ 29-frame-layout    │\n" + "k|rfdffffffffffffffffffffr\n" + "b|000000000000000000000000\n" +
		"t|│ waits on 21          │\n" + "k|rfddddddddddddddddddddfr\n" + "b|000000000000000000000000\n" +
		"t|╰──────────────────────╯\n" + "k|rrrrrrrrrrrrrrrrrrrrrrrr\n" + "b|000000000000000000000000\n"
	checkHandGrid(t, c.View(), want)
}

// TestTaskCardWorkingAndAccentLine checks tui-v2 decision 60, the border in
// working for the card of a working worker, and the accent line 2 of the
// form verify fail of rule 4.
func TestTaskCardWorkingAndAccentLine(t *testing.T) {
	c := TaskCard{ID: "7-x", Line: "verify fail · V4", LineColor: Accent, Dot: "○", DotColor: Dim, Working: true}
	want := "hive-grid 1 24 4\n" +
		"t|╭──────────────────────╮\n" + "k|wwwwwwwwwwwwwwwwwwwwwwww\n" + "b|000000000000000000000000\n" +
		"t|│ ○ 7-x                │\n" + "k|wfdffffffffffffffffffffw\n" + "b|000000000000000000000000\n" +
		"t|│ verify fail · V4     │\n" + "k|wfaaaaaaaaaaaaaaaaaaaafw\n" + "b|000000000000000000000000\n" +
		"t|╰──────────────────────╯\n" + "k|wwwwwwwwwwwwwwwwwwwwwwww\n" + "b|000000000000000000000000\n"
	checkHandGrid(t, c.View(), want)
}

// TestTaskCardCutsLongText checks the cut with … of the ID at 18 cells and
// of line 2 at 20 cells, as GfWorkCancelled row 27 shows.
func TestTaskCardCutsLongText(t *testing.T) {
	c := TaskCard{ID: "13-grid-diff-and-more-words", Line: "cancelled · merged in 9", Dot: "×", DotColor: Dim}
	want := "╭──────────────────────╮\n│ × 13-grid-diff-and-… │\n│ cancelled · merged … │\n╰──────────────────────╯"
	if got := ansi.Strip(c.View()); got != want {
		t.Errorf("View:\n%s\nwant:\n%s", got, want)
	}
}
