package component

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// densityBars are the bar symbols of golden-frames decision 7, low to high.
var densityBars = []rune("▁▂▃▄▅▆▇")

// densityLeft is the cells before the first strip cell: a blank, the start
// time, and a blank.
const densityLeft = 7

// DensityCell is one bucket of a DensityStrip: the event count and the
// highest level in the bucket, WARN or ERROR, or empty.
type DensityCell struct {
	Count int
	Level string
}

// DensityStrip is the density strip of logs.md rules 1 and 2: the strip of
// cells between the start and the end time, the base rule with ┴ under the
// cell Mark, and the marker text under the rule. A Mark below 0 draws no
// mark.
type DensityStrip struct {
	Start, End string
	Cells      []DensityCell
	Mark       int
	MarkText   string
	MarkLevel  string
}

// View renders the three rows at exactly width cells. The strip has one
// symbol for each cell: blank for a count of 0, else ▁ to ▇ by the count
// relative to the largest count. The marker text starts 6 cells before the
// ┴, so the level word sits under it, and stays inside the row.
func (s DensityStrip) View(width int) string {
	most := 0
	for _, cell := range s.Cells {
		most = max(most, cell.Count)
	}
	var strip, rule strings.Builder
	for index, cell := range s.Cells {
		bar := " "
		if cell.Count > 0 {
			bar = string(densityBars[(cell.Count*len(densityBars)+most-1)/most-1])
		}
		strip.WriteString(levelStyle(cell.Level, dimStyle).Render(bar))
		if index == s.Mark {
			rule.WriteString(levelStyle(s.MarkLevel, lipgloss.NewStyle()).Render("┴"))
			continue
		}
		rule.WriteString(borderStyle.Render("─"))
	}
	top := " " + dimStyle.Render(s.Start) + " " + strip.String() + " " + dimStyle.Render(s.End)
	base := strings.Repeat(" ", densityLeft) + rule.String()
	marker := ""
	if s.Mark >= 0 && s.Mark < len(s.Cells) {
		at := min(densityLeft+s.Mark-6, width-lipgloss.Width(s.MarkText))
		at = max(at, densityLeft)
		marker = strings.Repeat(" ", at) + levelStyle(s.MarkLevel, lipgloss.NewStyle()).Render(s.MarkText)
	}
	return fitRow(top, width) + "\n" + fitRow(base, width) + "\n" + fitRow(marker, width)
}

// levelStyle gives the style of a log level: WARN is accent, ERROR is error,
// DEBUG is dim, and each other level is other.
func levelStyle(level string, other lipgloss.Style) lipgloss.Style {
	switch level {
	case "WARN":
		return accentStyle
	case "ERROR":
		return lipgloss.NewStyle().Foreground(Error)
	case "DEBUG":
		return dimStyle
	}
	return other
}
