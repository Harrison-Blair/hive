package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// The sizes of the task popup (GfWorkTask) and the dependency popup
// (GfWorkDeps).
const (
	workTaskPopupWidth  = 80
	workTaskPopupHeight = 31
	workDepsPopupWidth  = 68
	workDepsPopupHeight = 17
)

// workLabelWidth is the cells of a label of the task popup with its gap.
const workLabelWidth = 11

// workCommandWidth is the cells of the command of an evidence line with its
// gap.
const workCommandWidth = 38

// workPopupView dims screen, the rows above the status rows, and draws the
// open popup centered on it (decision 5).
func (m *Model) workPopupView(screen string, width, height int) string {
	s := &m.tasks
	w, h := workTaskPopupWidth, workTaskPopupHeight
	var title, hint string
	var lines []string
	switch {
	case s.popup == workPopupDeps:
		w, h = workDepsPopupWidth, workDepsPopupHeight
		title = s.selectedID + " · dependencies · esc closes"
		lines = m.workDepsLines()
		hint = "enter opens the selected task · ↑↓ move · esc closes"
	case s.fileOpen:
		title = s.fileTitle + " · esc closes"
		text := s.fileContent
		if s.fileLoading {
			text = "loading…"
		}
		lines = strings.Split(ansi.Wrap(text, max(min(w, width)-4, 1), ""), "\n")
		hint = "↑↓ pgup/pgdn scroll · esc closes"
	default:
		title = s.selectedID + " · esc closes"
		lines = m.workTaskLines()
		hint = "b brief · v verified · d deps · ↑↓ pgup/pgdn scroll · esc closes"
	}
	w, h = min(w, width), min(h, height)
	body := max(h-3, 1)
	offset := 0
	if s.popup == workPopupTask {
		offset = min(s.popupOffset, max(len(lines)-body, 0))
	}
	shown := make([]string, body+1)
	copy(shown, lines[min(offset, len(lines)):])
	shown[body] = dimStyle.Render(hint)
	box := component.Panel{Title: title, Focused: true, Lines: shown, VMark: component.Thumb(body, len(lines), body, offset)}.View(w, h)
	rows := strings.Split(component.DimScreen(screen, width, height), "\n")
	plain := strings.Split(ansi.Strip(screen), "\n")
	x, y := (width-w)/2, (height-h)/2
	for index, line := range strings.Split(box, "\n") {
		if y+index >= len(rows) {
			break
		}
		row := fitText(plain[min(y+index, len(plain)-1)], width)
		if y+index >= len(plain) {
			row = strings.Repeat(" ", width)
		}
		rows[y+index] = workBorderStyle.Render(ansi.Cut(row, 0, x)) + line + workBorderStyle.Render(ansi.Cut(row, x+w, width))
	}
	return strings.Join(rows, "\n")
}

// workPopupMaxOffset gives the last scroll offset of the task popup.
func (m *Model) workPopupMaxOffset() int {
	_, height := m.contentArea()
	body := max(min(workTaskPopupHeight, height+2)-3, 1)
	return max(len(m.workTaskLines())-body, 0)
}

// workField gives a field line of the task popup: the label in dim in 11
// cells, then the value.
func workField(label, value string) string {
	return dimStyle.Render(label) + strings.Repeat(" ", workLabelWidth-ansi.StringWidth(label)) + value
}

