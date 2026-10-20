package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
)

const maxErrorBytes = 300

var levelCycle = []string{"debug", "info", "warn", "error"}

type logRow struct {
	event     logs.Event
	text      string
	key       string
	requestID string
}

func (m *Model) updateViewing(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "/":
		m.mode = editing
		m.input.Reset()
		return batch(m.input.Focus(), m.logAction("edit start"))
	case "l":
		return m.cycleLevel()
	case "c":
		m.componentFilter = cycleValue(m.componentFilter, m.logValues("component"))
		m.rebuild()
		return m.logAction("component filter")
	case "p":
		m.projectFilter = cycleValue(m.projectFilter, m.logValues("project_id"))
		m.rebuild()
		return m.logAction("project filter")
	case "a":
		m.agentFilter = cycleValue(m.agentFilter, m.logValues("agent_id"))
		m.rebuild()
		return m.logAction("agent filter")
	case "f":
		m.follow = !m.follow
		if m.follow {
			m.viewport.GotoBottom()
		}
		return m.logAction("follow")
	case "r":
		if m.requestFilter != "" {
			m.requestFilter = ""
		} else if row, ok := m.selectedLogRow(); ok {
			m.requestFilter = row.requestID
		}
		m.rebuild()
		return m.logAction("request filter")
	case "enter":
		if _, ok := m.selectedLogRow(); ok {
			m.logDetail = !m.logDetail
		}
		return m.logAction("event detail")
	case "[", "]":
		step := -1
		if press.String() == "]" {
			step = 1
		}
		m.logDate = startOfDay(m.logDate.AddDate(0, 0, step).Local())
		m.logDatePinned = true
		m.events = nil
		m.logSelection = 0
		m.logSelectedKey = ""
		m.logDetail = false
		m.rebuild()
		return batch(m.logAction("log day", slogTime("day", m.logDate)), m.startCurrentScreenLoad())
	case "up", "k":
		m.moveLogSelection(-1)
		return nil
	case "down", "j":
		m.moveLogSelection(1)
		return nil
	case "g":
		m.follow = false
		m.viewport.GotoTop()
		return m.logAction("top")
	case "G":
		m.follow = true
		m.viewport.GotoBottom()
		return m.logAction("bottom")
	case "esc", "escape":
		if m.requestFilter != "" {
			m.requestFilter = ""
			m.rebuild()
			return m.logAction("request filter clear")
		}
		if m.logDetail {
			m.logDetail = false
			return m.logAction("event detail close")
		}
		m.clearTextFilter()
		return m.logAction("filter clear")
	}
	return nil
}

func (m *Model) updateEditing(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "esc", "escape":
		m.requestFilter = ""
		m.clearTextFilter()
		return m.logAction("edit cancel")
	case "enter":
		m.textFilter = m.input.Value()
		m.input.Blur()
		m.mode = viewing
		m.rebuild()
		return m.logAction("filter apply")
	}
	updated, cmd := m.input.Update(press)
	m.input = updated
	return cmd
}

// clearTextFilter clears the applied text filter and the input, returns to
// viewing mode, and rebuilds the view with the other filters unchanged.
func (m *Model) clearTextFilter() {
	m.textFilter = ""
	m.input.Reset()
	m.input.Blur()
	m.mode = viewing
	m.rebuild()
}

func (m *Model) cycleLevel() tea.Cmd {
	for index, level := range levelCycle {
		if level == m.level {
			m.level = levelCycle[(index+1)%len(levelCycle)]
			m.rebuild()
			return m.logAction("level")
		}
	}
	m.level = "info"
	m.rebuild()
	return m.logAction("level")
}

