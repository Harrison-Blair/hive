package component

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// TaskCard is one task of the Work board (work.md rule 4): line 1 is the dot
// and the task ID, line 2 gives the state. The selected card has the heavy
// accent border and its ID in accent bold. A card of a working worker has
// the border in working (tui-v2 decision 60). LineColor is dim when it is
// not set.
type TaskCard struct {
	ID        string
	Line      string
	Dot       string
	DotColor  color.Color
	LineColor color.Color
	Selected  bool
	Working   bool
}

// View renders the card at exactly CardWidth×CardHeight cells. The ID is cut
// with … at 18 cells and line 2 at 20 cells. Each is padded in its own
// style.
func (c TaskCard) View() string {
	id := lipgloss.NewStyle()
	if c.Selected {
		id = titleStyle
	}
	line := dimStyle
	if c.LineColor != nil {
		line = lipgloss.NewStyle().Foreground(c.LineColor)
	}
	room := CardWidth - 4
	lines := []string{
		lipgloss.NewStyle().Foreground(c.DotColor).Render(c.Dot) + " " + id.Render(fitCell(c.ID, room-2)),
		line.Render(fitCell(c.Line, room)),
	}
	panel := Panel{Focused: c.Selected, Heavy: c.Selected, Lines: lines}
	if c.Working {
		panel.Edge = Working
	}
	return panel.View(CardWidth, CardHeight)
}
