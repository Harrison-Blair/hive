package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"log/slog"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type taskTab uint8

const (
	tasksActive taskTab = iota
	tasksQueued
	tasksLanded
	tasksFailed
	tasksAll
)

var taskTabs = [...]string{"active", "queued", "landed", "failed", "all"}
var verificationNumberPattern = regexp.MustCompile(`(?i)\bV([1-9][0-9]*)\b`)

type taskItem struct {
	record      records.Task
	projectPath string
	goalGrant   string
	evidence    []string
}

type tasksData struct {
	items    []taskItem
	projects []string
}

type tasksState struct {
	data            tasksData
	tab             taskTab
	selectedID      string
	projectIndex    int
	scrollOffset    int
	scrollHeight    int
	fileOpen        bool
	fileLoading     bool
	fileTitle       string
	fileContent     string
	fileViewport    viewport.Model
	fileViewportSet bool
}

type taskFileMsg struct {
	title   string
	content string
	err     error
}

func loadTasks(home string) (tasksData, error) {
	registry, err := store.Open(home)
	if err != nil {
		return tasksData{}, err
	}
	tasks, err := listRecords[records.Task](registry, "tasks")
	if err != nil {
		return tasksData{}, fmt.Errorf("tasks: %w", err)
	}
	projects, err := listRecords[records.Project](registry, "projects")
	if err != nil {
		return tasksData{}, fmt.Errorf("projects: %w", err)
	}
	goals, err := listRecords[records.Goal](registry, "goals")
	if err != nil {
		return tasksData{}, fmt.Errorf("goals: %w", err)
	}
	projectPaths := make(map[string]string, len(projects))
	projectIDs := make([]string, 0, len(projects))
	for _, project := range projects {
		projectPaths[project.ID] = project.Path
		projectIDs = append(projectIDs, project.ID)
	}
	sort.Strings(projectIDs)
	goalGrants := make(map[string]string)
	for _, goal := range goals {
		for _, wave := range goal.Waves {
			for _, taskID := range wave.Tasks {
				goalGrants[taskID] = goal.Grant
			}
		}
	}
	data := tasksData{projects: projectIDs, items: make([]taskItem, 0, len(tasks))}
	for _, task := range tasks {
		item := taskItem{record: task, projectPath: projectPaths[task.Project], goalGrant: goalGrants[task.ID]}
		if item.projectPath != "" && task.Feature != "" {
			dir := filepath.Join(item.projectPath, ".hive", "spec", task.Feature, "tasks", task.ID, "evidence")
			entries, readErr := os.ReadDir(dir)
			if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
				return tasksData{}, fmt.Errorf("task evidence %s: %w", task.ID, readErr)
			}
			for _, entry := range entries {
				if entry.Type().IsRegular() {
					item.evidence = append(item.evidence, entry.Name())
				}
			}
			sort.Strings(item.evidence)
			for index, name := range item.evidence {
				if name == "VERIFIED.md" {
					copy(item.evidence[1:index+1], item.evidence[0:index])
					item.evidence[0] = name
					break
				}
			}
		}
		data.items = append(data.items, item)
	}
	return data, nil
}

func (s *tasksState) accept(data tasksData) {
	selected := s.selectedID
	s.data = data
	for _, item := range s.visibleItems() {
		if item.record.ID == selected {
			return
		}
	}
	items := s.visibleItems()
	s.selectedID = ""
	if len(items) != 0 {
		s.selectedID = items[0].record.ID
	}
	s.scrollOffset = 0
}

func (s tasksState) projectFilter() string {
	if s.projectIndex <= 0 || s.projectIndex > len(s.data.projects) {
		return "all"
	}
	return s.data.projects[s.projectIndex-1]
}

func (s tasksState) projectItems() []taskItem {
	project := s.projectFilter()
	items := make([]taskItem, 0, len(s.data.items))
	for _, item := range s.data.items {
		if project == "all" || item.record.Project == project {
			items = append(items, item)
		}
	}
	return items
}

