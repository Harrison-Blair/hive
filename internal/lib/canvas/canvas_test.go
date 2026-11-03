package canvas

import (
	"slices"
	"testing"
)

func TestClipCutsToTheLimitWithAnEllipsis(t *testing.T) {
	for _, tc := range []struct {
		text  string
		limit int
		want  string
	}{
		{"claude", 6, "claude"},
		{"claude", 5, "clau…"},
		{"claude", 1, "…"},
		{"claude", 0, ""},
		{"claude", -3, ""},
	} {
		if got := Clip(tc.text, tc.limit); got != tc.want {
			t.Errorf("Clip(%q, %d) = %q; want %q", tc.text, tc.limit, got, tc.want)
		}
	}
}

func TestLinesJoinRunsOfOneStyleAndDropTrailingBlanks(t *testing.T) {
	c := New(8, 2)
	c.Box(0, 0, 4, 2, []rune("╭╮╰╯─│"), "2")
	c.Text(5, 0, "abcdef", "1", 2)
	c.Set(9, 0, 'x', "") // outside: ignored
	want := []string{"\x1b[2m╭──╮\x1b[0m \x1b[1mab\x1b[0m", "\x1b[2m╰──╯\x1b[0m"}
	if got := c.Lines(); !slices.Equal(got, want) {
		t.Fatalf("Lines = %q; want %q", got, want)
	}
}
