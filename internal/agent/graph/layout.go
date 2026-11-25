package graph

import (
	"github.com/Harrison-Blair/hive/internal/lib/canvas"
	"github.com/Harrison-Blair/hive/internal/lib/style"
)

// span is the place of one card in the pod area: its first row, its first
// column, and its width.
type span struct {
	top, left, width int
	status           string
}

// board is the drawn pod area: its lines, its width, and the place of each
// card.
type board struct {
	lines []string
	width int
	cards map[string]span
}

func depth(nodes []*Node) int {
	deepest := 0
	for _, n := range nodes {
		if len(n.Children) > 0 {
			deepest = max(deepest, 1+depth(n.Children))
		}
	}
	return deepest
}

func treeHeight(n *Node) int {
	h := cardHeight
	for _, child := range n.Children {
		h += 1 + treeHeight(child)
	}
	return h
}

func podHeight(p Pod) int {
	h := 2 + max(0, len(p.Roots)-1)
	for _, root := range p.Roots {
		h += treeHeight(root)
	}
	return h
}

// The track grid: a track is one card column, trackPitch cells apart. A card
// starts cardInset cells right of the first column of its track, and a pod
// box starts at the first column of its first track.
const (
	trackPitch = 40
	cardInset  = 3
)

// drawBoard draws the pods in rows of tracks that fit width, as trees when
// tree is true and else as indented lists. A pod goes beside the pod before
// it when its tracks fit in the row, else it starts a new row. Tree cards
// keep their full width; list cards narrow down to minCardWidth to fit. A
// board wider than the screen scrolls sideways.
func drawBoard(pods []Pod, width int, ages map[string]string, selected string, frame int, checkouts map[string]string, tree bool) board {
	d := drawer{ages: ages, selected: selected, frame: frame, cards: map[string]span{}, checkouts: checkouts}
	// sizes holds the tracks, the width, and the height of each pod.
	sizes := make([][3]int, len(pods))
	if tree {
		d.cw = cardWidth
		for i, p := range pods {
			n := leaves(p.Roots)
			sizes[i] = [3]int{n, trackPitch*n - 1, treePodHeight(p)}
		}
	} else {
		d.cw = min(cardWidth, max(minCardWidth, width-2*cardInset))
		for i, p := range pods {
			sizes[i] = [3]int{1, d.cw + 2*cardInset, podHeight(p)}
		}
	}
	perRow := max(1, (width+1)/trackPitch)
	// at holds the column and the row of pods of each pod.
	at := make([][2]int, len(pods))
	var rowHeights []int
	track, boardWidth := 0, 0
	for i, size := range sizes {
		if len(rowHeights) == 0 || track > 0 && track+size[0] > perRow {
			rowHeights = append(rowHeights, 0)
			track = 0
		}
		row := len(rowHeights) - 1
		at[i] = [2]int{track * trackPitch, row}
		rowHeights[row] = max(rowHeights[row], size[2])
		boardWidth = max(boardWidth, at[i][0]+size[1])
		track += size[0]
	}
	tops := make([]int, len(rowHeights))
	total := 0
	for i, h := range rowHeights {
		tops[i] = total
		total += h
	}
	d.c = canvas.New(boardWidth, total)
	for i, p := range pods {
		x, y, h := at[i][0], tops[at[i][1]], rowHeights[at[i][1]]
		d.box(p, x, y, sizes[i][1], h)
		if tree {
			d.forest(p.Roots, x+cardInset, y+1)
			continue
		}
		top := y + 1
		for _, root := range p.Roots {
			d.tree(root, x+cardInset+d.cw, 0, top)
			top += treeHeight(root) + 1
		}
	}
	return board{lines: d.c.Lines(), width: boardWidth, cards: d.cards}
}

type drawer struct {
	c         *canvas.Canvas
	ages      map[string]string
	selected  string
	frame     int
	cw        int
	cards     map[string]span
	checkouts map[string]string
}

// box draws the border and the title of p.
func (d *drawer) box(p Pod, x, y, w, h int) {
	d.c.Box(x, y, w, h, []rune("╭╮╰╯─│"), style.Border)
	d.c.Text(x+2, y, " "+canvas.Clip(p.Title, w-6)+" ", "", w-4)
}

// card draws the card of n at x, y with width w and records its span.
func (d *drawer) card(n *Node, x, y, w int) {
	drawCard(d.c, x, y, w, n.Entry, d.ages[n.Entry.Name], n.Entry.Name == d.selected, d.frame, d.checkouts[n.Entry.Name])
	d.cards[n.Entry.Name] = span{top: y, left: x, width: w, status: n.Entry.Status}
}

// listWidth is the width of a list card at depth: 3 narrower than its parent
// down to minCardWidth, and no narrower after depth 5.
func (d *drawer) listWidth(depth int) int {
	return max(minCardWidth, d.cw-cardInset*min(depth, 5))
}

// tree draws the card of n at depth in row y, ending in column end - 1, and
// its children below it, each ending in the same column. A rail in the
// second column of the card of n joins each child that starts 3 columns
// right of n with a tick in the middle row of that child.
func (d *drawer) tree(n *Node, end, depth, y int) {
	x := end - d.listWidth(depth)
	d.card(n, x, y, end-x)
	if len(n.Children) == 0 {
		return
	}
	rail := end-d.listWidth(depth+1)-x >= cardInset
	row := y + cardHeight
	var ticks []int
	for _, child := range n.Children {
		top := row + 1
		d.tree(child, end, depth+1, top)
		ticks = append(ticks, top+2)
		row = top + treeHeight(child)
	}
	if !rail {
		return
	}
	last := ticks[len(ticks)-1]
	for r := y + cardHeight; r < last; r++ {
		d.c.Set(x+1, r, '│', style.Border)
	}
	for _, r := range ticks {
		d.c.Set(x+1, r, '├', style.Border)
		d.c.Set(x+2, r, '─', style.Border)
	}
	d.c.Set(x+1, last, '╰', style.Border)
}
