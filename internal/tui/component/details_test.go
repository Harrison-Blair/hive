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
