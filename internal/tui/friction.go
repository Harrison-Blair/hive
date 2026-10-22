package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

type frictionTab uint8

const (
	frictionOpen frictionTab = iota
	frictionResolved
)

type frictionSection struct {
	title string
	body  []string
}

type frictionEntry struct {
	file     string
	path     string
	title    string
	status   string
	opened   string
	decision string
	body     string
	sections []frictionSection
	valid    bool
	resolved bool
}

type frictionData struct {
	entries []frictionEntry
}

// frictionState is the state of the Friction screen (decision 17). An open
// entry file shows in the details region and scrolls from fileOffset.
type frictionState struct {
	data         frictionData
	tab          frictionTab
	selectedPath string
	listOffset   int
	fileOpen     bool
	fileLoading  bool
	fileTitle    string
	fileContent  string
	fileOffset   int
}

type frictionFileMsg struct {
	title   string
	content string
	err     error
}

var frictionFrontMatterKey = regexp.MustCompile(`^([a-z][a-z0-9_-]*):(?:\s+(.*))?$`)
var frictionSectionHeader = regexp.MustCompile(`^\*\*(.+):\*\*$`)

func loadFriction(home string) (frictionData, error) {
	root := filepath.Join(home, "reference", "dogfood", "friction")
	if _, err := os.ReadFile(filepath.Join(root, "FRICTION.md")); err != nil {
		return frictionData{}, fmt.Errorf("friction index: %w", err)
	}
	data := frictionData{entries: make([]frictionEntry, 0)}
	for _, dir := range []struct {
		name     string
		resolved bool
	}{{name: "entries"}, {name: filepath.Join("entries", "resolved"), resolved: true}} {
		path := filepath.Join(root, dir.name)
		files, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return frictionData{}, err
		}
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
				continue
			}
			entryPath := filepath.Join(path, file.Name())
			contents, err := os.ReadFile(entryPath)
			if err != nil {
				return frictionData{}, fmt.Errorf("friction entry %s: %w", file.Name(), err)
			}
			entry := parseFrictionEntry(file.Name(), filepath.Join(dir.name, file.Name()), string(contents), dir.resolved)
			data.entries = append(data.entries, entry)
		}
	}
	sortFrictionEntries(data.entries)
	return data, nil
}

func sortFrictionEntries(entries []frictionEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].opened == entries[j].opened {
			return entries[i].file < entries[j].file
		}
		return entries[i].opened > entries[j].opened
	})
}

func parseFrictionEntry(file, path, contents string, resolved bool) frictionEntry {
	entry := frictionEntry{file: file, path: path, resolved: resolved}
	lines := strings.Split(contents, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return entry
	}
	closing := -1
	fields := make(map[string]string)
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			closing = index
			break
		}
		match := frictionFrontMatterKey.FindStringSubmatch(lines[index])
		if len(match) != 3 {
			return entry
		}
		if _, found := fields[match[1]]; found {
			return entry
		}
		value, ok := parseFrictionValue(match[2])
		if !ok {
			return entry
		}
		fields[match[1]] = value
	}
	if closing < 0 || fields["title"] == "" || fields["status"] == "" || fields["opened"] == "" {
		return entry
	}
	if _, found := fields["decision"]; !found {
		return entry
	}
	if _, err := time.Parse("2006-01-02", fields["opened"]); err != nil {
		return entry
	}
	entry.title = fields["title"]
	entry.status = fields["status"]
	entry.opened = fields["opened"]
	entry.decision = fields["decision"]
	entry.body = strings.Join(lines[closing+1:], "\n")
	entry.sections = frictionSections(strings.Split(entry.body, "\n"))
	entry.valid = true
	return entry
}

func parseFrictionValue(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, `"`) {
		unquoted, err := strconv.Unquote(value)
		return unquoted, err == nil
	}
	if strings.HasPrefix(value, "'") {
		if len(value) < 2 || !strings.HasSuffix(value, "'") {
			return "", false
		}
		return strings.ReplaceAll(value[1:len(value)-1], "''", "'"), true
	}
	return value, true
}

