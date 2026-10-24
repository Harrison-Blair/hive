package component

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// PanelMinWidth and PanelMinHeight are the smallest panel: the corners, one
// padding cell on each side, one content cell, and the two border rows.
const (
	PanelMinWidth  = 5
	PanelMinHeight = 2
)

// Panel is a region with a rounded border, one cell of padding on each side,
// and a title in the top border (decisions 6, 27). Lines are the body rows,
// styled by the caller. Panel cuts a long line with … and pads a short one.
// VMark and HMark draw the scroll marks of decision 5 on the right and the
// bottom border. Heavy draws the border with heavy lines (┏━┓┃┗┛). Edge,
// when set, is the colour of the border of a panel without focus. Bleed
// draws the lines over the padding cells too: each line has width-2 cells.
type Panel struct {
	Title   string
	Focused bool
	Heavy   bool
	Edge    color.Color
	Lines   []string
	VMark   ScrollMark
	HMark   ScrollMark
	Bleed   bool
}

// View renders the panel at exactly width×height cells.
func (p Panel) View(width, height int) string {
	width = max(width, PanelMinWidth)
	height = max(height, PanelMinHeight)
	b := lipgloss.RoundedBorder()
	if p.Heavy {
		b = lipgloss.ThickBorder()
	}
	edge, mark := borderStyle, dimStyle
	if p.Edge != nil {
		edge = lipgloss.NewStyle().Foreground(p.Edge)
	}
	if p.Focused {
		edge, mark = accentStyle, accentStyle
		if p.Heavy {
			// Ask a34: the mark is in fg, so it shows on the heavy border.
			mark = lipgloss.NewStyle()
		}
	}
	inner := width - 2
	body, room := lipgloss.NewStyle().Padding(0, 1).Width(inner), inner-2
	if p.Bleed {
		body, room = lipgloss.NewStyle().Width(inner), inner
	}

	rows := make([]string, 0, height)
	rows = append(rows, edge.Render(b.TopLeft)+p.top(inner, b, edge)+edge.Render(b.TopRight))
	for i := range height - 2 {
		line := ""
		if i < len(p.Lines) {
			line = ansi.Truncate(p.Lines[i], room, "…")
		}
		right := edge.Render(b.Right)
		if p.VMark.covers(i) {
			right = mark.Render(markV)
		}
		rows = append(rows, edge.Render(b.Left)+body.Render(line)+right)
	}
	start := min(max(p.HMark.Start, 0), inner)
	end := min(max(p.HMark.Start+p.HMark.Size, start), inner)
	bottom := edge.Render(b.BottomLeft+strings.Repeat(b.Bottom, start)) +
		mark.Render(strings.Repeat(markH, end-start)) +
		edge.Render(strings.Repeat(b.Bottom, inner-end)+b.BottomRight)
	rows = append(rows, bottom)
	return strings.Join(rows, "\n")
}

// top gives the inner cells of the top border: "─ title ─…". A title that
// does not fit is cut with …, and no title is drawn when one cell of it does
// not fit.
func (p Panel) top(inner int, b lipgloss.Border, edge lipgloss.Style) string {
	room := inner - 4 // "─ " before the title, " ─" after it
	if p.Title == "" || room < 1 {
		return edge.Render(strings.Repeat(b.Top, inner))
	}
	title := titleStyle.Render(" " + ansi.Truncate(p.Title, room, "…") + " ")
	rest := inner - 1 - lipgloss.Width(title)
	return edge.Render(b.Top) + title + edge.Render(strings.Repeat(b.Top, rest))
}
