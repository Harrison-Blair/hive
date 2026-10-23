package tui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// The card layout of decision 93: the columns of an option row in the
// card text, from the left padding cell.
const (
	askLabelCol     = 8
	askLabelWidth   = 27
	askDescCol      = 36
	askPreviewCol   = 42
	askDescRightPad = 2
)

var (
	askBoldDim = dimStyle.Bold(true)
)

// askRoute is the agent that gets the answer of an ask: the live asker, or
// the parent when the asker is not live.
type askRoute struct {
	stopped bool
	target  string
}

// askCard is the UI state of the card of one open ask.
type askCard struct {
	question int
	cursor   []int
	picks    [][]bool
	notes    []string
	others   []string
	noteOpen bool
	editing  string
}

// askRow maps a body row of the dock to an option of the focused card.
// Option is -1 for a row that is not an option row.
type askRow struct {
	question int
	option   int
}

type askDoneMsg struct {
	action   string
	id       string
	replies  []ask.Reply
	envelope output.Envelope
}

type askDeliveredMsg struct{ err error }

// buildAskRoutes gives the route of each ask.
func buildAskRoutes(asks []ask.Record, agents []records.Agent) map[string]askRoute {
	routes := make(map[string]askRoute, len(asks))
	for _, record := range asks {
		target, stopped, ok := ask.Route(record, agents)
		route := askRoute{stopped: stopped, target: target.Name}
		if !ok {
			route.target = record.Parent
		}
		routes[record.ID] = route
	}
	return routes
}

// openAsks gives the open asks, oldest first.
func (m *Model) openAsks() []ask.Record {
	var open []ask.Record
	for _, record := range m.chatData.asks {
		if record.State == ask.StateOpen {
			open = append(open, record)
		}
	}
	return open
}

// focusedAsk gives the open ask of the card in view and its place.
func (m *Model) focusedAsk() (ask.Record, int, bool) {
	open := m.openAsks()
	if len(open) == 0 {
		return ask.Record{}, 0, false
	}
	for index, record := range open {
		if record.ID == m.askFocusID {
			return record, index, true
		}
	}
	return open[0], 0, true
}

// askCardOf gives the UI state of the card of record.
func (m *Model) askCardOf(record ask.Record) *askCard {
	if m.askCards == nil {
		m.askCards = map[string]*askCard{}
	}
	card, ok := m.askCards[record.ID]
	if ok && len(card.cursor) == len(record.Questions) {
		return card
	}
	n := len(record.Questions)
	card = &askCard{cursor: make([]int, n), picks: make([][]bool, n), notes: make([]string, n), others: make([]string, n)}
	for index, question := range record.Questions {
		card.picks[index] = make([]bool, len(question.Options)+1)
	}
	m.askCards[record.ID] = card
	return card
}

// askFocused reports that the card has the focus: Chat is in normal mode,
// an ask is open, and the user did not give the feed the focus.
func (m *Model) askFocused() bool {
	_, _, ok := m.focusedAsk()
	return ok && !m.chatInputActive && !m.askFeedFocus
}

// askFocusState gives the focus state of the card: ask, note, or other.
func (m *Model) askFocusState() string {
	record, _, _ := m.focusedAsk()
	switch m.askCardOf(record).editing {
	case "note":
		return "note"
	case "other":
		return "other"
	}
	return "ask"
}