// frictionSections splits the body into its "**Title:**" sections. The blank
// lines at the start and the end of each section are dropped.
func frictionSections(lines []string) []frictionSection {
	sections := make([]frictionSection, 0)
	for _, line := range lines {
		match := frictionSectionHeader.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 2 {
			sections = append(sections, frictionSection{title: match[1]})
			continue
		}
		if len(sections) != 0 {
			sections[len(sections)-1].body = append(sections[len(sections)-1].body, line)
		}
	}
	for index := range sections {
		body := sections[index].body
		for len(body) > 0 && strings.TrimSpace(body[0]) == "" {
			body = body[1:]
		}
		for len(body) > 0 && strings.TrimSpace(body[len(body)-1]) == "" {
			body = body[:len(body)-1]
		}
		sections[index].body = body
	}
	return sections
}

// name gives the entry name of the list: the file name without ".md".
func (e frictionEntry) name() string { return strings.TrimSuffix(e.file, ".md") }

// resolvedDate gives the date of a "Resolved YYYY-MM-DD" status, or the
// opened date when the status has none.
func (e frictionEntry) resolvedDate() string {
	if date, ok := strings.CutPrefix(e.status, "Resolved "); ok {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(date)); err == nil {
			return strings.TrimSpace(date)
		}
	}
	return e.opened
}

// section gives the lines of the section with the title, in any case.
func (e frictionEntry) section(title string) []string {
	for _, section := range e.sections {
		if strings.EqualFold(section.title, title) {
			return section.body
		}
	}
	return nil
}

// seen gives the number of occurrences and the last date. An entry with no
// Occurrences section was seen one time, on its opened date.
func (e frictionEntry) seen() (int, string) {
	count, last := 0, e.opened
	for _, line := range e.section("Occurrences") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++
		if date, _, _ := strings.Cut(strings.TrimSpace(line), " "); len(date) == 10 && date > last {
			if _, err := time.Parse("2006-01-02", date); err == nil {
				last = date
			}
		}
	}
	return max(count, 1), last
}

func (s *frictionState) accept(data frictionData) {
	s.data = data
	s.acceptSelection()
}

// visibleEntries gives the entries of the view. Open entries keep the order
// of the data: the newest opened date first. Resolved entries show the
// newest resolved date first, then by name.
func (s frictionState) visibleEntries() []frictionEntry {
	entries := make([]frictionEntry, 0, len(s.data.entries))
	for _, entry := range s.data.entries {
		if (s.tab == frictionResolved) == entry.resolved {
			entries = append(entries, entry)
		}
	}
	if s.tab == frictionResolved {
		sort.SliceStable(entries, func(i, j int) bool {
			if a, b := entries[i].resolvedDate(), entries[j].resolvedDate(); a != b {
				return a > b
			}
			return entries[i].file < entries[j].file
		})
	}
	return entries
}

func (s frictionState) selectedIndex() int {
	for index, entry := range s.visibleEntries() {
		if entry.path == s.selectedPath {
			return index
		}
	}
	return 0
}

func (s frictionState) selected() (frictionEntry, bool) {
	for _, entry := range s.visibleEntries() {
		if entry.path == s.selectedPath {
			return entry, true
		}
	}
	return frictionEntry{}, false
}

// acceptSelection keeps the selected entry when the view has it, and selects
// the first entry when it does not.
func (s *frictionState) acceptSelection() {
	if _, ok := s.selected(); ok {
		return
	}
	entries := s.visibleEntries()
	s.selectedPath = ""
	if len(entries) != 0 {
		s.selectedPath = entries[0].path
	}
	s.listOffset = 0
}

// moveSelection moves the selection by delta entries, within the entries.
func (s *frictionState) moveSelection(delta int) {
	entries := s.visibleEntries()
	if len(entries) == 0 {
		s.selectedPath = ""
		return
	}
	index := min(max(s.selectedIndex()+delta, 0), len(entries)-1)
	s.selectedPath = entries[index].path
}

type frictionScreen struct{ model *Model }

