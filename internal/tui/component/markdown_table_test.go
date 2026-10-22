package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// The tests of decision 32 record what Glamour v2.0.1 does with a table. The
// user chose the Glamour output with no change on 2026-10-21. A change of
// Glamour that changes a table fails these tests.

// renderPlain renders src at width and gives the lines with no style. Each
// line must fit the width.
func renderPlain(t *testing.T, src string, width int) string {
	t.Helper()
	lines, err := RenderMarkdown(src, width)
	if err != nil {
		t.Fatal(err)
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w > width {
			t.Errorf("line %d is %d cells wide, wrap width %d: %q", i, w, width, ansi.Strip(l))
		}
		lines[i] = ansi.Strip(l)
	}
	return strings.Join(lines, "\n")
}

// TestMarkdownTableUneven renders a table of 51 cells at 40. The widest
// column shrinks, and its cell wraps to a second line.
func TestMarkdownTableUneven(t *testing.T) {
	src := "| id | description | state |\n| --- | --- | --- |\n| 1 | a long description of the task here | landed |\n"
	want := " id │ description              │ state\n" +
		"────┼──────────────────────────┼────────\n" +
		" 1  │ a long description of    │ landed\n" +
		"    │ the task here            │"
	if got := renderPlain(t, src, 40); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestMarkdownTableEqualColumns renders two columns of the same width. At 24
// cells only the last space of the right column goes. At 40 cells the
// columns stretch to the wrap width, and the left column grows first.
func TestMarkdownTableEqualColumns(t *testing.T) {
	src := "| aaaaaaaaaa | bbbbbbbbbb |\n| --- | --- |\n| x | y |\n"
	cases := []struct {
		width int
		want  string
	}{
		{24, " aaaaaaaaaa │ bbbbbbbbbb\n" +
			"────────────┼───────────\n" +
			" x          │ y"},
		{40, " aaaaaaaaaa         │ bbbbbbbbbb\n" +
			"────────────────────┼───────────────────\n" +
			" x                  │ y"},
	}
	for _, tc := range cases {
		if got := renderPlain(t, src, tc.width); got != tc.want {
			t.Errorf("width %d, got:\n%s\nwant:\n%s", tc.width, got, tc.want)
		}
	}
}

// TestMarkdownTableTooManyColumns renders 6 columns at 20 cells. Each column
// gets one or two cells: a header shows nothing or …, and a value wraps one
// character to a line. Glamour has no "Header: value" form.
func TestMarkdownTableTooManyColumns(t *testing.T) {
	src := "| c0 | c1 | c2 | c3 | c4 | c5 |\n|---|---|---|---|---|---|\n| v0 | v1 | v2 | v3 | v4 | v5 |\n"
	want := "  │  │  │ … │ … │ …\n" +
		"──┼──┼──┼───┼───┼───\n" +
		" v│ v│ v│ v │ v │ v\n" +
		"  │  │  │ 3 │ 4 │ 5"
	if got := renderPlain(t, src, 20); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestMarkdownTableTokens checks the style of a table: the header row and
// the cells in fg and not bold, the rules in border, and inline code in a
// cell in code.
func TestMarkdownTableTokens(t *testing.T) {
	want := "hive-grid 1 12 3\n" +
		"t| a    │ b   \n" + "k|ffffffrfffff\n" + "b|000000000000\n" +
		"t|──────┼─────\n" + "k|rrrrrrrrrrrr\n" + "b|000000000000\n" +
		"t| x    │ y   \n" + "k|fcffffrfffff\n" + "b|000000000000\n"
	lines, err := RenderMarkdown("| a | b |\n| --- | --- |\n| `x` | y |\n", 12)
	if err != nil {
		t.Fatal(err)
	}
	checkHandGrid(t, strings.Join(lines, "\n"), want)
}
