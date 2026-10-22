package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// listMarker is the selection marker of a List row.
const listMarker = "▶ "

// Column is one fixed column of a List. Width is the cells of the column,
// with the one-cell gap after it. The last column takes the cells that are
// left, and its Width is not used.
type Column struct {
	Title string
	Width int
}

// Cell is the text of one List cell and its style.
type Cell struct {
	Text  string
	Style lipgloss.Style
}

// List is a Panel of rows in fixed columns with a dim header row (decisions
// 8, 16 to 20). The selected row has the marker ▶ and its first cell in
// accent and bold over its full column, and the space after the marker is
// fg (user decisions 2026-10-21). Offset is the first visible row; View
// moves it so that the selected row is visible. When the rows do not fit, the panel draws the
// scroll mark of decision 5. A Hint shows in dim at the right of the last
// body row and takes one row line.
type List struct {
	Title    string
	Focused  bool
	Columns  []Column
	Rows     [][]Cell
	Selected int
	Offset   int
	Hint     string
}

// View renders the list at exactly width×height cells.
func (l List) View(width, height int) string {
	visible := l.visible(height)
	off := l.offset(visible)
	lines := []string{"  " + dimStyle.Render(l.header())}
	for i := off; i < min(off+visible, len(l.Rows)); i++ {
		lines = append(lines, l.row(i))
	}
	if l.Hint != "" {
		body := max(height, PanelMinHeight) - 2
		for len(lines) < body-1 {
			lines = append(lines, "")
		}
		hint := l.Hint + " "
		lines = append(lines, strings.Repeat(" ", max(width-4-lipgloss.Width(hint), 0))+dimStyle.Render(l.Hint)+" ")
	}
	vmark := Thumb(visible, len(l.Rows), visible, off)
	if vmark.Size > 0 {
		vmark.Start++ // the track starts below the header row
	}
	return Panel{Title: l.Title, Focused: l.Focused, Lines: lines, VMark: vmark}.View(width, height)
}

// Move moves the selection by delta rows, within the rows, and scrolls the
// list of the given height so that the selected row stays visible.
func (l *List) Move(delta, height int) {
	l.Selected = min(max(l.Selected+delta, 0), max(len(l.Rows)-1, 0))
	l.Offset = l.offset(l.visible(height))
}

// visible gives the number of row lines in a list of height cells: the body
// rows less the header row and the hint row.
func (l List) visible(height int) int {
	if l.Hint != "" {
		return max(height, PanelMinHeight) - 4
	}
	return max(height, PanelMinHeight) - 3
}

// offset gives the first visible row: Offset, moved the least so that the
// selected row is visible and no row line is empty without need.
func (l List) offset(visible int) int {
	if visible < 1 {
		return min(max(l.Selected, 0), max(len(l.Rows)-1, 0))
	}
	off := min(l.Offset, l.Selected)
	off = max(off, l.Selected-visible+1)
	return min(max(off, 0), max(len(l.Rows)-visible, 0))
}

func (l List) header() string {
	var b strings.Builder
	for i, c := range l.Columns {
		if i == len(l.Columns)-1 {
			b.WriteString(c.Title)
		} else {
			b.WriteString(fitCell(c.Title, c.Width))
		}
	}
	return b.String()
}

// row gives the line of row i: the marker, then each cell cut to its column
// with … and one cell of gap.
func (l List) row(i int) string {
	var b strings.Builder
	if i == l.Selected {
		b.WriteString(titleStyle.Render(strings.TrimSpace(listMarker)) + " ")
	} else {
		b.WriteString("  ")
	}
	for j, c := range l.Rows[i] {
		if j >= len(l.Columns) {
			break
		}
		if j == len(l.Columns)-1 {
			b.WriteString(c.Style.Render(c.Text))
			break
		}
		w := l.Columns[j].Width
		if j == 0 && i == l.Selected {
			b.WriteString(titleStyle.Render(fitCell(c.Text, w-1) + " "))
			continue
		}
		text := ansi.Truncate(c.Text, max(w-1, 0), "…")
		b.WriteString(c.Style.Render(text) + strings.Repeat(" ", max(w-lipgloss.Width(text), 0)))
	}
	return b.String()
}

// fitCell cuts s to width cells with …, or pads it with spaces to width
// cells.
func fitCell(s string, width int) string {
	s = ansi.Truncate(s, max(width, 0), "…")
	return s + strings.Repeat(" ", max(width-lipgloss.Width(s), 0))
}
