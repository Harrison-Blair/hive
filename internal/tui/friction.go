package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
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

// frictionFold is the selection key of the fold line (friction rule 3).
const frictionFold = "fold"

// frictionState is the state of the Friction screen. The selection is the
// path of an entry, frictionFold, or empty for no selection. Expanded shows
// the resolved entries below the fold line. Filter is the applied filter,
// and the draft is the text of the open filter line (friction rule 6).
type frictionState struct {
	data         frictionData
	selectedPath string
	expanded     bool
	listOffset   int
	filter       string
	filterOpen   bool
	filterDraft  string
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

// matches reports that the applied filter matches the entry name or the
// title, case-insensitive (friction rule 6).
func (s frictionState) matches(entry frictionEntry) bool {
	if s.filter == "" {
		return true
	}
	text := strings.ToLower(s.filter)
	return strings.Contains(strings.ToLower(entry.name()), text) || strings.Contains(strings.ToLower(entry.title), text)
}

// entries gives the open or the resolved entries that the filter matches.
// Open entries keep the order of the data: the newest opened date first.
// Resolved entries show the newest resolved date first, then by name.
func (s frictionState) entries(resolved bool) []frictionEntry {
	entries := make([]frictionEntry, 0, len(s.data.entries))
	for _, entry := range s.data.entries {
		if entry.resolved == resolved && s.matches(entry) {
			entries = append(entries, entry)
		}
	}
	if resolved {
		sort.SliceStable(entries, func(i, j int) bool {
			if a, b := entries[i].resolvedDate(), entries[j].resolvedDate(); a != b {
				return a > b
			}
			return entries[i].file < entries[j].file
		})
	}
	return entries
}

// counts gives the number of open and resolved entries, with no filter.
func (s frictionState) counts() (open, resolved int) {
	for _, entry := range s.data.entries {
		if entry.resolved {
			resolved++
		} else {
			open++
		}
	}
	return open, resolved
}

type frictionRowKind uint8

const (
	frictionEntryRow frictionRowKind = iota
	frictionNoteRow
	frictionBlankRow
	frictionFoldRow
)

// frictionRow is one row of the list: an entry, a note in dim, the blank
// row above the fold line, or the fold line.
type frictionRow struct {
	kind  frictionRowKind
	entry frictionEntry
	text  string
}

// key gives the selection key of a row, or empty for a row that cannot have
// the selection.
func (r frictionRow) key() string {
	switch r.kind {
	case frictionEntryRow:
		return r.entry.path
	case frictionFoldRow:
		return frictionFold
	}
	return ""
}

// rows gives the rows of the list: the open entries, one blank row, the fold
// line, and the resolved entries while they are expanded (friction rules 3,
// 6, and 7).
func (s frictionState) rows() []frictionRow {
	open, resolved := s.entries(false), s.entries(true)
	rows := make([]frictionRow, 0, len(open)+len(resolved)+3)
	switch {
	case len(open) == 0 && len(resolved) == 0 && s.filter != "":
		rows = append(rows, frictionRow{kind: frictionNoteRow, text: "no entry matches " + s.filter})
	case len(open) == 0:
		rows = append(rows, frictionRow{kind: frictionNoteRow, text: "no open entries"})
	}
	for _, entry := range open {
		rows = append(rows, frictionRow{kind: frictionEntryRow, entry: entry})
	}
	rows = append(rows, frictionRow{kind: frictionBlankRow}, frictionRow{kind: frictionFoldRow})
	if s.expanded {
		for _, entry := range resolved {
			rows = append(rows, frictionRow{kind: frictionEntryRow, entry: entry})
		}
	}
	return rows
}

// selectedRow gives the index of the selected row, or -1.
func (s frictionState) selectedRow(rows []frictionRow) int {
	if s.selectedPath == "" {
		return -1
	}
	for index, row := range rows {
		if row.key() == s.selectedPath {
			return index
		}
	}
	return -1
}

// selected gives the selected entry. The fold line is not an entry.
func (s frictionState) selected() (frictionEntry, bool) {
	rows := s.rows()
	if index := s.selectedRow(rows); index >= 0 && rows[index].kind == frictionEntryRow {
		return rows[index].entry, true
	}
	return frictionEntry{}, false
}

// acceptSelection keeps the selection when the list has it. Else it
// selects the first open entry, or nothing when no open entry shows
// (friction rule 2).
func (s *frictionState) acceptSelection() {
	rows := s.rows()
	if s.selectedRow(rows) >= 0 {
		return
	}
	s.selectedPath = ""
	if rows[0].kind == frictionEntryRow {
		s.selectedPath = rows[0].key()
	}
	s.listOffset = 0
}

// moveSelection moves the selection by delta selectable rows. With no
// selection, a move down starts at the first row and a move up at the last.
func (s *frictionState) moveSelection(delta int) {
	rows := s.rows()
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		if key := row.key(); key != "" {
			keys = append(keys, key)
		}
	}
	index := slices.Index(keys, s.selectedPath)
	switch {
	case index < 0 && delta > 0:
		index = delta - 1
	case index < 0:
		index = len(keys) + delta
	default:
		index += delta
	}
	s.selectedPath = keys[min(max(index, 0), len(keys)-1)]
}

