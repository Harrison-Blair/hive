package component

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Popup is a Details box centered over a dimmed screen (decision 5). The box
// has the focus and is Width×Height cells. Each line of the details has one
// more space at its start.
type Popup struct {
	Details
	Width  int
	Height int
}

// View dims screen, a text of width×height cells, and draws the box centered
// on it. The box is cut to the screen.
func (p Popup) View(screen string, width, height int) string {
	w, h := min(p.Width, width), min(p.Height, height)
	d := p.Details
	d.Focused = true
	d.Lines = make([]string, len(p.Lines))
	for i, l := range p.Lines {
		d.Lines[i] = " " + l
	}
	return Overlay(screen, d.View(w, h), width, height)
}

// Overlay dims screen, a text of width×height cells, and draws box, a text
// of at most width×height cells, centered on it.
func Overlay(screen, box string, width, height int) string {
	lines := strings.Split(box, "\n")
	w, h := ansi.StringWidth(lines[0]), len(lines)
	rows := plainRows(screen, width, height)
	x, y := (width-w)/2, (height-h)/2
	for i, row := range rows {
		if i < y || i >= y+h {
			rows[i] = borderStyle.Render(row)
			continue
		}
		rows[i] = borderStyle.Render(ansi.Truncate(row, x, "")) + lines[i-y] + borderStyle.Render(ansi.TruncateLeft(row, x+w, ""))
	}
	return strings.Join(rows, "\n")
}

// DimScreen gives screen, a text of width×height cells, with each cell in border
// and not bold (decision 5).
func DimScreen(screen string, width, height int) string {
	rows := plainRows(screen, width, height)
	for i, row := range rows {
		rows[i] = borderStyle.Render(row)
	}
	return strings.Join(rows, "\n")
}

// plainRows gives the rows of screen with no style, each cut or padded to
// width cells, and cut or padded to height rows.
func plainRows(screen string, width, height int) []string {
	lines := strings.Split(ansi.Strip(screen), "\n")
	rows := make([]string, height)
	for i := range rows {
		if i < len(lines) {
			rows[i] = fitRow(lines[i], width)
		} else {
			rows[i] = fitRow("", width)
		}
	}
	return rows
}
