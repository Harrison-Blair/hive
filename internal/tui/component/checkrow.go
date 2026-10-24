package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// checkAreaWidth is the width of the area name of a check row (health.md
// rule 1).
const checkAreaWidth = 11

// CheckTile is one tile of the Health check grid: a mark and a text.
// Passed gives ✓ in done, else ! in accent. Selected gives ▶ before the mark
// and the text in accent bold. Dim draws the mark and the text in dim, while
// hive doctor runs (rule 7). The ▶ and the text of a selected tile stay in
// accent.
type CheckTile struct {
	Text     string
	Passed   bool
	Selected bool
	Dim      bool
}

// CheckRow is one area row of the check grid: the area name in dim bold in
// 11 cells, then the tiles with 4 cells between tiles. The ▶ of a selected
// tile takes the last 2 of those cells.
type CheckRow struct {
	Area  string
	Tiles []CheckTile
}

// View renders the row at exactly width cells from column 1. A row wider
// than width is cut.
func (r CheckRow) View(width int) string {
	area := ansi.Truncate(r.Area, checkAreaWidth, "…")
	row := " " + dimStyle.Bold(true).Render(area+strings.Repeat(" ", checkAreaWidth-lipgloss.Width(area)))
	for index, tile := range r.Tiles {
		if index > 0 {
			row += "  "
		}
		row += tile.view()
	}
	return fitRow(row, width)
}

// view gives the marker cells, the mark, and the text of the tile.
func (t CheckTile) view() string {
	mark, markStyle := "!", titleStyle
	if t.Passed {
		mark, markStyle = "✓", lipgloss.NewStyle().Foreground(Done).Bold(true)
	}
	text := lipgloss.NewStyle()
	if t.Dim {
		markStyle, text = dimStyle.Bold(true), dimStyle
	}
	marker := "  "
	if t.Selected {
		marker, text = titleStyle.Render("▶")+" ", titleStyle
	}
	return marker + markStyle.Render(mark) + " " + text.Render(t.Text)
}
