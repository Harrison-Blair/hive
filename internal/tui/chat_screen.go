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
	"github.com/Harrison-Blair/hive/internal/agent"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const chatWindow = 7 * 24 * time.Hour

type chatScreenData struct {
	entries          []chatlib.Entry
	inbox            []records.Inbox
	coordinator      records.Agent
	coordinatorFound bool
}

type chatSendMsg struct {
	text     string
	envelope output.Envelope
}

type chatScreen struct{ model *Model }

func (s chatScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	return s.model.updateChat(press)
}

func (s chatScreen) View(width, height int) string { return s.model.chatView(width, height) }

func (s chatScreen) KeyHints() string { return s.model.chatKeyHints() }

func (chatScreen) DataRequest() dataRequest { return chatDataRequest }

func loadChatScreenData(home string, now time.Time) (chatScreenData, error) {
	cutoff := now.UTC().Add(-chatWindow)
	result := chatScreenData{entries: []chatlib.Entry{}, inbox: []records.Inbox{}}
	chatEnvelope := chatlib.Log(home, chatlib.LogOptions{Since: cutoff.Format("2006-01-02")})
	if chatEnvelope.Outcome != "ok" {
		return result, fmt.Errorf("chat history: %v", chatEnvelope.Error)
	}
	chatResult, ok := chatEnvelope.Result.(chatlib.LogResult)
	if !ok {
		return result, errors.New("chat history: unexpected result")
	}
	for _, entry := range chatResult.Entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err == nil && !at.Before(cutoff) {
			result.entries = append(result.entries, entry)
		}
	}

	inboxRecords, err := inbox.All(home, "coordinator")
	if err != nil {
		return result, fmt.Errorf("coordinator inbox: %w", err)
	}
	result.inbox = inboxRecords

	agentEnvelope := agent.List(home, agent.ListOptions{Role: "coordinator", All: true})
	if agentEnvelope.Outcome != "ok" {
		return result, fmt.Errorf("coordinator agent: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return result, errors.New("coordinator agent: unexpected result")
	}
	for _, record := range agents {
		live := record.Status == records.AgentLive
		currentLive := result.coordinator.Status == records.AgentLive
		if !result.coordinatorFound || (live && !currentLive) || (live == currentLive && record.Generation > result.coordinator.Generation) {
			result.coordinator = record
			result.coordinatorFound = true
		}
	}
	return result, nil
}

func newChatInput() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.Placeholder = "Type a message"
	input.CharLimit = 0
	input.MinHeight = 1
	input.MaxHeight = 4
	input.DynamicHeight = true
	input.Blur()
	return input
}

func (m *Model) updateChat(press tea.KeyPressMsg) tea.Cmd {
	if m.chatInputActive {
		switch press.String() {
		case "enter":
			return m.sendChat()
		case "shift+enter", "alt+enter":
			m.chatInput.InsertString("\n")
			return nil
		case "esc", "escape":
			m.chatInputActive = false
			m.chatInput.Blur()
			return m.logAction("chat input close")
		}
		updated, cmd := m.chatInput.Update(press)
		m.chatInput = updated
		return cmd
	}
	switch press.String() {
	case "i":
		m.chatInputActive = true
		return batch(m.chatInput.Focus(), m.logAction("chat input open"))
	case "n":
		switch m.chatNoise {
		case "messages":
			m.chatNoise = "off"
		case "off":
			m.chatNoise = "all"
		default:
			m.chatNoise = "messages"
		}
		m.rebuildChat()
		return m.logAction("chat noise", slogString("noise", m.chatNoise))
	case "e":
		m.chatEventsExpanded = !m.chatEventsExpanded
		m.rebuildChat()
		return m.logAction("chat events expand", slogBool("expanded", m.chatEventsExpanded))
	case "pgup", "pgdown":
		m.chatFollow = false
		updated, cmd := m.chatViewport.Update(press)
		m.chatViewport = updated
		if m.chatViewport.AtBottom() {
			m.chatFollow = true
		}
		return cmd
	}
	return nil
}

