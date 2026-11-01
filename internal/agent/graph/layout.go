package graph

import "time"

// span is the rows of one card in the pod area.
type span struct {
	top    int
	status string
}

// board is the drawn pod area: its lines and the rows of each card.
type board struct {
	lines []string
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

// drawBoard draws the pods in rows that fit width. Each card narrows down to
// minCardWidth to fit one pod in the width. A pod that is still too wide is
// cut at the right edge of the screen.
func drawBoard(pods []Pod, width int, now time.Time, selected string, frame int) board {
	deepest := 0
	for _, p := range pods {
		deepest = max(deepest, depth(p.Roots))
	}
	cw := min(cardWidth, max(minCardWidth, width-4-deepest))
	podWidth := deepest + cw + 4
	columns := max(1, (width+1)/(podWidth+1))
	var rowHeights []int
	for i, p := range pods {
		if i%columns == 0 {
			rowHeights = append(rowHeights, 0)
		}
		rowHeights[len(rowHeights)-1] = max(rowHeights[len(rowHeights)-1], podHeight(p))
	}
	total := 0
	for _, h := range rowHeights {
		total += h
	}
	c := newCanvas(columns*(podWidth+1)-1, total)
	d := drawer{c: c, now: now, selected: selected, frame: frame, cw: cw, cards: map[string]span{}}
	y := 0
	for row, h := range rowHeights {
		for col := range columns {
			i := row*columns + col
			if i >= len(pods) {
				break
			}
			x := col * (podWidth + 1)
			d.pod(pods[i], x, y, podWidth, h)
		}
		y += h
	}
	return board{lines: c.lines(), cards: d.cards}
}

type drawer struct {
	c        *canvas
	now      time.Time
	selected string
	frame    int
	cw       int
	cards    map[string]span
}

func (d *drawer) pod(p Pod, x, y, w, h int) {
	d.c.box(x, y, w, h, []rune("╭╮╰╯─│"), styleBorder)
	d.c.text(x+2, y, " "+clip(p.Title, w-6)+" ", "", w-4)
	top := y + 1
	for _, root := range p.Roots {
		d.tree(root, x+2, top)
		top += treeHeight(root) + 1
	}
}

// tree draws the card of n at x, y and its children below it, each one
// column to the right, joined to n by a trunk in column x.
func (d *drawer) tree(n *Node, x, y int) {
	selected := n.Entry.Name == d.selected
	drawCard(d.c, x, y, d.cw, n.Entry, d.now, selected, d.frame)
	d.cards[n.Entry.Name] = span{top: y, status: n.Entry.Status}
	if len(n.Children) == 0 {
		return
	}
	corner, cornerStyle := '├', styleBorder
	if selected {
		corner, cornerStyle = '┣', styleSelected
	} else if n.Entry.Status == "working" {
		cornerStyle = styleWorking
	}
	d.c.set(x, y+3, corner, cornerStyle)
	row := y + cardHeight
	for i, child := range n.Children {
		last := i == len(n.Children)-1
		top := row + 1
		join := top + 1
		for r := row; r < join; r++ {
			d.c.set(x, r, '│', styleBorder)
		}
		branch := '├'
		if last {
			branch = '╰'
		}
		d.c.set(x, join, branch, styleBorder)
		h := treeHeight(child)
		if !last {
			for r := join + 1; r < top+h; r++ {
				d.c.set(x, r, '│', styleBorder)
			}
		}
		d.tree(child, x+1, top)
		tee, teeStyle := '┤', styleBorder
		switch {
		case child.Entry.Name == d.selected:
			tee, teeStyle = '┨', styleSelected
		case child.Entry.Status == "working":
			teeStyle = styleWorking
		}
		d.c.set(x+1, join, tee, teeStyle)
		row = top + h
	}
}
