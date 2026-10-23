package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const gfChords = "^g agents  ^t work  ^f friction  ^y health  ^l logs"


// Decisions 10 and 11: GfComponents rows 7 to 9. Row 7 has no room for
// "? help", so the row removes it. Row 8 is the 80-column form, which removes
// the whole chords and hints segments.
func TestStatusRowGolden(t *testing.T) {
	t.Run("working", func(t *testing.T) {
		r := StatusRow{Segments: []string{GlyphText("● working 2m", Working), GlyphText("● supervisor", Working), StrongText("? 1 ask"), StrongText("needs you 1"), DimText(gfChords), DimText("? help")}, Clock: "14:22"}
		checkGolden(t, r.View(120), gfSheet, rect{7, 0, 120, 1})
	})
	t.Run("idle at 80 columns", func(t *testing.T) {
		r := StatusRow{Segments: []string{GlyphText("○ idle 5m", Dim), GlyphText("● supervisor", Working), DimText("? 0 asks"), DimText("needs you 0"), DimText(gfChords), DimText("? help")}, Clock: "14:22"}
		checkGolden(t, r.View(80), gfSheet, rect{8, 0, 80, 1})
	})
	t.Run("blocked", func(t *testing.T) {
		r := StatusRow{Segments: []string{GlyphText("▲ blocked 1m", Accent), GlyphText("● supervisor", Working), DimText("? 0 asks"), DimText("needs you 0"), DimText(gfChords), DimText("? help")}, Clock: "14:22"}
		checkGolden(t, r.View(120), gfSheet, rect{9, 0, 120, 1})
	})
}

// Decision 10: the TUI never removes the first segment. A first segment that
// is wider than the row is cut. A row with no clock uses the full width.
func TestStatusRowFit(t *testing.T) {
	segs := []string{DimText("first"), DimText("second"), DimText("third")}
	cases := []struct {
		width int
		clock string
		want  string
	}{
		{30, "", " first · second · third       "},
		{20, "", " first · second     "},
		{10, "", " first    "},
		{4, "", " fir"},
		{30, "9:41", " first · second          9:41 "},
		{20, "9:41", " first         9:41 "},
	}
	for _, tc := range cases {
		got := StatusRow{Segments: segs, Clock: tc.clock}.View(tc.width)
		if s := ansi.Strip(got); s != tc.want {
			t.Errorf("View(%d) clock %q = %q, want %q", tc.width, tc.clock, s, tc.want)
		}
	}
}
