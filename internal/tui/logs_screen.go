package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

const maxErrorBytes = 300

// defaultLogFilter is the applied filter of a new Logs screen. It hides the
// debug events, as the old level filter did.
const defaultLogFilter = "level:info"

// The cells of the event list columns: the marker, the time, the level, and
// the component, each with its gap. The message takes the cells that are
// left.
const (
	logMarkerCells    = 2
	logTimeCells      = 9
	logLevelCells     = 6
	logComponentCells = 10
)

// The cells after the status of the filter bar, from 120 columns and below
// (the Logs and CompactLogs boards).
const (
	logStatusGapWide    = 10
	logStatusGapCompact = 6
)

var (
	accentTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
)

type logsScreen struct{ model *Model }

func (s logsScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if s.model.mode == editing {
		return s.model.updateLogFilter(press)
	}
	return s.model.updateLogKey(press)
}

// View is not used: the frame draws the Logs content with logsContent.
func (s logsScreen) View(width, height int) string {
	return s.model.logsContent(width, height)
}

func (s logsScreen) KeyHints() string { return "" }

func (s logsScreen) DataRequest() dataRequest { return logsDataRequest }

// logRow is one row of the event list. A command start and its command end
// fold to one row.
type logRow struct {
	event logs.Event
	text  string
	key   string
}

// logFilter is the parsed filter bar text (decision 25).
type logFilter struct {
	match logs.Filter
	proj  string
	agent string
	words []string
}

func parseLogFilter(text string) logFilter {
	var f logFilter
	for _, word := range strings.Fields(text) {
		name, value, found := strings.Cut(word, ":")
		if !found || value == "" {
			f.words = append(f.words, word)
			continue
		}
		switch name {
		case "level":
			f.match.Level = strings.ToLower(value)
		case "comp":
			f.match.Component = value
		case "proj":
			f.proj = value
		case "agent":
			f.agent = value
		case "task":
			f.match.TaskID = value
		case "req":
			f.match.RequestID = value
		default:
			f.words = append(f.words, word)
		}
	}
	if f.match.Level == "debug" {
		f.match.Level = ""
	}
	return f
}

// matches reports that row passes the filter. An agent value matches the
// agent ID or the agent name before its generation, for example
// manager-hive for manager-hive-g2.
func (f logFilter) matches(row logRow) bool {
	if !f.match.Match(row.event) {
		return false
	}
	if f.proj != "" && logField(row.event, "project_id") != f.proj {
		return false
	}
	if agent := logField(row.event, "agent_id"); f.agent != "" && agent != f.agent && !strings.HasPrefix(agent, f.agent+"-g") {
		return false
	}
	for _, word := range f.words {
		if !strings.Contains(row.text, word) {
			return false
		}
	}
	return true
}

// updateLogKey handles a key of the event list or the JSON panel that the
// frame gives to the screen.
func (m *Model) updateLogKey(press tea.KeyPressMsg) tea.Cmd {
	jsonFocused := m.detailsFocused()
	switch press.String() {
	case "/":
		m.mode = editing
		m.filterDraft = m.textFilter
		return m.logAction("edit start")
	case "f":
		m.follow = !m.follow
		if m.follow {
			m.selectLastLogRow()
		}
		return m.logAction("follow")
	case "[", "]":
		// The frame gives < and > to the screen as [ and ].
		step := -1
		if press.String() == "]" {
			step = 1
		}
		m.logDate = startOfDay(m.logDate.AddDate(0, 0, step))
		m.logDatePinned = true
		m.events = nil
		m.logSelection = 0
		m.logSelectedKey = ""
		m.logOffset = 0
		return batch(m.logAction("log day", slogTime("day", m.logDate)), m.startCurrentScreenLoad())
	case "enter":
		if _, ok := m.selectedLogRow(); ok && !jsonFocused {
			m.regionFocus[tabLogs] = 1
			m.detailsOffset[tabLogs] = 0
			return m.logAction("event detail")
		}
		return nil
	case "esc", "escape":
		m.textFilter = defaultLogFilter
		m.refreshLogSelection()
		return m.logAction("filter clear")
	case "g":
		if jsonFocused {
			m.detailsOffset[tabLogs] = 0
			return nil
		}
		m.follow = false
		m.moveLogSelection(-len(m.events))
		return m.logAction("top")
	case "G":
		if jsonFocused {
			m.detailsOffset[tabLogs] = 1 << 30
			return nil
		}
		m.follow = true
		m.selectLastLogRow()
		return m.logAction("bottom")
	case "up", "k":
		m.follow = false
		m.moveLogSelection(-1)
	case "down", "j":
		m.moveLogSelection(1)
	case "pgup":
		m.follow = false
		m.moveLogSelection(-m.logPage())
	case "pgdown":
		m.moveLogSelection(m.logPage())
	}
	return nil
}