// updateAskKey handles a key while the card has the focus. It reports false
// for a key that the frame handles.
func (m *Model) updateAskKey(press tea.KeyPressMsg) (tea.Cmd, bool) {
	record, _, _ := m.focusedAsk()
	card := m.askCardOf(record)
	question := record.Questions[card.question]
	other := len(question.Options)
	key := press.String()
	m.invalidateView()
	if card.editing != "" {
		text := &card.notes[card.question]
		if card.editing == "other" {
			text = &card.others[card.question]
		}
		switch key {
		case "enter":
			if card.editing == "other" && strings.TrimSpace(*text) == "" {
				return nil, true
			}
			if card.editing == "other" {
				card.setPick(other, true, question.MultiSelect)
			}
			card.editing = ""
			return m.askEnter(record, card), true
		case "tab":
			if card.editing == "note" {
				*text = ""
				card.noteOpen = false
				card.editing = ""
			}
		case "ctrl+]", "esc":
			card.editing = ""
		case "up", "down":
			if card.editing == "note" {
				card.move(key, other)
			}
		case "backspace":
			runes := []rune(*text)
			if len(runes) > 0 {
				*text = string(runes[:len(runes)-1])
			}
		default:
			if press.Text != "" {
				*text += safeChatText(press.Text)
			}
		}
		return nil, true
	}
	switch key {
	case "up", "down":
		card.move(key, other)
	case "space":
		on := !question.MultiSelect || !card.picks[card.question][card.cursor[card.question]]
		card.setPick(card.cursor[card.question], on, question.MultiSelect)
		if on && card.cursor[card.question] == other {
			card.editing = "other"
		}
	case "enter":
		cursor := card.cursor[card.question]
		if !question.MultiSelect || !slices.Contains(card.picks[card.question], true) {
			card.setPick(cursor, true, question.MultiSelect)
		}
		return m.askEnter(record, card), true
	case "tab":
		card.noteOpen = true
		card.editing = "note"
	case "x":
		return m.closeAsk(record, "dismiss", nil), true
	case "J", "K":
		return m.moveAskCard(key), true
	case "pgup", "pgdown":
		return m.scrollChatFeed(press), true
	case "i":
		return m.startChatInput(), true
	case "ctrl+o":
		return m.updateChat(press), true
	default:
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' && int(key[0]-'1') <= other {
			return m.askPick(record, card, int(key[0]-'1')), true
		}
		return nil, false
	}
	return nil, true
}

// askPick picks option index of the current question, as a digit or a
// click does. In a multi-select question it marks the option.
func (m *Model) askPick(record ask.Record, card *askCard, index int) tea.Cmd {
	question := record.Questions[card.question]
	card.cursor[card.question] = index
	m.invalidateView()
	if question.MultiSelect {
		on := !card.picks[card.question][index]
		card.setPick(index, on, true)
		if on && index == len(question.Options) {
			card.editing = "other"
		}
		return nil
	}
	card.setPick(index, true, false)
	return m.askEnter(record, card)
}

// askEnter moves to the next question, or sends the answer on the last
// question. A picked Other row first opens its text line.
func (m *Model) askEnter(record ask.Record, card *askCard) tea.Cmd {
	q := card.question
	other := len(record.Questions[q].Options)
	if card.picks[q][other] && strings.TrimSpace(card.others[q]) == "" {
		card.editing = "other"
		return nil
	}
	card.editing, card.noteOpen = "", false
	if q < len(record.Questions)-1 {
		card.question++
		return nil
	}
	replies := make([]ask.Reply, len(record.Questions))
	for index := range record.Questions {
		reply := ask.Reply{Note: card.notes[index]}
		for option, picked := range card.picks[index] {
			if picked {
				reply.Picks = append(reply.Picks, option+1)
			}
		}
		if card.picks[index][len(record.Questions[index].Options)] {
			reply.Other = card.others[index]
		}
		replies[index] = reply
	}
	return m.closeAsk(record, "answer", replies)
}

func (c *askCard) move(key string, last int) {
	step := 1
	if key == "up" {
		step = -1
	}
	c.cursor[c.question] = min(max(c.cursor[c.question]+step, 0), last)
}

// setPick sets the mark of option index. A single-select question keeps one
// pick.
func (c *askCard) setPick(index int, on, multi bool) {
	picks := c.picks[c.question]
	if !multi {
		clear(picks)
	}
	picks[index] = on
}

// moveAskCard gives the focus to the next card with J or the earlier card
// with K.
func (m *Model) moveAskCard(key string) tea.Cmd {
	open := m.openAsks()
	_, index, _ := m.focusedAsk()
	if key == "J" {
		index = min(index+1, len(open)-1)
	} else {
		index = max(index-1, 0)
	}
	m.askFocusID = open[index].ID
	m.rebuildChat()
	return m.logAction("ask card", slogString("ask_id", m.askFocusID))
}