// workTaskLines gives the lines of the task popup (rule 9): the fields, the
// acceptance criteria of the brief, and the evidence of VERIFIED.md.
func (m *Model) workTaskLines() []string {
	s := &m.tasks
	item, ok := s.data.item(s.selectedID)
	if !ok {
		return nil
	}
	task := item.record
	status := task.Status
	if taskLanded(task) {
		status = "landed"
	}
	goal := "none"
	if item.goal != "" {
		goal = item.goal + " · wave " + item.wave
	}
	worker := "none"
	if task.ActingAgent != "" {
		worker = task.ActingAgent
		if agent, ok := s.data.agents[task.ActingAgent]; ok {
			for _, part := range []string{agent.AgentStatus, coordinatorStatusAge(agent.StatusChangedAt, m.now())} {
				if part != "" {
					worker += " · " + part
				}
			}
		}
	}
	verification := latestTaskVerification(task)
	verifier := "none yet"
	if verification != nil {
		verifier = verification.Agent + " · " + verification.Result
	}
	branch := "none"
	if task.Worktree != nil && task.Worktree.Branch != "" {
		branch = task.Worktree.Branch
	}
	merge := "none"
	if taskLanded(task) {
		merge = shortSHA(task.Delivery.Commit)
		if task.Delivery.Target != "" {
			merge += " on " + task.Delivery.Target
		}
	}
	depends := make([]string, 0, len(task.DependsOn))
	for _, id := range task.DependsOn {
		state := "missing"
		if dependency, ok := s.data.item(id); ok {
			state = dependency.record.Status
			if taskLanded(dependency.record) {
				state = "landed"
			}
		}
		depends = append(depends, id+" · "+state)
	}
	if len(depends) == 0 {
		depends = []string{"none"}
	}
	lines := []string{
		workField("title", task.Title),
		workField("status", status),
		workField("goal", goal),
		workField("worker", worker),
		workField("verifier", verifier),
		workField("branch", branch),
		workField("merge", merge),
		workField("depends", strings.Join(depends, ", ")),
		"",
		component.DetailSection("ACCEPTANCE"),
	}
	// tui-v2 decision 16: one mark for each criterion from the last
	// verification.
	mark, markStyle := "○", dimStyle
	if verification != nil && verification.Result == "pass" {
		mark, markStyle = "✓", doneStyle
	}
	criteria := markdownRows(item.brief, "AC")
	if len(criteria) == 0 {
		lines = append(lines, dimStyle.Render("no acceptance criteria"))
	}
	for _, row := range criteria {
		label := mark + " " + row[0]
		text := ""
		if len(row) > 1 {
			text = row[1]
		}
		lines = append(lines, markStyle.Render(label)+strings.Repeat(" ", max(workLabelWidth-ansi.StringWidth(label), 1))+text)
	}
	key := func(k string) string { return titleStyle.Render(k) }
	lines = append(lines, "", component.DetailSection("EVIDENCE"),
		key("b")+" brief.md    "+key("v")+" VERIFIED.md   "+key("d")+" deps graph", "")
	for _, row := range markdownRows(item.verified, "V") {
		if len(row) < 3 {
			continue
		}
		lines = append(lines, workField(row[0], padCell(row[1], workCommandWidth)+row[2]))
	}
	return lines
}

// workDepsLines gives the lines of the dependency popup (rule 10 and the
// board GfWorkDeps): the dependencies, the tasks that depend on the
// selected task, and the ready when line.
func (m *Model) workDepsLines() []string {
	s := &m.tasks
	item, ok := s.data.item(s.selectedID)
	if !ok {
		return nil
	}
	index := 0
	row := func(other taskItem) string {
		lead, id := "  ", padCell(other.record.ID, 20)
		if index == s.depsRow {
			lead, id = titleStyle.Render("▶")+" ", titleStyle.Render(id)
		}
		index++
		card := s.card(other, m.now())
		state, _ := s.cardLine(other, m.now())
		if strings.HasPrefix(state, "waits on ") {
			state = other.record.Status
		}
		wave := "no goal"
		if other.wave != "" {
			wave = "wave " + other.wave
		}
		return lead + lipgloss.NewStyle().Foreground(card.DotColor).Render(card.Dot) + " " + id + dimStyle.Render(wave+" · "+state)
	}
	lines := []string{component.DetailSection("WAITS ON")}
	dependents := s.dependents(item.record.ID)
	rows := s.depsRows()
	for _, other := range rows[:len(rows)-len(dependents)] {
		lines = append(lines, row(other))
	}
	if len(rows) == len(dependents) {
		lines = append(lines, dimStyle.Render("none"))
	}
	lines = append(lines, "", component.DetailSection("WAITED ON BY"))
	for _, other := range dependents {
		lines = append(lines, row(other))
	}
	if len(dependents) == 0 {
		lines = append(lines, dimStyle.Render("none"))
	}
	if waits := s.waitsOn(item.record); len(waits) > 0 {
		verb := "lands"
		if len(waits) > 1 {
			verb = "land"
		}
		lines = append(lines, "", dimStyle.Render("ready when")+fmt.Sprintf(" %s %s", strings.Join(waits, ", "), verb))
	}
	return lines
}
