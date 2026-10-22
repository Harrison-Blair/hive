package component

import (
	"fmt"
	"slices"
)

// detailLabelWidth is the cells of a details label with its gap.
const detailLabelWidth = 10

// Details is the details panel of a list (decisions 8, 9): fields, section
// headings, and other lines that the caller styles. The body scrolls from
// Offset and draws the scroll mark of decision 5 when the lines do not fit.
// When lines stay below the view, the last body row shows "↓ N more lines"
// (user decision 2026-10-21).
type Details struct {
	Title   string
	Focused bool
	Lines   []string
	Offset  int
}

// View renders the details at exactly width×height cells.
func (d Details) View(width, height int) string {
	body := max(height, PanelMinHeight) - 2
	off := min(max(d.Offset, 0), max(len(d.Lines)-body, 0))
	lines := d.Lines[off:]
	shown := body
	if body > 0 && len(lines) > body {
		shown = body - 1
		more := fmt.Sprintf("↓ %d more lines", len(lines)-shown)
		lines = append(slices.Clone(lines[:shown]), dimStyle.Render(more))
	}
	p := Panel{Title: d.Title, Focused: d.Focused, Lines: lines, VMark: Thumb(body, len(d.Lines), shown, off)}
	return p.View(width, height)
}

// DetailField gives a field line: the label in dim in 10 cells, then the
// value, which the caller styles.
func DetailField(label, value string) string {
	return dimStyle.Render(fitCell(label, detailLabelWidth-1)+" ") + value
}

// DetailSection gives a section heading in accent and bold.
func DetailSection(name string) string { return titleStyle.Render(name) }
