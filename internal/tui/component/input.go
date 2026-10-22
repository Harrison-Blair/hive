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
	Text      string
	Focused   bool
	Height    int
	CursorRow int
	CursorCol int
	HasCursor bool
	Status    string
	Age       string
	Queued    bool
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
	cursorRow, cursorCol := in.CursorRow, in.CursorCol
	if in.Focused && !in.HasCursor {
		cursorRow = len(textRows) - 1
		cursorCol = 2 + lipgloss.Width(textRows[cursorRow])
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
