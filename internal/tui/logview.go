package tui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
)

const maxErrorBytes = 300

var levelCycle = []string{"debug", "info", "warn", "error"}

func (m *Model) updateViewing(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "q", "ctrl+c":
		m.logAction("quit")
		return tea.Quit
	case "/":
		m.mode = editing
		m.input.Reset()
		m.logAction("edit start")
		return m.input.Focus()
	case "l":
		m.cycleLevel()
	case "f":
		m.follow = !m.follow
		if m.follow {
			m.viewport.GotoBottom()
		}
		m.logAction("follow")
	case "g":
		m.follow = false
		m.viewport.GotoTop()
		m.logAction("top")
	case "G":
		m.follow = true
		m.viewport.GotoBottom()
		m.logAction("bottom")
	}
	return nil
}

func (m *Model) updateEditing(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "ctrl+c":
		m.logAction("quit")
		return tea.Quit
	case "esc", "escape":
		m.textFilter = ""
		m.input.Reset()
		m.input.Blur()
		m.mode = viewing
		m.rebuild()
		m.logAction("edit cancel")
		return nil
	case "enter":
		m.textFilter = m.input.Value()
		m.input.Blur()
		m.mode = viewing
		m.rebuild()
		m.logAction("filter apply")
		return nil
	}
	updated, cmd := m.input.Update(press)
	m.input = updated
	return cmd
}

func (m *Model) cycleLevel() {
	for index, level := range levelCycle {
		if level == m.level {
			m.level = levelCycle[(index+1)%len(levelCycle)]
			m.rebuild()
			m.logAction("level")
			return
		}
	}
	m.level = levelCycle[0]
	m.rebuild()
	m.logAction("level")
}

func (m *Model) appendEvent(event logs.Event) {
	m.events = append(m.events, event)
	m.rebuild()
}

func (m *Model) rebuild() {
	lines := make([]string, 0, len(m.events))
	if len(m.events) == 0 {
		lines = append(lines, "no events today")
	}
	for _, event := range m.events {
		text := renderEvent(event)
		if !m.eventMatches(event, text) {
			continue
		}
		lines = append(lines, text)
	}
	m.viewport.SetContentLines(lines)
}

func renderEvent(event logs.Event) string {
	if event.Malformed() {
		return logs.MalformedText(event)
	}
	return logs.TextLine(event)
}

func (m *Model) eventMatches(event logs.Event, text string) bool {
	level := m.level
	if level == "debug" {
		level = ""
	}
	if !(logs.Filter{Level: level}).Match(event) {
		return false
	}
	return m.textFilter == "" || strings.Contains(text, m.textFilter)
}

func (m *Model) statusLine() string {
	text := m.textFilter
	if text == "" {
		text = "-"
	}
	follow := "off"
	if m.follow {
		follow = "on"
	}
	return "level: " + m.level + "  text: " + sanitizeText(text) + "  follow: " + follow
}

func (m *Model) bodyView() string {
	if m.errText != "" {
		return "error: " + m.errText
	}
	return m.viewport.View()
}

func (m *Model) setError(message string) {
	m.errText = boundedError(message)
	m.logError("view", errString(m.errText))
}

func boundedError(message string) string {
	message = sanitizeText(message)
	if len(message) <= maxErrorBytes {
		return message
	}
	cut := maxErrorBytes
	for cut > 0 && !utf8.RuneStart(message[cut]) {
		cut--
	}
	return message[:cut]
}

func sanitizeText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		switch r {
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
				builder.WriteString(`?`)
			} else {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}

type errorText string

func (e errorText) Error() string { return string(e) }

func errString(message string) error { return errorText(message) }
