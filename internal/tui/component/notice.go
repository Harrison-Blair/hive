package component

import "github.com/charmbracelet/x/ansi"

// Notice is a one-line notice, for example "no project agents" (decision 10)
// or the restart notice of decision 19. The text is dim.
type Notice struct {
	Text string
}

// View renders the notice at exactly width cells: one space, then the text
// in dim, cut with … when it does not fit.
func (n Notice) View(width int) string {
	return fitRow(" "+dimStyle.Render(ansi.Truncate(n.Text, max(width-1, 0), "…")), width)
}
