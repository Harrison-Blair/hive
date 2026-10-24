package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// The two composers of GfComponents rows 19 to 21 (golden-frames chat.md
// rules 9 and 10): the placeholder, and typed text with the cursor.
func TestGfComposerGolden(t *testing.T) {
	checkGolden(t, Composer{Placeholder: "type a message · ? help", Cursor: true, InputRows: 3}.View(58), gfSheet, rect{19, 0, 58, 3})
	text := "restart the supervisor please"
	checkGolden(t, Composer{Text: text, Cursor: true, CursorColumn: len(text), InputRows: 3}.View(58), gfSheet, rect{19, 62, 58, 3})
}

// The frame of GfComponents rows 24 and 32 to 34 with a card: the title and
// the note on the top border, the ├─┤ rule, and the input line, which keeps
// its text with no cursor while the card has the focus (rules 14 and 17).
// The card rows are the TUI's; the Chat boards check them.
func TestGfComposerCardFrameGolden(t *testing.T) {
	card := &ComposerCard{
		Title: "? manager-hive asks · Model",
		Note:  titleStyle.Render(" open · 1 of 1 "),
		Lines: make([]string, 7),
		Rows:  7,
	}
	rows := strings.Split(Composer{Text: "typed text stays while the card has the focus", InputRows: 3, Card: card}.View(120), "\n")
	if len(rows) != 11 {
		t.Fatalf("%d rows, want 11", len(rows))
	}
	checkGolden(t, rows[0], gfSheet, rect{24, 0, 120, 1})
	checkGolden(t, strings.Join(rows[8:], "\n"), gfSheet, rect{32, 0, 120, 3})
}

// Rule 11: beyond InputRows, the text scrolls inside the frame, the cursor
// row stays visible, and the ┃ thumb is on the right border.
func TestComposerScrollsToTheCursor(t *testing.T) {
	text := strings.TrimSpace(strings.Repeat("word ", 40))
	for _, tc := range []struct {
		name   string
		column int
		want   string
	}{
		{"cursor at the start", 0, "> █ord word"},
		{"cursor at the end", len(text), "█"},
	} {
		rows := strings.Split(Composer{Text: text, Cursor: true, CursorColumn: tc.column, InputRows: 2}.View(30), "\n")
		if len(rows) != 4 {
			t.Fatalf("%s: %d rows, want 4", tc.name, len(rows))
		}
		inner := ansi.Strip(rows[1]) + ansi.Strip(rows[2])
		if !strings.Contains(inner, tc.want) {
			t.Errorf("%s: rows %q do not show %q", tc.name, rows[1:3], tc.want)
		}
		thumbs := strings.Count(ansi.Strip(rows[1])[len(ansi.Strip(rows[1]))-len("┃"):], "┃") +
			strings.Count(ansi.Strip(rows[2])[len(ansi.Strip(rows[2]))-len("┃"):], "┃")
		if thumbs == 0 {
			t.Errorf("%s: no ┃ thumb on the right border: %q", tc.name, rows[1:3])
		}
	}
	rows := strings.Split(Composer{Text: "short", Cursor: true, InputRows: 2}.View(30), "\n")
	if len(rows) != 3 || strings.Contains(ansi.Strip(rows[1]), "┃") {
		t.Errorf("short text: rows %q, want 3 rows and no thumb", rows)
	}
}

// InputHeight grows with the wrapped text up to the cap (rule 11).
func TestInputHeight(t *testing.T) {
	text := strings.TrimSpace(strings.Repeat("word ", 40))
	for _, tc := range []struct {
		text  string
		rows  int
		width int
		want  int
	}{
		{"", 5, 30, 1},
		{"short", 5, 30, 1},
		{text, 3, 30, 3},
		{text, 20, 30, 8}, // 25 cells of room hold 5 words a row
	} {
		if got := InputHeight(tc.text, tc.width, tc.rows); got != tc.want {
			t.Errorf("InputHeight(%d words, %d, %d) = %d, want %d", len(strings.Fields(tc.text)), tc.width, tc.rows, got, tc.want)
		}
	}
}
