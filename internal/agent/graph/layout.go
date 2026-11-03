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

// drawBoard draws the pods in rows that fit width, as trees when tree is
// true and else as indented lists. A pod goes beside the pod before it when
// it fits, else it starts a new row. Each card narrows down to minCardWidth to
// fit the widest pod in the width. A board that is still too wide is wider
// than the screen, and the view scrolls it sideways.
func drawBoard(pods []Pod, width int, ages map[string]string, selected string, frame int, checkouts map[string]string, tree bool) board {
	d := drawer{ages: ages, selected: selected, frame: frame, cards: map[string]span{}, checkouts: checkouts}
	sizes := make([][2]int, len(pods))
	if tree {
		d.cw = treeCardWidth(pods, width)
		for i, p := range pods {
			sizes[i] = [2]int{forestWidth(p.Roots, d.cw) + 4, treePodHeight(p)}
		}
	} else {
		deepest := 0
		for _, p := range pods {
			deepest = max(deepest, depth(p.Roots))
		}
		d.cw = min(cardWidth, max(minCardWidth, width-4-deepest))
		for i, p := range pods {
			sizes[i] = [2]int{deepest + d.cw + 4, podHeight(p)}
		}
	}
	// at holds the column and the row of pods of each pod.
	at := make([][2]int, len(pods))
	var rowHeights []int
	x, boardWidth := 0, 0
	for i, size := range sizes {
		if len(rowHeights) == 0 || x > 0 && x+size[0] > width {
			rowHeights = append(rowHeights, 0)
			x = 0
		}
		row := len(rowHeights) - 1
		at[i] = [2]int{x, row}
		rowHeights[row] = max(rowHeights[row], size[1])
		boardWidth = max(boardWidth, x+size[0])
		x += size[0] + 1
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
		d.box(p, x, y, sizes[i][0], h)
		if tree {
			d.forest(p.Roots, x+2, y+1)
			continue
		}
		top := y + 1
		for _, root := range p.Roots {
			d.tree(root, x+2, top)
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

// card draws the card of n at x, y and records its span.
func (d *drawer) card(n *Node, x, y int) {
	drawCard(d.c, x, y, d.cw, n.Entry, d.ages[n.Entry.Name], n.Entry.Name == d.selected, d.frame, d.checkouts[n.Entry.Name])
	d.cards[n.Entry.Name] = span{top: y, left: x, width: d.cw, status: n.Entry.Status}
}

// tree draws the card of n at x, y and its children below it, each one
// column to the right, joined to n by a trunk in column x.
func (d *drawer) tree(n *Node, x, y int) {
	d.card(n, x, y)
	if len(n.Children) == 0 {
		return
	}
	row := y + cardHeight
	for i, child := range n.Children {
		last := i == len(n.Children)-1
		top := row + 1
		join := top + 1
		for r := row; r < join; r++ {
			d.c.Set(x, r, '│', style.Border)
		}
		branch := '├'
		if last {
			branch = '╰'
		}
		d.c.Set(x, join, branch, style.Border)
		h := treeHeight(child)
		if !last {
			for r := join + 1; r < top+h; r++ {
				d.c.Set(x, r, '│', style.Border)
			}
		}
		d.tree(child, x+1, top)
		row = top + h
	}
}
