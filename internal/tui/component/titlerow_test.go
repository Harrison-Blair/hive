package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Golden-frames decision 2: GfComponents rows 15 and 16, the title row of a
// screen and the rule below it.
func TestTitleRowGolden(t *testing.T) {
	r := TitleRow{Name: "work · g2-tui-v2", Facts: "· TUI v2: boards, components, and screens", Right: "[ ] goal 2 of 2"}
	checkGolden(t, r.View(120), gfSheet, rect{15, 0, 120, 2})
}

// The facts are cut with … before the right text. A title row with no facts
// and no right text shows only the name.
func TestTitleRowFit(t *testing.T) {
	cases := []struct {
		row   TitleRow
		width int
		want  string
	}{
		{TitleRow{Name: "logs"}, 12, " logs       \n ────────── "},
		{TitleRow{Name: "logs", Facts: "312 events", Right: "← day →"}, 30, " logs 312 events      ← day → \n ──────────────────────────── "},
		{TitleRow{Name: "logs", Facts: "312 events", Right: "← day →"}, 20, " logs 312 … ← day → \n ────────────────── "},
	}
	for _, tc := range cases {
		if got := ansi.Strip(tc.row.View(tc.width)); got != tc.want {
			t.Errorf("View(%d) =\n%q\nwant\n%q", tc.width, got, tc.want)
		}
	}
}
