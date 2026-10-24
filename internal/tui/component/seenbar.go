package component

import "strings"

// seenBarMax is the most marks of a seen bar (friction rule 1).
const seenBarMax = 10

// SeenBar is the seen bar of a friction entry: one ▮ for each occurrence, up
// to 10. The bar is accent from 3 occurrences, else dim.
type SeenBar struct {
	Count int
}

// View renders the bar. It is as wide as its marks.
func (s SeenBar) View() string {
	bar := strings.Repeat("▮", min(max(s.Count, 0), seenBarMax))
	if s.Count >= 3 {
		return accentStyle.Render(bar)
	}
	return dimStyle.Render(bar)
}
