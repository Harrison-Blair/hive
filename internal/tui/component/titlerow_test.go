package component

import (
	"strings"
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

// Friction rule 6: while a filter applies, the title row shows the tag
// "· / TEXT" in accent after the facts. The tag stays when the facts are cut.
func TestTitleRowTag(t *testing.T) {
	want := "hive-grid 1 40 1\n" +
		"t| friction · 2 open · / agent   / filter \n" +
		"k|faaaaaaaafddddddddfaaaaaaaaafffddddddddf\n" +
		"b|0111111110000000000000000000000000000000\n"
	got := TitleRow{Name: "friction", Facts: "· 2 open", Tag: "· / agent", Right: "/ filter"}.View(40)
	checkHandGrid(t, strings.Split(got, "\n")[0], want)
	cut := TitleRow{Name: "friction", Facts: "· 15 open · 23 resolved", Tag: "· / agent", Right: "/ filter"}.View(40)
	if row := ansi.Strip(strings.Split(cut, "\n")[0]); row != " friction · 15 open… · / agent / filter " {
		t.Errorf("cut row = %q", row)
	}
}
