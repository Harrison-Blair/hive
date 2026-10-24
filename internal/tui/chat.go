package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/ask"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

const chatWindow = 7 * 24 * time.Hour

// The noise levels of the feed. Each level shows the inbox kinds that the
// earlier Chat screen showed at the same level.
const (
	chatNoiseQuiet  = "quiet"
	chatNoiseNormal = "normal"
	chatNoiseAll    = "all"
)

type chatScreenData struct {
	entries []chatlib.Entry
	inbox   []records.Inbox
	asks    []ask.Record
	routes  map[string]askRoute
}

// chatItem is one message of the transcript, its time, and the inbox
// sequence of an event.
type chatItem struct {
	at  time.Time
	seq uint64
	msg component.ChatMessage
}

type chatSendMsg struct {
	text     string
	envelope output.Envelope
}

// chatFeed is the scroll state of the transcript and the lines of its last
// render. Height is the rows of the transcript and composerTop is the first
// row of the composer. Top is the first line in view. TopItem is a message to
// put at the top on the next render, or -1. Follow keeps the last line in
// view. Msgs and seqs are the messages of the lines and the inbox sequence of
// each event.
type chatFeed struct {
	top         int
	topItem     int
	follow      bool
	lines       []string
	starts      []int
	msgs        []component.ChatMessage
	seqs        []uint64
	width       int
	visible     int
	height      int
	composerTop int
}

// chatScreen is Chat, the home screen of golden-frames decision 1: the
// transcript and the composer.
type chatScreen struct{ model *Model }

func (s chatScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	return s.model.updateChat(press)
}

func (s chatScreen) View(width, height int) string { return s.model.chatView(width, height) }

func (chatScreen) KeyHints() string { return "" }

func (chatScreen) DataRequest() dataRequest { return chatDataRequest }

func loadChatScreenData(home string, now time.Time) (chatScreenData, error) {
	cutoff := now.UTC().Add(-chatWindow)
	chatEnvelope := chatlib.Log(home, chatlib.LogOptions{Since: cutoff.Format("2006-01-02")})
	if chatEnvelope.Outcome != "ok" {
		return chatScreenData{}, fmt.Errorf("chat history: %v", chatEnvelope.Error)
	}
	chatResult, ok := chatEnvelope.Result.(chatlib.LogResult)
	if !ok {
		return chatScreenData{}, errors.New("chat history: unexpected result")
	}
	inboxRecords, err := inbox.All(home, "coordinator")
	if err != nil {
		return chatScreenData{}, fmt.Errorf("coordinator inbox: %w", err)
	}
	asks, err := ask.List(home)
	if err != nil {
		return chatScreenData{}, fmt.Errorf("asks: %w", err)
	}
	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if agentEnvelope.Outcome != "ok" || !ok {
		return chatScreenData{}, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	return buildChatScreenData(chatResult.Entries, inboxRecords, asks, agents, now), nil
}

// buildChatScreenData keeps the chat entries of the chat window, the asks in
// the order of their IDs, and the route of each ask.
func buildChatScreenData(entries []chatlib.Entry, inboxRecords []records.Inbox, asks []ask.Record, agents []records.Agent, now time.Time) chatScreenData {
	cutoff := now.UTC().Add(-chatWindow)
	asks = slices.Clone(asks)
	sortAsks(asks)
	result := chatScreenData{entries: []chatlib.Entry{}, inbox: inboxRecords, asks: asks, routes: buildAskRoutes(asks, agents)}
	for _, entry := range entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err == nil && !at.Before(cutoff) {
			result.entries = append(result.entries, entry)
		}
	}
	return result
}

// newChatInput gives the edit buffer of the input line. The Composer
// component draws it. Golden-frames decision 16 takes ctrl+g, ctrl+t, and
// ctrl+f from the key map, and right keeps character forward.
func newChatInput() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.CharLimit = 0
	input.MaxHeight = 0
	input.SetVirtualCursor(false)
	input.KeyMap.CharacterForward.SetKeys("right")
	input.KeyMap.TransposeCharacterBackward.SetKeys()
	input.KeyMap.SelectAll.SetKeys()
	input.Focus()
	return input
}

func (m *Model) updateChat(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "ctrl+o":
		return m.toggleChatEvents()
	case "enter":
		return m.sendChat()
	case "shift+enter", "alt+enter":
		m.chatInput.InsertString("\n")
		m.invalidateView()
		return nil
	}
	updated, cmd := m.chatInput.Update(press)
	m.chatInput = updated
	m.invalidateView()
	return cmd
}

