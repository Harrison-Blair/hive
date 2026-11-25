package graph

import "github.com/Harrison-Blair/hive/internal/lib/style"

// leaves is the number of tracks of the subtrees of nodes side by side: one
// for each card with no children.
func leaves(nodes []*Node) int {
	n := 0
	for _, node := range nodes {
		n += max(1, leaves(node.Children))
	}
	return n
}

// treePodHeight is the height of a pod with one row of cards for each depth
// and one row of joins between two rows of cards.
func treePodHeight(p Pod) int {
	levels := 1 + depth(p.Roots)
	return 2 + levels*cardHeight + levels - 1
}

// forest draws the subtrees of nodes side by side from the card column x,
// each on as many tracks as it has leaves, with their roots in row y. It
// returns the card column of each subtree.
func (d *drawer) forest(nodes []*Node, x, y int) []int {
	xs := make([]int, len(nodes))
	for i, n := range nodes {
		xs[i] = x
		d.subtree(n, x, y)
		x += trackPitch * max(1, leaves(n.Children))
	}
	return xs
}

// subtree draws the subtree of n with the card of n at x, y, on the track
// of its first child.
func (d *drawer) subtree(n *Node, x, y int) {
	d.card(n, x, y, d.cw)
	if len(n.Children) > 0 {
		d.joins(d.forest(n.Children, x, y+cardHeight+1), y)
	}
}

// joins draws the lines in the row below the card row y from the parent,
// over the first child, to the children at the card columns xs.
func (d *drawer) joins(xs []int, y int) {
	half := d.cw / 2
	bar := y + cardHeight
	first, last := xs[0]+half, xs[len(xs)-1]+half
	if len(xs) == 1 {
		d.c.Set(first, bar, '│', style.Border)
		return
	}
	for col := first + 1; col < last; col++ {
		d.c.Set(col, bar, '─', style.Border)
	}
	d.c.Set(first, bar, '├', style.Border)
	for _, cx := range xs[1 : len(xs)-1] {
		d.c.Set(cx+half, bar, '┬', style.Border)
	}
	d.c.Set(last, bar, '╮', style.Border)
}

// vertical moves the selection delta rows of cards down the screen in tree
// mode: to the card of the next row whose center is closest, the left one on
// a tie. List mode and the details step through the pre-order of the cards.
func (m *Model) vertical(delta int) {
	if m.list || m.details {
		m.step(delta)
		return
	}
	cards := m.board().cards
	from, ok := cards[m.selected]
	if !ok {
		return
	}
	row, found := 0, false
	for _, c := range cards {
		if (c.top-from.top)*delta > 0 && (!found || (c.top-row)*delta < 0) {
			row, found = c.top, true
		}
	}
	if !found {
		return
	}
	center := 2*from.left + from.width
	best, bestSpan := "", span{}
	for name, c := range cards {
		if c.top != row {
			continue
		}
		dist, bestDist := abs(2*c.left+c.width-center), abs(2*bestSpan.left+bestSpan.width-center)
		if best == "" || dist < bestDist || dist == bestDist && c.left < bestSpan.left {
			best, bestSpan = name, c
		}
	}
	m.selected = best
}

// horizontal moves the selection to the closest card delta's way across the
// screen in tree mode whose rows overlap the rows of the selected card. List
// mode and the details jump to the next pod.
func (m *Model) horizontal(delta int) {
	if m.list || m.details {
		m.jump(delta)
		return
	}
	cards := m.board().cards
	from, ok := cards[m.selected]
	if !ok {
		return
	}
	best, bestSpan := "", span{}
	for name, c := range cards {
		if (c.left-from.left)*delta <= 0 || c.top >= from.top+cardHeight || from.top >= c.top+cardHeight {
			continue
		}
		dist, bestDist := abs(c.left-from.left), abs(bestSpan.left-from.left)
		if best == "" || dist < bestDist || dist == bestDist && c.top < bestSpan.top {
			best, bestSpan = name, c
		}
	}
	if best != "" {
		m.selected = best
	}
}

func abs(n int) int {
	return max(n, -n)
}