func (m *Model) rebuild() {
	rows := m.filteredLogRows()
	if len(rows) == 0 {
		m.logSelection = 0
		m.logSelectedKey = ""
		message := "no events on " + m.logDate.Format("2006-01-02")
		if sameDay(m.logDate, m.now()) {
			message = "no events today"
		}
		m.viewport.SetContentLines([]string{message})
		return
	}
	if m.logSelectedKey != "" {
		for index, row := range rows {
			if row.key == m.logSelectedKey {
				m.logSelection = index
				break
			}
		}
	} else if m.follow {
		m.logSelection = len(rows) - 1
	}
	if m.logSelection >= len(rows) {
		m.logSelection = len(rows) - 1
	}
	if m.logSelection < 0 {
		m.logSelection = 0
	}
	m.logSelectedKey = rows[m.logSelection].key
	lines := make([]string, len(rows))
	for index, row := range rows {
		lines[index] = row.text
		if index == m.logSelection {
			lines[index] = selectedStyle.Render("▶ " + row.text)
		}
	}
	m.viewport.SetContentLines(lines)
	m.viewport.EnsureVisible(m.logSelection, 0, 1)
}

func (m *Model) filteredLogRows() []logRow {
	rows := make([]logRow, 0, len(m.events))
	for _, row := range m.makeLogRows() {
		if !m.rowMatches(row) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func (m *Model) makeLogRows() []logRow {
	ends := make(map[string]int)
	for index, event := range m.events {
		if logField(event, "msg") == "command end" {
			key := commandRequestKey(event)
			if key != "" {
				ends[key] = index
			}
		}
	}
	skipped := make(map[int]bool)
	rows := make([]logRow, 0, len(m.events))
	for index, event := range m.events {
		if skipped[index] {
			continue
		}
		message := logField(event, "msg")
		if message == "command start" {
			key := commandRequestKey(event)
			if end, ok := ends[key]; key != "" && ok && end > index {
				endEvent := m.events[end]
				command := logField(event, "command")
				row := makeLogRow(event, command, durationValue(endEvent))
				skipped[end] = true
				rows = append(rows, row)
				continue
			}
		}
		rows = append(rows, makeLogRow(event, message, durationValue(event)))
	}
	return rows
}

func makeLogRow(event logs.Event, message, duration string) logRow {
	if event.Malformed() {
		message = logs.MalformedText(event)
	}
	parts := []string{logTime(event), logLevel(event), fieldOrDash(event, "component"), sanitizeText(message)}
	if duration != "" {
		parts = append(parts, duration)
	}
	return logRow{
		event: event, text: strings.Join(parts, " "),
		key: fmt.Sprintf("%s:%d", event.Path(), event.Line()), requestID: logField(event, "request_id"),
	}
}

func (m *Model) rowMatches(row logRow) bool {
	level := m.level
	if level == "debug" {
		level = ""
	}
	if !(logs.Filter{Level: level, Component: m.componentFilter, RequestID: m.requestFilter}).Match(row.event) {
		return false
	}
	if m.projectFilter != "" && logField(row.event, "project_id") != m.projectFilter {
		return false
	}
	if m.agentFilter != "" && logField(row.event, "agent_id") != m.agentFilter {
		return false
	}
	return m.textFilter == "" || strings.Contains(row.text, m.textFilter)
}

func (m *Model) logValues(field string) []string {
	found := map[string]bool{}
	for _, event := range m.events {
		if value := logField(event, field); value != "" {
			found[value] = true
		}
	}
	values := make([]string, 0, len(found))
	for value := range found {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func cycleValue(current string, values []string) string {
	if current == "" {
		if len(values) == 0 {
			return ""
		}
		return values[0]
	}
	for index, value := range values {
		if value == current {
			if index+1 < len(values) {
				return values[index+1]
			}
			return ""
		}
	}
	return ""
}

func (m *Model) moveLogSelection(delta int) {
	rows := m.filteredLogRows()
	if len(rows) == 0 {
		return
	}
	m.logSelection += delta
	if m.logSelection < 0 {
		m.logSelection = 0
	}
	if m.logSelection >= len(rows) {
		m.logSelection = len(rows) - 1
	}
	m.logSelectedKey = rows[m.logSelection].key
	m.rebuild()
}

func (m *Model) selectedLogRow() (logRow, bool) {
	rows := m.filteredLogRows()
	if m.logSelection < 0 || m.logSelection >= len(rows) {
		return logRow{}, false
	}
	return rows[m.logSelection], true
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
	return fmt.Sprintf("level: %s  component: %s  project: %s  agent: %s  text: %s  follow: %s  day: %s  request: %s",
		m.level, filterValue(m.componentFilter), filterValue(m.projectFilter), filterValue(m.agentFilter),
		sanitizeText(text), follow, m.logDate.Format("2006-01-02"), filterValue(m.requestFilter))
}

func (m *Model) bodyView() string {
	if m.logErr != "" {
		return "error: " + m.logErr
	}
	return m.viewport.View()
}

func (m *Model) detailView(width, height int) string {
	row, ok := m.selectedLogRow()
	if !ok {
		return ""
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, row.event.Raw, "", "  "); err != nil {
		pretty.WriteString(string(row.event.Raw))
	}
	if pretty.Len() == 0 {
		pretty.WriteString("{}")
	}
	lines := wrapJSON(pretty.String(), max(1, width-4))
	if len(lines) > height-2 && height > 2 {
		lines = lines[:height-2]
	}
	return strings.Join(append([]string{titleStyle.Render("event · " + logTime(row.event))}, lines...), "\n")
}

func wrapFilterLine(text string, width int) []string {
	if width < 1 {
		return []string{text}
	}
	words := strings.Fields(text)
	lines := make([]string, 0, 2)
	line := ""
	for _, word := range words {
		if line != "" && lipgloss.Width(line+" "+word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line == "" {
			line = word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func wrapJSON(text string, width int) []string {
	var wrapped []string
	for _, line := range strings.Split(text, "\n") {
		current := ""
		for _, character := range line {
			if current != "" && lipgloss.Width(current+string(character)) > width {
				wrapped = append(wrapped, current)
				current = ""
			}
			current += string(character)
		}
		wrapped = append(wrapped, current)
	}
	return wrapped
}

func filterValue(value string) string {
	if value == "" {
		return "all"
	}
	return sanitizeText(value)
}

func logField(event logs.Event, key string) string {
	if event.Fields == nil || event.Fields[key] == nil {
		return ""
	}
	return fmt.Sprint(event.Fields[key])
}

func fieldOrDash(event logs.Event, key string) string {
	if value := logField(event, key); value != "" {
		return sanitizeText(value)
	}
	return "-"
}

func logTime(event logs.Event) string {
	value, err := time.Parse(time.RFC3339Nano, logField(event, "time"))
	if err != nil {
		return "??:??:??"
	}
	return value.Local().Format("15:04:05")
}

func logLevel(event logs.Event) string {
	level := strings.ToUpper(logField(event, "level"))
	if level == "" {
		return "????"
	}
	return level
}

func durationValue(event logs.Event) string {
	value, ok := event.Fields["duration_ms"]
	if !ok {
		return ""
	}
	milliseconds, ok := value.(float64)
	if !ok {
		return ""
	}
	if milliseconds >= 1000 {
		return fmt.Sprintf("%.1fs", milliseconds/1000)
	}
	return fmt.Sprintf("%gms", milliseconds)
}

func commandRequestKey(event logs.Event) string {
	request := logField(event, "request_id")
	command := logField(event, "command")
	if request == "" || command == "" {
		return ""
	}
	return request + "\x00" + command
}

func eventsOnDay(events []logs.Event, day time.Time) []logs.Event {
	date := day.Local().Format("2006-01-02")
	filtered := make([]logs.Event, 0, len(events))
	for _, event := range events {
		when, err := time.Parse(time.RFC3339Nano, logField(event, "time"))
		if err == nil {
			if when.Local().Format("2006-01-02") == date {
				filtered = append(filtered, event)
			}
			continue
		}
		if filepath.Base(event.Path()) == date+".jsonl" {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func sameDay(left, right time.Time) bool {
	left, right = left.Local(), right.Local()
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}

func startOfDay(day time.Time) time.Time {
	day = day.Local()
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
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

func slogTime(key string, value time.Time) slog.Attr {
	return slog.String(key, value.Format(time.RFC3339))
}
