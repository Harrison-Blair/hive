package graph

import (
	"fmt"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/canvas"
	"github.com/Harrison-Blair/hive/internal/lib/style"
)

// The card size of the historical Agents graph, widened so that a name of 16
// cells, such as hive-coordinator, and a time in state with the ≥ mark and
// seconds up to ≥99h59m59s fit on row 1.
const (
	cardWidth    = 33
	minCardWidth = 16
	cardHeight   = 5
)

// frameInterval is the frame interval of the Pi working indicator.
const frameInterval = 80 * time.Millisecond

// workingFrames are the frames of the Pi working indicator. The working glyph
// is always one of them.
var workingFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// statuses gives the glyph and the glyph style of each herdr status, in
// legend order.
var statuses = []struct{ status, glyph, style string }{
	{"working", workingFrames[0], style.Working},
	{"idle", "○", style.Accent},
	{"blocked", "×", style.Error},
	{"done", "✓", style.Dim},
	{"unknown", "·", style.Dim},
	{"gone", "–", style.Dim},
}

// glyph gives the status glyph and its style. The working glyph is the
// braille frame of frame.
func glyph(status string, frame int) (string, string) {
	if status == "working" {
		return workingFrames[frame%len(workingFrames)], style.Working
	}
	for _, s := range statuses {
		if s.status == status {
			return s.glyph, s.style
		}
	}
	return "·", style.Dim
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
// glyph, the name, and age, the time in the current state. The age shows only
// when it fits after the glyph, and the name gets the room that is left. Row 2 has the checkout, and row 3
// has the harness in its harness colour and the model.
// A selected card has a heavy border.
func drawCard(c *canvas.Canvas, x, y, w int, e list.Entry, age string, selected bool, frame int, checkout string) {
	border, edge := style.Border, "╭╮╰╯─│"
	switch {
	case selected:
		border, edge = style.Selected, "┏┓┗┛━┃"
	case e.Status == "working":
		border = style.Working
	}
	corners := []rune(edge)
	c.Box(x, y, w, cardHeight, corners, border)
	inner := w - 4
	mark, markStyle := glyph(e.Status, frame)
	c.Text(x+2, y+1, mark, markStyle, 1)
	room := inner - 2
	if cells := len([]rune(age)); cells > 0 && cells <= room {
		room -= cells + 1
		c.Text(x+w-2-cells, y+1, age, style.Dim, cells)
	}
	nameStyle := ""
	switch {
	case selected:
		nameStyle = style.Selected
	case e.Status == "done" || e.Status == "gone":
		nameStyle = style.Dim
	}
	c.Text(x+4, y+1, canvas.Clip(e.Name, room), nameStyle, room)
	c.Text(x+2, y+2, canvas.Clip(checkout, inner), style.Dim, inner)
	sub := e.Harness
	if e.Model != "" {
		model := e.Model
		if _, name, ok := strings.Cut(model, "/"); ok {
			model = name
		}
		if strings.HasPrefix(model, "gpt-") {
			model = strings.TrimPrefix(model, "gpt-")
		} else {
			model = strings.TrimPrefix(model, "claude-")
		}
		sub += " · " + model
	}
	sub = canvas.Clip(sub, inner)
	c.Text(x+2, y+3, sub, style.Dim, inner)
	c.Text(x+2, y+3, sub, style.Harness(e.Harness), min(len([]rune(e.Harness)), len([]rune(sub))))
}

// cardLines renders one card alone.
func cardLines(e list.Entry, age string, selected bool, frame, w int) []string {
	c := canvas.New(w, cardHeight)
	drawCard(c, 0, 0, w, e, age, selected, frame, "")
	return c.Lines()
}
