package component

import "testing"

// TestNoticeGrid renders the one-line notice of decision 19 at 50 cells and
// compares it with a hand-written grid: one space, the text in dim, then
// spaces in fg. No board draws a sample notice.
func TestNoticeGrid(t *testing.T) {
	want := "hive-grid 1 50 1\n" +
		"t| restart: coordinator and supervisor only         \n" +
		"k|fddddddddddddddddddddddddddddddddddddddddfffffffff\n" +
		"b|00000000000000000000000000000000000000000000000000\n"
	checkHandGrid(t, Notice{Text: "restart: coordinator and supervisor only"}.View(50), want)
}

// TestNoticeCutsLongText checks that a text longer than the width is cut
// with … in dim.
func TestNoticeCutsLongText(t *testing.T) {
	want := "hive-grid 1 12 1\n" +
		"t| no project…\n" +
		"k|fddddddddddd\n" +
		"b|000000000000\n"
	checkHandGrid(t, Notice{Text: "no project agents"}.View(12), want)
}
