package component

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestDetailsGolden renders the Details section, Components.dc.html lines
// 60-64.
func TestDetailsGolden(t *testing.T) {
	done := lipgloss.NewStyle().Foreground(Done)
	d := Details{Title: "3-tui-restart", Lines: []string{
		DetailField("status", done.Render("landed")),
		DetailField("goal", "g1-tui-restart · wave B"),
		DetailSection("ACCEPTANCE"),
		done.Render("✓") + " AC1 r restarts, confirms only when live",
	}}
	checkGolden(t, d.View(56, 6), sheet, rect{33, 64, 56, 6})
}

// TestDetailsScrolls checks that the details body scrolls from Offset and
// draws the scroll mark when the lines do not fit.
func TestDetailsScrolls(t *testing.T) {
	d := Details{Title: "x"}
	for i := range 10 {
		d.Lines = append(d.Lines, fmt.Sprintf("l%d", i))
	}
	d.Offset = 8
	want := "╭─ x ──╮\n│ l8   │\n│ l9   ┃\n╰──────╯"
	if got := ansi.Strip(d.View(8, 4)); got != want {
		t.Errorf("View:\n%s\nwant:\n%s", got, want)
	}
}

// TestDetailsShowsMoreLinesOnOverflow checks the user decision 2 of
// 2026-10-21: the details show each line that fits, and the last body row
// shows "↓ N more lines" only when lines stay below the view. The scroll mark
// is the thumb of the lines that show.
func TestDetailsShowsMoreLinesOnOverflow(t *testing.T) {
	d := Details{Title: "x"}
	for i := range 10 {
		d.Lines = append(d.Lines, fmt.Sprintf("l%d", i))
	}
	want := "╭─ x ──────────────╮\n│ l0               ┃\n│ l1               │\n│ ↓ 8 more lines   │\n╰──────────────────╯"
	if got := ansi.Strip(d.View(20, 5)); got != want {
		t.Errorf("top:\n%s\nwant:\n%s", got, want)
	}
	d.Offset = 3
	want = "╭─ x ──────────────╮\n│ l3               ┃\n│ l4               │\n│ ↓ 5 more lines   │\n╰──────────────────╯"
	if got := ansi.Strip(d.View(20, 5)); got != want {
		t.Errorf("offset 3:\n%s\nwant:\n%s", got, want)
	}
	d.Lines = d.Lines[:4]
	d.Offset = 0
	want = "╭─ x ──────────────╮\n│ l0               ┃\n│ l1               │\n│ ↓ 2 more lines   │\n╰──────────────────╯"
	if got := ansi.Strip(d.View(20, 5)); got != want {
		t.Errorf("one line over (Thumb(3, 4, 2, 0) is one cell):\n%s\nwant:\n%s", got, want)
	}
	d.Lines = d.Lines[:3]
	want = "╭─ x ──────────────╮\n│ l0               │\n│ l1               │\n│ l2               │\n╰──────────────────╯"
	if got := ansi.Strip(d.View(20, 5)); got != want {
		t.Errorf("lines that fit:\n%s\nwant:\n%s", got, want)
	}
}
