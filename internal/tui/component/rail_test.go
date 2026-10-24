package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestRailUp checks work.md rule 5 on the gap between columns A and B of
// GfWork (rows 3 to 36, columns 29 and 30): a rail from row 13, the name line
// of 9-cellgrid, up to row 3, the name line of 10-board-export, in border.
func TestRailUp(t *testing.T) {
	r := Rail{Height: 34, Links: []RailLink{{From: 13, To: 3}}}
	checkTestdataGrid(t, r.View(), "rail-up.grid")
}

// TestRailJoinsLinks checks that rails from one row join in one column: up,
// straight, and down from row 2.
func TestRailJoinsLinks(t *testing.T) {
	r := Rail{Height: 6, Links: []RailLink{{From: 2, To: 0}, {From: 2, To: 2}, {From: 2, To: 4}}}
	want := strings.Join([]string{" ╭", " │", "─┼", " │", " ╰", "  "}, "\n")
	if got := ansi.Strip(r.View()); got != want {
		t.Errorf("View:\n%q\nwant:\n%q", got, want)
	}
}

// TestRailEmpty checks that a gap with no rail is two blank cells on each
// row.
func TestRailEmpty(t *testing.T) {
	want := strings.Join([]string{"  ", "  ", "  "}, "\n")
	if got := (Rail{Height: 3}).View(); got != want {
		t.Errorf("View: %q, want %q", got, want)
	}
}
