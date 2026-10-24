package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// terminalScreen replays a terminal write stream onto the final text of a
// width×height screen. The TUI writes only the cells that changed in a row,
// so the raw stream is not the screen and ansi.Strip cannot rebuild it. The
// stream comes from a non-terminal output, where the renderer maps a newline
// to a carriage return and a line feed.
func terminalScreen(raw string, width, height int) []string {
	// The program leaves the alternate screen at exit. The state before that
	// sequence is the screen the user saw.
	if index := strings.Index(raw, "\x1b[?1049l"); index >= 0 {
		raw = raw[:index]
	}
	screen := &screenGrid{width: width, height: height}
	screen.reset()
	parser := ansi.NewParser()
	parser.SetHandler(ansi.Handler{
		Print:     screen.print,
		Execute:   screen.execute,
		HandleCsi: screen.csi,
		HandleEsc: screen.esc,
	})
	parser.Parse([]byte(raw))
	lines := make([]string, height)
	for row := range screen.rows {
		lines[row] = strings.TrimRight(string(screen.rows[row]), " ")
	}
	return lines
}

// screenGrid is the cell grid of terminalScreen.
type screenGrid struct {
	width, height int
	rows          [][]rune
	row, col      int
	wrap          bool
	savedRow      int
	savedCol      int
}

func (s *screenGrid) reset() {
	s.rows = make([][]rune, s.height)
	for row := range s.rows {
		s.rows[row] = []rune(strings.Repeat(" ", s.width))
	}
	s.row, s.col, s.wrap = 0, 0, false
}

func (s *screenGrid) print(r rune) {
	if s.wrap {
		s.row++
		s.col = 0
		s.wrap = false
	}
	if s.row >= s.height {
		s.row = s.height - 1
	}
	width := ansi.StringWidth(string(r))
	if width <= 0 {
		return
	}
	if s.col >= s.width {
		s.col = s.width - 1
	}
	s.rows[s.row][s.col] = r
	for cell := 1; cell < width && s.col+cell < s.width; cell++ {
		s.rows[s.row][s.col+cell] = ' '
	}
	s.col += width
	if s.col >= s.width {
		s.col = s.width - 1
		s.wrap = true
	}
}

func (s *screenGrid) execute(code byte) {
	switch code {
	case '\r':
		s.col, s.wrap = 0, false
	case '\n':
		s.row, s.col, s.wrap = min(s.row+1, s.height-1), 0, false
	case '\b':
		s.col, s.wrap = max(s.col-1, 0), false
	case '\t':
		s.col, s.wrap = min((s.col/8+1)*8, s.width-1), false
	}
}

func (s *screenGrid) csi(command ansi.Cmd, params ansi.Params) {
	count := func(index, fallback int) int {
		value, _, _ := params.Param(index, fallback)
		return value
	}
	switch command.Final() {
	case 'H', 'f':
		s.moveTo(count(0, 1)-1, count(1, 1)-1)
	case 'A':
		s.moveTo(s.row-count(0, 1), s.col)
	case 'B':
		s.moveTo(s.row+count(0, 1), s.col)
	case 'C':
		s.moveTo(s.row, s.col+count(0, 1))
	case 'D':
		s.moveTo(s.row, s.col-count(0, 1))
	case 'E':
		s.moveTo(s.row+count(0, 1), 0)
	case 'F':
		s.moveTo(s.row-count(0, 1), 0)
	case 'G', '`':
		s.moveTo(s.row, count(0, 1)-1)
	case 'd':
		s.moveTo(count(0, 1)-1, s.col)
	case 'J':
		switch count(0, 0) {
		case 0:
			s.clearFrom(s.row, s.col)
			for row := s.row + 1; row < s.height; row++ {
				s.clearFrom(row, 0)
			}
		case 1:
			for row := 0; row < s.row; row++ {
				s.clearFrom(row, 0)
			}
			for col := 0; col <= s.col && col < s.width; col++ {
				s.rows[s.row][col] = ' '
			}
		case 2, 3:
			for row := range s.rows {
				s.clearFrom(row, 0)
			}
		}
	case 'K':
		switch count(0, 0) {
		case 0:
			s.clearFrom(s.row, s.col)
		case 1:
			for col := 0; col <= s.col && col < s.width; col++ {
				s.rows[s.row][col] = ' '
			}
		case 2:
			s.clearFrom(s.row, 0)
		}
	case 'X':
		for col := s.col; col < s.col+count(0, 1) && col < s.width; col++ {
			s.rows[s.row][col] = ' '
		}
	}
}

func (s *screenGrid) esc(command ansi.Cmd) {
	switch command.Final() {
	case '7':
		s.savedRow, s.savedCol = s.row, s.col
	case '8':
		s.moveTo(s.savedRow, s.savedCol)
	case 'D':
		s.moveTo(s.row+1, s.col)
	case 'E':
		s.moveTo(s.row+1, 0)
	case 'M':
		s.moveTo(s.row-1, s.col)
	}
}

func (s *screenGrid) moveTo(row, col int) {
	s.row = min(max(row, 0), s.height-1)
	s.col = min(max(col, 0), s.width-1)
	s.wrap = false
}

func (s *screenGrid) clearFrom(row, col int) {
	for cell := col; cell < s.width; cell++ {
		s.rows[row][cell] = ' '
	}
}

// TestWorkTitleSurvivesIncrementalRedraw proves the terminal-state check of
// TestRunHomeShowsTasksFromStringReader. The first Work frame has no task
// data. The frame after the load writes only the title suffix at row 1,
// column 6, so the stripped stream holds no contiguous title.
func TestWorkTitleSurvivesIncrementalRedraw(t *testing.T) {
	stream := "\x1b[H\x1b[2J work · no goals" +
		"\n ────────────────────────────────────────────────────────────────────────────────" +
		"\x1b[12;23Hno goals · hive goal create starts one" +
		"\x1b[1;6H · g1-tui · hive TUI" +
		"\x1b[9;6H● 79-tui-shell" +
		"\x1b[18;4H┏━ wave-c · done · 1/1 ━━━━━━━━━━━┓"
	if strings.Contains(ansi.Strip(stream), "work · g1-tui") {
		t.Fatal("the test stream holds the title in one write")
	}
	screen := strings.Join(terminalScreen(stream, defaultWidth, defaultHeight), "\n")
	for _, text := range []string{"work · g1-tui", "wave-c · done", "79-tui-shell"} {
		if !strings.Contains(screen, text) {
			t.Errorf("final screen does not contain %q:\n%s", text, screen)
		}
	}
}
