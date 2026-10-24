package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// ChatKind is the kind of a chat message.
type ChatKind int

// The kinds of chat message of golden-frames chat.md rules 2, 3, and 6.
const (
	ChatUser  ChatKind = iota // text of the user, in user
	ChatAgent                 // an agent reply, rendered as markdown
	ChatEvent                 // an event, folded to one dim line
	ChatAsk                   // an ask line, one line that the caller styles
)

// ChatMessage is one message of the transcript. Note is styled text after
// the time on the first line, such as the delivery of a user message. The
// Text of a ChatAsk message is styled by the caller.
type ChatMessage struct {
	Time string
	Name string
	Kind ChatKind
	Text string
	Note string
	Open bool
}

var (
	infoStyle    = lipgloss.NewStyle().Foreground(Info)
	dimBoldStyle = dimStyle.Bold(true)
)

// ChatLines gives the rows of the transcript at width cells (golden-frames
// chat.md rules 1 to 6). Each row starts at column 1, and the text wraps at
// column width-2. One blank row separates two messages. Two folded events, or
// two ask lines, have no blank row between them.
func ChatLines(msgs []ChatMessage, width int) ([]string, error) {
	var out []string
	for i, m := range msgs {
		lines, err := chatMessage(m, width)
		if err != nil {
			return nil, err
		}
		if i > 0 && !chatJoined(msgs[i-1], m) {
			out = append(out, "")
		}
		out = append(out, lines...)
	}
	return out, nil
}

// ChatStarts gives the first row of each message in the rows of ChatLines.
func ChatStarts(msgs []ChatMessage, width int) ([]int, error) {
	starts := make([]int, 0, len(msgs))
	row := 0
	for i, m := range msgs {
		lines, err := chatMessage(m, width)
		if err != nil {
			return nil, err
		}
		if i > 0 && !chatJoined(msgs[i-1], m) {
			row++
		}
		starts = append(starts, row)
		row += len(lines)
	}
	return starts, nil
}

func chatJoined(prev, next ChatMessage) bool {
	folded := func(m ChatMessage) bool { return m.Kind == ChatEvent && !m.Open }
	return folded(prev) && folded(next) || prev.Kind == ChatAsk && next.Kind == ChatAsk
}

func chatMessage(m ChatMessage, width int) ([]string, error) {
	wrap := max(width-2, 1)
	switch m.Kind {
	case ChatEvent:
		marker := "▸"
		if m.Open {
			marker = "▾"
		}
		head := " " + dimStyle.Render(marker) + " " + dimBoldStyle.Render(m.Name) + " "
		if !m.Open {
			first, _, _ := strings.Cut(m.Text, "\n")
			rest := max(width-3-lipgloss.Width(head), 0)
			return []string{head + dimStyle.Render(fitCell(ansi.Truncate(m.Time+"  "+first, rest, "…"), rest))}, nil
		}
		lines := []string{head + dimStyle.Render(m.Time)}
		for _, line := range wordWrap(m.Text, wrap) {
			lines = append(lines, " "+dimStyle.Render(fitCell(line, width-3)))
		}
		return lines, nil
	case ChatAsk:
		head := " " + titleStyle.Render("?") + " " + titleStyle.Render(m.Name) + " "
		return []string{ansi.Truncate(head+dimStyle.Render(m.Time+"  ")+m.Text, width-1, "…")}, nil
	}
	marker, nameStyle := "●", infoStyle.Bold(true)
	if m.Kind == ChatUser {
		marker, nameStyle = "›", userStyle.Bold(true)
	}
	head := " " + nameStyle.Render(marker) + " " + nameStyle.Render(m.Name) + " " + dimStyle.Render(m.Time)
	if m.Note != "" {
		head += m.Note
	}
	lines := []string{head}
	if m.Kind == ChatAgent {
		body, err := RenderMarkdown(m.Text, wrap)
		if err != nil {
			return nil, err
		}
		for _, line := range compactHeadingGaps(body) {
			lines = append(lines, " "+line)
		}
		return lines, nil
	}
	for _, line := range wordWrap(m.Text, wrap) {
		lines = append(lines, " "+userStyle.Render(line))
	}
	return lines, nil
}

// compactHeadingGaps removes the blank row that Glamour puts between a
// heading and the list or the code block below it, as the boards draw it.
func compactHeadingGaps(lines []string) []string {
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		if ansi.Strip(line) == "" && len(out) > 0 && i+1 < len(lines) && isHeading(out[len(out)-1]) &&
			(strings.HasPrefix(ansi.Strip(lines[i+1]), "• ") || strings.HasPrefix(ansi.Strip(lines[i+1]), "  ")) {
			continue
		}
		out = append(out, line)
	}
	return out
}

// isHeading reports that line starts with the style of a heading.
func isHeading(line string) bool {
	return headingPrefix != "" && strings.HasPrefix(line, headingPrefix)
}

// headingPrefix is the SGR sequence that starts a rendered heading.
var headingPrefix = func() string {
	lines, err := RenderMarkdown("# x", 10)
	if err != nil || len(lines) == 0 {
		return ""
	}
	prefix, _, _ := strings.Cut(lines[0], "x")
	return prefix
}()

// wordWrap breaks s at spaces into lines of at most width cells. A word
// longer than width is cut into parts.
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
