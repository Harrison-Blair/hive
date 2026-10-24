package component

// Fold is a fold line (friction rule 3): ▸ when folded or ▾ when open in
// dim, the text in dim bold, then the hint in dim. A selected line shows ▶
// in accent bold in place of its glyph, and its text in accent bold.
type Fold struct {
	Open     bool
	Selected bool
	Text     string
	Hint     string
}

// View renders the line at exactly width cells. The line starts at column 1.
func (f Fold) View(width int) string {
	glyph, text := dimStyle.Render("▸"), dimStyle.Bold(true).Render(f.Text)
	if f.Open {
		glyph = dimStyle.Render("▾")
	}
	if f.Selected {
		glyph, text = titleStyle.Render("▶"), titleStyle.Render(f.Text)
	}
	row := " " + glyph + " " + text
	if f.Hint != "" {
		row += "  " + dimStyle.Render(f.Hint)
	}
	return fitRow(row, width)
}
