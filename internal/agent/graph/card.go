package graph

import (
	"fmt"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/list"
)

// The card size of the historical Agents graph, widened so that the elapsed
// time with seconds fits beside the name.
const (
	cardWidth    = 28
	minCardWidth = 16
	cardHeight   = 4
)

// The SGR parameters of the historical palette. Each state also has a glyph
// or a border shape, so colour only adds.
const (
	styleBorder   = "38;2;61;58;51"
	styleDim      = "38;2;163;157;144"
	styleAccent   = "38;2;245;185;66"
	styleSelected = "1;38;2;245;185;66"
	styleWorking  = "38;2;110;230;154"
	styleError    = "38;2;232;112;95"
)

// frameInterval is the frame interval of the Pi working indicator.
const frameInterval = 80 * time.Millisecond

// workingFrames are the frames of the Pi working indicator. Frame 0 is the
// static working glyph.
var workingFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// statuses gives the glyph and the glyph style of each herdr status, in
// legend order.
var statuses = []struct{ status, glyph, style string }{
	{"working", "●", styleWorking},
	{"idle", "◐", styleAccent},
	{"blocked", "▲", styleError},
	{"done", "✓", styleDim},
	{"unknown", "◌", styleDim},
	{"gone", "○", styleDim},
}

// glyph gives the status glyph and its style. A working glyph at frame 1 or
// later is a braille frame.
func glyph(status string, frame int) (string, string) {
	if status == "working" && frame > 0 {
		return workingFrames[(frame-1)%len(workingFrames)], styleWorking
	}
	for _, s := range statuses {
		if s.status == status {
			return s.glyph, s.style
		}
	}
	return "◌", styleDim
}

// elapsed formats a duration with seconds: 7s, 4m05s, 26h04m03s.
func elapsed(age time.Duration) string {
	seconds := max(0, int(age/time.Second))
	h, m, s := seconds/3600, seconds/60%60, seconds%60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	case m > 0:
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func since(e list.Entry, now time.Time) string {
	if e.StartedAt.IsZero() {
		return ""
	}
	return elapsed(now.Sub(e.StartedAt))
}

// drawCard draws the card of e at x, y with width w. Row 1 has the status
// glyph, the name, and the elapsed time. Row 2 has the harness and the model.
// A selected card has a heavy border.
func drawCard(c *canvas, x, y, w int, e list.Entry, now time.Time, selected bool, frame int) {
	border, edge := styleBorder, "╭╮╰╯─│"
	switch {
	case selected:
		border, edge = styleSelected, "┏┓┗┛━┃"
	case e.Status == "working":
		border = styleWorking
	}
	corners := []rune(edge)
	c.box(x, y, w, cardHeight, corners, border)
	inner := w - 4
	mark, markStyle := glyph(e.Status, frame)
	c.text(x+2, y+1, mark, markStyle, 1)
	age := since(e, now)
	room := inner - 2
	if age != "" {
		room -= len(age) + 1
		c.text(x+w-2-len(age), y+1, age, styleDim, len(age))
	}
	nameStyle := ""
	switch {
	case selected:
		nameStyle = styleSelected
	case e.Status == "done" || e.Status == "gone":
		nameStyle = styleDim
	}
	c.text(x+4, y+1, clip(e.Name, room), nameStyle, room)
	sub := e.Harness
	if e.Model != "" {
		sub += " · " + e.Model
	}
	c.text(x+2, y+2, clip(sub, inner), styleDim, inner)
}

// cardLines renders one card alone.
func cardLines(e list.Entry, now time.Time, selected bool, frame, w int) []string {
	c := newCanvas(w, cardHeight)
	drawCard(c, 0, 0, w, e, now, selected, frame)
	return c.lines()
}

// clip cuts text to limit cells with … at the end.
func clip(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 {
		return ""
	}
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}

type cell struct {
	r     rune
	style string
}

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

// text writes at most limit runes of s from x.
func (c *canvas) text(x, y int, s, style string, limit int) {
	for i, r := range []rune(s) {
		if i >= limit {
			return
		}
		c.set(x+i, y, r, style)
	}
}

// box draws a border of w×h. Corners holds ╭╮╰╯ then the horizontal and the
// vertical edge.
func (c *canvas) box(x, y, w, h int, corners []rune, style string) {
	for i := 1; i < w-1; i++ {
		c.set(x+i, y, corners[4], style)
		c.set(x+i, y+h-1, corners[4], style)
	}
	for j := 1; j < h-1; j++ {
		c.set(x, y+j, corners[5], style)
		c.set(x+w-1, y+j, corners[5], style)
	}
	c.set(x, y, corners[0], style)
	c.set(x+w-1, y, corners[1], style)
	c.set(x, y+h-1, corners[2], style)
	c.set(x+w-1, y+h-1, corners[3], style)
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
		style := ""
		for _, cl := range row[:end] {
			if cl.r == 0 {
				cl = cell{' ', ""}
			}
			if cl.style != style {
				if style != "" {
					b.WriteString("\x1b[0m")
				}
				if cl.style != "" {
					b.WriteString("\x1b[" + cl.style + "m")
				}
				style = cl.style
			}
			b.WriteRune(cl.r)
		}
		if style != "" {
			b.WriteString("\x1b[0m")
		}
		out[y] = b.String()
	}
	return out
}