func (m *Model) sendChat() tea.Cmd {
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
		envelope := chatlib.Send(context.Background(), home, chatlib.SendOptions{Text: text})
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

func (m *Model) chatKeyHints() string {
	if m.chatInputActive {
		return "enter send · shift+enter newline · esc keep draft · ctrl+c×2 quit"
	}
	return "i write · pgup/pgdown scroll · e expand · n noise · 1–8 screens · ctrl+c×2 quit"
}

func (m *Model) chatView(width, height int) string {
	if width < 1 {
		width = 1
	}
	eventWidth := width / 3
	if eventWidth < 18 {
		eventWidth = 18
	}
	if eventWidth > 32 {
		eventWidth = 32
	}
	if eventWidth > width-12 {
		eventWidth = width - 12
	}
	if eventWidth < 1 {
		eventWidth = 1
	}
	conversationWidth := width - eventWidth
	if conversationWidth < 1 {
		conversationWidth = 1
	}
	feedHeight := height - 8
	if feedHeight < 1 {
		feedHeight = 1
	}
	m.chatViewport.SetWidth(conversationWidth - 2)
	m.chatViewport.SetHeight(feedHeight)
	m.chatInput.SetWidth(conversationWidth - 2)
	m.chatInput.SetHeight(4)
	if m.chatFollow {
		m.chatViewport.GotoBottom()
	}

	status := coordinatorLine(m.chatData)
	conversation := titleStyle.Render("coordinator") + "\n" + status + "\n" + m.chatViewport.View()
	if m.chatInputActive || m.chatInput.Value() != "" {
		conversation += "\n" + m.chatInput.View()
	}
	main := panelStyle.Width(conversationWidth).Height(height).Render(conversation)
	events := m.chatEventsView(eventWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, main, events)
}

func coordinatorLine(data chatScreenData) string {
	if !data.coordinatorFound {
		return dimStyle.Render("not live · no coordinator")
	}
	record := data.coordinator
	live := "not live"
	stateStyle := dimStyle
	if record.Status == records.AgentLive {
		live = "live"
		stateStyle = workingStyle
	}
	harnessStatus := record.AgentStatus
	if harnessStatus == "" {
		harnessStatus = "unknown"
	}
	model := record.Model
	if model == "" {
		model = "—"
	}
	harness := record.Harness
	if harness == "" {
		harness = "unknown harness"
	}
	pane := record.PaneID
	if pane == "" {
		pane = "no pane"
	}
	return fmt.Sprintf("%s · %s · %s\n%s · %s", stateStyle.Render(live), harness, harnessStatus, model, pane)
}

func (m *Model) chatEventsView(width, height int) string {
	lines := []string{titleStyle.Render("hive events")}
	if m.chatNoise == "off" {
		lines = append(lines, "off")
	} else {
		events := m.visibleChatEvents()
		for _, event := range events {
			lines = append(lines, renderInboxEvent(event, false))
		}
		if len(events) == 0 {
			lines = append(lines, "no events")
		}
		lines = append(lines, "", "noise: "+m.chatNoise)
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m *Model) visibleChatEvents() []records.Inbox {
	visible := make([]records.Inbox, 0, len(m.chatData.inbox))
	for _, event := range m.chatData.inbox {
		if m.chatNoise == "messages" && event.Kind != records.InboxMessage {
			continue
		}
		visible = append(visible, event)
	}
	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].CreatedAt > visible[j].CreatedAt
	})
	return visible
}

func (m *Model) rebuildChat() {
	items := make([]chatTimelineItem, 0, len(m.chatData.entries)+len(m.chatData.inbox))
	for _, entry := range m.chatData.entries {
		at, err := time.Parse(time.RFC3339Nano, entry.At)
		if err == nil {
			copy := entry
			items = append(items, chatTimelineItem{at: at, entry: &copy})
		}
	}
	for _, event := range m.visibleChatEvents() {
		at, err := time.Parse(time.RFC3339Nano, event.CreatedAt)
		if err == nil {
			copy := event
			items = append(items, chatTimelineItem{at: at, event: &copy})
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].at.Before(items[j].at) })
	lines := make([]string, 0, len(items))
	pendingEvents := 0
	flushEvents := func() {
		if pendingEvents == 0 {
			return
		}
		if m.chatEventsExpanded {
			for index := range items {
				item := &items[index]
				if item.event == nil || !item.used {
					continue
				}
				lines = append(lines, renderInboxEvent(*item.event, true))
				item.used = false
			}
		} else {
			lines = append(lines, fmt.Sprintf("── %d hive events ──", pendingEvents))
		}
		pendingEvents = 0
	}
	for index := range items {
		item := &items[index]
		if item.event != nil {
			item.used = true
			pendingEvents++
			continue
		}
		flushEvents()
		lines = append(lines, renderChatEntry(*item.entry))
	}
	flushEvents()
	if len(lines) == 0 {
		lines = []string{"no chat yet"}
	}
	m.chatViewport.SetContentLines(lines)
}

type chatTimelineItem struct {
	at    time.Time
	entry *chatlib.Entry
	event *records.Inbox
	used  bool
}

func renderChatEntry(entry chatlib.Entry) string {
	from := "coordinator"
	if entry.From == "user" {
		from = "you"
	}
	at := formatChatTime(entry.At)
	delivery := ""
	if entry.From == "user" && entry.Delivery != "sent" {
		delivery = " [" + entry.Delivery + "]"
	}
	reply := ""
	if entry.ReplyTo != "" {
		reply = " ↳ " + entry.ReplyTo
	}
	text := safeChatText(entry.Text)
	text = strings.ReplaceAll(text, "\n", "\n  ")
	return fmt.Sprintf("%s · %s%s%s\n  %s", from, at, delivery, reply, text)
}

func renderInboxEvent(event records.Inbox, expanded bool) string {
	at := formatChatTime(event.CreatedAt)
	text := safeChatText(event.Text)
	if expanded {
		return fmt.Sprintf("hive event · %s · %s: %s", at, event.Agent, strings.ReplaceAll(text, "\n", "\n  "))
	}
	firstLine, _, _ := strings.Cut(text, "\n")
	return fmt.Sprintf("%s · %s: %s", at, event.Agent, firstLine)
}

func formatChatTime(value string) string {
	at, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "unknown time"
	}
	return at.Local().Format("Jan 02 15:04")
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

func slogBool(key string, value bool) slog.Attr { return slog.Bool(key, value) }