func (s tasksState) visibleItems() []taskItem {
	items := s.projectItems()
	if s.tab == tasksAll {
		return items
	}
	visible := make([]taskItem, 0, len(items))
	for _, item := range items {
		if taskInTab(item.record, s.tab) {
			visible = append(visible, item)
		}
	}
	return visible
}

func taskInTab(task records.Task, tab taskTab) bool {
	landed := taskIsLanded(task)
	switch tab {
	case tasksActive:
		return !landed && (task.Status == "assigned" || task.Status == "completed" || task.Status == "verified")
	case tasksQueued:
		return task.Status == "created"
	case tasksLanded:
		return landed
	case tasksFailed:
		verification, _ := latestTaskVerification(task)
		return !landed && task.Status != "verified" && verification != nil && verification.Result == "fail"
	default:
		return true
	}
}

func taskIsLanded(task records.Task) bool {
	return task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed
}

func latestTaskVerification(task records.Task) (*records.Verification, int) {
	var latest *records.Verification
	index := 0
	for i, history := range task.VerificationHistory {
		if history.Verification == nil {
			continue
		}
		latest = history.Verification
		index = i + 1
	}
	if task.Verification != nil {
		latest = task.Verification
		index = len(task.VerificationHistory) + 1
	}
	return latest, index
}

func taskVerificationText(task records.Task) string {
	verification, index := latestTaskVerification(task)
	if verification != nil {
		if verification.Result == "fail" {
			return fmt.Sprintf("V%d fail", verificationLabelNumber(verification.Reason, index))
		}
		if verification.Result == "pass" {
			if task.Delivery != nil && task.Delivery.Commit != "" {
				return "pass " + shortCommit(task.Delivery.Commit)
			}
			return "pass"
		}
	}
	if task.Status == "assigned" || task.Status == "completed" || task.Status == "verified" {
		return "running"
	}
	return "-"
}

func verificationLabelNumber(reason string, fallback int) int {
	if match := verificationNumberPattern.FindStringSubmatch(reason); len(match) == 2 {
		var number int
		_, _ = fmt.Sscan(match[1], &number)
		if number > 0 {
			return number
		}
	}
	if fallback > 0 {
		return fallback
	}
	return 1
}

func shortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

type tasksScreen struct{ model *Model }

func (tasksScreen) DataRequest() dataRequest { return tasksDataRequest }

func (tasksScreen) KeyHints() string {
	return "tab filter · enter evidence · b brief · v verified · d deps graph · p project · ask in Chat · ↑/↓ move"
}

func (s tasksScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.tasks
	key := press.String()
	if state.fileOpen {
		if key == "esc" || key == "escape" {
			state.fileOpen = false
			state.fileLoading = false
			state.fileContent = ""
			return s.model.logAction("task file close", slog.String("file", state.fileTitle))
		}
		updated, cmd := state.fileViewport.Update(press)
		state.fileViewport = updated
		return cmd
	}
	switch key {
	case "tab":
		state.tab = (state.tab + 1) % taskTab(len(taskTabs))
		state.acceptSelection()
		return s.model.logAction("tasks tab", slog.String("tab", taskTabs[state.tab]))
	case "p":
		state.projectIndex = (state.projectIndex + 1) % (len(state.data.projects) + 1)
		state.acceptSelection()
		return s.model.logAction("tasks project", slog.String("project", state.projectFilter()))
	case "up", "k":
		previous := state.selectedID
		state.moveSelection(-1)
		if previous != state.selectedID {
			return s.model.logAction("tasks selection", slog.String("task_id", state.selectedID))
		}
	case "down", "j":
		previous := state.selectedID
		state.moveSelection(1)
		if previous != state.selectedID {
			return s.model.logAction("tasks selection", slog.String("task_id", state.selectedID))
		}
	case "enter":
		if selected, ok := state.selected(); ok && len(selected.evidence) != 0 {
			read := s.model.startTaskFileRead(state, filepath.Join(selected.projectPath, ".hive", "spec", selected.record.Feature, "tasks", selected.record.ID, "evidence", selected.evidence[0]), selected.evidence[0])
			return batch(s.model.logAction("task evidence open", slog.String("task_id", selected.record.ID), slog.String("file", selected.evidence[0])), read)
		}
	case "b":
		if selected, ok := state.selected(); ok && selected.projectPath != "" && selected.record.Feature != "" {
			path := filepath.Join(selected.projectPath, ".hive", "spec", selected.record.Feature, "tasks", selected.record.ID, "brief.md")
			read := s.model.startTaskFileRead(state, path, "brief.md")
			return batch(s.model.logAction("task brief open", slog.String("task_id", selected.record.ID)), read)
		}
	case "v":
		if selected, ok := state.selected(); ok && selected.projectPath != "" && selected.record.Feature != "" {
			path := filepath.Join(selected.projectPath, ".hive", "spec", selected.record.Feature, "tasks", selected.record.ID, "evidence", "VERIFIED.md")
			read := s.model.startTaskFileRead(state, path, "VERIFIED.md")
			return batch(s.model.logAction("task verification evidence open", slog.String("task_id", selected.record.ID)), read)
		}
	case "d":
		if selected, ok := state.selected(); ok {
			state.fileOpen = true
			state.fileLoading = false
			state.fileTitle = "dependency graph · " + selected.record.ID
			state.fileContent = strings.Join(taskDependencyGraphLines(state.data, selected), "\n")
			state.fileViewportSet = false
			return s.model.logAction("task dependency graph open", slog.String("task_id", selected.record.ID))
		}
	}
	return nil
}

