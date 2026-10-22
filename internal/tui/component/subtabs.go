package component

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// subTabsHint is the hint at the right of the views, with its right margin.
const subTabsHint = "[ ] switch view"

// SubTabs is the row of the views of a screen (decisions 16 to 18). Active
// is the zero-based index of the active view. Hint, when set, takes the place
// of the view hint and ends 2 cells before the right edge, as the Compact
// board shows "list › details" while the details overlay is open.
type SubTabs struct {
	Views  []string
	Active int
	Hint   string
}

// View renders the views at exactly width cells. The active view is in
// brackets, accent and bold. Each other view is dim with one space on each
// side. Two spaces separate the views. The hint ends 3 cells before the
// right edge.
func (s SubTabs) View(width int) string {
	var b strings.Builder
	b.WriteString(" ")
	for i, v := range s.Views {
		if i > 0 {
			b.WriteString("  ")
		}
		if i == s.Active {
			b.WriteString(titleStyle.Render("[" + v + "]"))
		} else {
			b.WriteString(dimStyle.Render(" " + v + " "))
		}
	}
	hint := dimStyle.Render(subTabsHint) + "   "
	if s.Hint != "" {
		hint = dimStyle.Render(s.Hint) + "  "
	}
	gap := width - lipgloss.Width(b.String()) - lipgloss.Width(hint)
	return fitRow(b.String()+strings.Repeat(" ", max(gap, 1))+hint, width)
}
