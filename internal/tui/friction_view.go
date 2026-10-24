package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// The columns of the friction list at 120 columns: the name from column 3,
// then AGE, SEEN, LAST, and DECISION.
const (
	frictionNameWidth = 37
	frictionAgeWidth  = 6
	frictionSeenWidth = 11
	frictionLastWidth = 7
	frictionDecision  = 3 + frictionNameWidth + frictionAgeWidth + frictionSeenWidth + frictionLastWidth
)

// frictionFilterHint is the hint of the open filter line. It starts at
// column frictionHintColumn, as GfFrictionFilter draws it.
const (
	frictionFilterHint = "enter applies · esc closes and keeps the applied filter"
	frictionHintColumn = 61
)

// frictionListItem starts a new line in the details text: a numbered or a
// bulleted item.
var frictionListItem = regexp.MustCompile(`^\s*(\d+\.|[-*•])\s`)

// frictionSplit gives the list rows and the details box rows of a body of
// height rows: the header row, the list, one blank row, the box, and one
// blank row. At 120×40 the list has rows 3 to 19 and the box rows 21 to 37.
func frictionSplit(height int) (listRows, boxRows int) {
	boxRows = max((height-3)/2, component.PanelMinHeight)
	return max(height-boxRows-3, 1), boxRows
}

// frictionLayout gives the list rows and the box rows of the current size.
func (m *Model) frictionLayout() (listRows, boxRows int) {
	_, height := m.contentArea()
	return frictionSplit(height)
}

// frictionTitle gives the title row of the screen (friction layout and rule
// 6): the counts, the applied filter in accent, and the right text.
func (m *Model) frictionTitle() component.TitleRow {
	open, resolved := m.friction.counts()
	row := component.TitleRow{Name: "friction", Facts: fmt.Sprintf("· %d open · %d resolved", open, resolved), Right: "/ filter"}
	if m.friction.filter != "" {
		row.Tag = "· / " + m.friction.filter
	}
	return row
}

// frictionBody draws the header row or the filter line, the list, and the
// details box at exactly width×height cells.
func (m *Model) frictionBody(width, height int) string {
	listRows, boxRows := frictionSplit(height)
	lines := append([]string{m.frictionHeader(width)}, m.frictionListRows(width, listRows)...)
	lines = append(lines, "")
	lines = append(lines, strings.Split(m.frictionBox(width, boxRows), "\n")...)
	return fitBlock(strings.Join(lines, "\n"), width, height)
}

// frictionHeader gives the column headers in dim bold, or the open filter
// line in their place (friction rule 6).
func (m *Model) frictionHeader(width int) string {
	state := &m.friction
	if state.filterOpen {
		line := " " + component.StrongText("/") + " " + state.filterDraft + " █"
		if w := lipgloss.Width(line); w < frictionHintColumn-1 && frictionHintColumn+lipgloss.Width(frictionFilterHint) < width {
			line = padCell(line, frictionHintColumn) + component.DimText(frictionFilterHint)
		}
		return fitLine(line, width)
	}
	head := lipgloss.NewStyle().Foreground(component.Dim).Bold(true)
	line := "   " + padCell(head.Render("ENTRY"), frictionNameWidth) + padCell(head.Render("AGE"), frictionAgeWidth) +
		padCell(head.Render("SEEN"), frictionSeenWidth) + padCell(head.Render("LAST"), frictionLastWidth) + head.Render("DECISION")
	return fitLine(line, width)
}

// frictionListRows gives the visible rows of the list. The list scrolls so
// that the selected row shows, with the ┃ thumb on the last column when its
// rows do not fit (friction rule 3).
func (m *Model) frictionListRows(width, height int) []string {
	state := &m.friction
	rows := state.rows()
	selected := state.selectedRow(rows)
	offset := state.listOffset
	if selected >= 0 {
		offset = min(max(offset, selected-height+1), selected)
	}
	offset = min(max(offset, 0), max(len(rows)-height, 0))
	state.listOffset = offset
	thumb := component.Thumb(height, len(rows), height, offset)
	mark := component.DimText("┃")
	if m.regionFocus[tabFriction] == 0 {
		mark = component.AccentText("┃")
	}
	_, resolvedCount := state.counts()
	out := make([]string, 0, height)
	for index := range height {
		line := ""
		if row := offset + index; row < len(rows) {
			line = m.frictionRow(rows[row], row == selected, resolvedCount, width-1)
		}
		line = fitLine(line, width-1)
		if index >= thumb.Start && index < thumb.Start+thumb.Size {
			line += mark
		} else {
			line += " "
		}
		out = append(out, line)
	}
	return out
}