// toggleChatEvents opens or folds all events (chat.md rule 3).
func (m *Model) toggleChatEvents() tea.Cmd {
	m.chatEventsOpen = !m.chatEventsOpen
	m.chatEventOpen = nil
	m.rebuildChat()
	return m.logAction("chat events toggle", slog.Bool("open", m.chatEventsOpen))
}

// scrollChatFeed pages the transcript (chat.md rule 7).
func (m *Model) scrollChatFeed(press tea.KeyPressMsg) tea.Cmd {
	page := max(m.chatFeed.visible, 1)
	if press.String() == "pgup" {
		page = -page
	}
	m.scrollChat(page)
	return nil
}

// scrollChat moves the transcript by lines. The transcript follows its end
// when the last line comes into view.
func (m *Model) scrollChat(lines int) {
	feed := &m.chatFeed
	end := max(len(feed.lines)-feed.visible, 0)
	if feed.follow {
		feed.top = end
	}
	feed.top = min(max(feed.top+lines, 0), end)
	feed.follow = feed.top >= end
	m.invalidateView()
}

// sendChat sends the input line (chat.md rule 12). With no live
// coordinator, it shows the error and keeps the text.
func (m *Model) sendChat() tea.Cmd {
	if m.header.coordinator == "" {
		m.chatErr = "no live coordinator"
		m.invalidateView()
		return m.logError("chat send", errors.New(m.chatErr))
	}
	if m.demo != nil {
		return m.demoBlocked("chat send")
	}
	if m.chatSending {
		return nil
	}
	text := m.chatInput.Value()
	if text == "" {
		return nil
	}
	m.chatSending = true
	m.chatErr = ""
	home := m.home
	return batch(m.logAction("chat send"), func() tea.Msg {
		envelope := chatlib.Send(context.Background(), home, chatlib.SendOptions{
			PaneID: m.chatPaneID, WorkingDir: m.chatCallerCwd, Text: text,
		})
		return chatSendMsg{text: text, envelope: envelope}
	})
}

func (m *Model) updateChatSend(msg chatSendMsg) tea.Cmd {
	m.chatSending = false
	result, ok := msg.envelope.Result.(chatlib.SendResult)
	if msg.envelope.Outcome != "ok" || !ok || result.Delivery != "sent" {
		reason := fmt.Sprint(msg.envelope.Error)
		if msg.envelope.Error == nil || reason == "<nil>" {
			if ok && result.Delivery == "unknown" {
				reason = "chat delivery is unknown"
			} else {
				reason = "chat send did not complete"
			}
		}
		m.chatErr = boundedError(reason)
		return m.logError("chat send", errors.New(reason))
	}
	if m.chatInput.Value() == msg.text {
		m.chatInput.Reset()
	}
	m.chatErr = ""
	return batch(m.logAction("chat sent", slogString("chat_id", result.ID)), m.startCurrentScreenLoad())
}

// chatView draws Chat in rows rows above the status rows (golden-frames
// chat.md): the transcript, the notice line while a notice exists, and the
// composer with the ask card.
func (m *Model) chatView(width, rows int) string {
	inputRows := max(rows/3-2, 1)
	composer := component.Composer{
		Text: m.chatInput.Value(), Cursor: m.chatInputActive,
		CursorLine: m.chatInput.Line(), CursorColumn: m.chatInput.Column(),
		Placeholder: m.chatPlaceholder(), InputRows: inputRows,
	}
	if record, _, ok := m.focusedAsk(); ok {
		used := component.InputHeight(composer.Text, width, inputRows) + 3
		composer.Card = m.askCardView(record, width, max(rows/2-used, 1))
	}
	frame := composer.View(width)
	frameRows := strings.Count(frame, "\n") + 1
	notice := m.noticeLine(width)
	transcript := rows - frameRows
	if notice != "" {
		transcript--
	}
	m.chatFeed.height = max(transcript, 0)
	m.chatFeed.composerTop = m.chatFeed.height
	if notice != "" {
		m.chatFeed.composerTop++
	}
	out := m.transcriptRows(width, max(transcript, 0))
	if notice != "" {
		out = append(out, notice)
	}
	return strings.Join(append(out, frame), "\n")
}

