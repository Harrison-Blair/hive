package coordinator

import (
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/style"
)

type cell struct {
	r     rune
	style string
}

// span is one run of title text with its SGR style.
type span struct{ text, style string }

// canvas is a grid of one-cell runes, each with an SGR style.
type canvas struct {
	w, h  int
	cells [][]cell
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w: w, h: h, cells: make([][]cell, h)}
	for y := range c.cells {
		c.cells[y] = make([]cell, w)
	}
	return c
}

func (c *canvas) set(x, y int, r rune, style string) {
	if x >= 0 && y >= 0 && x < c.w && y < c.h {
		c.cells[y][x] = cell{r, style}
	}
}

// text writes s from x. A text longer than limit cells ends in … at limit.
func (c *canvas) text(x, y int, s, style string, limit int) {
	runes := []rune(s)
	if limit <= 0 {
		return
	}
	if len(runes) > limit {
		runes = append(runes[:limit-1], '…')
	}
	for i, r := range runes {
		c.set(x+i, y, r, style)
	}
}

// box draws the rounded border of w×h at the top left.
func (c *canvas) box(w, h int) {
	for x := 1; x < w-1; x++ {
		c.set(x, 0, '─', style.Border)
		c.set(x, h-1, '─', style.Border)
	}
	for y := 1; y < h-1; y++ {
		c.set(0, y, '│', style.Border)
		c.set(w-1, y, '│', style.Border)
	}
	c.set(0, 0, '╭', style.Border)
	c.set(w-1, 0, '╮', style.Border)
	c.set(0, h-1, '╰', style.Border)
	c.set(w-1, h-1, '╯', style.Border)
}

// title writes the spans into the top edge, joined by a dim " · ".
func (c *canvas) title(spans []span) {
	x, end := 3, c.w-2
	c.set(2, 0, ' ', "")
	for i, s := range spans {
		if i > 0 {
			c.text(x, 0, " · ", style.Dim, end-x)
			x += 3
		}
		c.text(x, 0, s.text, s.style, end-x)
		x += len([]rune(s.text))
	}
	if x < end {
		c.set(x, 0, ' ', "")
	}
}

// lines renders each row with SGR sequences, without trailing blanks.
func (c *canvas) lines() []string {
	out := make([]string, c.h)
	for y, row := range c.cells {
		end := len(row)
		for end > 0 && row[end-1].r == 0 {
			end--
		}
		var b strings.Builder
		current := ""
		for _, cl := range row[:end] {
			if cl.r == 0 {
				cl = cell{' ', ""}
			}
			if cl.style != current {
				if current != "" {
					b.WriteString("\x1b[0m")
				}
				if cl.style != "" {
					b.WriteString("\x1b[" + cl.style + "m")
				}
				current = cl.style
			}
			b.WriteRune(cl.r)
		}
		if current != "" {
			b.WriteString("\x1b[0m")
		}
		out[y] = b.String()
	}
	return out
}
