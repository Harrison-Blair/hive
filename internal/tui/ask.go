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

// The card layout of golden-frames chat.md rule 15 and the boards: the
// columns of an option row from the first card column, and the preview box.
// The preview box shows from askPreviewRows card rows (rule 16).
const (
	askLabelWidth   = 26
	askDescCol      = 35
	askDescRightPad = 2
	askPreviewCol   = 78
	askPreviewWidth = 38
	askPreviewRows  = 8
)

var (
	askBoldDim    = dimStyle.Bold(true)
	accentStyle   = lipgloss.NewStyle().Foreground(component.Accent)
	userTextStyle = lipgloss.NewStyle().Foreground(component.User)
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
	// review shows the review of all answers before the send. reviewRow
	// is the highlighted question, or the number of questions for Send.
	// fromReview returns the edit of a question to the review. A
	// highlighted row taller than the dock scrolls: reviewScroll is its
	// first visible line, and the last render sets reviewMax and
	// reviewStep.
	review       bool
	reviewRow    int
	fromReview   bool
	reviewScroll int
	reviewMax    int
	reviewStep   int
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

// askFocused reports that the card has the focus: an ask is open and the
// input line does not have the focus (chat.md rule 17).
func (m *Model) askFocused() bool {
	_, _, ok := m.focusedAsk()
	return ok && !m.chatInputActive
}

// askFocusState gives the focus state of the card: ask, note, other, or
// review.
func (m *Model) askFocusState() string {
	record, _, _ := m.focusedAsk()
	card := m.askCardOf(record)
	switch {
	case card.review:
		return "review"
	case card.editing == "note":
		return "note"
	case card.editing == "other":
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
	if card.review {
		return m.updateAskReviewKey(press, record, card)
	}
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
			} else if !slices.Contains(card.picks[card.question], true) {
				card.setPick(card.cursor[card.question], true, question.MultiSelect)
			}
			card.editing = ""
			return m.askEnter(record, card), true
		case "ctrl+r", "esc":
			card.editing = ""
			return m.startChatInput(), true
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
	case "n":
		card.noteOpen = true
		card.editing = "note"
	case "x":
		return m.closeAsk(record, "dismiss", nil), true
	case "J", "K":
		return m.moveAskCard(key), true
	case "pgup", "pgdown":
		return m.scrollChatFeed(press), true
	case "ctrl+r", "esc":
		return m.startChatInput(), true
	case "ctrl+o":
		return m.toggleChatEvents(), true
	case "tab":
		// chat.md rule 18: tab does nothing on the card.
	default:
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' && int(key[0]-'1') <= other {
			return m.askPick(record, card, int(key[0]-'1')), true
		}
		return nil, false
	}
	return nil, true
}