// chatPlaceholder gives the placeholder of chat.md rule 10.
func (m *Model) chatPlaceholder() string {
	switch m.header.coordinator {
	case "working", "idle":
		return "type a message · " + chordText(m.tuiConfig.Keys["help"]) + " help"
	case "blocked":
		return "type a message · the coordinator is blocked"
	case "":
		return "no coordinator · " + chordText(m.tuiConfig.Keys["health"]) + " health, r restarts it"
	}
	return "coordinator starting · a message waits here"
}

// transcriptRows gives the rows of the transcript (chat.md rules 1, 7, and
// 8). When the lines do not fit, the view follows the end, row 0 shows the
// count of the messages above the view, and column width-1 shows the thumb.
func (m *Model) transcriptRows(width, height int) []string {
	feed := &m.chatFeed
	if feed.lines == nil || feed.width != width {
		feed.width = width
		feed.lines, feed.starts = m.chatLines(width)
	}
	out := make([]string, 0, height)
	if len(feed.lines) == 0 {
		feed.visible = height
		for _, line := range m.welcomeLines() {
			out = append(out, line)
		}
		for len(out) < height {
			out = append(out, "")
		}
		return fitRows(out[:height], width)
	}
	if len(feed.lines) <= height {
		feed.visible, feed.top, feed.follow = height, 0, true
		out = append(out, feed.lines...)
		for len(out) < height {
			out = append(out, "")
		}
		return fitRows(out, width)
	}
	feed.visible = max(height-1, 1)
	end := len(feed.lines) - feed.visible
	if feed.topItem >= 0 {
		feed.top = feed.starts[min(feed.topItem, len(feed.starts)-1)]
		feed.topItem = -1
		feed.follow = feed.top >= end
	}
	if feed.follow {
		feed.top = end
	}
	feed.top = min(max(feed.top, 0), end)
	earlier := sort.SearchInts(feed.starts, feed.top)
	head := ""
	if earlier > 0 {
		head = strings.Repeat(" ", chatEarlierCol) + dimStyle.Render(fmt.Sprintf("· · · %d earlier · · ·", earlier))
	}
	out = append(out, head)
	out = append(out, feed.lines[feed.top:feed.top+feed.visible]...)
	out = fitRows(out, width)
	mark := component.Thumb(feed.visible, len(feed.lines), feed.visible, feed.top)
	for row := mark.Start; row < mark.Start+mark.Size; row++ {
		out[row+1] = ansi.Truncate(out[row+1], width-1, "") + dimStyle.Render("┃")
	}
	return out
}

// chatEarlierCol is the column of the earlier row of the layout of chat.md.
const chatEarlierCol = 52

// fitRows cuts or pads each row to width cells.
func fitRows(rows []string, width int) []string {
	for i, row := range rows {
		rows[i] = fitLine(row, width)
	}
	return rows
}

// welcomeLines gives the welcome block of chat.md rule 8, from row 1.
func (m *Model) welcomeLines() []string {
	home := m.headerView().Home
	state := ansi.Strip(m.stateSegment())
	glyph, rest, _ := strings.Cut(state, " ")
	glyphStyle := lipgloss.NewStyle().Foreground(stateColor(m.header.coordinator))
	chords := make([]string, 0, len(chordTabs))
	for _, chord := range chordTabs {
		chords = append(chords, chordText(m.tuiConfig.Keys[chord.action])+" "+chord.action)
	}
	return []string{
		"",
		" " + dimStyle.Render("hive · "+home),
		" " + glyphStyle.Render(glyph) + dimStyle.Render(" "+rest),
		" " + dimStyle.Render("type a message to the coordinator."),
		" " + dimStyle.Render(strings.Join(chords, "  ")),
		" " + dimStyle.Render(chordText(m.tuiConfig.Keys["help"])+" help"),
	}
}

func coordinatorStatusAge(changedAt string, now time.Time) string {
	changed, err := time.Parse(time.RFC3339Nano, changedAt)
	if err != nil {
		return ""
	}
	minutes := int(now.Sub(changed) / time.Minute)
	if minutes < 0 {
		minutes = 0
	}
	return fmt.Sprintf("%dm", minutes)
}

