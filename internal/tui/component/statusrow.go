package component

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// StatusRow is one status row of golden-frames decisions 8 to 11: segments
// joined by " · ", and the clock at the right edge. Each segment is text that
// the caller renders with GlyphText, DimText, StrongText, or AccentText.
type StatusRow struct {
	Segments []string
	Clock    string
}

// View renders the row at exactly width cells. The row starts at column 1.
// When the segments do not fit, the row removes segments from the right
// (decision 10). It never removes the first segment, and cuts it when it is
// wider than the row. The clock is in dim with one cell after it.
func (s StatusRow) View(width int) string {
	room := width - 2
	if s.Clock != "" {
		room -= lipgloss.Width(s.Clock) + 3
	}
	segs := s.Segments
	for len(segs) > 1 && segmentsWidth(segs) > room {
		segs = segs[:len(segs)-1]
	}
	row := " " + strings.Join(segs, dimStyle.Render(" · "))
	if s.Clock == "" {
		return fitRow(row, width)
	}
	clock := dimStyle.Render(s.Clock) + " "
	left := width - lipgloss.Width(clock)
	return fitRow(ansi.Truncate(row, left, ""), left) + clock
}

func segmentsWidth(segs []string) int {
	w := 3 * (len(segs) - 1)
	for _, s := range segs {
		w += lipgloss.Width(s)
	}
	return w
}

// GlyphText renders a state segment of decision 11: the glyph in bold in its
// colour, then the rest of the text in dim.
func GlyphText(text string, glyph color.Color) string {
	r := []rune(text)
	if len(r) == 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(glyph).Bold(true).Render(string(r[0])) + dimStyle.Render(string(r[1:]))
}

// DimText renders text in dim.
func DimText(text string) string { return dimStyle.Render(text) }

// StrongText renders text in accent bold, for a count more than 0 or a
// screen name.
func StrongText(text string) string { return titleStyle.Render(text) }

// AccentText renders text in accent.
func AccentText(text string) string { return accentStyle.Render(text) }