func (s tasksScreen) View(width, height int) string {
	state := &s.model.tasks
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if state.fileOpen {
		return state.fileView(contentWidth, height-4)
	}
	visible := state.visibleItems()
	state.keepSelectedVisible()
	counts := state.counts()
	listLines := []string{titleStyle.Render(taskTabLine(state.tab, counts)), "project: " + state.projectFilter(), "ID             title                       project   status     agent · verification"}
	rowLines := make([]string, 0, len(visible))
	for _, item := range visible {
		prefix := "  "
		if item.record.ID == state.selectedID {
			prefix = "> "
		}
		line := fmt.Sprintf("%-14s %-27s %-9s %-10s %s", clipCell(item.record.ID, 14), clipCell(item.record.Title, 27), item.record.Project, taskDisplayStatus(item.record), item.record.ActingAgent+" · "+taskVerificationText(item.record))
		rowLines = append(rowLines, prefix+strings.TrimSpace(line))
	}
	if len(rowLines) == 0 {
		rowLines = append(rowLines, "  no tasks")
	}
	listHeight := height - 5
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
	listLines = append(listLines, rowLines[state.scrollOffset:end]...)
	for index, line := range listLines {
		listLines[index] = trimGraphLine(line, contentWidth)
	}

	detailLines := append([]string{titleStyle.Render("task details")}, state.detailLines()...)
	evidenceLines := append([]string{dimStyle.Render("┤ evidence ├")}, state.evidenceLines()...)
	note := "landed = delivery commit set and pushed · verified = verification passed"
	if contentWidth >= 120 {
		listWidth := contentWidth * 55 / 100
		if listWidth < 1 {
			listWidth = 1
		}
		rightWidth := contentWidth - listWidth - 1
		if rightWidth < 1 {
			rightWidth = 1
		}
		panelHeight := height - 4
		if panelHeight < 1 {
			panelHeight = 1
		}
		detailHeight := panelHeight / 3
		if detailHeight < 10 {
			detailHeight = 10
		}
		if detailHeight >= panelHeight {
			detailHeight = panelHeight - 1
		}
		if detailHeight < 1 {
			detailHeight = 1
		}
		evidenceHeight := panelHeight - detailHeight
		if evidenceHeight < 1 {
			evidenceHeight = 1
			detailHeight = panelHeight - evidenceHeight
		}
		for index, line := range listLines {
			listLines[index] = trimGraphLine(line, listWidth)
		}
		for index, line := range detailLines {
			detailLines[index] = trimGraphLine(line, rightWidth-4)
		}
		for index, line := range evidenceLines {
			evidenceLines[index] = trimGraphLine(line, rightWidth-4)
		}
		left := lipgloss.NewStyle().Width(listWidth).Height(panelHeight).Render(strings.Join(listLines, "\n"))
		details := panelStyle.Width(rightWidth).Height(detailHeight).Render(strings.Join(detailLines, "\n"))
		evidence := panelStyle.Width(rightWidth).Height(evidenceHeight).Render(strings.Join(evidenceLines, "\n"))
		panels := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", lipgloss.JoinVertical(lipgloss.Left, details, evidence))
		return panels + "\n" + trimGraphLine(note, contentWidth)
	}

	lines := append([]string(nil), listLines...)
	lines = append(lines, "┤ task details ├")
	lines = append(lines, state.detailLines()...)
	lines = append(lines, "┤ evidence ├")
	lines = append(lines, state.evidenceLines()...)
	lines = append(lines, trimGraphLine(note, contentWidth))
	for index, line := range lines {
		lines[index] = trimGraphLine(line, contentWidth)
	}
	return strings.Join(lines, "\n")
}