// closeAsk sends the answer or the dismissal of record.
func (m *Model) closeAsk(record ask.Record, action string, replies []ask.Reply) tea.Cmd {
	if m.demo != nil {
		return m.demoBlocked("ask " + action)
	}
	// The answer has the authority of the user, as chat has (rule 2).
	if m.chatCallerBlocked {
		m.invalidateView()
		return nil
	}
	if m.askBusy {
		return nil
	}
	m.askBusy = true
	answer, dismiss, home := m.askAnswer, m.askDismiss, m.home
	return batch(m.logAction("ask "+action, slogString("ask_id", record.ID)), func() tea.Msg {
		var envelope output.Envelope
		if action == "answer" {
			envelope = answer(context.Background(), home, record.ID, replies)
		} else {
			envelope = dismiss(context.Background(), home, record.ID)
		}
		return askDoneMsg{action: action, id: record.ID, replies: replies, envelope: envelope}
	})
}

// updateAskDone closes the card after the store keeps the answer. A failed
// delivery keeps the answer, and the delivery tries again on a later tick.
func (m *Model) updateAskDone(msg askDoneMsg) tea.Cmd {
	m.askBusy = false
	if msg.envelope.Outcome == "failed" {
		reason := fmt.Sprint(msg.envelope.Error)
		m.chatErr = boundedError(reason)
		return m.logError("ask "+msg.action, errors.New(reason))
	}
	state := ask.StateAnswered
	if msg.action == "dismiss" {
		state = ask.StateDismissed
	}
	for index := range m.chatData.asks {
		if m.chatData.asks[index].ID == msg.id {
			m.chatData.asks[index].State = state
			m.chatData.asks[index].Replies = msg.replies
			m.chatData.asks[index].ClosedAt = m.now().UTC().Format(time.RFC3339Nano)
		}
	}
	delete(m.askCards, msg.id)
	m.header.asks = len(m.openAsks())
	m.rebuildChat()
	m.invalidateView()
	cmds := []tea.Cmd{m.startHeaderLoad(), m.startCurrentScreenLoad()}
	if msg.envelope.Outcome != "ok" {
		cmds = append(cmds, m.logError("ask delivery", fmt.Errorf("%v", msg.envelope.Error)))
	}
	return batch(cmds...)
}

// startAskDelivery tries again to send the answers that wait for delivery.
func (m *Model) startAskDelivery() tea.Cmd {
	if m.demo != nil || m.askDelivering || m.header.askUndelivered == 0 {
		return nil
	}
	m.askDelivering = true
	deliver, home := m.askDeliver, m.home
	return func() tea.Msg { return askDeliveredMsg{err: deliver(context.Background(), home)} }
}

// askHints gives the footer hints of the card.
func (m *Model) askHints(focus string, wide bool) []string {
	switch focus {
	case "note":
		return []string{"type a note", "enter send with note", "tab close note", "ctrl+] leave note", "↑↓ option"}
	case "other":
		return []string{"type the answer", "enter send", "ctrl+] leave"}
	}
	if !wide {
		return []string{"↑↓ move", "enter send", "tab note", "x dismiss", "J/K card"}
	}
	record, _, _ := m.focusedAsk()
	card := m.askCardOf(record)
	pick := fmt.Sprintf("1..%d pick", len(record.Questions[card.question].Options)+1)
	if m.chatData.routes[record.ID].stopped {
		return []string{"↑↓ move", pick, "enter send to parent", "tab note", "x dismiss", "J/K next/prev card"}
	}
	return []string{"↑↓ move", pick, "enter send", "tab note", "x dismiss", "J/K next/prev card", "wheel/pgup scroll feed", "i type"}
}

