// Package canvas is a grid of one-cell runes, each with an SGR style, that
// the agent graph and the picker of bare hive draw on.
package canvas

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Cell is one column of the grid with its SGR style. R is 0 for an unset
// cell. G holds a grapheme of more than one rune, of which R is the first. A
// wide grapheme takes its cell and the next one, which has cont set.
type Cell struct {
	R     rune
	Style string
	G     string
	cont  bool
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

// Set writes one cell of one column. A cell outside the grid is ignored.
func (c *Canvas) Set(x, y int, r rune, style string) {
	c.put(x, y, Cell{R: r, Style: style})
}

// put writes cell at x, y. When it covers half of a wide grapheme, the other
// half becomes a blank in the style of the grapheme.
func (c *Canvas) put(x, y int, cell Cell) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	row := c.Cells[y]
	if row[x].cont && x > 0 {
		row[x-1] = Cell{R: ' ', Style: row[x-1].Style}
	}
	if x+1 < c.W && row[x+1].cont {
		row[x+1] = Cell{R: ' ', Style: row[x].Style}
	}
	row[x] = cell
}

// Text writes s from x, one grapheme at a time, in at most limit columns. A
// wide grapheme takes two columns. A grapheme that does not fit in limit or
// in the grid ends the text.
func (c *Canvas) Text(x, y int, s, style string, limit int) {
	col := 0
	for s != "" {
		g, width := ansi.FirstGraphemeCluster(s, ansi.GraphemeWidth)
		s = s[len(g):]
		if width == 0 {
			continue
		}
		if col+width > limit || x+col+width > c.W {
			return
		}
		cell := Cell{R: []rune(g)[0], Style: style}
		if len([]rune(g)) > 1 {
			cell.G = g
		}
		c.put(x+col, y, cell)
		for i := 1; i < width; i++ {
			c.put(x+col+i, y, Cell{Style: style, cont: true})
		}
		col += width
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
			if cl.cont {
				continue
			}
			if cl.R == 0 {
				cl = Cell{R: ' '}
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
			if cl.G != "" {
				b.WriteString(cl.G)
			} else {
				b.WriteRune(cl.R)
			}
		}
		if style != "" {
			b.WriteString("\x1b[0m")
		}
		out[y] = b.String()
	}
	return out
}

// Clip cuts text to limit columns with … at the end.
func Clip(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	return ansi.Truncate(text, limit, "…")
}
