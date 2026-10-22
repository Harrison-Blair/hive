package component

import "github.com/charmbracelet/x/ansi"

// Viewport is a Panel whose body scrolls (decision 5). It shows the lines of
// the panel from line YOffset and from cell XOffset of each line. The caller
// sets the scroll marks of the panel, for example with Thumb.
type Viewport struct {
	Panel
	YOffset int
	XOffset int
}

// View renders the viewport at exactly width×height cells.
func (v Viewport) View(width, height int) string {
	p := v.Panel
	lines := p.Lines[min(max(v.YOffset, 0), len(p.Lines)):]
	if v.XOffset > 0 {
		cut := make([]string, len(lines))
		for i, l := range lines {
			cut[i] = ansi.TruncateLeft(l, v.XOffset, "")
		}
		lines = cut
	}
	p.Lines = lines
	return p.View(width, height)
}
