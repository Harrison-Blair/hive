package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// ChatKind is the kind of a chat message.
type ChatKind int

// The kinds of chat message (decision 15).
const (
	ChatUser  ChatKind = iota // text of the user, in user
	ChatAgent                 // an agent reply, rendered as markdown
	ChatEvent                 // an event, folded to one dim line
)

// ChatMessage is one message of the chat feed.
type ChatMessage struct {
	Time string
	Name string
	Kind ChatKind
	Text string
	Open bool
}

// The gutter of the chat feed from LongFormWidth (decision 30): the time,
// then the name in 13 cells, then the text from col 21 of a line. A line
// starts at the panel padding, so the text starts 23 cells from the left
// edge of the panel.
const (
	chatNameCol   = 7
	chatNameWidth = 13
	chatTextCol   = 21
)

var infoStyle = lipgloss.NewStyle().Foreground(Info)

// ChatLines gives the lines of the chat feed for a feed panel of width
// cells (decisions 15, 30, 31). From LongFormWidth the time and the name
// are in a gutter and the text wraps at width-26 cells. Below it the time and
// the name use their own line and the text wraps at the inner width of the
// panel. A blank line separates two messages.
func ChatLines(msgs []ChatMessage, width int) ([]string, error) {
	var out []string
	for _, m := range msgs {
		lines, err := chatMessage(m, width)
		if err != nil {
			return nil, err
		}
		if len(out) > 0 && ansi.Strip(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, lines...)
	}
	return out, nil
}

func chatMessage(m ChatMessage, width int) ([]string, error) {
	nameStyle := infoStyle.Bold(true)
	if m.Kind == ChatUser {
		nameStyle = userStyle.Bold(true)
	}
	wide := width >= LongFormWidth
	wrap := width - 4
	if wide {
		wrap = width - 26
	}

	if m.Kind == ChatEvent {
		marker := "▸"
		if m.Open {
			marker = "▾"
		}
		first, _, _ := strings.Cut(m.Text, "\n")
		if !m.Open {
			if wide {
				return []string{dimStyle.Render(m.Time) + "  " + dimStyle.Render(fitCell(marker+" "+m.Name, chatNameWidth)) +
					" " + dimStyle.Render(fitCell(first, wrap))}, nil
			}
			line := ansi.Truncate(m.Time+" "+marker+" "+m.Name+" "+first, wrap, "…")
			return []string{dimStyle.Render(fitCell(line, wrap))}, nil
		}
		body := wordWrap(m.Text, wrap)
		if wide {
			lines := []string{dimStyle.Render(m.Time) + "  " + dimStyle.Render(fitCell(marker+" "+m.Name, chatNameWidth)) + " " + dimStyle.Render(body[0])}
			for _, line := range body[1:] {
				lines = append(lines, strings.Repeat(" ", chatTextCol)+dimStyle.Render(line))
			}
			return lines, nil
		}
		lines := []string{dimStyle.Render(m.Time + " " + marker + " " + m.Name)}
		for _, line := range body {
			lines = append(lines, dimStyle.Render(line))
		}
		return lines, nil
	}

	var body []string
	if m.Kind == ChatAgent {
		var err error
		if body, err = RenderMarkdown(m.Text, wrap); err != nil {
			return nil, err
		}
		body = compactHeadingListGaps(body)
	} else {
		for _, l := range wordWrap(m.Text, wrap) {
			body = append(body, userStyle.Render(l))
		}
	}

	if !wide {
		name := ansi.Truncate(m.Name, max(wrap-lipgloss.Width(m.Time)-1, 0), "…")
		head := dimStyle.Render(m.Time) + " " + nameStyle.Render(name)
		return append([]string{head}, body...), nil
	}
	gutter := dimStyle.Render(m.Time) + "  " + nameStyle.Render(fitCell(m.Name, chatNameWidth)) + " "
	indent := strings.Repeat(" ", chatTextCol)
	for i, l := range body {
		if i == 0 {
			body[i] = gutter + l
		} else if ansi.Strip(l) != "" {
			body[i] = indent + l
		}
	}
	if len(body) == 0 {
		body = []string{gutter}
	}
	return body, nil
}

// compactHeadingListGaps removes the blank row that Glamour puts between a
// bold section label and its list.
func compactHeadingListGaps(lines []string) []string {
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		if ansi.Strip(line) == "" && len(out) > 0 && i+1 < len(lines) &&
			strings.Contains(out[len(out)-1], "\x1b[1m") && strings.HasPrefix(ansi.Strip(lines[i+1]), "• ") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// wordWrap breaks s at spaces into lines of at most width cells, as the
// board script does. A word longer than width is cut into parts.
func wordWrap(s string, width int) []string {
	var lines []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		for _, w := range strings.Split(para, " ") {
			switch {
			case cur == "":
				cur = w
			case lipgloss.Width(cur+" "+w) <= width:
				cur += " " + w
			default:
				lines = append(lines, cur)
				cur = w
			}
			for lipgloss.Width(cur) > width {
				lines = append(lines, ansi.Truncate(cur, width, ""))
				cur = ansi.TruncateLeft(cur, width, "")
			}
		}
		lines = append(lines, cur)
	}
	return lines
}
