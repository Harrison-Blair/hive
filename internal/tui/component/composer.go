package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Composer is the rounded frame of golden-frames chat.md rules 9 to 11 and
// 14 to 16: the input line, and the ask card above it while an ask is open.
type Composer struct {
	Text string
	// Cursor draws the cursor █ at CursorLine and CursorColumn, a line of
	// Text and a rune column in that line.
	Cursor       bool
	CursorLine   int
	CursorColumn int
	// Placeholder shows in dim after "> " while Text is empty.
	Placeholder string
	// InputRows is the most rows of the input line. Beyond it the text
	// scrolls and keeps the cursor row in view.
	InputRows int
	Card      *ComposerCard
}

// ComposerCard is the ask card in the composer frame. Lines are the styled
// card rows at width-3 cells, from column 2. Title and Note are the title and
// the right text of the top border, which the card styles. Offset is the
// first row in view, and Rows is the most rows in view.
type ComposerCard struct {
	Title  string
	Note   string
	Lines  []string
	Offset int
	Rows   int
}

var (
	borderAccent = accentStyle
	userStyle    = lipgloss.NewStyle().Foreground(User)
)

// View renders the frame at exactly width cells. The border is accent.
func (c Composer) View(width int) string {
	inner := width - 2
	var rows []string
	if c.Card == nil {
		rows = append(rows, borderAccent.Render("╭"+strings.Repeat("─", inner)+"╮"))
	} else {
		rows = append(rows, c.Card.top(width))
		visible := max(min(len(c.Card.Lines)-c.Card.Offset, c.Card.Rows), 0)
		mark := Thumb(visible, len(c.Card.Lines), visible, c.Card.Offset)
		for i := range visible {
			rows = append(rows, frameRow(c.Card.Lines[c.Card.Offset+i], width, mark.covers(i)))
		}
		rows = append(rows, borderAccent.Render("├"+strings.Repeat("─", inner)+"┤"))
	}
	text, mark := c.inputLines(width)
	for i, line := range text {
		rows = append(rows, frameRow(line, width, mark.covers(i)))
	}
	rows = append(rows, borderAccent.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(rows, "\n")
}

// InputHeight gives the rows of the input line at width cells, up to rows.
func InputHeight(text string, width, rows int) int {
	return min(len(wordWrap(text, max(width-5, 1))), max(rows, 1))
}

func (card *ComposerCard) top(width int) string {
	title := titleStyle.Render(" " + card.Title + " ")
	note := ""
	if card.Note != "" {
		note = card.Note
	}
	fill := width - 3 - lipgloss.Width(title) - lipgloss.Width(note)
	if fill < 0 {
		note, fill = "", max(width-3-lipgloss.Width(title), 0)
	}
	return ansi.Truncate(borderAccent.Render("╭─")+title+borderAccent.Render(strings.Repeat("─", fill))+note+borderAccent.Render("╮"), width, "")
}

// frameRow puts line between the side borders, from column 2. thumb puts
// the dim ┃ on the right border.
func frameRow(line string, width int, thumb bool) string {
	right := borderAccent.Render("│")
	if thumb {
		right = dimStyle.Render("┃")
	}
	return borderAccent.Render("│") + " " + fitRow(line, width-3) + right
}

// inputLines gives the visible text rows of the input line and the
// scroll mark.
func (c Composer) inputLines(width int) ([]string, ScrollMark) {
	room := max(width-5, 1)
	if c.Text == "" {
		line := userStyle.Render("> ")
		if c.Placeholder != "" {
			line += dimStyle.Render(c.Placeholder)
		}
		if c.Cursor && c.Placeholder == "" {
			line += userStyle.Render("█")
		}
		return []string{line}, ScrollMark{}
	}
	wrapped := wordWrap(c.Text, room)
	visible := min(len(wrapped), max(c.InputRows, 1))
	cursorRow, cursorCol := len(wrapped)-1, lipgloss.Width(wrapped[len(wrapped)-1])
	if c.Cursor {
		cursorRow, cursorCol = inputCursor(c.Text, room, c.CursorLine, c.CursorColumn)
		cursorCol -= 2
	}
	offset := min(max(cursorRow-visible+1, 0), max(len(wrapped)-visible, 0))
	lines := make([]string, visible)
	for i := range visible {
		row := offset + i
		prompt := "  "
		if row == 0 {
			prompt = "> "
		}
		text := wrapped[row]
		if c.Cursor && row == cursorRow {
			before := ansi.Truncate(text, cursorCol, "")
			after := ansi.TruncateLeft(text, cursorCol+1, "")
			text = before + "█" + after
			if cursorCol >= lipgloss.Width(wrapped[row]) {
				text = wrapped[row] + strings.Repeat(" ", cursorCol-lipgloss.Width(wrapped[row])) + "█"
			}
		}
		lines[i] = userStyle.Render(prompt + text)
	}
	return lines, Thumb(visible, len(wrapped), visible, offset)
}

// inputCursor gives the wrapped row and the cell column, after the two-cell
// prompt, of rune column col in line of text. Each wrapped row is a part of
// its line. A cursor on the space at a soft break shows at the end of the row
// before it.
func inputCursor(text string, width, line, col int) (int, int) {
	row := 0
	for i, para := range strings.Split(text, "\n") {
		runes := []rune(para)
		paraRows := wordWrap(para, width)
		if i < line {
			row += len(paraRows)
			continue
		}
		pos, cell := 0, 0
		for k, r := range paraRows {
			for pos < len(runes) && !strings.HasPrefix(string(runes[pos:]), r) {
				pos++
			}
			if k > 0 && pos > col {
				break
			}
			end := pos + len([]rune(r))
			row += min(k, 1)
			cell = lipgloss.Width(string(runes[pos:min(max(col, pos), end)]))
			pos = end
		}
		return row, 2 + cell
	}
	return row, 2
}
