package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// LongFormWidth is the width from which the header shows the long form and
// the tab bar spaces its labels (decisions 2, 3).
const LongFormWidth = 120

// Header is the first row of each screen (decision 2). Coordinator and
// Supervisor are state words, for example "working", "live", "idle", or
// "stopped".
type Header struct {
	Home        string
	Coordinator string
	Supervisor  string
	NeedsYou    int
	Clock       string
}

var headerStyle = lipgloss.NewStyle().Bold(true)

// View renders the header at exactly width cells. From LongFormWidth it
// shows the home on the left and the state words on the right. Below it, it
// shows only the dots of the states. A needs-you count above 0 is accent.
// The header cuts a long home with ….
func (h Header) View(width int) string {
	coord, sup := stateDot(h.Coordinator), stateDot(h.Supervisor)
	bold := headerStyle.Render
	needs := bold(fmt.Sprint(h.NeedsYou))
	if h.NeedsYou > 0 {
		needs = headerStyle.Foreground(Accent).Render(fmt.Sprint(h.NeedsYou))
	}
	var left, right string
	if width >= LongFormWidth {
		right = bold("coordinator ") + coord + bold(" "+h.Coordinator+"   supervisor ") + sup +
			bold(" "+h.Supervisor+"   needs you ") + needs + bold("   "+h.Clock+" ")
		room := width - lipgloss.Width(right) - lipgloss.Width(" hive  ") - 1
		left = bold(" hive  " + ansi.Truncate(h.Home, max(room, 0), "…"))
	} else {
		right = bold("coord ") + coord + bold(" · sup ") + sup + bold(" · needs you ") + needs + bold(" · "+h.Clock+" ")
		left = bold(" hive")
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	return ansi.Truncate(left+strings.Repeat(" ", max(gap, 0))+right, width, "")
}

// stateDot gives the dot of an agent state: ● working for a working or live
// agent, ◐ accent for an idle agent (decision 10), and ○ dim for each other
// state. The dot is bold, as the header is.
func stateDot(state string) string {
	switch state {
	case "working", "live":
		return headerStyle.Foreground(Working).Render("●")
	case "idle":
		return headerStyle.Foreground(Accent).Render("◐")
	default:
		return headerStyle.Foreground(Dim).Render("○")
	}
}