// updateLogFilter handles a key of the filter bar (decision 21): enter
// applies the text, and esc closes the bar and keeps the applied filter.
func (m *Model) updateLogFilter(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "esc", "escape":
		m.mode = viewing
		m.filterDraft = ""
		return m.logAction("edit cancel")
	case "enter":
		m.textFilter = m.filterDraft
		m.mode = viewing
		m.refreshLogSelection()
		return m.logAction("filter apply")
	case "backspace":
		if _, size := utf8.DecodeLastRuneInString(m.filterDraft); size > 0 {
			m.filterDraft = m.filterDraft[:len(m.filterDraft)-size]
		}
		return nil
	}
	if press.Text != "" {
		m.filterDraft += sanitizeText(press.Text)
	}
	return nil
}

// pasteLogFilter adds pasted text to the open filter bar.
func (m *Model) pasteLogFilter(text string) {
	if m.activeTab() == tabLogs && m.mode == editing {
		m.filterDraft += sanitizeText(text)
		m.invalidateView()
	}
}

// acceptLogEvents takes the events of a load. The selection keeps its event
// across a refresh. Follow selects the newest row, and a fixture selection
// selects the event with that time.
func (m *Model) acceptLogEvents(events []logs.Event) {
	m.events = events
	rows := m.filteredLogRows()
	if m.logPendingTime != "" {
		for _, row := range rows {
			if logField(row.event, "time") == m.logPendingTime {
				m.logSelectedKey = row.key
			}
		}
		m.logPendingTime = ""
	} else if m.follow {
		m.logSelectedKey = ""
		if len(rows) > 0 {
			m.logSelectedKey = rows[len(rows)-1].key
		}
	}
	m.refreshLogSelection()
}

// refreshLogSelection finds the selected event in the filtered rows. When
// the filter hides it, the selection moves to the nearest row.
func (m *Model) refreshLogSelection() {
	rows := m.filteredLogRows()
	for index, row := range rows {
		if row.key == m.logSelectedKey {
			m.logSelection = index
			return
		}
	}
	if len(rows) == 0 {
		m.logSelection = 0
		m.logSelectedKey = ""
		return
	}
	m.logSelection = min(max(m.logSelection, 0), len(rows)-1)
	m.logSelectedKey = rows[m.logSelection].key
}

func (m *Model) moveLogSelection(delta int) {
	rows := m.filteredLogRows()
	if len(rows) == 0 {
		return
	}
	m.logSelection = min(max(m.logSelection+delta, 0), len(rows)-1)
	m.logSelectedKey = rows[m.logSelection].key
}

func (m *Model) selectLastLogRow() {
	m.moveLogSelection(len(m.events))
}

// logPage gives the event rows of one page at the current size.
func (m *Model) logPage() int {
	return max(m.height-frameRows-3-3, 1)
}

func (m *Model) selectedLogRow() (logRow, bool) {
	rows := m.filteredLogRows()
	if m.logSelection < 0 || m.logSelection >= len(rows) {
		return logRow{}, false
	}
	return rows[m.logSelection], true
}

func (m *Model) filteredLogRows() []logRow {
	filter := parseLogFilter(m.textFilter)
	all := m.makeLogRows()
	rows := make([]logRow, 0, len(all))
	for _, row := range all {
		if filter.matches(row) {
			rows = append(rows, row)
		}
	}
	return rows
}

// makeLogRows gives one row for each event. A command start with its
// command end of the same request folds to one row with the command. An
// event with a reason adds " · " and the reason text before the first ";".
func (m *Model) makeLogRows() []logRow {
	ends := make(map[string]int)
	for index, event := range m.events {
		if logField(event, "msg") == "command end" {
			if key := commandRequestKey(event); key != "" {
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
				skipped[end] = true
				message = logField(event, "command")
			}
		}
		if reason, _, _ := strings.Cut(logField(event, "reason"), ";"); reason != "" {
			message += " · " + reason
		}
		if event.Malformed() {
			message = logs.MalformedText(event)
		}
		key := fmt.Sprintf("%s:%d", event.Path(), event.Line())
		if event.Path() == "" {
			key = fmt.Sprintf("#%d", index)
		}
		rows = append(rows, logRow{event: event, text: sanitizeText(message), key: key})
	}
	return rows
}