// askFeedText gives the styled text of the feed line of an ask.
func (m *Model) askFeedText(record ask.Record, wide bool) string {
	text := "? asked · " + ask.Headers(record)
	if m.chatData.routes[record.ID].stopped && record.State == ask.StateOpen {
		text += " · asker stopped"
	}
	if record.State == ask.StateOpen {
		focused, index, _ := m.focusedAsk()
		switch {
		case focused.ID == record.ID && wide:
			text += " · open below"
		case focused.ID == record.ID:
			text += " · below"
		default:
			for place, open := range m.openAsks() {
				if open.ID == record.ID {
					index = place
				}
			}
			text += fmt.Sprintf(" · waits · %d of %d", index+1, len(m.openAsks()))
		}
	}
	return titleStyle.UnsetBold().Render(text)
}

// askAnswerText gives the styled text of the feed line of the answer.
func askAnswerText(record ask.Record) string {
	if record.State == ask.StateDismissed {
		return "dismissed " + record.Asker + " · " + ask.Headers(record)
	}
	var text strings.Builder
	text.WriteString("answered " + record.Asker + " ·")
	for index, question := range record.Questions {
		if index >= len(record.Replies) {
			break
		}
		if index > 0 {
			text.WriteString(";")
		}
		reply := record.Replies[index]
		picks := make([]string, 0, len(reply.Picks))
		for _, pick := range reply.Picks {
			if pick <= len(question.Options) {
				picks = append(picks, fmt.Sprintf("%d %s", pick, question.Options[pick-1].Label))
			} else {
				picks = append(picks, reply.Other)
			}
		}
		text.WriteString(" " + question.Header + " " + titleStyle.Render("→ "+strings.Join(picks, ", ")))
		if reply.Note != "" {
			text.WriteString(" · note: " + reply.Note)
		}
	}
	return text.String()
}

// askMessages gives the feed lines of the asks: the ask, and the answer or
// the dismissal.
func (m *Model) askMessages(wide bool, zone *time.Location) []chatItem {
	var items []chatItem
	for _, record := range m.chatData.asks {
		if at, err := time.Parse(time.RFC3339Nano, record.CreatedAt); err == nil {
			items = append(items, chatItem{at, component.ChatMessage{
				Time: at.In(zone).Format("15:04"), Name: record.Asker, Kind: component.ChatAgent,
				Text: m.askFeedText(record, wide), Styled: true,
			}})
		}
		if record.State == ask.StateOpen {
			continue
		}
		if at, err := time.Parse(time.RFC3339Nano, record.ClosedAt); err == nil {
			items = append(items, chatItem{at, component.ChatMessage{
				Time: at.In(zone).Format("15:04"), Name: "you", Kind: component.ChatUser,
				Text: askAnswerText(record), Styled: true,
			}})
		}
	}
	return items
}

// askDockView draws the card of the focused ask at width cells and at most
// limit rows. It records the option of each row for a click.
func (m *Model) askDockView(width, limit int) string {
	record, index, ok := m.focusedAsk()
	m.askRows = nil
	if !ok {
		return ""
	}
	card := m.askCardOf(record)
	focused := m.askFocused()
	room := width - 4
	wide := width >= component.LongFormWidth
	lines, rows, current := askCardLines(record, card, room)

	visible := max(min(len(lines), limit-2), 1)
	offset := max(current[1]-visible+1, 0)
	offset = min(offset, current[0])
	end := min(offset+visible, len(lines))
	m.askRows = rows[offset:end]
	panel := component.Panel{
		Title: "? " + record.Asker + " asks · " + ask.Headers(record), Focused: focused, Heavy: true,
		Lines: lines[offset:end], VMark: component.Thumb(visible, len(lines), visible, offset),
	}
	out := strings.Split(panel.View(width, visible+2), "\n")

	route := m.chatData.routes[record.ID]
	note := dimStyle.Render(fmt.Sprintf(" %d of %d ", index+1, len(m.openAsks())))
	if wide {
		next := ""
		if index+1 < len(m.openAsks()) {
			next = " · J next"
		}
		note = dimStyle.Render(fmt.Sprintf(" open · %d of %d%s ", index+1, len(m.openAsks()), next))
	}
	if route.stopped {
		note = titleStyle.Render(" asker stopped · answer goes to " + route.target + " ")
	}
	out[0] = overlayCells(out[0], note, width-2-lipgloss.Width(note))
	out[len(out)-1] = overlayCells(out[len(out)-1], dimStyle.Render(" "+askCardHint(record, card, wide, route)+" "), 2)
	return strings.Join(out, "\n")
}

