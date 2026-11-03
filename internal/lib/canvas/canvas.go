// Package canvas is a grid of one-cell runes, each with an SGR style, that
// the agent graph and the picker of bare hive draw on.
package canvas

import "strings"

// Cell is one rune of the grid with its SGR style. R is 0 for an unset cell.
type Cell struct {
	R     rune
	Style string
}

// Canvas is a grid of W × H cells.
type Canvas struct {
	W, H  int
	Cells [][]Cell
}

// New returns an empty canvas of w × h.
func New(w, h int) *Canvas {
	c := &Canvas{W: w, H: h, Cells: make([][]Cell, h)}
	for y := range c.Cells {
		c.Cells[y] = make([]Cell, w)
	}
	return c
}

// Set writes one cell. A cell outside the grid is ignored.
func (c *Canvas) Set(x, y int, r rune, style string) {
	if x >= 0 && y >= 0 && x < c.W && y < c.H {
		c.Cells[y][x] = Cell{r, style}
	}
}

// Text writes at most limit runes of s from x.
func (c *Canvas) Text(x, y int, s, style string, limit int) {
	for i, r := range []rune(s) {
		if i >= limit {
			return
		}
		c.Set(x+i, y, r, style)
	}
}

// Box draws a border of w×h. Corners holds ╭╮╰╯ then the horizontal and the
// vertical edge.
func (c *Canvas) Box(x, y, w, h int, corners []rune, style string) {
	for i := 1; i < w-1; i++ {
		c.Set(x+i, y, corners[4], style)
		c.Set(x+i, y+h-1, corners[4], style)
	}
	for j := 1; j < h-1; j++ {
		c.Set(x, y+j, corners[5], style)
		c.Set(x+w-1, y+j, corners[5], style)
	}
	c.Set(x, y, corners[0], style)
	c.Set(x+w-1, y, corners[1], style)
	c.Set(x, y+h-1, corners[2], style)
	c.Set(x+w-1, y+h-1, corners[3], style)
}

// Lines renders each row with SGR sequences, without trailing blanks.
func (c *Canvas) Lines() []string {
	out := make([]string, c.H)
	for y, row := range c.Cells {
		end := len(row)
		for end > 0 && row[end-1].R == 0 {
			end--
		}
		var b strings.Builder
		style := ""
		for _, cl := range row[:end] {
			if cl.R == 0 {
				cl = Cell{' ', ""}
			}
			if cl.Style != style {
				if style != "" {
					b.WriteString("\x1b[0m")
				}
				if cl.Style != "" {
					b.WriteString("\x1b[" + cl.Style + "m")
				}
				style = cl.Style
			}
			b.WriteRune(cl.R)
		}
		if style != "" {
			b.WriteString("\x1b[0m")
		}
		out[y] = b.String()
	}
	return out
}

// Clip cuts text to limit runes with … at the end.
func Clip(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 {
		return ""
	}
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}
