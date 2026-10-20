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

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

type frictionState struct {
	data            frictionData
	tab             frictionTab
	selectedPath    string
	scrollOffset    int
	scrollHeight    int
	fileOpen        bool
	fileLoading     bool
	fileTitle       string
	fileContent     string
	fileViewport    viewport.Model
	fileViewportSet bool
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
	sort.SliceStable(data.entries, func(i, j int) bool {
		if data.entries[i].opened == data.entries[j].opened {
			return data.entries[i].file < data.entries[j].file
		}
		return data.entries[i].opened > data.entries[j].opened
	})
	return data, nil
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

func frictionOneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

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
	return sections
}

func (s *frictionState) accept(data frictionData) {
	selected := s.selectedPath
	s.data = data
	for _, entry := range s.visibleEntries() {
		if entry.path == selected {
			return
		}
	}
	entries := s.visibleEntries()
	s.selectedPath = ""
	if len(entries) != 0 {
		s.selectedPath = entries[0].path
	}
	s.scrollOffset = 0
}

func (s frictionState) visibleEntries() []frictionEntry {
	entries := make([]frictionEntry, 0, len(s.data.entries))
	for _, entry := range s.data.entries {
		if (s.tab == frictionResolved) == entry.resolved {
			entries = append(entries, entry)
		}
	}
	return entries
}

type frictionScreen struct{ model *Model }

func (frictionScreen) DataRequest() dataRequest { return frictionDataRequest }

func (frictionScreen) KeyHints() string {
	return "tab open/resolved · enter read · ask in Chat · ↑/↓ move"
}

func (s frictionScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.friction
	key := press.String()
	if state.fileOpen {
		if key == "esc" || key == "escape" {
			state.fileOpen = false
			state.fileLoading = false
			state.fileContent = ""
			return s.model.logAction("friction file close", slog.String("file", state.fileTitle))
		}
		updated, cmd := state.fileViewport.Update(press)
		state.fileViewport = updated
		return cmd
	}
	switch key {
	case "tab":
		state.tab = (state.tab + 1) % 2
		state.acceptSelection()
		return s.model.logAction("friction tab", slog.String("tab", []string{"open", "resolved"}[state.tab]))
	case "up", "k":
		previous := state.selectedPath
		state.moveSelection(-1)
		if previous != state.selectedPath {
			return s.model.logAction("friction selection", slog.String("file", state.selectedPath))
		}
	case "down", "j":
		previous := state.selectedPath
		state.moveSelection(1)
		if previous != state.selectedPath {
			return s.model.logAction("friction selection", slog.String("file", state.selectedPath))
		}
	case "enter":
		if selected, ok := state.selected(); ok {
			path := filepath.Join(s.model.home, "reference", "dogfood", "friction", selected.path)
			read := s.model.startFrictionFileRead(state, path, selected.file)
			return batch(s.model.logAction("friction entry open", slog.String("file", selected.file)), read)
		}
	}
	return nil
}

func (s frictionScreen) View(width, height int) string {
	state := &s.model.friction
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if state.fileOpen {
		return state.fileView(contentWidth, height-4)
	}
	entries := state.visibleEntries()
	openCount, resolvedCount := 0, 0
	for _, entry := range state.data.entries {
		if entry.resolved {
			resolvedCount++
		} else {
			openCount++
		}
	}
	var tabLine string
	if state.tab == frictionOpen {
		tabLine = "[open " + strconv.Itoa(openCount) + "]  resolved " + strconv.Itoa(resolvedCount)
	} else {
		tabLine = "open " + strconv.Itoa(openCount) + "  [resolved " + strconv.Itoa(resolvedCount) + "]"
	}

	split := width >= 72
	listWidth := width
	titleWidth, decisionWidth := 42, 40
	detailWidth := 0
	if split {
		detailWidth = width / 3
		if detailWidth < 28 {
			detailWidth = 28
		}
		if detailWidth > 56 {
			detailWidth = 56
		}
		if detailWidth > width-36 {
			detailWidth = width - 36
		}
		listWidth = width - detailWidth - 1
		columnWidth := listWidth - 16
		titleWidth = columnWidth * 2 / 3
		if titleWidth > 42 {
			titleWidth = 42
		}
		if titleWidth < 1 {
			titleWidth = 1
		}
		decisionWidth = columnWidth - titleWidth
	}
	header := "opened       title                                      decision"
	if split {
		header = fmt.Sprintf("%-12s %-*s %s", "opened", titleWidth, "title", "decision")
	}
	lines := []string{titleStyle.Render(tabLine), header}
	rowLines := make([]string, 0, len(entries))
	for _, entry := range entries {
		prefix := "  "
		if entry.path == state.selectedPath {
			prefix = "> "
		}
		opened := entry.opened
		title := entry.title
		decision := entry.decision
		if !entry.valid {
			opened, title, decision = "?", "? "+entry.file, "?"
		}
		row := fmt.Sprintf("%s%-12s %-42s %s", prefix, opened, clipCell(title, 42), clipCell(frictionOneLine(decision), 40))
		if split {
			row = fmt.Sprintf("%s%-12s %-*s %-*s", prefix, opened, titleWidth, clipCell(title, titleWidth), decisionWidth, clipCell(frictionOneLine(decision), decisionWidth))
		}
		if entry.resolved {
			row = dimStyle.Render(row)
		}
		rowLines = append(rowLines, row)
	}
	if len(rowLines) == 0 {
		rowLines = append(rowLines, "  no entries")
	}
	state.keepSelectedVisible()
	detailLines := state.detailLines()
	listHeight := height - 5 - len(detailLines)
	if split {
		listHeight = height - 5
	}
	if listHeight < 1 {
		listHeight = 1
	}
	state.scrollHeight = listHeight
	state.keepSelectedVisible()
	end := state.scrollOffset + listHeight
	if end > len(rowLines) {
		end = len(rowLines)
	}
	if state.scrollOffset > len(rowLines) {
		state.scrollOffset = 0
	}
	lines = append(lines, rowLines[state.scrollOffset:end]...)
	if split {
		for index, line := range lines {
			lines[index] = trimGraphLine(line, listWidth)
		}
		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorAccent)).
			Padding(0, 1).
			Width(max(1, detailWidth-4)).
			Height(max(1, height-3))
		detail := detailStyle.Render(strings.Join(detailLines, "\n"))
		return lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(lines, "\n"), detail)
	}
	lines = append(lines, detailLines...)
	for index, line := range lines {
		lines[index] = trimGraphLine(line, contentWidth)
	}
	return strings.Join(lines, "\n")
}