// chatCursor puts the real terminal cursor on the cursor cell of the input
// line, while the input line has the focus. An empty input line has the
// cursor after "> ".
func (m *Model) chatCursor(content string) *tea.Cursor {
	if m.activeTab() != tabChat || !m.chatInputActive || m.helpOpen {
		return nil
	}
	lines := strings.Split(ansi.Strip(content), "\n")
	place := func(col, row int) *tea.Cursor {
		cursor := tea.NewCursor(col, row)
		cursor.Shape = tea.CursorBar
		cursor.Color = lipgloss.Color(colorAccent)
		return cursor
	}
	for row, line := range lines {
		if index := strings.LastIndex(line, "█"); index >= 0 {
			return place(ansi.StringWidth(line[:index]), row)
		}
	}
	for row := len(lines) - 1; row >= 0; row-- {
		if strings.HasPrefix(lines[row], "│ > ") {
			return place(4, row)
		}
	}
	return nil
}

// chatLines gives the lines of the transcript at width cells and the first
// line of each message.
func (m *Model) chatLines(width int) ([]string, []int) {
	msgs := m.chatMessages()
	lines, err := component.ChatLines(msgs, width)
	if err != nil {
		lines = nil
		for _, msg := range msgs {
			lines = append(lines, " "+safeChatText(msg.Text))
		}
	}
	starts, err := component.ChatStarts(msgs, width)
	if err != nil {
		starts = make([]int, len(msgs))
		for i := range starts {
			starts[i] = i
		}
	}
	m.chatFeed.msgs = msgs
	return lines, starts
}

// chatMessages gives the chat entries, the visible inbox records, and the
// ask lines in time order (chat.md rules 2 to 6).
func (m *Model) chatMessages() []component.ChatMessage {
	zone := m.now().Location()
	var items []chatItem
	for _, entry := range m.chatData.entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err != nil {
			continue
		}
		msg := component.ChatMessage{Time: at.In(zone).Format("15:04"), Name: "coordinator", Kind: component.ChatAgent, Text: safeChatText(entry.Text)}
		if entry.From == "user" {
			msg.Name, msg.Kind = "you", component.ChatUser
			switch entry.Delivery {
			case "pending":
				msg.Note = dimStyle.Render(" · pending")
			case "unknown":
				msg.Note = accentStyle.Render(" · not delivered")
			}
		}
		items = append(items, chatItem{at: at, msg: msg})
	}
	for _, event := range m.visibleChatEvents() {
		at, err := time.Parse(time.RFC3339Nano, event.CreatedAt)
		if err != nil {
			continue
		}
		open, ok := m.chatEventOpen[event.Seq]
		if !ok {
			open = m.chatEventsOpen
		}
		items = append(items, chatItem{at: at, seq: event.Seq, msg: component.ChatMessage{
			Time: at.In(zone).Format("15:04"), Name: event.Agent, Kind: component.ChatEvent,
			Text: safeChatText(event.Text), Open: open,
		}})
	}
	items = append(items, m.askMessages(zone)...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].at.Before(items[j].at) })
	msgs := make([]component.ChatMessage, len(items))
	m.chatFeed.seqs = make([]uint64, len(items))
	for index, it := range items {
		msgs[index] = it.msg
		m.chatFeed.seqs[index] = it.seq
	}
	return msgs
}

// visibleChatEvents gives the inbox records of the noise level.
func (m *Model) visibleChatEvents() []records.Inbox {
	visible := make([]records.Inbox, 0, len(m.chatData.inbox))
	for _, event := range m.chatData.inbox {
		switch m.chatNoise {
		case chatNoiseNormal:
			switch event.Kind {
			case records.InboxMessage, records.InboxBlocked, records.InboxUnresolved,
				records.InboxCompleted, records.InboxVerified, records.InboxQuiet:
			default:
				continue
			}
		case chatNoiseQuiet:
			switch event.Kind {
			case records.InboxMessage, records.InboxBlocked, records.InboxUnresolved:
			default:
				continue
			}
		}
		visible = append(visible, event)
	}
	return visible
}

// rebuildChat drops the lines of the last render after new data.
func (m *Model) rebuildChat() {
	m.chatFeed.lines = nil
}

func safeChatText(text string) string {
	var safe strings.Builder
	for _, r := range text {
		if r == '\n' {
			safe.WriteRune(r)
		} else if r == '\t' {
			safe.WriteString("  ")
		} else if unicode.IsControl(r) {
			safe.WriteRune('?')
		} else {
			safe.WriteRune(r)
		}
	}
	return safe.String()
}

func slogString(key, value string) slog.Attr { return slog.String(key, value) }