type frictionScreen struct{ model *Model }

func (frictionScreen) DataRequest() dataRequest { return frictionDataRequest }

func (frictionScreen) KeyHints() string { return "" }

// Update takes the keys of the friction key table. The frame takes ctrl+]
// and esc. The filter line takes the keys while it is open.
func (s frictionScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	state := &m.friction
	key := press.String()
	if state.filterOpen {
		return m.updateFrictionFilter(press)
	}
	listRows, boxRows := m.frictionLayout()
	if m.regionFocus[tabFriction] == 1 {
		page := max(boxRows-2, 1)
		offset := &m.detailsOffset[tabFriction]
		before := *offset
		switch key {
		case "up":
			*offset--
		case "down":
			*offset++
		case "pgup":
			*offset -= page
		case "pgdown":
			*offset += page
		case "g":
			*offset = 0
		case "G":
			*offset = 1 << 16
		default:
			return nil
		}
		*offset = min(max(*offset, 0), m.frictionDetailsMax())
		if *offset == before {
			return nil
		}
		return m.logAction("friction details scroll", slog.Int("offset", *offset))
	}
	delta := 0
	switch key {
	case "up":
		delta = -1
	case "down":
		delta = 1
	case "pgup":
		delta = -max(listRows, 1)
	case "pgdown":
		delta = max(listRows, 1)
	case "g":
		delta = -(1 << 16)
	case "G":
		delta = 1 << 16
	case "enter":
		if state.selectedPath != frictionFold {
			return nil
		}
		state.expanded = !state.expanded
		action := "friction resolved fold"
		if state.expanded {
			action = "friction resolved expand"
		}
		return m.logAction(action)
	case "/":
		state.filterOpen = true
		state.filterDraft = state.filter
		return m.logAction("friction filter open")
	default:
		return nil
	}
	return m.moveFrictionSelection(delta)
}

// moveFrictionSelection moves the selection and logs a change. The details
// box shows a new selection from its top.
func (m *Model) moveFrictionSelection(delta int) tea.Cmd {
	state := &m.friction
	previous := state.selectedPath
	state.moveSelection(delta)
	if previous == state.selectedPath {
		return nil
	}
	m.detailsOffset[tabFriction] = 0
	return m.logAction("friction selection", slog.String("file", state.selectedPath))
}

// updateFrictionFilter edits the filter line (friction rule 6). Enter
// applies the text and closes the line. Esc closes the line and keeps the
// applied filter.
func (m *Model) updateFrictionFilter(press tea.KeyPressMsg) tea.Cmd {
	state := &m.friction
	switch press.String() {
	case "enter":
		state.filter = strings.TrimSpace(state.filterDraft)
		state.filterOpen = false
		state.acceptSelection()
		m.detailsOffset[tabFriction] = 0
		return m.logAction("friction filter apply", slog.String("filter", state.filter))
	case "esc":
		state.filterOpen = false
		return m.logAction("friction filter close")
	case "backspace":
		if runes := []rune(state.filterDraft); len(runes) > 0 {
			state.filterDraft = string(runes[:len(runes)-1])
		}
		return nil
	}
	if press.Text != "" {
		state.filterDraft += press.Text
	}
	return nil
}

// View gives the body of the screen. The frame draws the title row and the
// status rows.
func (s frictionScreen) View(width, height int) string {
	return s.model.frictionBody(width, height)
}