func (s frictionState) selected() (frictionEntry, bool) {
	for _, entry := range s.visibleEntries() {
		if entry.path == s.selectedPath {
			return entry, true
		}
	}
	return frictionEntry{}, false
}

func (s *frictionState) acceptSelection() {
	for _, entry := range s.visibleEntries() {
		if entry.path == s.selectedPath {
			s.keepSelectedVisible()
			return
		}
	}
	entries := s.visibleEntries()
	s.selectedPath = ""
	if len(entries) != 0 {
		s.selectedPath = entries[0].path
	}
	s.scrollOffset = 0
}

func (s *frictionState) moveSelection(delta int) {
	entries := s.visibleEntries()
	if len(entries) == 0 {
		s.selectedPath = ""
		return
	}
	index := 0
	for i, entry := range entries {
		if entry.path == s.selectedPath {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = 0
	}
	if index >= len(entries) {
		index = len(entries) - 1
	}
	s.selectedPath = entries[index].path
	s.keepSelectedVisible()
}

func (s *frictionState) keepSelectedVisible() {
	if s.scrollHeight < 1 {
		return
	}
	line := 0
	for index, entry := range s.visibleEntries() {
		if entry.path == s.selectedPath {
			line = index
			break
		}
	}
	if line < s.scrollOffset {
		s.scrollOffset = line
	} else if line >= s.scrollOffset+s.scrollHeight {
		s.scrollOffset = line - s.scrollHeight + 1
	}
	maxOffset := len(s.visibleEntries()) - s.scrollHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.scrollOffset > maxOffset {
		s.scrollOffset = maxOffset
	}
}

func (s frictionState) detailLines() []string {
	entry, ok := s.selected()
	if !ok {
		return []string{"", "no entry selected"}
	}
	if !entry.valid {
		return []string{"", titleStyle.Render("? " + entry.file), "status: ?", "opened: ?", "file: " + entry.file}
	}
	lines := []string{"", titleStyle.Render(entry.title), "status: " + entry.status, "opened: " + entry.opened, "file: " + entry.file}
	for _, section := range entry.sections {
		lines = append(lines, dimStyle.Render(section.title))
		for _, text := range section.body {
			if strings.TrimSpace(text) != "" {
				lines = append(lines, strings.TrimSpace(text))
			}
		}
	}
	return lines
}

func (s *frictionState) fileView(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if !s.fileViewportSet {
		s.fileViewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
		s.fileViewport.SoftWrap = true
		s.fileViewportSet = true
	}
	s.fileViewport.SetWidth(width)
	s.fileViewport.SetHeight(height)
	if s.fileLoading {
		s.fileViewport.SetContent("loading…")
	} else {
		s.fileViewport.SetContent(s.fileContent)
	}
	return s.fileTitle + "\n" + s.fileViewport.View()
}

func (m *Model) startFrictionFileRead(state *frictionState, path, title string) tea.Cmd {
	state.fileOpen = true
	state.fileLoading = true
	state.fileTitle = title
	state.fileContent = ""
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		return frictionFileMsg{title: title, content: string(data), err: err}
	}
}

func (m *Model) acceptFrictionFile(msg frictionFileMsg) {
	m.friction.fileLoading = false
	if msg.err != nil {
		m.friction.fileContent = "error: " + boundedError(msg.err.Error())
		return
	}
	m.friction.fileTitle = msg.title
	m.friction.fileContent = msg.content
	m.friction.fileViewport.SetContent(msg.content)
	m.friction.fileViewport.GotoTop()
}
