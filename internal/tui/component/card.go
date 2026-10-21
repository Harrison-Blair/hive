package component

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// CardWidth and CardHeight are the size of an agent card (decision 10).
const (
	CardWidth  = 24
	CardHeight = 4
)

// Card is one agent node of the Agents graph (decision 10). State is a state
// word, as for Header. A selected card has an accent border and its name in
// accent and bold. A stopped agent has its name in dim.
type Card struct {
	Name     string
	Sub      string
	State    string
	Selected bool
}

// View renders the card at exactly CardWidth×CardHeight cells: the state dot
// and the name, then the sub line in dim. Each is cut with … to fit.
func (c Card) View() string {
	dot, name := cardDot(c.State), lipgloss.NewStyle()
	switch {
	case c.Selected:
		name = titleStyle
	case dot == cardDot("stopped"):
		name = dimStyle
	}
	room := CardWidth - 4
	lines := []string{
		dot + " " + name.Render(ansi.Truncate(c.Name, room-2, "…")),
		dimStyle.Render(ansi.Truncate(c.Sub, room, "…")),
	}
	return Panel{Focused: c.Selected, Lines: lines}.View(CardWidth, CardHeight)
}

// cardDot gives the dot of an agent state, as stateDot does, but not bold.
func cardDot(state string) string {
	switch state {
	case "working", "live":
		return lipgloss.NewStyle().Foreground(Working).Render("●")
	case "idle":
		return accentStyle.Render("◐")
	default:
		return dimStyle.Render("○")
	}
}
