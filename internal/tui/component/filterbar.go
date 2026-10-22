package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// filterFields are the typed fields of the filter bar (decision 25).
var filterFields = []string{"level:", "comp:", "proj:", "agent:", "task:", "req:"}

// cursor is the text cursor of the filter bar and the chat input.
var cursor = accentStyle.Render("█")

// FilterBar is the fixed filter bar above a list (decisions 20, 25). Text is
// the typed filter. While Editing, the bar shows the cursor after the text;
// otherwise it shows the applied filter. A typed field name is dim and its
// value is fg. Free text is fg. Status is text that the caller styles; it
// ends at the right padding of the bar. A focused bar has the accent border.
type FilterBar struct {
	Text    string
	Editing bool
	Focused bool
	Status  string
}

// View renders the filter bar at exactly width×3 cells. A text that does not
// fit is cut with …: at the start while Editing, so the cursor stays
// visible, and at the end otherwise. The text keeps one cell of gap before
// the status.
func (f FilterBar) View(width int) string {
	width = max(width, PanelMinWidth)
	prompt := titleStyle.Render("/") + " "
	text := filterText(f.Text)
	room := width - 4 - lipgloss.Width(prompt)
	status := ""
	if f.Status != "" && lipgloss.Width(f.Status) < room {
		status = f.Status
		room -= lipgloss.Width(status) + 1
	}
	if f.Editing {
		text += cursor
		if over := lipgloss.Width(text) - room; over > 0 {
			text = ansi.TruncateLeft(text, over+1, "…")
		}
	} else if status != "" {
		text = ansi.Truncate(text, room, "…")
	}
	line := prompt + text
	if status != "" {
		line += strings.Repeat(" ", max(width-4-lipgloss.Width(line)-lipgloss.Width(status), 1)) + status
	}
	return Panel{Title: "filter", Focused: f.Focused, Lines: []string{line}}.View(width, 3)
}

// filterText styles each word of a filter: the name of a typed field in dim.
func filterText(s string) string {
	words := strings.Split(s, " ")
	for i, w := range words {
		for _, f := range filterFields {
			if strings.HasPrefix(w, f) {
				words[i] = dimStyle.Render(f) + w[len(f):]
				break
			}
		}
	}
	return strings.Join(words, " ")
}