func (frictionScreen) DataRequest() dataRequest { return frictionDataRequest }

func (frictionScreen) KeyHints() string { return "" }

// Update takes the list keys of the screen (decisions 21 to 23). The frame
// takes the other keys. An open entry file takes the scroll keys, and esc
// closes it.
func (s frictionScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	state := &m.friction
	key := press.String()
	page := max(m.height-frameRows-1-3, 1)
	if state.fileOpen {
		switch key {
		case "esc":
			state.fileOpen = false
			state.fileLoading = false
			state.fileContent = ""
			return m.logAction("friction file close", slog.String("file", state.fileTitle))
		case "up", "k":
			state.fileOffset = max(state.fileOffset-1, 0)
		case "down", "j":
			state.fileOffset++
		case "pgup":
			state.fileOffset = max(state.fileOffset-page, 0)
		case "pgdown":
			state.fileOffset += page
		case "g":
			state.fileOffset = 0
		case "G":
			state.fileOffset = 1 << 16
		}
		return nil
	}
	delta := 0
	switch key {
	case "up", "k":
		delta = -1
	case "down", "j":
		delta = 1
	case "pgup":
		delta = -page
	case "pgdown":
		delta = page
	case "g":
		delta = -len(state.data.entries)
	case "G":
		delta = len(state.data.entries)
	case "enter":
		if selected, ok := state.selected(); ok {
			path := filepath.Join(m.home, "reference", "dogfood", "friction", selected.path)
			read := m.startFrictionFileRead(state, path, selected.file)
			return batch(m.logAction("friction entry open", slog.String("file", selected.file)), read)
		}
		return nil
	default:
		return nil
	}
	previous := state.selectedPath
	state.moveSelection(delta)
	if previous == state.selectedPath {
		return nil
	}
	m.detailsOffset[tabFriction] = 0
	return m.logAction("friction selection", slog.String("file", state.selectedPath))
}

// View gives the list. The frame draws the screen with frictionPanels.
func (s frictionScreen) View(width, height int) string {
	return s.model.frictionList(width, height, true)
}

// frictionPanels gives the list and the details regions of the screen.
func frictionPanels(m *Model) (func(int, int, bool) string, func(int, int, bool, string) string) {
	return m.frictionList, m.frictionDetails
}

// frictionList draws the entries of the view. From 120 columns the date is
// MM-DD. Below 120 columns the list shows the full date, the seen count, and
// the hint that enter opens the details.
func (m *Model) frictionList(width, height int, focused bool) string {
	state := &m.friction
	entries := state.visibleEntries()
	view, dateTitle := "open", "OPENED"
	if state.tab == frictionResolved {
		view, dateTitle = "resolved", "RESOLVED"
	}
	wide := m.width >= component.LongFormWidth
	list := component.List{
		Title:    fmt.Sprintf("%s · %d", view, len(entries)),
		Focused:  focused,
		Columns:  []component.Column{{Title: "ENTRY", Width: 34}, {Title: dateTitle}},
		Selected: state.selectedIndex(),
		Offset:   state.listOffset,
	}
	if !wide {
		list.Columns = []component.Column{{Title: "ENTRY", Width: 38}, {Title: dateTitle, Width: 12}, {Title: "SEEN"}}
		list.Hint = "enter opens details"
	}
	dim := lipgloss.NewStyle().Foreground(component.Dim)
	for _, entry := range entries {
		date := entry.opened
		if entry.resolved {
			date = entry.resolvedDate()
		}
		if !entry.valid {
			date = "?"
		}
		row := []component.Cell{{Text: entry.name()}}
		if wide {
			if len(date) == len("2006-01-02") {
				date = date[5:]
			}
			row = append(row, component.Cell{Text: date, Style: dim})
		} else {
			count, _ := entry.seen()
			row = append(row, component.Cell{Text: date, Style: dim}, component.Cell{Text: strconv.Itoa(count), Style: dim})
		}
		list.Rows = append(list.Rows, row)
	}
	list.Move(0, height)
	state.listOffset = list.Offset
	return list.View(width, height)
}