// askCardHint gives the hint in the bottom border of the card.
func askCardHint(record ask.Record, card *askCard, wide bool, route askRoute) string {
	question := record.Questions[card.question]
	multi := slices.ContainsFunc(record.Questions, func(q ask.Question) bool { return q.MultiSelect })
	switch {
	case !wide:
		return "enter send · tab note"
	case route.stopped:
		return "enter sends to " + route.target + " · x dismiss"
	case askHasPreview(question):
		return "↑↓ move · the preview follows the option · enter send"
	case len(record.Questions) > 1:
		hint := "tab note · enter next question · enter on last sends"
		if multi {
			hint = "space mark · " + hint
		}
		return hint
	case question.MultiSelect:
		return fmt.Sprintf("↑↓ move · 1..%d mark · space mark · enter send · tab note · x dismiss · J/K card", len(question.Options)+1)
	}
	return fmt.Sprintf("↑↓ move · 1..%d pick · enter send · tab note · x dismiss · J/K card", len(question.Options)+1)
}

func askHasPreview(question ask.Question) bool {
	return slices.ContainsFunc(question.Options, func(option ask.Option) bool { return option.Preview != "" })
}

// askCardLines gives the body lines of the card at room cells, the option
// of each line, and the first and the last line of the current question.
func askCardLines(record ask.Record, card *askCard, room int) ([]string, []askRow, [2]int) {
	var lines []string
	var rows []askRow
	var current [2]int
	add := func(line string, question, option int) {
		lines = append(lines, line)
		rows = append(rows, askRow{question, option})
	}
	for qi, question := range record.Questions {
		if qi > 0 {
			add("", qi, -1)
		}
		if qi == card.question {
			current[0] = len(lines)
		}
		text := question.Question
		if len(record.Questions) > 1 {
			text = fmt.Sprintf("%d. %s", qi+1, text)
		}
		for _, line := range strings.Split(lipgloss.NewStyle().Width(room).Render(safeChatText(text)), "\n") {
			add(headerStyle.Render(strings.TrimRight(line, " ")), qi, -1)
		}
		add("", qi, -1)
		first := len(lines)
		preview := qi == card.question && askHasPreview(question)
		labelWidth := askLabelWidth
		if preview {
			labelWidth = askPreviewCol - askLabelCol - 1
		}
		for oi := range len(question.Options) + 1 {
			add(askOptionLine(question, card, qi, oi, labelWidth, room, preview), qi, oi)
		}
		if card.notes[qi] != "" || (card.noteOpen && qi == card.question) {
			add(askTextLine("note", card.notes[qi], card.editing == "note" && qi == card.question), qi, -1)
		}
		if card.picks[qi][len(question.Options)] || (card.editing == "other" && qi == card.question) {
			add(askTextLine("other", card.others[qi], card.editing == "other" && qi == card.question), qi, -1)
		}
		if preview {
			box := askPreviewBox(question, card.cursor[qi], room-askPreviewCol)
			for index, line := range box {
				row := first + index
				if row >= len(lines) {
					add("", qi, -1)
				}
				lines[row] = padCell(ansi.Truncate(lines[row], askPreviewCol, ""), askPreviewCol) + line
			}
		}
		if qi == card.question {
			current[1] = len(lines) - 1
		}
	}
	return lines, rows, current
}

