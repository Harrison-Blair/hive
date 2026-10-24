package component

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// WaveColumnWidth is the width of a wave column (work.md layout).
const WaveColumnWidth = 28

// cardStride is the rows of one card and the blank row below it.
const cardStride = CardHeight + 1

// WaveColumn is one wave of the Work board (work.md rules 3, 5, 7, and 12).
// The title sits in the top border. Cards are the cards on the screen from
// the top. The column of the selected card has the heavy accent border. The
// foot shows Foot in dim, and ✓ done in done for a Done wave. Hint shows on
// the row above the foot in dim. Thumb is the ┃ thumb on the right border.
// RailsOut and RailsIn are the rows where a rail leaves the right border and
// enters the left border.
type WaveColumn struct {
	Title    string
	Selected bool
	Cards    []TaskCard
	Foot     string
	Done     bool
	Hint     string
	Height   int
	Thumb    ScrollMark
	RailsOut []int
	RailsIn  []int
}

// WaveColumnSlots gives the number of cards that fit in a column of height
// rows: from row 2, at a stride of 5 rows, above the row before the foot,
// as the board script of GfWork places them.
func WaveColumnSlots(height int) int {
	return max((height-3)/cardStride, 0)
}

// View renders the column at exactly WaveColumnWidth×Height cells.
func (c WaveColumn) View() string {
	var footer strings.Builder
	footer.WriteString(dimStyle.Render(c.Foot))
	if c.Done {
		footer.WriteString(" " + lipgloss.NewStyle().Foreground(Done).Render("✓ done"))
	}
	inner := c.Height - 2
	lines := make([]string, inner)
	slots := WaveColumnSlots(c.Height)
	if c.Hint != "" {
		slots = WaveColumnSlots(c.Height - 1)
		lines[inner-2] = dimStyle.Render(c.Hint)
	}
	for index, card := range c.Cards[:min(len(c.Cards), slots)] {
		for row, text := range strings.Split(card.View(), "\n") {
			lines[1+index*cardStride+row] = text
		}
	}
	lines[inner-1] = footer.String()
	panel := Panel{Title: c.Title, Focused: c.Selected, Heavy: c.Selected, Lines: lines, VMark: c.Thumb}
	rows := strings.Split(panel.View(WaveColumnWidth, c.Height), "\n")
	edge, out, in := borderStyle, "┤", "├"
	if c.Selected {
		edge, out, in = accentStyle, "┨", "┠"
	}
	for index := 1; index < c.Height-1; index++ {
		if slices.Contains(c.RailsIn, index) {
			rows[index] = edge.Render(in) + borderStyle.Render("▶") + ansi.Cut(rows[index], 2, WaveColumnWidth)
		}
		if slices.Contains(c.RailsOut, index) {
			rows[index] = ansi.Cut(rows[index], 0, WaveColumnWidth-1) + edge.Render(out)
		}
	}
	return strings.Join(rows, "\n")
}