func taskTabLine(selected taskTab, counts [5]int) string {
	parts := make([]string, 0, len(taskTabs))
	for index, label := range taskTabs {
		text := fmt.Sprintf("%s %d", label, counts[index])
		if taskTab(index) == selected {
			text = selectedStyle.Render("[" + text + "]")
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "  ")
}

func (s *tasksState) counts() [5]int {
	var counts [5]int
	for _, item := range s.projectItems() {
		for index := range taskTabs {
			if taskInTab(item.record, taskTab(index)) {
				counts[index]++
			}
		}
	}
	counts[tasksAll] = len(s.projectItems())
	return counts
}

func taskDisplayStatus(task records.Task) string {
	if taskIsLanded(task) {
		return "landed"
	}
	return task.Status
}

func clipCell(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	}
	if width < 2 {
		return text[:width]
	}
	return text[:width-1] + "…"
}

func (s tasksState) selected() (taskItem, bool) {
	for _, item := range s.visibleItems() {
		if item.record.ID == s.selectedID {
			return item, true
		}
	}
	return taskItem{}, false
}

func (s *tasksState) acceptSelection() {
	for _, item := range s.visibleItems() {
		if item.record.ID == s.selectedID {
			s.keepSelectedVisible()
			return
		}
	}
	items := s.visibleItems()
	s.selectedID = ""
	if len(items) != 0 {
		s.selectedID = items[0].record.ID
	}
	s.scrollOffset = 0
}

