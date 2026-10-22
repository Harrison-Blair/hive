package component

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestPanelGolden renders the two panels of the Panel section,
// Components.dc.html lines 46-47.
func TestPanelGolden(t *testing.T) {
	dim := lipgloss.NewStyle().Foreground(Dim)
	cases := []struct {
		name  string
		panel Panel
		r     rect
	}{
		{"unfocused", Panel{Title: "unfocused", Lines: []string{"", dim.Render("border #3d3a33"), ""}}, rect{17, 0, 40, 5}},
		{"focused", Panel{Title: "focused", Focused: true, Lines: []string{"", dim.Render("border #f5b942 · takes the keys"), ""}}, rect{17, 44, 40, 5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkGolden(t, tc.panel.View(tc.r.w, tc.r.h), sheet, tc.r)
		})
	}
}

// TestPanelBleedFillsThePadding: with Bleed, a line also fills the padding
// cell on each side, so a clipped graph meets the border.
func TestPanelBleedFillsThePadding(t *testing.T) {
	got := ansi.Strip(Panel{Bleed: true, Lines: []string{"──────", "ab"}}.View(8, 4))
	want := "╭──────╮\n│──────│\n│ab    │\n╰──────╯"
	if got != want {
		t.Fatalf("view:\n%s\nwant:\n%s", got, want)
	}
}
