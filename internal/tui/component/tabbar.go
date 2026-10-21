package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// tabNames are the screens of the tab bar, in key order (decision 3).
var tabNames = []string{"Agents", "Chat", "Work", "Friction", "Health", "Logs"}

// TabBar is the row of screen labels and the rule below it (decision 3).
// Active is the zero-based index of the active screen.
type TabBar struct {
	Active int
}

// View renders the two rows of the tab bar at exactly width cells. The
// labels start after one space. From LongFormWidth one more space separates
// the labels; below it the labels touch. The active label is accent and bold
// with a ━ underline. The rule is border.
func (b TabBar) View(width int) string {
	gap := ""
	if width >= LongFormWidth {
		gap = " "
	}
	var labels, rule strings.Builder
	labels.WriteString(" ")
	rule.WriteString(borderStyle.Render("─"))
	for i, name := range tabNames {
		lab := fmt.Sprintf(" %d %s ", i+1, name)
		n := lipgloss.Width(lab)
		if i == b.Active {
			labels.WriteString(titleStyle.Render(lab))
			rule.WriteString(accentStyle.Render(strings.Repeat("━", n)))
		} else {
			labels.WriteString(dimStyle.Render(lab))
			rule.WriteString(borderStyle.Render(strings.Repeat("─", n)))
		}
		labels.WriteString(gap)
		rule.WriteString(borderStyle.Render(strings.Repeat("─", len(gap))))
	}
	return fitRow(labels.String(), width) + "\n" + fitRow(rule.String()+borderStyle.Render(strings.Repeat("─", width)), width)
}

// fitRow cuts s to width cells, or pads it with plain spaces to width cells.
func fitRow(s string, width int) string {
	s = ansi.Truncate(s, width, "")
	return s + strings.Repeat(" ", max(width-lipgloss.Width(s), 0))
}