// logsContent draws the content area of Logs (decision 20): the filter bar,
// then the event list and the JSON panel side by side from 120 columns.
// Below 120 columns the JSON panel is a full-width overlay (decision 9).
func (m *Model) logsContent(width, height int) string {
	focus := m.focusState()
	wide := width >= component.LongFormWidth
	status := m.logDayStatus()
	gap := logStatusGapCompact
	if wide {
		status = dimStyle.Render("follow") + " " + m.followMark() + "  " + status
		gap = logStatusGapWide
	}
	text := m.textFilter
	if m.mode == editing {
		text = m.filterDraft
	}
	bar := component.FilterBar{Text: text, Editing: m.mode == editing, Focused: focus == "filter", Status: status + strings.Repeat(" ", gap)}.View(width)
	bodyHeight := max(height-3, component.PanelMinHeight)
	var body string
	switch {
	case wide:
		listWidth := width * 8 / 15
		list := strings.Split(m.logEventsPanel(listWidth, bodyHeight, focus == "events"), "\n")
		details := strings.Split(m.logJSONPanel(width-listWidth, bodyHeight, focus == "json", "", true), "\n")
		for index := range list {
			list[index] += details[index]
		}
		body = strings.Join(list, "\n")
	case m.detailsFocused():
		body = m.logJSONPanel(width, bodyHeight, focus == "json", " · esc back", false)
	default:
		body = m.logEventsPanel(width, bodyHeight, focus == "events")
	}
	return fitBlock(bar+"\n"+body, width, height)
}

func (m *Model) logDayStatus() string {
	return titleStyle.Render("<") + " " + m.logDate.Format("2006-01-02") + " " + titleStyle.Render(">")
}

// followMark gives ● when follow is on and ○ when it is off.
func (m *Model) followMark() string {
	if m.follow {
		return workingStyle.Render("●")
	}
	return dimStyle.Render("○")
}

// logEventsPanel draws the event list. The first body row is blank, as the
// header row of a list with no column titles. The list keeps the selected
// row visible and shows no empty row without need. A log error takes the
// place of the rows.
func (m *Model) logEventsPanel(width, height int, focused bool) string {
	rows := m.filteredLogRows()
	visible := max(height-3, 1)
	lines := []string{""}
	if m.logErr != "" {
		// The log error takes the place of the rows, as on the old screen.
		lines = append(lines, "error: "+m.logErr)
		rows = nil
	} else if len(rows) == 0 {
		message := "no events on " + m.logDate.Format("2006-01-02")
		if sameDay(m.logDate, m.now()) {
			message = "no events today"
		}
		lines = append(lines, dimStyle.Render(message))
	}
	offset := min(m.logOffset, m.logSelection)
	offset = max(offset, m.logSelection-visible+1)
	offset = min(max(offset, 0), max(len(rows)-visible, 0))
	m.logOffset = offset
	message := max(width-4-logMarkerCells-logTimeCells-logLevelCells-logComponentCells, 1)
	for index := offset; index < min(offset+visible, len(rows)); index++ {
		lines = append(lines, m.logRowLine(rows[index], index == m.logSelection, message))
	}
	mark := component.Thumb(visible, len(rows), visible, offset)
	if mark.Size > 0 {
		mark.Start++ // the track starts below the blank first row
	}
	panel := component.Panel{Title: fmt.Sprintf("events · %d", len(rows)), Focused: focused, Lines: lines, VMark: mark}
	return panel.View(width, height)
}

// logRowLine gives one row of the event list. The selected row has the
// marker ▶, its time in accent, and its message in accent and bold. A warn
// or error level is accent and bold. Each column style covers its padding,
// as on the Logs board.
func (m *Model) logRowLine(row logRow, selected bool, messageCells int) string {
	marker, timeStyle, messageStyle := " ", dimStyle, lipgloss.NewStyle()
	if selected {
		marker, timeStyle, messageStyle = titleStyle.Render("▶"), accentTextStyle, titleStyle
	}
	level := logLevel(row.event)
	levelStyle := dimStyle
	if level == "WARN" || level == "ERROR" {
		levelStyle = titleStyle
	}
	return marker + " " +
		timeStyle.Render(logCell(logTime(row.event, m.now().Location()), logTimeCells-1)) + " " +
		levelStyle.Render(logCell(level, logLevelCells)) +
		infoStyle.Render(logCell(fieldOrDash(row.event, "component"), logComponentCells)) +
		messageStyle.Render(logCell(row.text, messageCells))
}