// frictionRow draws one row of the list at width cells.
func (m *Model) frictionRow(row frictionRow, selected bool, resolvedCount, width int) string {
	state := &m.friction
	switch row.kind {
	case frictionNoteRow:
		return "   " + component.DimText(row.text)
	case frictionFoldRow:
		text := fmt.Sprintf("resolved · %d", resolvedCount)
		if state.filter != "" {
			text += fmt.Sprintf(" · %d match", len(state.entries(true)))
		}
		hint := "enter opens"
		if state.expanded {
			hint = "enter folds"
		}
		return component.Fold{Open: state.expanded, Selected: selected, Text: text, Hint: hint}.View(width)
	case frictionEntryRow:
		return m.frictionEntryRow(row.entry, selected, width)
	}
	return ""
}

// frictionEntryRow draws an entry row (friction rules 1 to 3): the name,
// the age since opened in days, the seen bar, the last date, and the
// decision cut with … at the row end. The decision column is dim also where
// it has no text, as the boards draw it. A resolved row shows its resolved
// date in LAST, and its name and bar in dim, as GfFrictionResolved draws it.
// The style of the name also covers its padding, as the boards draw it.
func (m *Model) frictionEntryRow(entry frictionEntry, selected bool, width int) string {
	marker, name := " ", ansi.Truncate(entry.name(), frictionNameWidth-1, "…")
	switch {
	case selected:
		marker, name = component.StrongText("▶"), component.StrongText(padCell(name, frictionNameWidth-1))
	case entry.resolved:
		name = component.DimText(padCell(name, frictionNameWidth-1))
	}
	age, bar, last := "?", "", "?"
	if entry.valid {
		count, seen := entry.seen()
		age = fmt.Sprintf("%dd", frictionDays(entry.opened, m.now()))
		bar = component.SeenBar{Count: count}.View()
		if entry.resolved {
			bar = component.DimText(ansi.Strip(bar))
		}
		last = seen
		if entry.resolved {
			last = entry.resolvedDate()
		}
		if len(last) == len("2006-01-02") {
			last = last[5:]
		}
	}
	line := " " + marker + " " + padCell(name, frictionNameWidth) + padCell(component.DimText(age), frictionAgeWidth) +
		padCell(bar, frictionSeenWidth) + padCell(component.DimText(last), frictionLastWidth)
	if room := width - lipgloss.Width(line) - 1; room > 0 {
		line += component.DimText(padCell(ansi.Truncate(entry.decision, room, "…"), room))
	}
	return line
}

// frictionDays gives the whole days from the date to the day of now.
func frictionDays(date string, now time.Time) int {
	opened, err := time.ParseInLocation("2006-01-02", date, now.Location())
	if err != nil {
		return 0
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return max(int(today.Sub(opened).Hours()/24), 0)
}

// frictionBox draws the details box of the selected entry at width×height
// cells. The box scrolls with the ┃ thumb on its right border (friction
// rules 4 and 5).
func (m *Model) frictionBox(width, height int) string {
	title, lines, dim := m.frictionDetails(width)
	body := max(height-2, 0)
	offset := &m.detailsOffset[tabFriction]
	*offset = min(max(*offset, 0), max(len(lines)-body, 0))
	return component.Panel{
		Title: title, DimTitle: dim, Focused: m.regionFocus[tabFriction] == 1,
		Lines: lines[*offset:], VMark: component.Thumb(body, len(lines), body, *offset),
	}.View(width, height)
}

// frictionDetailsMax gives the largest scroll offset of the details box.
func (m *Model) frictionDetailsMax() int {
	_, boxRows := m.frictionLayout()
	_, lines, _ := m.frictionDetails(m.width)
	return max(len(lines)-max(boxRows-2, 0), 0)
}

// frictionDetails gives the title and the lines of the details box for a
// box of width cells, and reports a dim title. The title is NAME · STATUS ·
// opened DATE · seen N · PATH. The body has the sections SUMMARY,
// REPRODUCTION, OCCURRENCES, and DECISION that the entry has (friction rule
// 4). With no selected entry, the box shows no entry (friction rule 7).
func (m *Model) frictionDetails(width int) (string, []string, bool) {
	entry, ok := m.friction.selected()
	if !ok {
		return "no entry", []string{component.DimText("select an entry to read it")}, true
	}
	if !entry.valid {
		return entry.name() + " · ? · " + entry.path, []string{component.DimText("the front matter of this entry is not valid")}, false
	}
	count, _ := entry.seen()
	title := fmt.Sprintf("%s · %s · opened %s · seen %d · %s", entry.name(), entry.status, entry.opened, count, entry.path)
	room := max(width-4, 1)
	var groups [][]string
	for _, section := range []struct{ heading, name string }{{"SUMMARY", "Summary"}, {"REPRODUCTION", "Reproduction"}} {
		if text := frictionReflow(entry.sectionPrefix(section.name), room); len(text) > 0 {
			groups = append(groups, append([]string{component.StrongText(section.heading)}, text...))
		}
	}
	if occurrences := frictionOccurrences(entry.section("Occurrences"), room); len(occurrences) > 0 {
		group := []string{component.StrongText("OCCURRENCES")}
		for _, line := range occurrences {
			group = append(group, component.DimText(line))
		}
		groups = append(groups, group)
	}
	if entry.decision != "" {
		wrapped := strings.Split(ansi.Wrap("DECISION "+entry.decision, room, ""), "\n")
		wrapped[0] = component.StrongText("DECISION") + strings.TrimPrefix(wrapped[0], "DECISION")
		groups = append(groups, wrapped)
	}
	var lines []string
	for index, group := range groups {
		if index > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, group...)
	}
	return title, lines, false
}