// updateAskReviewKey handles a key in the review. Only ↑↓, enter, x, J/K,
// ctrl+r, and esc act. Each other key does nothing. ↑↓ first pages through a
// highlighted row that is taller than the dock. A row gets the highlight
// at its first line.
func (m *Model) updateAskReviewKey(press tea.KeyPressMsg, record ask.Record, card *askCard) (tea.Cmd, bool) {
	key := press.String()
	switch key {
	case "up":
		switch {
		case card.reviewScroll > 0:
			card.reviewScroll = max(card.reviewScroll-card.reviewStep, 0)
		case card.reviewRow > 0:
			card.reviewRow, card.reviewMax = card.reviewRow-1, 0
		}
	case "down":
		switch {
		case card.reviewScroll < card.reviewMax:
			card.reviewScroll = min(card.reviewScroll+card.reviewStep, card.reviewMax)
		case card.reviewRow < len(record.Questions):
			card.reviewRow, card.reviewScroll, card.reviewMax = card.reviewRow+1, 0, 0
		}
	case "enter":
		if card.reviewRow == len(record.Questions) {
			return m.sendAsk(record, card), true
		}
		card.review, card.fromReview, card.question = false, true, card.reviewRow
	case "x":
		return m.closeAsk(record, "dismiss", nil), true
	case "J", "K":
		return m.moveAskCard(key), true
	case "ctrl+r", "esc":
		return m.startChatInput(), true
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

// askEnter moves to the next question, or opens the review on the last
// question and after the edit of a question from the review. A picked Other
// row first opens its text line.
func (m *Model) askEnter(record ask.Record, card *askCard) tea.Cmd {
	q := card.question
	other := len(record.Questions[q].Options)
	if card.picks[q][other] && strings.TrimSpace(card.others[q]) == "" {
		card.editing = "other"
		return nil
	}
	card.editing, card.noteOpen = "", false
	if q < len(record.Questions)-1 && !card.fromReview {
		card.question++
		return nil
	}
	card.review, card.fromReview, card.reviewRow, card.reviewScroll = true, false, len(record.Questions), 0
	return nil
}

// sendAsk sends the answer of the card.
func (m *Model) sendAsk(record ask.Record, card *askCard) tea.Cmd {
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
	var focus tea.Cmd
	if m.header.asks == 0 {
		// chat.md rule 19: the frame returns to the composer alone.
		focus = m.startChatInput()
	}
	m.rebuildChat()
	m.invalidateView()
	cmds := []tea.Cmd{focus, m.startHeaderLoad(), m.startCurrentScreenLoad()}
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

// askHints gives the keys of the focus of the card (golden-frames decision
// 33).
func (m *Model) askHints(focus string) []string {
	record, _, _ := m.focusedAsk()
	card := m.askCardOf(record)
	question := record.Questions[card.question]
	switch focus {
	case "review":
		return []string{"↑↓ question", "enter send", "x dismiss", "J/K card", "^r input"}
	case "note":
		hints := []string{"type the note", "enter keep note", "↑↓ option"}
		if slices.ContainsFunc(record.Questions, func(q ask.Question) bool { return q.MultiSelect }) {
			hints = append(hints, "space mark")
		}
		return append(hints, "^r input")
	case "other":
		return []string{"type the answer", "enter keep answer", "^r input"}
	}
	review := "enter review"
	if m.chatData.routes[record.ID].stopped {
		review = "enter review for parent"
	}
	return []string{"↑↓ move", fmt.Sprintf("1..%d pick", len(question.Options)+1), review, "n note", "x dismiss", "J/K card", "^r input", "esc input"}
}

// askLineText gives the styled text of the ask line after its time
// (chat.md rule 6).
func (m *Model) askLineText(record ask.Record) string {
	text := "asked · " + askHeaders(record)
	if record.State == ask.StateOpen {
		if m.chatData.routes[record.ID].stopped {
			text += " · asker stopped"
		}
		focused, _, _ := m.focusedAsk()
		if focused.ID == record.ID {
			text += " · in the composer, ^r to answer"
		} else {
			open := m.openAsks()
			for place, other := range open {
				if other.ID == record.ID {
					text += fmt.Sprintf(" · waits · %d of %d", place+1, len(open))
				}
			}
		}
	}
	line := dimStyle.Render(text)
	if record.LastError != "" {
		line += accentStyle.Render(" · delivery failed · retrying")
	}
	return line
}

// askAnswerText gives the text of the line from you after an answer or a
// dismiss (chat.md rule 6).
func askAnswerText(record ask.Record) string {
	if record.State == ask.StateDismissed {
		return "dismissed " + record.Asker + " · " + askHeaders(record)
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
				picks = append(picks, question.Options[pick-1].Label)
			} else {
				picks = append(picks, reply.Other)
			}
		}
		text.WriteString(" " + question.Header + " → " + strings.Join(picks, ", "))
		if reply.Note != "" {
			text.WriteString(" · note: " + reply.Note)
		}
	}
	return safeChatText(text.String())
}

// askMessages gives the transcript lines of the asks: the ask line, and the
// line from you after an answer or a dismiss.
func (m *Model) askMessages(zone *time.Location) []chatItem {
	var items []chatItem
	for _, record := range m.chatData.asks {
		if at, err := time.Parse(time.RFC3339Nano, record.CreatedAt); err == nil {
			items = append(items, chatItem{at: at, msg: component.ChatMessage{
				Time: at.In(zone).Format("15:04"), Name: record.Asker, Kind: component.ChatAsk,
				Text: m.askLineText(record),
			}})
		}
		if record.State == ask.StateOpen {
			continue
		}
		if at, err := time.Parse(time.RFC3339Nano, record.ClosedAt); err == nil {
			items = append(items, chatItem{at: at, msg: component.ChatMessage{
				Time: at.In(zone).Format("15:04"), Name: "you", Kind: component.ChatUser,
				Text: askAnswerText(record),
			}})
		}
	}
	return items
}

// askCardView gives the card of record in the composer frame at width
// cells and at most rows card rows (chat.md rules 14 to 16). It records the
// option of each visible row for a click.
func (m *Model) askCardView(record ask.Record, width, rows int) *component.ComposerCard {
	card := m.askCardOf(record)
	focused := m.askFocused()
	room := width - 3
	route := m.chatData.routes[record.ID]
	title := "? " + record.Asker + " asks · " + askHeaders(record)
	var lines []string
	var refs []askRow
	var current [2]int
	if card.review {
		title = "? " + record.Asker + " asks · confirm"
		lines, refs, current = askReviewLines(record, card, room, route, focused)
	} else {
		lines, refs, current = askCardLines(record, card, room, focused, rows >= askPreviewRows)
	}
	visible := max(min(len(lines), rows), 1)
	offset := max(current[1]-visible+1, 0)
	offset = min(offset, current[0])
	if card.review {
		// A highlighted row taller than the card scrolls with ↑↓, and its
		// first visible line has the › mark.
		card.reviewMax = max(current[1]-current[0]+1-visible, 0)
		card.reviewStep = max(visible-1, 1)
		card.reviewScroll = min(card.reviewScroll, card.reviewMax)
		if card.reviewMax > 0 {
			offset = current[0] + card.reviewScroll
			lines[offset] = askMarkerStyle(focused).Render("›") + ansi.TruncateLeft(lines[offset], 1, "")
		}
	}
	m.askRows = refs[offset:min(offset+visible, len(refs))]

	_, index, _ := m.focusedAsk()
	open := len(m.openAsks())
	note := fmt.Sprintf(" open · %d of %d ", index+1, open)
	if index+1 < open {
		note = fmt.Sprintf(" open · %d of %d · J next ", index+1, open)
	}
	if route.stopped {
		// Without room for the parent, the note keeps the state. Send names
		// the parent.
		note = ""
		for _, text := range []string{" asker stopped · answer goes to " + route.target + " ", " asker stopped · to " + route.target + " ", " asker stopped "} {
			if 2+lipgloss.Width(title)+2+1+lipgloss.Width(text)+1 <= width {
				note = text
				break
			}
		}
	}
	if note != "" {
		note = titleStyle.Render(note)
	}
	return &component.ComposerCard{Title: title, Note: note, Lines: lines, Offset: offset, Rows: visible}
}

// askMarkerStyle gives the style of the › row marker: accent bold while the
// card has the focus, else dim bold (chat.md rule 17).
func askMarkerStyle(focused bool) lipgloss.Style {
	if focused {
		return titleStyle
	}
	return askBoldDim
}

func askHasPreview(question ask.Question) bool {
	return slices.ContainsFunc(question.Options, func(option ask.Option) bool { return option.Preview != "" })
}

// askCardLines gives the card rows at room cells, the option of each row,
// and the first and the last row of the current question. box draws the
// preview box of the current option.
func askCardLines(record ask.Record, card *askCard, room int, focused, box bool) ([]string, []askRow, [2]int) {
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
			add(strings.TrimRight(line, " "), qi, -1)
		}
		add("", qi, -1)
		first := len(lines)
		preview := qi == card.question && askHasPreview(question)
		for oi := range len(question.Options) + 1 {
			add(askOptionLine(question, card, qi, oi, room, preview, box, focused), qi, oi)
		}
		if qi == card.question || card.notes[qi] != "" {
			add(askTextLine("note", card.notes[qi], card.editing == "note" && qi == card.question, "none · n adds one"), qi, -1)
		}
		if card.picks[qi][len(question.Options)] || (card.editing == "other" && qi == card.question) {
			add(askTextLine("other", card.others[qi], card.editing == "other" && qi == card.question, ""), qi, -1)
		}
		if preview && box {
			for index, line := range askPreviewBox(question, card.cursor[qi]) {
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

// askReviewLines gives the rows of the review at room cells: the answer of
// each question in full, wrapped, the Send row, and a blank row. It also
// gives the first and the last row of the highlighted row.
func askReviewLines(record ask.Record, card *askCard, room int, route askRoute, focused bool) ([]string, []askRow, [2]int) {
	var lines []string
	var rows []askRow
	var current [2]int
	marker := askMarkerStyle(focused)
	plain := lipgloss.NewStyle()
	add := func(index int, text string) {
		style, start := plain, len(lines)
		if index == card.reviewRow {
			style = titleStyle
		}
		for row, line := range wrapAskWords(styledWords(style, text), room-2) {
			mark := " "
			if row == 0 && index == card.reviewRow {
				mark = "›"
			}
			lines = append(lines, marker.Render(mark)+" "+line)
			rows = append(rows, askRow{index, -1})
		}
		if index == card.reviewRow {
			current = [2]int{start, len(lines) - 1}
		}
	}
	for qi, question := range record.Questions {
		text := safeChatText(question.Header) + " → " + strings.Join(askReviewPicks(question, card, qi), ", ")
		if card.notes[qi] != "" {
			text += " · note: " + safeChatText(card.notes[qi])
		}
		add(qi, text)
	}
	lines = append(lines, marker.Render(" "))
	rows = append(rows, askRow{len(record.Questions), -1})
	send := "Send"
	if route.stopped {
		send = "Send to " + route.target
	}
	add(len(record.Questions), send)
	lines = append(lines, "")
	rows = append(rows, askRow{len(record.Questions), -1})
	return lines, rows, current
}

// askReviewPicks gives the picked labels of question qi, with the Other
// text in place of the Other label.
func askReviewPicks(question ask.Question, card *askCard, qi int) []string {
	var picks []string
	for option, picked := range card.picks[qi] {
		switch {
		case !picked:
		case option < len(question.Options):
			picks = append(picks, safeChatText(question.Options[option].Label))
		default:
			picks = append(picks, safeChatText(card.others[qi]))
		}
	}
	return picks
}

// askWord is one word of a review line and its style. A word with
// lineFeed starts a new line before its text.
type askWord struct {
	text     string
	style    lipgloss.Style
	lineFeed bool
}

func styledWords(style lipgloss.Style, text string) []askWord {
	var words []askWord
	for index, line := range strings.Split(text, "\n") {
		for column, word := range strings.Split(line, " ") {
			words = append(words, askWord{word, style, index > 0 && column == 0})
		}
	}
	return words
}

// wrapAskWords puts the words on lines of at most width cells. A word that
// is longer than width breaks at width.
func wrapAskWords(words []askWord, width int) []string {
	width = max(width, 1)
	var lines []string
	line, used := "", 0
	for _, word := range words {
		if word.lineFeed {
			lines, line, used = append(lines, line), "", 0
		}
		size := ansi.StringWidth(word.text)
		if used > 0 && used+1+size > width {
			lines, line, used = append(lines, line), "", 0
		}
		for size > width {
			head := ansi.Truncate(word.text, width, "")
			lines = append(lines, word.style.Render(head))
			word.text = ansi.TruncateLeft(word.text, ansi.StringWidth(head), "")
			size = ansi.StringWidth(word.text)
		}
		if used > 0 {
			line += word.style.Render(" ")
			used++
		}
		line += word.style.Render(word.text)
		used += size
	}
	return append(lines, line)
}

// askOptionLine gives one option row: the › row marker of the current
// question, the number, the pick mark, the label, and the description. The
// preview text replaces the description of the current option when the card
// has no room for the preview box (chat.md rule 16).
func askOptionLine(question ask.Question, card *askCard, qi, oi, room int, preview, box, focused bool) string {
	inQuestion := qi == card.question
	isCurrent := inQuestion && card.cursor[qi] == oi
	marker := " "
	if inQuestion {
		marker = " "
		if isCurrent {
			marker = "›"
		}
		marker = askMarkerStyle(focused).Render(marker)
	}
	mark := "[ ]"
	if card.picks[qi][oi] {
		mark = "[x]"
	}
	label, description := ask.OtherLabel, ""
	if oi < len(question.Options) {
		label, description = safeChatText(question.Options[oi].Label), safeChatText(question.Options[oi].Description)
		if isCurrent && preview && !box && question.Options[oi].Preview != "" {
			description = strings.ReplaceAll(safeChatText(question.Options[oi].Preview), "\n", " ")
		}
	}
	label = ansi.Truncate(label, askLabelWidth, "…")
	gap := strings.Repeat(" ", askLabelWidth-lipgloss.Width(label)+1)
	number := fmt.Sprintf("%d %s ", oi+1, mark)
	var body string
	switch {
	case isCurrent && focused:
		body = titleStyle.Render(number+label) + gap
	case card.picks[qi][oi]:
		body = accentStyle.Render(number) + label + gap
	default:
		body = number + label + gap
	}
	line := marker + " " + body
	if description != "" {
		end := room - askDescRightPad
		if preview && box {
			end = askPreviewCol - 1
		}
		line += dimStyle.Render(ansi.Truncate(description, max(end-askDescCol, 0), "…"))
	}
	return line
}

// askTextLine gives the line of a note or of an Other answer. The cursor
// █ shows while the line takes the keys. empty shows in dim while the text
// is empty.
func askTextLine(name, text string, editing bool, empty string) string {
	line := dimStyle.Render(name+" ›") + " "
	switch {
	case editing:
		line += userTextStyle.Render(text + "█")
	case text == "":
		line += dimStyle.Render(empty)
	default:
		line += userTextStyle.Render(text)
	}
	return line
}

// askPreviewBox gives the rows of the preview box of the current option. Its
// height holds the longest preview of the question.
func askPreviewBox(question ask.Question, cursor int) []string {
	height := 0
	for _, option := range question.Options {
		height = max(height, len(strings.Split(option.Preview, "\n")))
	}
	var text []string
	if cursor < len(question.Options) && question.Options[cursor].Preview != "" {
		text = strings.Split(safeChatText(question.Options[cursor].Preview), "\n")
	}
	inner := askPreviewWidth - 2
	title := titleStyle.Render(fmt.Sprintf(" preview · %d ", cursor+1))
	edge := lipgloss.NewStyle().Foreground(component.Border)
	code := lipgloss.NewStyle().Foreground(component.Code)
	box := []string{edge.Render("╭─") + title + edge.Render(strings.Repeat("─", max(inner-1-lipgloss.Width(title), 0))+"╮")}
	for index := range height {
		line := ""
		if index < len(text) {
			line = text[index]
		}
		box = append(box, edge.Render("│")+" "+code.Render(" "+fitText(line, inner-3))+" "+edge.Render("│"))
	}
	return append(box, edge.Render("╰"+strings.Repeat("─", inner)+"╯"))
}

// clickAsk handles a click on a card row. A click on an option row of the
// current question picks the option. A click gives the card the focus.
func (m *Model) clickAsk(row int) tea.Cmd {
	record, _, ok := m.focusedAsk()
	if !ok {
		return nil
	}
	m.chatInputActive = false
	m.invalidateView()
	if row < 0 || row >= len(m.askRows) {
		return nil
	}
	card := m.askCardOf(record)
	ref := m.askRows[row]
	if ref.option < 0 || ref.question != card.question || card.review {
		return nil
	}
	card.editing = ""
	return m.askPick(record, card, ref.option)
}

// askHeaders gives the headers of record joined with ", ", as the boards
// show them.
func askHeaders(record ask.Record) string {
	headers := make([]string, len(record.Questions))
	for index, question := range record.Questions {
		headers[index] = safeChatText(question.Header)
	}
	return strings.Join(headers, ", ")
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