func (s *tasksState) moveSelection(delta int) {
	items := s.visibleItems()
	if len(items) == 0 {
		s.selectedID = ""
		return
	}
	index := 0
	for i, item := range items {
		if item.record.ID == s.selectedID {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}
	s.selectedID = items[index].record.ID
	s.keepSelectedVisible()
}

func (s *tasksState) keepSelectedVisible() {
	if s.scrollHeight < 1 {
		return
	}
	line := 0
	for index, item := range s.visibleItems() {
		if item.record.ID == s.selectedID {
			line = index
			break
		}
	}
	if line < s.scrollOffset {
		s.scrollOffset = line
	} else if line >= s.scrollOffset+s.scrollHeight {
		s.scrollOffset = line - s.scrollHeight + 1
	}
	maxOffset := len(s.visibleItems()) - s.scrollHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.scrollOffset > maxOffset {
		s.scrollOffset = maxOffset
	}
}

func (s tasksState) detailLines() []string {
	item, ok := s.selected()
	if !ok {
		return []string{"no task selected"}
	}
	task := item.record
	needs := make([]string, 0, len(task.DependsOn))
	for _, id := range task.DependsOn {
		mark := "?"
		for _, dependency := range s.data.items {
			if dependency.record.ID == id {
				mark = "○"
				if dependency.record.Status == "verified" || taskIsLanded(dependency.record) {
					mark = "✓"
				}
				break
			}
		}
		needs = append(needs, id+" "+mark)
	}
	blocks := make([]string, 0)
	for _, candidate := range s.data.items {
		for _, dependency := range candidate.record.DependsOn {
			if dependency == task.ID {
				blocks = append(blocks, candidate.record.ID)
				break
			}
		}
	}
	branch := "-"
	if task.Worktree != nil && task.Worktree.Branch != "" {
		branch = task.Worktree.Branch
	}
	feature := task.Feature
	if feature == "" {
		feature = "-"
	}
	grant := item.goalGrant
	if grant == "" && task.Delivery != nil {
		grant = task.Delivery.Grant
	}
	if grant == "" {
		grant = "-"
	}
	needsUser := task.NeedsUser
	if needsUser == "" {
		needsUser = "-"
	}
	return []string{
		titleStyle.Render(task.ID + " · " + task.Title),
		"needs: " + joinOrDash(needs),
		"blocks: " + joinOrDash(blocks),
		"branch: " + branch,
		"feature: " + feature,
		"grant: " + grant,
		"needs user: " + needsUser,
	}
}

func (s tasksState) evidenceLines() []string {
	item, ok := s.selected()
	if !ok {
		return []string{"no task selected"}
	}
	task := item.record
	brief := "-"
	if task.Feature != "" {
		brief = filepath.ToSlash(filepath.Join("tasks", task.ID, "brief.md"))
	}
	evidence := append([]string(nil), item.evidence...)
	if len(evidence) == 0 {
		evidence = []string{"-"}
	}
	lines := []string{"brief: " + brief, "evidence: " + strings.Join(evidence, ", ")}
	for index, history := range task.VerificationHistory {
		if history.Verification == nil {
			continue
		}
		reason := history.Verification.Reason
		if reason == "" {
			reason = history.Reason
		} else if history.Reason != "" && history.Reason != reason {
			reason += " · reopen: " + history.Reason
		}
		at := history.Verification.At
		if at == "" {
			at = history.At
		}
		lines = append(lines, fmt.Sprintf("V%d %s · %s · %s", verificationLabelNumber(reason, index+1), history.Verification.Result, reason, at))
	}
	if task.Verification != nil {
		verification := task.Verification
		reason := verification.Reason
		if reason == "" {
			reason = "-"
		}
		lines = append(lines, fmt.Sprintf("current %s · %s · %s", verification.Result, reason, verification.At))
	}
	if len(task.VerificationHistory) == 0 && task.Verification == nil {
		lines = append(lines, "verification history: -")
	}
	return lines
}

func taskDependencyGraphLines(data tasksData, selected taskItem) []string {
	task := selected.record
	lines := []string{task.ID + " · " + task.Title, "needs:"}
	if len(task.DependsOn) == 0 {
		lines = append(lines, "  -")
	} else {
		for _, dependency := range task.DependsOn {
			lines = append(lines, task.ID+" ← "+dependency+" (needs)")
		}
	}
	blocks := make([]string, 0)
	for _, candidate := range data.items {
		for _, dependency := range candidate.record.DependsOn {
			if dependency == task.ID {
				blocks = append(blocks, candidate.record.ID)
				break
			}
		}
	}
	sort.Strings(blocks)
	lines = append(lines, "blocks:")
	if len(blocks) == 0 {
		lines = append(lines, "  -")
	} else {
		for _, blocked := range blocks {
			lines = append(lines, task.ID+" → "+blocked+" (blocks)")
		}
	}
	return lines
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func (s *tasksState) fileView(width, height int) string {
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
	title := s.fileTitle
	if s.fileLoading {
		s.fileViewport.SetContent("loading…")
	} else {
		s.fileViewport.SetContent(s.fileContent)
	}
	return title + "\n" + s.fileViewport.View()
}

func (m *Model) startTaskFileRead(state *tasksState, path, title string) tea.Cmd {
	state.fileOpen = true
	state.fileLoading = true
	state.fileTitle = title
	state.fileContent = ""
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		return taskFileMsg{title: title, content: string(data), err: err}
	}
}

func (m *Model) acceptTaskFile(msg taskFileMsg) {
	m.tasks.fileLoading = false
	if msg.err != nil {
		m.tasks.fileContent = "error: " + boundedError(msg.err.Error())
		return
	}
	m.tasks.fileTitle = msg.title
	m.tasks.fileContent = msg.content
	m.tasks.fileViewport.SetContent(msg.content)
	m.tasks.fileViewport.GotoTop()
}
