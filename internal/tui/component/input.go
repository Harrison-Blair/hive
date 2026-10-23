package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var userStyle = lipgloss.NewStyle().Foreground(User)

// Input is the fixed chat input below the feed (decision 15). It can grow to
// the height that the caller gives it. Text uses the user token.
type Input struct {
	Text    string
	Focused bool
	Height  int
	// CursorLine and CursorColumn give the cursor as a line of Text and a
	// rune column in that line.
	CursorLine   int
	CursorColumn int
	HasCursor    bool
	Status       string
	Age          string
	Queued       bool
	// Placeholder shows in dim, with no prompt, when the input has no text
	// and no focus.
	Placeholder string
}

// View renders the input at exactly width cells. Height is the maximum panel
// height. When Height is zero, the input grows to fit its text.
func (in Input) View(width int) string {
	width = max(width, PanelMinWidth)
	textRows := inputTextLines(in.Text, max(width-6, 1))
	panelHeight := len(textRows) + 2
	if in.Height > 0 {
		panelHeight = min(max(in.Height, 3), panelHeight)
	}
	visibleRows := max(panelHeight-2, 1)
	cursorRow, cursorCol := len(textRows)-1, 2+lipgloss.Width(textRows[len(textRows)-1])
	if in.HasCursor {
		cursorRow, cursorCol = inputCursor(in.Text, max(width-6, 1), in.CursorLine, in.CursorColumn)
	}
	cursorRow = min(max(cursorRow, 0), len(textRows)-1)
	cursorCol = min(cursorCol, 2+lipgloss.Width(textRows[cursorRow]))
	offset := max(cursorRow-visibleRows+1, 0)
	offset = min(offset, max(len(textRows)-visibleRows, 0))
	visible := textRows[offset:min(offset+visibleRows, len(textRows))]
	lines := make([]string, len(visible))
	for i, line := range visible {
		prompt := "  "
		if offset+i == 0 {
			prompt = "> "
		}
		line = prompt + line
		if in.Focused && offset+i == cursorRow {
			if cursorCol >= lipgloss.Width(line) {
				gap := strings.Repeat(" ", cursorCol-lipgloss.Width(line))
				line = userStyle.Render(line) + gap + accentStyle.Render("█")
			} else {
				before := ansi.Truncate(line, cursorCol, "")
				after := ansi.TruncateLeft(line, cursorCol+1, "")
				line = userStyle.Render(before) + accentStyle.Render("█") + userStyle.Render(after)
			}
		} else {
			line = userStyle.Render(line)
		}
		lines[i] = line
	}
	if !in.Focused && in.Text == "" && in.Placeholder != "" {
		lines[0] = dimStyle.Render(in.Placeholder)
	}
	mark := Thumb(visibleRows, len(textRows), len(visible), offset)
	status := ""
	if in.Status != "" || in.Age != "" || in.Queued {
		status = inputStatus(in.Status, in.Age, in.Queued)
	}
	panel := Panel{Title: "message", Focused: in.Focused, Lines: lines, VMark: mark}.View(width, panelHeight)
	rows := strings.Split(panel, "\n")
	for row := max(mark.Start, 0); row < min(mark.Start+mark.Size, panelHeight-2); row++ {
		rows[row+1] = ansi.Truncate(rows[row+1], width-1, "") + dimStyle.Render("┃")
	}
	if status != "" {
		start := lipgloss.Width("╭─")
		title := titleStyle.Render(" message ")
		end := start + lipgloss.Width(" message ")
		if width >= end+1 {
			status = ansi.Truncate(status, width-1-end, "")
			rows[0] = ansi.Truncate(rows[0], start, "") + title + status + ansi.TruncateLeft(rows[0], end+lipgloss.Width(status), "")
		}
	}
	return strings.Join(rows, "\n")
}

func inputStatus(state, age string, queued bool) string {
	glyph, style := "○", dimStyle
	switch state {
	case "working":
		glyph, style = "●", lipgloss.NewStyle().Foreground(Working)
	case "blocked":
		glyph, style = "▲", accentStyle
	}
	if state == "" {
		state = "none"
	}
	status := dimStyle.Render("· ") + style.Render(glyph) + lipgloss.NewStyle().Render(" "+state)
	if age != "" {
		status += lipgloss.NewStyle().Render(" " + age)
	}
	if queued {
		status += " " + dimStyle.Render("· queued ")
	} else {
		status += " "
	}
	return status
}

func inputTextLines(text string, width int) []string {
	return wordWrap(text, width)
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
