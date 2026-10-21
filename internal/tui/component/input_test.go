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
		checkGolden(t, Input{Text: "1A, 2A, 3B", Focused: true}.View(120), "Chat", rect{36, 0, 120, 3})
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

// TestInputShowsEndOfText checks that a long text and a text with newlines
// show the end of the last line and the cursor.
func TestInputShowsEndOfText(t *testing.T) {
	cases := []struct{ text, want string }{
		{strings.Repeat("x", 40) + "END", "│ > …xxxxxEND█ │"},
		{"first line\nsecond", "│ > second█    │"},
	}
	for _, tc := range cases {
		row := ansi.Strip(strings.Split(Input{Text: tc.text, Focused: true}.View(16), "\n")[1])
		if row != tc.want {
			t.Errorf("Input %q: row %q, want %q", tc.text, row, tc.want)
		}
	}
}