// frictionDetails draws the selected entry, or the open entry file.
func (m *Model) frictionDetails(width, height int, focused bool, suffix string) string {
	state := &m.friction
	if state.fileOpen {
		text := state.fileContent
		if state.fileLoading {
			text = "loading…"
		}
		lines := wrapLines(strings.Split(text, "\n"), width-4)
		state.fileOffset = min(state.fileOffset, max(len(lines)-(height-2), 0))
		return component.Details{Title: state.fileTitle + " · esc closes", Focused: focused, Lines: lines, Offset: state.fileOffset}.View(width, height)
	}
	title, lines := "details", []string{"", "no entry selected"}
	if entry, ok := state.selected(); ok {
		title, lines = entry.name(), frictionDetailLines(entry, width)
	}
	offset := &m.detailsOffset[tabFriction]
	*offset = min(max(*offset, 0), max(len(lines)-(height-2), 0))
	return component.Details{Title: title + suffix, Focused: focused, Lines: lines, Offset: *offset}.View(width, height)
}

// frictionDetailLines gives the fields of the entry, then each section with
// its heading in upper case. An open entry shows how many times it was seen,
// and a resolved entry shows its decision.
func frictionDetailLines(entry frictionEntry, width int) []string {
	valueWidth := max(width-14, 1)
	field := func(label, value string, style lipgloss.Style) string {
		return component.DetailField(label, style.Render(padCell(ansi.Truncate(value, valueWidth, "…"), valueWidth)))
	}
	plain := lipgloss.NewStyle()
	dim := lipgloss.NewStyle().Foreground(component.Dim)
	if !entry.valid {
		return []string{"", field("title", "? "+entry.file, plain), field("status", "?", plain), field("file", entry.path, plain)}
	}
	status := lipgloss.NewStyle().Foreground(component.Accent).Bold(true)
	fourth := field("decision", entry.decision, plain)
	if entry.resolved {
		status = lipgloss.NewStyle().Foreground(component.Done)
	} else {
		count, last := entry.seen()
		times := "times"
		if count == 1 {
			times = "time"
		}
		fourth = field("seen", fmt.Sprintf("%d %s · last %s", count, times, last), plain)
	}
	lines := []string{
		"",
		field("title", entry.title, plain),
		field("status", entry.status, status),
		field("opened", entry.opened, plain),
		fourth,
		field("file", entry.path, plain),
	}
	for _, section := range entry.sections {
		lines = append(lines, "", component.DetailSection(strings.ToUpper(section.title)))
		for _, line := range wrapLines(section.body, width-4) {
			switch {
			case strings.EqualFold(section.title, "Occurrences"):
				line = dim.Render(line)
			case strings.EqualFold(section.title, "Resolved by"):
				if name, rest, ok := strings.Cut(line, "  "); ok {
					line = padCell(name, 16) + dim.Render(strings.TrimSpace(rest))
				}
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// wrapLines wraps each line to width cells and keeps its indent.
func wrapLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.Split(ansi.Wrap(line, max(width, 1), ""), "\n")...)
	}
	return out
}

func (m *Model) startFrictionFileRead(state *frictionState, path, title string) tea.Cmd {
	state.fileOpen = true
	state.fileLoading = true
	state.fileTitle = title
	state.fileContent = ""
	state.fileOffset = 0
	return func() tea.Msg {
		content, err := m.readScreenFile(path)
		return frictionFileMsg{title: title, content: content, err: err}
	}
}

func (m *Model) acceptFrictionFile(msg frictionFileMsg) {
	m.friction.fileLoading = false
	if errors.Is(msg.err, errDemoNoFile) {
		m.friction.fileContent = demoNoFile
		m.demo.notice = demoNoFile
		return
	}
	if msg.err != nil {
		m.friction.fileContent = "error: " + boundedError(msg.err.Error())
		return
	}
	m.friction.fileTitle = msg.title
	m.friction.fileContent = msg.content
	m.friction.fileOffset = 0
}
