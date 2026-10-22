package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
}

type chatSendMsg struct {
	text     string
	envelope output.Envelope
}

// chatFeed is the scroll state of the feed and the lines of its last render.
// Top is the first line of the feed in view. TopItem is a message to put at
// the top on the next render, or -1. Follow keeps the last line in view.
type chatFeed struct {
	top     int
	topItem int
	follow  bool
	lines   []string
	starts  []int
	width   int
	visible int
}

// chatScreen is the tab 2 Chat (decisions 15, 24, 30, 31): a feed that
// scrolls and a fixed input below it.
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
	return buildChatScreenData(chatResult.Entries, inboxRecords, now), nil
}

// buildChatScreenData keeps the chat entries of the chat window.
func buildChatScreenData(entries []chatlib.Entry, inboxRecords []records.Inbox, now time.Time) chatScreenData {
	cutoff := now.UTC().Add(-chatWindow)
	result := chatScreenData{entries: []chatlib.Entry{}, inbox: inboxRecords}
	for _, entry := range entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err == nil && !at.Before(cutoff) {
			result.entries = append(result.entries, entry)
		}
	}
	return result
}

// newChatInput gives the edit buffer of the input. The Input component draws
// it.
func newChatInput() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.SetPromptFunc(2, func(info textarea.PromptInfo) string {
		if info.LineNumber == 0 {
			return "> "
		}
		return "  "
	})
	input.ShowLineNumbers = false
	input.CharLimit = 0
	input.MaxHeight = 0
	input.EndOfBufferCharacter = ' '
	input.SetVirtualCursor(false)
	styles := input.Styles()
	styles.Focused.Text = lipgloss.NewStyle().Foreground(component.User)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(component.User)
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(component.User)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(component.User)
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	input.SetStyles(styles)
	input.Blur()
	return input
}

func (m *Model) updateChat(press tea.KeyPressMsg) tea.Cmd {
	if press.String() == "ctrl+o" {
		m.chatEventsOpen = !m.chatEventsOpen
		m.rebuildChat()
		return m.logAction("chat events toggle", slog.Bool("open", m.chatEventsOpen))
	}
	if m.chatInputActive {
		switch press.String() {
		case "enter":
			return m.sendChat()
		case "shift+enter", "alt+enter":
			m.chatInput.InsertString("\n")
			return nil
		case "esc":
			m.chatInputActive = false
			m.chatInput.Blur()
			return m.logAction("focus region", slogString("region", "feed"))
		}
		updated, cmd := m.chatInput.Update(press)
		m.chatInput = updated
		return cmd
	}
	switch press.String() {
	case "up", "k":
		m.scrollChat(-1)
	case "down", "j":
		m.scrollChat(1)
	case "g":
		m.chatFeed.follow = false
		m.chatFeed.top = 0
	case "G":
		m.chatFeed.follow = true
	}
	return nil
}

// scrollChatFeed pages the feed (decision 22).
func (m *Model) scrollChatFeed(press tea.KeyPressMsg) tea.Cmd {
	page := max(m.chatFeed.visible, 1)
	if press.String() == "pgup" {
		page = -page
	}
	m.scrollChat(page)
	return nil
}

// scrollChat moves the feed by lines. The feed follows its end when the last
// line comes into view.
func (m *Model) scrollChat(lines int) {
	feed := &m.chatFeed
	end := max(len(feed.lines)-feed.visible, 0)
	if feed.follow {
		feed.top = end
	}
	feed.top = min(max(feed.top+lines, 0), max(end, feed.top))
	feed.follow = feed.top >= end
}

