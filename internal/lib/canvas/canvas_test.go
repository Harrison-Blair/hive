package canvas

import (
	"slices"
	"testing"

	"github.com/charmbracelet/x/ansi"
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

func TestTextGivesAWideCharacterTwoCells(t *testing.T) {
	c := New(10, 1)
	c.Text(0, 0, "日本語", "", 10)
	c.Text(6, 0, "ab", "", 2)
	if got := c.Lines()[0]; got != "日本語ab" || ansi.StringWidth(got) != 8 {
		t.Fatalf("line = %q (%d cells); want 日本語ab in 8 cells", got, ansi.StringWidth(got))
	}
}

func TestTextDropsAWideCharacterThatDoesNotFit(t *testing.T) {
	limit := New(10, 1)
	limit.Text(0, 0, "日本語", "", 5)
	edge := New(5, 1)
	edge.Text(0, 0, "日本語", "", 10)
	for name, c := range map[string]*Canvas{"limit": limit, "edge": edge} {
		if got := c.Lines()[0]; got != "日本" {
			t.Errorf("%s: line = %q; want 日本 with the third character dropped", name, got)
		}
	}
}

func TestTextKeepsAnEmojiClusterWhole(t *testing.T) {
	for _, emoji := range []string{"👍🏽", "👨‍👩‍👧"} {
		c := New(6, 1)
		c.Text(0, 0, emoji+"x", "", 6)
		if got := c.Lines()[0]; got != emoji+"x" || ansi.StringWidth(got) != 3 {
			t.Errorf("line = %q (%d cells); want %q in 3 cells", got, ansi.StringWidth(got), emoji+"x")
		}
	}
}

func TestSetOverAHalfOfAWideCharacterBlanksTheOtherHalf(t *testing.T) {
	c := New(6, 2)
	c.Text(0, 0, "日本", "", 6)
	c.Set(1, 0, 'x', "")
	c.Text(0, 1, "日本", "", 6)
	c.Set(2, 1, 'y', "")
	want := []string{" x本", "日y "}
	if got := c.Lines(); !slices.Equal(got, want) {
		t.Fatalf("Lines = %q; want %q", got, want)
	}
}

func TestClipCountsDisplayWidth(t *testing.T) {
	for _, tc := range []struct {
		text  string
		limit int
		want  string
	}{
		{"日本語", 6, "日本語"},
		{"日本語", 5, "日本…"},
		{"日本語", 4, "日…"},
	} {
		if got := Clip(tc.text, tc.limit); got != tc.want {
			t.Errorf("Clip(%q, %d) = %q; want %q", tc.text, tc.limit, got, tc.want)
		}
	}
}
