package component

// detailLabelWidth is the cells of a details label with its gap.
const detailLabelWidth = 10

// Details is the details panel of a list (decisions 8, 9): fields, section
// headings, and other lines that the caller styles. The body scrolls from
// Offset and draws the scroll mark of decision 5 when the lines do not fit.
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
	p := Panel{Title: d.Title, Focused: d.Focused, Lines: d.Lines, VMark: Thumb(body, len(d.Lines), body, off)}
	return Viewport{Panel: p, YOffset: off}.View(width, height)
}

// DetailField gives a field line: the label in dim in 10 cells, then the
// value, which the caller styles.
func DetailField(label, value string) string {
	return dimStyle.Render(fitCell(label, detailLabelWidth-1)+" ") + value
}

// DetailSection gives a section heading in accent and bold.
func DetailSection(name string) string { return titleStyle.Render(name) }