func (m *Model) sendChat() tea.Cmd {
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

// chatView draws the feed panel and the input at exactly width×height cells.
// From 120 columns the first body row of the feed is blank, and the top
// border shows the count of the messages above the view.
func (m *Model) chatView(width, height int) string {
	wide := width >= component.LongFormWidth
	input := m.chatInputView(width, height)
	inputHeight := len(strings.Split(input, "\n"))
	feedHeight := max(height-inputHeight, component.PanelMinHeight)
	body := feedHeight - 2
	pad := 0
	if wide {
		pad = 1
	}
	feed := &m.chatFeed
	feed.visible = max(body-pad, 1)
	if feed.lines == nil || feed.width != width {
		feed.lines, feed.starts = m.chatLines(width)
		feed.width = width
	}
	end := max(len(feed.lines)-feed.visible, 0)
	switch {
	case feed.topItem >= 0:
		feed.top = feed.starts[min(feed.topItem, len(feed.starts)-1)]
		feed.topItem = -1
		feed.follow = false
	case feed.follow:
		feed.top = end
	}
	feed.top = min(max(feed.top, 0), max(len(feed.lines)-1, 0))
	earlier := sort.SearchInts(feed.starts, feed.top)

	lines := make([]string, pad, pad+feed.visible)
	lines = append(lines, feed.lines[feed.top:min(feed.top+feed.visible, len(feed.lines))]...)
	title := "chat"
	if wide {
		title = "chat · noise " + m.chatNoise
	}
	mark := component.Thumb(body, len(feed.lines), body, feed.top)
	panel := component.Panel{Title: title, Focused: !m.chatInputActive, Lines: lines, VMark: mark}
	rows := strings.Split(panel.View(width, feedHeight), "\n")
	for row := max(mark.Start, 0); row < min(mark.Start+mark.Size, feedHeight-2); row++ {
		rows[row+1] = ansi.Truncate(rows[row+1], width-1, "") + dimStyle.Render("┃")
	}
	if wide && earlier > 0 {
		rows[0] = borderNote(rows[0], fmt.Sprintf(" ↑ %d earlier ", earlier), width/2)
	}
	return strings.Join(rows, "\n") + "\n" + input
}

func (m *Model) chatInputView(width, contentHeight int) string {
	panelCap := max(contentHeight/3, 3)
	m.chatInput.DynamicHeight = true
	m.chatInput.MinHeight = 1
	m.chatInput.MaxHeight = 0
	// The textarea keeps its two-cell prompt inside this width, so its rows
	// for up and down are width-6 cells, as the rows of the Input component.
	m.chatInput.SetWidth(max(width-4, 3))
	status := m.header.coordinator
	if status == "" {
		status = "none"
	}
	age := coordinatorStatusAge(m.header.coordinatorChangedAt, m.now())
	return component.Input{
		Text: m.chatInput.Value(), Focused: m.chatInputActive, Height: panelCap,
		CursorLine: m.chatInput.Line(), CursorColumn: m.chatInput.Column(),
		HasCursor: true, Status: status, Age: age,
		Queued: m.lastUserChatPending(),
	}.View(width)
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

func (m *Model) lastUserChatPending() bool {
	var latest time.Time
	pending := false
	for _, entry := range m.chatData.entries {
		if entry.From != "user" {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err != nil || at.Before(latest) {
			continue
		}
		latest = at
		pending = entry.Delivery == "pending"
	}
	return pending
}

// chatCursor puts the real terminal cursor on the cursor cell of the input
// text row, while the input has the focus. The input bottom border and the
// footer are below that row.
func (m *Model) chatCursor(content string) *tea.Cursor {
	if m.activeTab() != tabChat || !m.chatInputActive || m.helpOpen {
		return nil
	}
	for row, line := range strings.Split(ansi.Strip(content), "\n") {
		index := strings.LastIndex(line, "█")
		if index >= 0 {
			cursor := tea.NewCursor(ansi.StringWidth(line[:index]), row)
			cursor.Shape = tea.CursorBar
			cursor.Color = lipgloss.Color(colorAccent)
			return cursor
		}
	}
	return nil
}

// borderNote draws note in dim over the top border row from cell col.
func borderNote(row, note string, col int) string {
	end := col + lipgloss.Width(note)
	if end >= lipgloss.Width(row) {
		return row
	}
	return ansi.Truncate(row, col, "") + dimStyle.Render(note) + ansi.TruncateLeft(row, end, "")
}

// chatLines gives the lines of the feed for a feed panel of width cells, and
// the first line of each message. A blank line separates two messages, as
// in component.ChatLines.
func (m *Model) chatLines(width int) ([]string, []int) {
	var lines []string
	var starts []int
	previousFoldedEvent := false
	for _, msg := range m.chatMessages() {
		got, err := component.ChatLines([]component.ChatMessage{msg}, width)
		if err != nil {
			got = []string{safeChatText(msg.Text)}
		}
		foldedEvent := msg.Kind == component.ChatEvent && !msg.Open
		if len(lines) > 0 && ansi.Strip(lines[len(lines)-1]) != "" && (!previousFoldedEvent || !foldedEvent) {
			lines = append(lines, "")
		}
		starts = append(starts, len(lines))
		lines = append(lines, got...)
		previousFoldedEvent = foldedEvent
	}
	if len(lines) == 0 {
		lines = []string{dimStyle.Render("no chat yet")}
		starts = []int{0}
	}
	return lines, starts
}

// chatMessages gives the chat entries and the visible inbox records in time
// order. A user entry that is not sent shows its delivery after the name.
func (m *Model) chatMessages() []component.ChatMessage {
	type item struct {
		at  time.Time
		msg component.ChatMessage
	}
	zone := m.now().Location()
	var items []item
	for _, entry := range m.chatData.entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err != nil {
			continue
		}
		msg := component.ChatMessage{Time: at.In(zone).Format("15:04"), Name: "coordinator", Kind: component.ChatAgent, Text: safeChatText(entry.Text)}
		if entry.From == "user" {
			msg.Name, msg.Kind = "you", component.ChatUser
			if entry.Delivery != "sent" {
				msg.Name += " · " + entry.Delivery
			}
		}
		items = append(items, item{at, msg})
	}
	for _, event := range m.visibleChatEvents() {
		at, err := time.Parse(time.RFC3339Nano, event.CreatedAt)
		if err != nil {
			continue
		}
		items = append(items, item{at, component.ChatMessage{
			Time: at.In(zone).Format("15:04"), Name: event.Agent, Kind: component.ChatEvent,
			Text: safeChatText(event.Text), Open: m.chatEventsOpen,
		}})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].at.Before(items[j].at) })
	msgs := make([]component.ChatMessage, len(items))
	for index, it := range items {
		msgs[index] = it.msg
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