// sectionPrefix gives the lines of the first section whose title starts
// with prefix, in any case: "Reproduction steps" for REPRODUCTION.
func (e frictionEntry) sectionPrefix(prefix string) []string {
	for _, section := range e.sections {
		if strings.HasPrefix(strings.ToLower(section.title), strings.ToLower(prefix)) {
			return section.body
		}
	}
	return nil
}

// frictionReflow joins the lines of each paragraph and wraps it to width
// cells. A blank line stays, and a numbered or bulleted item starts a new
// line.
func frictionReflow(lines []string, width int) []string {
	var out []string
	paragraph := ""
	flush := func() {
		if paragraph != "" {
			out = append(out, strings.Split(ansi.Wrap(paragraph, width, ""), "\n")...)
		}
		paragraph = ""
	}
	for _, line := range lines {
		text := strings.TrimSpace(line)
		switch {
		case text == "":
			flush()
			out = append(out, "")
		case frictionListItem.MatchString(line):
			flush()
			paragraph = text
		case paragraph == "":
			paragraph = text
		default:
			paragraph += " " + text
		}
	}
	flush()
	return out
}

// frictionOccurrences packs the occurrence lines into rows of width cells,
// with five spaces between two occurrences.
func frictionOccurrences(lines []string, width int) []string {
	var out []string
	row := ""
	for _, line := range lines {
		item := strings.TrimSpace(line)
		switch {
		case item == "":
			continue
		case row == "":
			row = item
		case lipgloss.Width(row)+5+lipgloss.Width(item) <= width:
			row += "     " + item
		default:
			out = append(out, row)
			row = item
		}
	}
	if row != "" {
		out = append(out, row)
	}
	return out
}

// frictionClick handles a click on the body row: a click on a list row
// selects it and gives the list the focus, and a click on the details box
// gives the box the focus (README decision 20).
func (m *Model) frictionClick(row int) tea.Cmd {
	listRows, boxRows := m.frictionLayout()
	var cmds []tea.Cmd
	var region int
	switch {
	case row >= 1 && row <= listRows:
		region = 0
		rows := m.friction.rows()
		if index := m.friction.listOffset + row - 1; index < len(rows) && rows[index].key() != "" && rows[index].key() != m.friction.selectedPath {
			m.friction.selectedPath = rows[index].key()
			m.detailsOffset[tabFriction] = 0
			cmds = append(cmds, m.logAction("friction selection", slogString("file", m.friction.selectedPath)))
		}
	case row >= listRows+2 && row < listRows+2+boxRows:
		region = 1
	default:
		return nil
	}
	if m.regionFocus[tabFriction] != region {
		m.regionFocus[tabFriction] = region
		cmds = append(cmds, m.logAction("focus region", slogString("region", screenRegistry[tabFriction].regions[region])))
	}
	return batch(cmds...)
}

// frictionWheel scrolls the box below the pointer and keeps the focus
// (README decision 6): the list moves its selection, and the details box
// scrolls by three lines.
func (m *Model) frictionWheel(row, step int) tea.Cmd {
	listRows, boxRows := m.frictionLayout()
	switch {
	case row >= 1 && row <= listRows:
		return m.moveFrictionSelection(step)
	case row >= listRows+2 && row < listRows+2+boxRows:
		offset := &m.detailsOffset[tabFriction]
		*offset = min(max(*offset+step*wheelLines, 0), m.frictionDetailsMax())
	}
	return nil
}
