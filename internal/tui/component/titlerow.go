package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TitleRow is the first two rows of each screen other than Chat
// (golden-frames decision 2): the screen name in accent bold, the facts in
// dim, the right text in dim, and a rule below. Note shows in accent two
// cells after the facts. Line, a styled row, takes the place of the rule.
type TitleRow struct {
	Name, Facts, Right string
	Note, Line         string
}

// View renders the two rows at exactly width cells. The row starts at column
// 1 and the right text ends one cell before the edge. The facts are cut with
// … one cell before the right text. The rule is in border from column 1 to
// the cell before the last.
func (r TitleRow) View(width int) string {
	right := ""
	if r.Right != "" {
		right = dimStyle.Render(r.Right) + " "
	}
	left := width - lipgloss.Width(right)
	row := " " + titleStyle.Render(r.Name)
	if r.Facts != "" {
		room := left - lipgloss.Width(row) - 1
		if right != "" {
			room--
		}
		if room > 0 {
			row += " " + dimStyle.Render(ansi.Truncate(r.Facts, room, "…"))
		}
	}
	if r.Note != "" {
		row += "  " + accentStyle.Render(r.Note)
	}
	rule := " " + borderStyle.Render(strings.Repeat("─", max(width-2, 0))) + " "
	if r.Line != "" {
		rule = " " + r.Line
	}
	return fitRow(row, left) + right + "\n" + fitRow(rule, width)
}