// logJSONPanel draws the JSON of the selected event. The panel scrolls on
// its own from the details offset of Logs (decision 20). When the lines do
// not fit, the last body row shows how many lines are below it. From 120
// columns the first body row is blank, level with the blank first row of the
// event list.
func (m *Model) logJSONPanel(width, height int, focused bool, suffix string, blankFirst bool) string {
	title, lines := "event", []string{dimStyle.Render("no event selected")}
	if row, ok := m.selectedLogRow(); ok {
		title = "event " + logTime(row.event, m.now().Location())
		lines = logJSONLines(row.event, max(width-6, 1))
	}
	if blankFirst {
		lines = append([]string{""}, lines...)
	}
	body := max(height-2, 1)
	offset := min(max(m.detailsOffset[tabLogs], 0), max(len(lines)-body, 0))
	m.detailsOffset[tabLogs] = offset
	shown := lines[offset:]
	if len(shown) > body {
		more := len(shown) - body + 1
		shown = append(shown[:body-1:body-1], dimStyle.Render(fmt.Sprintf("↓ %d more lines", more)))
	}
	panel := component.Panel{Title: title + suffix, Focused: focused, Lines: shown, VMark: component.Thumb(body, len(lines), body, offset)}
	return panel.View(width, height)
}

// logJSONLines gives the event JSON with two spaces of indent and each key
// in info. Each line is cut to width cells with ….
func logJSONLines(event logs.Event, width int) []string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, event.Raw, "", "  "); err != nil {
		pretty.Reset()
		pretty.Write(event.Raw)
	}
	if pretty.Len() == 0 {
		pretty.WriteString("{}")
	}
	var lines []string
	for line := range strings.SplitSeq(pretty.String(), "\n") {
		line = ansi.Truncate(sanitizeText(line), width, "…")
		lines = append(lines, styleJSONKey(line))
	}
	return lines
}

// styleJSONKey styles the key of a "key": value line in info.
func styleJSONKey(line string) string {
	body := strings.TrimLeft(line, " ")
	indent := line[:len(line)-len(body)]
	if !strings.HasPrefix(body, `"`) {
		return line
	}
	end := strings.Index(body[1:], `":`)
	if end < 0 {
		return line
	}
	key := body[:end+2]
	return indent + infoStyle.Render(key) + body[len(key):]
}

// logCell cuts text to width cells with … and pads it to width cells. A
// column with a gap after it keeps the gap: its text has width-1 cells.
func logCell(text string, width int) string {
	return padCell(ansi.Truncate(text, width, "…"), width)
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

// logTime gives the time of the event in the clock zone.
func logTime(event logs.Event, zone *time.Location) string {
	value, err := time.Parse(time.RFC3339Nano, logField(event, "time"))
	if err != nil {
		return "??:??:??"
	}
	return value.In(zone).Format("15:04:05")
}

func logLevel(event logs.Event) string {
	level := strings.ToUpper(logField(event, "level"))
	if level == "" {
		return "????"
	}
	return level
}

func commandRequestKey(event logs.Event) string {
	request := logField(event, "request_id")
	command := logField(event, "command")
	if request == "" || command == "" {
		return ""
	}
	return request + "\x00" + command
}

// eventsOnDay gives the events of day, in the zone of day.
func eventsOnDay(events []logs.Event, day time.Time) []logs.Event {
	date := day.Format("2006-01-02")
	filtered := make([]logs.Event, 0, len(events))
	for _, event := range events {
		when, err := time.Parse(time.RFC3339Nano, logField(event, "time"))
		if err == nil {
			if when.In(day.Location()).Format("2006-01-02") == date {
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

// sameDay reports that left and right are on the same date in the zone of
// left.
func sameDay(left, right time.Time) bool {
	return left.Format("2006-01-02") == right.In(left.Location()).Format("2006-01-02")
}

// startOfDay gives midnight of day in the zone of day.
func startOfDay(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
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
