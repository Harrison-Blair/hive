package component

import "strings"

// RailLink is one rail of work.md rule 5, from the row of the name line of a
// card in one column to the row of the name line of a card in the next
// column.
type RailLink struct {
	From, To int
}

// Rail is the gap of two cells between two wave columns. Each link draws ─
// in the first cell and its turn in the second cell. Links that share a cell
// join into one box glyph. The rail is in border.
type Rail struct {
	Height int
	Links  []RailLink
}

// View renders the gap at exactly 2×Height cells.
func (r Rail) View() string {
	first := make([]bool, r.Height)
	second := make([]Arms, r.Height)
	add := func(row int, arms Arms) {
		if row < 0 || row >= r.Height {
			return
		}
		a := &second[row]
		a.Up, a.Down = max(a.Up, arms.Up), max(a.Down, arms.Down)
		a.Left, a.Right = max(a.Left, arms.Left), max(a.Right, arms.Right)
	}
	for _, link := range r.Links {
		if link.From >= 0 && link.From < r.Height {
			first[link.From] = true
		}
		switch {
		case link.To == link.From:
			add(link.From, Arms{Left: Light, Right: Light})
		case link.To < link.From:
			add(link.From, Arms{Left: Light, Up: Light})
			add(link.To, Arms{Down: Light, Right: Light})
		default:
			add(link.From, Arms{Left: Light, Down: Light})
			add(link.To, Arms{Up: Light, Right: Light})
		}
		for row := min(link.From, link.To) + 1; row < max(link.From, link.To); row++ {
			add(row, Arms{Up: Light, Down: Light})
		}
	}
	rows := make([]string, r.Height)
	for row := range rows {
		left, right := " ", " "
		if first[row] {
			left = borderStyle.Render("─")
		}
		if glyph, ok := BoxRune(second[row]); ok {
			right = borderStyle.Render(string(glyph))
		}
		rows[row] = left + right
	}
	return strings.Join(rows, "\n")
}
