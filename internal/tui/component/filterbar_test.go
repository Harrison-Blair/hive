package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestFilterBarGolden renders the FilterBar section, Components.dc.html
// lines 66-67. The board draws the cursor one cell after the text, so the
// sample text ends with a space.
func TestFilterBarGolden(t *testing.T) {
	f := FilterBar{Text: "level:warn agent:manager-hive ", Editing: true}
	checkGolden(t, f.View(120), sheet, rect{41, 0, 120, 3})
}

// TestFilterBarFields checks each typed field of decision 25: the field name
// is dim and its value and free text are fg. The applied filter has no
// cursor.
func TestFilterBarFields(t *testing.T) {
	f := FilterBar{Text: "level:info comp:tui proj:hive agent:a task:t1 req:9 free text"}
	got := f.View(80)
	row := strings.Split(got, "\n")[1]
	for _, field := range []string{"level:", "comp:", "proj:", "agent:", "task:", "req:"} {
		if !strings.Contains(row, dimStyle.Render(field)) {
			t.Errorf("field %q is not dim in %q", field, row)
		}
	}
	if strings.Contains(row, dimStyle.Render("free")) || strings.Contains(row, dimStyle.Render("text")) {
		t.Errorf("free text is dim in %q", row)
	}
	want := "│ / level:info comp:tui proj:hive agent:a task:t1 req:9 free text" + strings.Repeat(" ", 14) + "│"
	if plain := ansi.Strip(row); plain != want {
		t.Errorf("row %q, want %q", plain, want)
	}
}

// TestFilterBarLongTextKeepsCursor checks that a long typed filter shows
// its end and the cursor.
func TestFilterBarLongTextKeepsCursor(t *testing.T) {
	f := FilterBar{Text: strings.Repeat("x", 50) + "END", Editing: true}
	row := ansi.Strip(strings.Split(f.View(30), "\n")[1])
	if !strings.HasSuffix(row, "END█ │") || !strings.HasPrefix(row, "│ / …") {
		t.Errorf("row %q, want the end of the text and the cursor", row)
	}
}

// TestFilterBarStatus checks the status text at the end of the bar: the
// filter text keeps the start of the row, the status ends at the padding,
// and a long applied filter is cut with … before the status.
func TestFilterBarStatus(t *testing.T) {
	f := FilterBar{Text: "level:debug proj:hive", Status: "< 2026-10-21 >  "}
	row := ansi.Strip(strings.Split(f.View(44), "\n")[1])
	if want := "│ / level:debug proj:hive < 2026-10-21 >   │"; row != want {
		t.Errorf("row %q, want %q", row, want)
	}
	f.Text = strings.Repeat("word ", 10)
	row = ansi.Strip(strings.Split(f.View(40), "\n")[1])
	if want := "│ / word word word w… < 2026-10-21 >   │"; row != want {
		t.Errorf("row %q, want %q", row, want)
	}
}

// TestFilterBarFocused checks that the focused bar has the accent border
// (decision 6).
func TestFilterBarFocused(t *testing.T) {
	focused := strings.Split(FilterBar{Text: "x", Editing: true, Focused: true}.View(30), "\n")[0]
	applied := strings.Split(FilterBar{Text: "x"}.View(30), "\n")[0]
	if !strings.HasPrefix(focused, accentStyle.Render("╭")) {
		t.Errorf("focused border %q, want accent", focused)
	}
	if !strings.HasPrefix(applied, borderStyle.Render("╭")) {
		t.Errorf("applied border %q, want border", applied)
	}
}
