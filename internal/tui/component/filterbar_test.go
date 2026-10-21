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
