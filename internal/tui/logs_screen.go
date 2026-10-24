package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
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

var (
	accentTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	codeStyle       = lipgloss.NewStyle().Foreground(component.Code)
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

// logFilterFields are the typed fields of tui-v2 decision 25.
var logFilterFields = []string{"level", "comp", "proj", "agent", "task", "req"}

// unknownLogField gives the name of the first typed field of text that is
// not a field of decision 25, or empty.
func unknownLogField(text string) string {
	for _, word := range strings.Fields(text) {
		if name, value, found := strings.Cut(word, ":"); found && value != "" && !slices.Contains(logFilterFields, name) {
			return sanitizeText(name)
		}
	}
	return ""
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

// updateLogKey handles a key of the event list or the field panel that the
// frame gives to the screen (logs.md keys).
func (m *Model) updateLogKey(press tea.KeyPressMsg) tea.Cmd {
	panel := m.detailsFocused()
	switch key := press.String(); key {
	case "/":
		m.mode = editing
		m.filterDraft = m.textFilter
		return m.logAction("edit start")
	case "f":
		m.follow = !m.follow
		if m.follow {
			m.selectLastLogRow()
		}
		return m.logAction("follow", slog.Bool("on", m.follow))
	case "left", "right":
		step := -1
		if key == "right" {
			step = 1
		}
		m.logDate = startOfDay(m.logDate.AddDate(0, 0, step))
		m.logDatePinned = true
		m.events = nil
		m.logSelection = 0
		m.logSelectedKey = ""
		m.logOffset = 0
		m.detailsOffset[tabLogs] = 0
		return batch(m.logAction("log day", slogTime("day", m.logDate)), m.startCurrentScreenLoad())
	case "j":
		m.logRaw = !m.logRaw
		m.detailsOffset[tabLogs] = 0
		return m.logAction("raw json", slog.Bool("on", m.logRaw))
	case "r":
		row, ok := m.selectedLogRow()
		request := logField(row.event, "request_id")
		if !ok || request == "" {
			return nil
		}
		m.textFilter = "req:" + sanitizeText(request)
		m.refreshLogSelection()
		return m.logAction("filter request", slog.String("request_id", request))
	case "enter":
		if _, ok := m.selectedLogRow(); ok && !panel && m.width < component.LongFormWidth {
			m.regionFocus[tabLogs] = 1
			m.detailsOffset[tabLogs] = 0
			return m.logAction("event detail")
		}
		return nil
	case "esc", "escape":
		// The frame gives esc to the screen only to close the overlay of
		// the field panel below 120 columns.
		m.regionFocus[tabLogs] = 0
		return m.logAction("event detail close")
	}
	if panel {
		m.scrollLogPanel(press.String())
		return nil
	}
	switch press.String() {
	case "g":
		m.follow = false
		m.moveLogSelection(-len(m.events))
		return m.logAction("top")
	case "G":
		m.follow = true
		m.selectLastLogRow()
		return m.logAction("bottom")
	case "up":
		m.follow = false
		m.moveLogSelection(-1)
	case "down":
		m.moveLogSelection(1)
	case "pgup":
		m.follow = false
		m.moveLogSelection(-m.logPage())
	case "pgdown":
		m.moveLogSelection(m.logPage())
	}
	return nil
}

// scrollLogPanel scrolls the field panel by the key: one row, one page, or
// to the top or the end.
func (m *Model) scrollLogPanel(key string) {
	switch key {
	case "up":
		m.detailsOffset[tabLogs]--
	case "down":
		m.detailsOffset[tabLogs]++
	case "pgup":
		m.detailsOffset[tabLogs] -= m.logPage()
	case "pgdown":
		m.detailsOffset[tabLogs] += m.logPage()
	case "g":
		m.detailsOffset[tabLogs] = 0
	case "G":
		m.detailsOffset[tabLogs] = 1 << 30
	}
	m.detailsOffset[tabLogs] = max(m.detailsOffset[tabLogs], 0)
}

// updateLogFilter handles a key of the open filter line (logs.md rule 7):
// enter applies the text, and esc closes the line and keeps the applied
// filter. A field with a bad name keeps the line open and shows the error
// on the notice line.
func (m *Model) updateLogFilter(press tea.KeyPressMsg) tea.Cmd {
	switch press.String() {
	case "esc", "escape":
		m.mode = viewing
		m.filterDraft = ""
		m.logFilterErr = ""
		return m.logAction("edit cancel")
	case "enter":
		if name := unknownLogField(m.filterDraft); name != "" {
			m.logFilterErr = "unknown field " + name
			return m.logAction("filter refused", slog.String("field", name))
		}
		m.logFilterErr = ""
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
// across a refresh. When the load brings a new newest row, follow selects
// it. A fixture selection selects the event with that time.
func (m *Model) acceptLogEvents(events []logs.Event) {
	newest := ""
	if before := m.filteredLogRows(); len(before) > 0 {
		newest = before[len(before)-1].key
	}
	m.events = events
	rows := m.filteredLogRows()
	if m.logPendingTime != "" {
		for _, row := range rows {
			if logField(row.event, "time") == m.logPendingTime {
				m.logSelectedKey = row.key
			}
		}
		m.logPendingTime = ""
	} else if m.follow && (len(rows) == 0 || rows[len(rows)-1].key != newest) {
		m.logSelectedKey = ""
		if len(rows) > 0 {
			m.logSelectedKey = rows[len(rows)-1].key
		}
	}
	m.refreshLogSelection()
}

// refreshLogSelection finds the selected event in the filtered rows. When
// the filter hides it, the selection moves to the nearest row. With no rows,
// the selected event stays, so it shows again with the rows.
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

// The rows of the Logs content area above the boxes (logs.md layout): the
// filter line, a blank row, and the three rows of the density strip. A
// blank row follows the boxes.
const (
	logBoxTop = 5
	logBoxGap = 1
)

// The cells of the event list columns (logs.md rule 3): the marker and a
// blank, the time and a blank, the level and two blanks, and the component
// and two blanks. The message takes the cells that are left.
const (
	logLeadCells      = 2
	logTimeCells      = 9
	logLevelCells     = 7
	logComponentCells = 11
)

// logListHeader is the title row of the event list.
const logListHeader = "  TIME     LEVEL  COMP       MESSAGE"

// logFollowCells is the cells after the follow text of the filter line.
const logFollowCells = 4

// logsContent draws the content area of Logs (logs.md): the filter line,
// the density strip, and the event list with the field panel. From 120
// columns the list and the panel sit side by side. Below 120 columns the
// panel is a full-width overlay while it has the focus.
func (m *Model) logsContent(width, height int) string {
	focus := m.focusState()
	rows := m.filteredLogRows()
	strip := logDensity(rows, width-logStripMargin, m.logSelection, m.now().Location())
	if len(m.events) == 0 || m.logErr != "" {
		strip = logDensity(nil, width-logStripMargin, -1, m.now().Location())
	}
	top := m.logFilterLine(width) + "\n\n" + strip.View(width)
	boxHeight := max(height-logBoxTop-logBoxGap, component.PanelMinHeight)
	var body string
	switch {
	case width >= component.LongFormWidth:
		listWidth := listRegionWidth(tabLogs, width)
		list := strings.Split(m.logEventsPanel(rows, listWidth, boxHeight, focus == "events"), "\n")
		details := strings.Split(m.logFieldPanel(width-listWidth, boxHeight, focus == "json"), "\n")
		for index := range list {
			list[index] += details[index]
		}
		body = strings.Join(list, "\n")
	case m.detailsFocused():
		body = m.logFieldPanel(width, boxHeight, focus == "json")
	default:
		body = m.logEventsPanel(rows, width, boxHeight, focus == "events")
	}
	return fitBlock(top+"\n"+body, width, height)
}

// logsTitle gives the facts and the right text of the Logs title row
// (logs.md layout): the event count, the warn and error counts when more
// than 0, and the path of the day file, or the no file text.
func (m *Model) logsTitle() (string, string) {
	date := m.logDate.Format("2006-01-02")
	rows := m.makeLogRows()
	facts := fmt.Sprintf("· %d events", len(rows))
	if len(rows) == 0 && m.logErr == "" {
		return facts + " · no log file for " + date, "← " + date + " → day"
	}
	warn, errs := 0, 0
	for _, row := range rows {
		switch logLevel(row.event) {
		case "WARN":
			warn++
		case "ERROR":
			errs++
		}
	}
	if warn > 0 {
		facts += fmt.Sprintf(" · %d warn", warn)
	}
	if errs > 0 {
		facts += fmt.Sprintf(" · %d error", errs)
	}
	return facts + " · " + m.homeLabel() + "/.hive/logs/" + date + ".jsonl", "← " + date + " → day"
}

// logFilterLine gives the filter line of logs.md rule 7: / in accent bold,
// the filter text, the cursor while the line is open, and the follow state
// at the right.
func (m *Model) logFilterLine(width int) string {
	text := m.textFilter
	if m.mode == editing {
		text = m.filterDraft + " █"
	}
	mark := dimStyle.Render("○")
	if m.follow {
		mark = workingStyle.Render("●")
	}
	right := dimStyle.Render("follow ") + mark + dimStyle.Render("  f toggles follow") + strings.Repeat(" ", logFollowCells)
	room := width - lipgloss.Width(right) - 4
	if room < 1 {
		right, room = "", width-3
	}
	if m.mode == editing {
		text = ansi.TruncateLeft(text, max(lipgloss.Width(text)-room, 0), "…")
	} else {
		text = ansi.Truncate(text, room, "…")
	}
	line := " " + titleStyle.Render("/") + " " + text
	return fitLine(line, width-lipgloss.Width(right)) + right
}

// logEventsPanel draws the event list of logs.md rule 3. The list keeps the
// selected row visible. A log error, a day with no event, and a filter with
// no match each take the place of the rows.
func (m *Model) logEventsPanel(rows []logRow, width, height int, focused bool) string {
	visible := max(height-3, 1)
	var lines []string
	switch {
	case m.logErr != "":
		lines = []string{"", "error: " + m.logErr}
		rows = nil
	case len(m.events) == 0:
		lines = []string{"", dimStyle.Render("no log file for " + m.logDate.Format("2006-01-02"))}
	case len(rows) == 0:
		lines = []string{"", dimStyle.Render("no event matches " + sanitizeText(m.textFilter))}
	default:
		lines = []string{dimStyle.Bold(true).Render(logListHeader)}
	}
	offset := min(m.logOffset, m.logSelection)
	offset = max(offset, m.logSelection-visible+1)
	offset = min(max(offset, 0), max(len(rows)-visible, 0))
	m.logOffset = offset
	message := max(width-4-logLeadCells-logTimeCells-logLevelCells-logComponentCells, 1)
	for index := offset; index < min(offset+visible, len(rows)); index++ {
		lines = append(lines, m.logRowLine(rows[index], index == m.logSelection, message))
	}
	title := fmt.Sprintf("events · %d", len(rows))
	if len(rows) > 0 {
		title += " · newest at the end"
	}
	panel := component.Panel{Title: title, Focused: focused, Lines: lines, VMark: component.Thumb(height-2, len(rows), visible, offset)}
	return panel.View(width, height)
}

// logRowLine gives one row of the event list (logs.md rule 3). The time
// and the component are dim. A DEBUG row is dim. The level of a WARN row
// is accent bold and of an ERROR row error bold. The selected row starts
// with ▶ and has its time and message in accent bold.
func (m *Model) logRowLine(row logRow, selected bool, messageCells int) string {
	level := logLevel(row.event)
	lead, timeStyle, levelStyle, messageStyle := "  ", dimStyle, lipgloss.NewStyle(), lipgloss.NewStyle()
	switch level {
	case "DEBUG":
		levelStyle, messageStyle = dimStyle, dimStyle
	case "WARN":
		levelStyle = titleStyle
	case "ERROR":
		levelStyle = errorTextStyle.Bold(true)
	}
	if selected {
		lead, timeStyle, messageStyle = titleStyle.Render("▶")+" ", titleStyle, titleStyle
	}
	return lead +
		timeStyle.Render(logTime(row.event, m.now().Location())) + " " +
		levelStyle.Render(logCell(level, logLevelCells-2)) + "  " +
		dimStyle.Render(logCell(fieldOrDash(row.event, "component"), logComponentCells-2)) + "  " +
		messageStyle.Render(logCell(row.text, messageCells))
}

// logFieldPanel draws the selected event in the field panel: its fields
// (logs.md rule 4) or its raw JSON (rule 5). The lines scroll on their own
// from the details offset of Logs. The key hint sits on the row above the
// last body row. Below 120 columns the panel is an overlay, and its title
// ends with "esc back".
func (m *Model) logFieldPanel(width, height int, focused bool) string {
	row, ok := m.selectedLogRow()
	if !ok || m.logErr != "" {
		return component.Panel{Title: "event", Focused: focused}.View(width, height)
	}
	room := max(width-4, 1)
	title := "event " + logTime(row.event, m.now().Location())
	hint := "j raw JSON · r filter by this request"
	var lines []string
	if m.logRaw {
		title += " · raw"
		hint = "j fields · r filter by this request"
		lines = logRawLines(row.event.Raw, room)
	} else {
		title += " · " + logLevel(row.event)
		lines = logFieldLines(row.event.Raw, room-logLabelCells)
	}
	body := max(height-2, 1)
	region := max(body-2, 1)
	offset := min(max(m.detailsOffset[tabLogs], 0), max(len(lines)-region, 0))
	m.detailsOffset[tabLogs] = offset
	shown := make([]string, body)
	copy(shown, lines[offset:min(offset+region, len(lines))])
	if body > region {
		shown[region] = dimStyle.Render(hint)
	}
	if width < component.LongFormWidth {
		title += " · esc back"
	}
	panel := component.Panel{Title: title, Focused: focused, Lines: shown, VMark: component.Thumb(body, len(lines), region, offset)}
	return panel.View(width, height)
}

// logRawLines gives the raw JSON of an event with two spaces of indent, in
// code. A line that is wider than width wraps at a word, and each next row
// indents by two more cells.
func logRawLines(raw []byte, width int) []string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		pretty.Reset()
		pretty.Write(raw)
	}
	if pretty.Len() == 0 {
		pretty.WriteString("{}")
	}
	var lines []string
	for line := range strings.SplitSeq(pretty.String(), "\n") {
		line = sanitizeText(line)
		text := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(text)]
		first := wrapLogValue(text, max(width-len(indent), 1))[0]
		lines = append(lines, codeStyle.Render(indent+first))
		if rest := strings.TrimPrefix(text[len(first):], " "); rest != "" {
			for _, next := range wrapLogValue(rest, max(width-len(indent)-2, 1)) {
				lines = append(lines, codeStyle.Render(indent+"  "+next))
			}
		}
	}
	return lines
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