// askOptionLine gives one option row: the current mark ›, the number, the
// pick mark, the label, and the description.
func askOptionLine(question ask.Question, card *askCard, qi, oi, labelWidth, room int, preview bool) string {
	isCurrent := qi == card.question && card.cursor[qi] == oi
	marker, number := titleStyle.Render(" "), askBoldDim.Render(fmt.Sprint(oi+1))
	if isCurrent {
		marker, number = titleStyle.Render("›"), titleStyle.Render(fmt.Sprint(oi+1))
	}
	mark := dimStyle.Render("[ ]")
	if card.picks[qi][oi] {
		mark = titleStyle.Render("[x]")
	}
	label, description := ask.OtherLabel, ""
	if oi < len(question.Options) {
		label, description = safeChatText(question.Options[oi].Label), safeChatText(question.Options[oi].Description)
	}
	label = ansi.Truncate(label, labelWidth, "…")
	gap := strings.Repeat(" ", labelWidth-lipgloss.Width(label)+1)
	if isCurrent {
		label = titleStyle.Render(label)
	}
	line := marker + " " + number + " " + mark + " " + label + gap
	if !preview && description != "" {
		line += dimStyle.Render(ansi.Truncate(description, max(room-askDescCol-askDescRightPad, 0), "…"))
	}
	return line
}

// askTextLine gives the line of a note or of an Other answer. The cursor
// █ shows while the line takes the keys.
func askTextLine(name, text string, editing bool) string {
	line := "    " + titleStyle.Render(name+" ›") + " " + text
	if editing {
		line += titleStyle.UnsetBold().Render("█")
	}
	return line
}

// askPreviewBox gives the rows of the preview box of the current option at
// width cells. Its height holds the longest preview of the question.
func askPreviewBox(question ask.Question, cursor, width int) []string {
	height := 0
	for _, option := range question.Options {
		height = max(height, len(strings.Split(option.Preview, "\n")))
	}
	var text []string
	if cursor < len(question.Options) && question.Options[cursor].Preview != "" {
		text = strings.Split(safeChatText(question.Options[cursor].Preview), "\n")
	}
	inner := width - 2
	title := dimStyle.Render(fmt.Sprintf(" preview · %d ", cursor+1))
	edge := lipgloss.NewStyle().Foreground(component.Border)
	box := []string{edge.Render("╭─") + title + edge.Render(strings.Repeat("─", max(inner-1-lipgloss.Width(title), 0))+"╮")}
	for index := range height {
		line := ""
		if index < len(text) {
			line = text[index]
		}
		box = append(box, edge.Render("│")+" "+fitText(line, inner-2)+" "+edge.Render("│"))
	}
	return append(box, edge.Render("╰"+strings.Repeat("─", inner)+"╯"))
}

// overlayCells draws text over row from cell col.
func overlayCells(row, text string, col int) string {
	end := col + lipgloss.Width(text)
	if col < 0 || end > lipgloss.Width(row) {
		return row
	}
	return ansi.Truncate(row, col, "") + text + ansi.TruncateLeft(row, end, "")
}

// askDockHeight gives the rows of the dock in the last render.
func (m *Model) askDockHeight() int {
	if len(m.askRows) == 0 {
		return 0
	}
	return len(m.askRows) + 2
}

// clickAsk handles a click on body row of the dock. A click on an option
// row picks the option.
func (m *Model) clickAsk(row int) tea.Cmd {
	m.askFeedFocus = false
	cmd := m.stopChatInput()
	record, _, ok := m.focusedAsk()
	if !ok || row < 0 || row >= len(m.askRows) {
		return cmd
	}
	card := m.askCardOf(record)
	ref := m.askRows[row]
	if ref.option < 0 || ref.question != card.question {
		return cmd
	}
	card.editing = ""
	return batch(cmd, m.askPick(record, card, ref.option))
}

// sortAsks puts the asks in the order of their IDs.
func sortAsks(asks []ask.Record) {
	sort.SliceStable(asks, func(i, j int) bool {
		var a, b int
		_, _ = fmt.Sscanf(asks[i].ID, "a%d", &a)
		_, _ = fmt.Sscanf(asks[j].ID, "a%d", &b)
		return a < b
	})
}
