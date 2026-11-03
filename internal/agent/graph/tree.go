package graph

import "github.com/Harrison-Blair/hive/internal/lib/style"

// treeGap is the number of columns between two subtrees side by side.
const treeGap = 2

// treeCardWidth is the largest card width up to cardWidth at which the
// widest pod fits width, and at least minCardWidth.
func treeCardWidth(pods []Pod, width int) int {
	for cw := cardWidth; cw > minCardWidth; cw-- {
		fits := true
		for _, p := range pods {
			fits = fits && forestWidth(p.Roots, cw)+4 <= width
		}
		if fits {
			return cw
		}
	}
	return minCardWidth
}

// forestWidth is the width of the subtrees of nodes side by side.
func forestWidth(nodes []*Node, cw int) int {
	w := -treeGap
	for _, n := range nodes {
		w += treeGap + max(cw, forestWidth(n.Children, cw))
	}
	return max(0, w)
}

// treePodHeight is the height of a pod with one row of cards for each depth
// and one row of joins between two rows of cards.
func treePodHeight(p Pod) int {
	levels := 1 + depth(p.Roots)
	return 2 + levels*cardHeight + levels - 1
}

// forest draws the subtrees of nodes side by side from x, with their roots in
// row y.
func (d *drawer) forest(nodes []*Node, x, y int) {
	for _, n := range nodes {
		d.subtree(n, x, y)
		x += max(d.cw, forestWidth(n.Children, d.cw)) + treeGap
	}
}

// subtree draws the subtree of n with its left edge at x and the card of n
// in row y. The card is centered over the first and the last child. It
// returns the column of the card.
func (d *drawer) subtree(n *Node, x, y int) int {
	if len(n.Children) == 0 {
		d.card(n, x, y)
		return x
	}
	xs := make([]int, len(n.Children))
	at := x
	for i, child := range n.Children {
		xs[i] = d.subtree(child, at, y+cardHeight+1)
		at += max(d.cw, forestWidth(child.Children, d.cw)) + treeGap
	}
	cx := (xs[0] + xs[len(xs)-1]) / 2
	d.card(n, cx, y)
	d.joins(n, cx, y, xs)
	return cx
}

// joins draws the lines from the card of n at x, y to the cards of its
// children at xs in the next row of cards.
func (d *drawer) joins(n *Node, x, y int, xs []int) {
	half := d.cw / 2
	center, bar := x+half, y+cardHeight
	first, last := xs[0]+half, xs[len(xs)-1]+half
	for col := first; col <= last; col++ {
		d.c.Set(col, bar, '─', style.Border)
	}
	if len(xs) == 1 {
		d.c.Set(center, bar, '│', style.Border)
	} else {
		d.c.Set(first, bar, '╭', style.Border)
		d.c.Set(last, bar, '╮', style.Border)
		for _, cx := range xs[1 : len(xs)-1] {
			d.c.Set(cx+half, bar, '┬', style.Border)
		}
		up := '┴'
		for _, cx := range xs[1 : len(xs)-1] {
			if cx+half == center {
				up = '┼'
			}
		}
		d.c.Set(center, bar, up, style.Border)
	}
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
