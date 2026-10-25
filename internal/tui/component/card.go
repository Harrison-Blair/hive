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
// word, as for Header. A selected card has a heavy accent border and its name
// in accent and bold (user decision 2026-10-21, option A). A stopped agent has its name in dim.
// A working card that is not selected has its border in working (decision 60).
type Card struct {
	Name     string
	Sub      string
	State    string
	Selected bool
	// Frame is the working marker frame of the view. Frame 0 keeps the
	// static state glyph of the approved boards.
	Frame int
}

// View renders the card at exactly CardWidth×CardHeight cells: the state dot
// and the name, then the sub line in dim. Each is cut with … to fit.
func (c Card) View() string {
	dot, name := cardDot(c.State, c.Frame), lipgloss.NewStyle()
	switch {
	case c.Selected:
		name = titleStyle
	case dot == cardDot("stopped", c.Frame):
		name = dimStyle
	}
	room := CardWidth - 4
	lines := []string{
		dot + " " + name.Render(ansi.Truncate(c.Name, room-2, "…")),
		dimStyle.Render(ansi.Truncate(c.Sub, room, "…")),
	}
	panel := Panel{Focused: c.Selected, Heavy: c.Selected, Lines: lines}
	if c.State == "working" {
		panel.Edge = Working
	}
	return panel.View(CardWidth, CardHeight)
}

// cardDot gives the dot of an agent state, as stateDot does, but not bold.
// A working dot uses the working marker frame.
func cardDot(state string, frame int) string {
	switch state {
	case "working":
		return lipgloss.NewStyle().Foreground(Working).Render(WorkingGlyph(frame, "●"))
	case "live":
		return lipgloss.NewStyle().Foreground(Working).Render("●")
	case "idle":
		return accentStyle.Render("◐")
	case "blocked":
		return accentStyle.Render("▲")
	default:
		return dimStyle.Render("○")
	}
}
