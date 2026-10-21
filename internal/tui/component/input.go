package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var userStyle = lipgloss.NewStyle().Foreground(User)

// Input is the fixed chat input below the feed (decision 15). The text uses
// the token user. With the focus, the border is accent and the cursor shows
// after the text.
type Input struct {
	Text    string
	Focused bool
}

// View renders the input at exactly width×3 cells. It shows the last line of
// the text. A line that does not fit is cut with … at the start while the
// input has the focus, so the cursor stays visible.
func (in Input) View(width int) string {
	width = max(width, PanelMinWidth)
	line := in.Text[strings.LastIndex(in.Text, "\n")+1:]
	text := userStyle.Render("> " + line)
	if in.Focused {
		text += cursor
		if over := lipgloss.Width(text) - (width - 4); over > 0 {
			text = userStyle.Render("> ") + ansi.TruncateLeft(userStyle.Render(line)+cursor, over+1, "…")
		}
	}
	return Panel{Title: "message", Focused: in.Focused, Lines: []string{text}}.View(width, 3)
}
