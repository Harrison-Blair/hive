package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestInputGolden renders the Input section, Components.dc.html line 69, and
// the input of Chat.dc.html lines 46-48.
func TestInputGolden(t *testing.T) {
	t.Run("components", func(t *testing.T) {
		checkGolden(t, Input{Text: "restart the supervisor please", Focused: true}.View(120), sheet, rect{46, 0, 120, 3})
	})
	t.Run("chat", func(t *testing.T) {
		checkGolden(t, Input{Text: "1A, 2A, 3B", Focused: true, Status: "working", Age: "1m"}.View(120), "Chat", rect{36, 0, 120, 3})
	})
}

// TestInputUnfocused checks that the input without focus has a border
// border and no cursor.
func TestInputUnfocused(t *testing.T) {
	got := Input{Text: "hi"}.View(10)
	if strings.Contains(got, "█") || strings.Contains(got, accentStyle.Render("╭")) {
		t.Errorf("unfocused input has a cursor or an accent border:\n%s", got)
	}
}

func TestInputGrowsToFitWrappedText(t *testing.T) {
	view := Input{Text: "one two three four five six seven", Focused: true}.View(14)
	if rows := len(strings.Split(view, "\n")); rows != 7 {
		t.Fatalf("input has %d rows; want 7 rows for five wrapped text rows and borders:\n%s", rows, ansi.Strip(view))
	}
}

// TestInputShowsWrappedTextAndCursor checks that long words and newlines
// stay visible with the cursor at the end of the text.
func TestInputShowsWrappedTextAndCursor(t *testing.T) {
	long := strings.Split(ansi.Strip(Input{Text: strings.Repeat("x", 40) + "END", Focused: true}.View(16)), "\n")
	if len(long) != 7 || !strings.Contains(long[5], "END█") {
		t.Fatalf("long input rows = %q; want five text rows and END█ on the last row", long)
	}
	multi := strings.Split(ansi.Strip(Input{Text: "first line\nsecond", Focused: true}.View(16)), "\n")
	if len(multi) != 4 || !strings.Contains(multi[2], "second█") {
		t.Fatalf("multiline input rows = %q; want the second line and cursor", multi)
	}
}

// TestInputCursorFollowsLineAndColumn checks that the cursor shows on the
// wrapped row of its line and rune column. The text width is 8 cells.
func TestInputCursorFollowsLineAndColumn(t *testing.T) {
	for _, tc := range []struct {
		name      string
		text      string
		line, col int
		row       int
		want      string
	}{
		{"on the soft break space", "one two three", 0, 7, 1, "> one two█"},
		{"after the soft break", "one two three", 0, 8, 2, "  █hree"},
		{"first row", "one two three", 0, 1, 1, "> o█e two"},
		{"second line", "ab\ncd", 1, 1, 2, "  c█"},
		{"inside a broken word", "xxxxxxxxxxyy", 0, 9, 2, "  x█yy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := strings.Split(ansi.Strip(Input{Text: tc.text, Focused: true, HasCursor: true, CursorLine: tc.line, CursorColumn: tc.col}.View(14)), "\n")
			if !strings.Contains(rows[tc.row], tc.want) {
				t.Errorf("row %d = %q; want %q in rows %q", tc.row, rows[tc.row], tc.want, rows)
			}
		})
	}
}
