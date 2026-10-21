package component

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestViewportGolden renders the two regions of the Viewport + ScrollMark
// section, Components.dc.html lines 49-52. The board draws fixed thumbs, so
// the test passes the thumb cells of the board (vline rows 26-27, 42 cells of
// ━ from col 65).
func TestViewportGolden(t *testing.T) {
	lines := make([]string, 120)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d of 120", i+1)
	}
	dim := lipgloss.NewStyle().Foreground(Dim)
	cases := []struct {
		name string
		vp   Viewport
		r    rect
	}{
		{"vertical", Viewport{Panel: Panel{Title: "viewport", Focused: true, Lines: lines, VMark: ScrollMark{Start: 1, Size: 2}}, YOffset: 40}, rect{24, 0, 60, 7}},
		{"horizontal", Viewport{Panel: Panel{Title: "horizontal", Lines: []string{"", dim.Render("projects 1–4 of 5 · ←→ at the edge")}, HMark: ScrollMark{Start: 0, Size: 42}}}, rect{24, 64, 56, 7}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkGolden(t, tc.vp.View(tc.r.w, tc.r.h), sheet, tc.r)
		})
	}
}

func TestViewportXOffset(t *testing.T) {
	vp := Viewport{Panel: Panel{Lines: []string{"abcdefghij"}}, XOffset: 3}
	got := ansi.Strip(vp.View(9, 3))
	want := "╭───────╮\n│ defg… │\n╰───────╯"
	if got != want {
		t.Errorf("View:\n%s\nwant:\n%s", got, want)
	}
}

func TestThumb(t *testing.T) {
	cases := []struct {
		track, total, visible, offset int
		want                          ScrollMark
	}{
		{5, 5, 5, 0, ScrollMark{}},
		{5, 3, 5, 0, ScrollMark{}},
		{5, 120, 5, 0, ScrollMark{Start: 0, Size: 1}},
		{5, 120, 5, 115, ScrollMark{Start: 4, Size: 1}},
		{5, 120, 5, 500, ScrollMark{Start: 4, Size: 1}},
		{10, 20, 10, 5, ScrollMark{Start: 2, Size: 5}},
		{54, 5, 4, 1, ScrollMark{Start: 11, Size: 43}},
	}
	for _, tc := range cases {
		if got := Thumb(tc.track, tc.total, tc.visible, tc.offset); got != tc.want {
			t.Errorf("Thumb(%d, %d, %d, %d) = %+v, want %+v", tc.track, tc.total, tc.visible, tc.offset, got, tc.want)
		}
	}
}
