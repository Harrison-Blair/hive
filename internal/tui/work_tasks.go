package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// taskItem is one task of the Tasks view with the data that its details
// show. Brief and verified are the text of brief.md and VERIFIED.md, or ""
// when the file is missing.
type taskItem struct {
	record      records.Task
	projectPath string
	goal        string
	wave        string
	evidence    []string
	brief       string
	verified    string
}

// tasksData is the data of the Tasks view. Agents gives the record status of
// each agent by key and by name.
type tasksData struct {
	items  []taskItem
	agents map[string]string
}

// tasksState is the state of the Tasks view. A file of the selected task, or
// its dependencies, can take the place of its details.
type tasksState struct {
	data         tasksData
	selectedID   string
	scrollOffset int
	fileOpen     bool
	fileLoading  bool
	fileTitle    string
	fileContent  string
}

type taskFileMsg struct {
	title   string
	content string
	err     error
}

// taskFiles reads the evidence names and the text of one file of a task.
type taskFiles struct {
	evidence func(taskItem) ([]string, error)
	read     func(item taskItem, rel string) (string, error)
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
	agents, err := listRecords[records.Agent](registry, "agents")
	if err != nil {
		return tasksData{}, fmt.Errorf("agents: %w", err)
	}
	return buildTasks(tasks, projects, goals, agents, taskFiles{evidence: readTaskEvidence, read: readTaskFile})
}

// taskDir gives the path of the task folder below a project directory.
func taskDir(task records.Task) string {
	return filepath.Join(".hive", "spec", task.Feature, "tasks", task.ID)
}

// readTaskEvidence lists the regular files in the evidence folder of a task.
func readTaskEvidence(item taskItem) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(item.projectPath, taskDir(item.record), "evidence"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// readTaskFile reads a file of a task. A missing file gives "".
func readTaskFile(item taskItem, rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(item.projectPath, rel))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}

func buildTasks(tasks []records.Task, projects []records.Project, goals []records.Goal, agents []records.Agent, files taskFiles) (tasksData, error) {
	projectPaths := make(map[string]string, len(projects))
	for _, project := range projects {
		projectPaths[project.ID] = project.Path
	}
	type place struct{ goal, wave string }
	places := make(map[string]place)
	for _, goal := range goals {
		for _, wave := range goal.Waves {
			for _, taskID := range wave.Tasks {
				places[taskID] = place{goal.ID, wave.ID}
			}
		}
	}
	data := tasksData{items: make([]taskItem, 0, len(tasks)), agents: make(map[string]string, 2*len(agents))}
	for _, agent := range agents {
		data.agents[agent.Name] = agent.Status
		data.agents[agent.Key] = agent.Status
	}
	for _, task := range tasks {
		at := places[task.ID]
		item := taskItem{record: task, projectPath: projectPaths[task.Project], goal: at.goal, wave: at.wave}
		if item.projectPath != "" && task.Feature != "" {
			names, err := files.evidence(item)
			if err != nil {
				return tasksData{}, fmt.Errorf("task evidence %s: %w", task.ID, err)
			}
			item.evidence = append(item.evidence, names...)
			sort.Strings(item.evidence)
			if item.brief, err = files.read(item, filepath.Join(taskDir(task), "brief.md")); err != nil {
				return tasksData{}, fmt.Errorf("task brief %s: %w", task.ID, err)
			}
			if item.verified, err = files.read(item, filepath.Join(taskDir(task), "evidence", "VERIFIED.md")); err != nil {
				return tasksData{}, fmt.Errorf("task verification %s: %w", task.ID, err)
			}
		}
		data.items = append(data.items, item)
	}
	return data, nil
}

func (s *tasksState) accept(data tasksData) {
	s.data = data
	s.acceptSelection()
}

// acceptSelection keeps the selected task, or selects the first task when the
// selected task is gone.
func (s *tasksState) acceptSelection() {
	if _, ok := s.selected(); ok {
		return
	}
	s.selectedID = ""
	if len(s.data.items) != 0 {
		s.selectedID = s.data.items[0].record.ID
	}
	s.scrollOffset = 0
}

func (s tasksState) selectedIndex() int {
	for index, item := range s.data.items {
		if item.record.ID == s.selectedID {
			return index
		}
	}
	return -1
}

func (s tasksState) selected() (taskItem, bool) {
	if index := s.selectedIndex(); index >= 0 {
		return s.data.items[index], true
	}
	return taskItem{}, false
}

func taskIsLanded(task records.Task) bool {
	return task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed
}

func taskDisplayStatus(task records.Task) string {
	if taskIsLanded(task) {
		return "landed"
	}
	return task.Status
}

// taskStatusStyle gives the token of a task status: done for a landed task,
// info for each other state.
func taskStatusStyle(task records.Task) lipgloss.Style {
	if taskIsLanded(task) {
		return doneStyle
	}
	return infoStyle
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

// orDash gives text, or — when text is empty.
func orDash(text string) string {
	if text == "" {
		return "—"
	}
	return text
}

type workTasksScreen struct{ model *Model }

func (workTasksScreen) DataRequest() dataRequest { return tasksDataRequest }

func (workTasksScreen) KeyHints() string { return "" }

func (s workTasksScreen) View(width, height int) string {
	return s.model.workTasksList(width, height, false)
}

func (s workTasksScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	state := &m.tasks
	key := press.String()
	switch key {
	case "esc":
		if state.closeFile() {
			return m.logAction("task file close", slog.String("file", state.fileTitle))
		}
	case "up", "k", "down", "j", "pgup", "pgdown", "g", "G", "home", "end":
		return m.workTasksMove(key)
	case "enter":
		if _, ok := state.selected(); ok {
			m.regionFocus[tabWork] = 1
			m.detailsOffset[tabWork] = 0
			return m.logAction("details open", slog.String("region", "details"))
		}
	case "b", "v":
		selected, ok := state.selected()
		if !ok || selected.projectPath == "" || selected.record.Feature == "" {
			return nil
		}
		name, rel := "brief.md", filepath.Join(taskDir(selected.record), "brief.md")
		if key == "v" {
			name, rel = "VERIFIED.md", filepath.Join(taskDir(selected.record), "evidence", "VERIFIED.md")
		}
		read := m.startTaskFileRead(state, filepath.Join(selected.projectPath, rel), name)
		return batch(m.logAction("task file open", slog.String("task_id", selected.record.ID), slog.String("file", name)), read)
	case "d":
		if selected, ok := state.selected(); ok {
			m.openTaskDetails()
			state.fileOpen = true
			state.fileLoading = false
			state.fileTitle = "dependency graph · " + selected.record.ID
			state.fileContent = strings.Join(taskDependencyLines(state.data, selected), "\n")
			return m.logAction("task dependency graph open", slog.String("task_id", selected.record.ID))
		}
	}
	return nil
}

// openTaskDetails shows the details from the top. Below 120 columns it opens
// the details overlay (decision 9).
func (m *Model) openTaskDetails() {
	m.detailsOffset[tabWork] = 0
	if m.width < component.LongFormWidth {
		m.regionFocus[tabWork] = 1
	}
}

// closeFile closes an open file and reports that one was open.
func (s *tasksState) closeFile() bool {
	if !s.fileOpen {
		return false
	}
	s.fileOpen = false
	s.fileLoading = false
	s.fileContent = ""
	return true
}

// workTasksMove moves the selection by key. A new selection closes an open
// file and shows its details from the top.
func (m *Model) workTasksMove(key string) tea.Cmd {
	state := &m.tasks
	list := m.workTasksListModel(false)
	height := m.workTasksBodyHeight()
	switch key {
	case "up", "k":
		list.Move(-1, height)
	case "down", "j":
		list.Move(1, height)
	case "pgup":
		list.Move(-max(height-3, 1), height)
	case "pgdown":
		list.Move(max(height-3, 1), height)
	case "g", "home":
		list.Move(-len(list.Rows), height)
	case "G", "end":
		list.Move(len(list.Rows), height)
	}
	state.scrollOffset = list.Offset
	if len(state.data.items) == 0 || state.data.items[list.Selected].record.ID == state.selectedID {
		return nil
	}
	state.selectedID = state.data.items[list.Selected].record.ID
	state.closeFile()
	m.detailsOffset[tabWork] = 0
	return m.logAction("tasks selection", slog.String("task_id", state.selectedID))
}

// workTasksBodyHeight gives the height of the list panel: the content area
// less the view row.
func (m *Model) workTasksBodyHeight() int {
	return max(m.height-frameRows-1, 1)
}

func (m *Model) workTasksListModel(focused bool) component.List {
	state := &m.tasks
	dim := dimStyle
	rows := make([][]component.Cell, 0, len(state.data.items))
	for _, item := range state.data.items {
		rows = append(rows, []component.Cell{
			{Text: item.record.ID},
			{Text: taskDisplayStatus(item.record), Style: taskStatusStyle(item.record)},
			{Text: orDash(item.wave), Style: dim},
		})
	}
	return component.List{
		Title:    fmt.Sprintf("tasks · %d", len(rows)),
		Focused:  focused,
		Columns:  []component.Column{{Title: "ID", Width: 26}, {Title: "STATUS", Width: 10}, {Title: "WAVE"}},
		Rows:     rows,
		Selected: max(state.selectedIndex(), 0),
		Offset:   state.scrollOffset,
		Fill:     true,
	}
}

// workTasksList renders the list region of the Tasks view.
func (m *Model) workTasksList(width, height int, focused bool) string {
	return m.workTasksListModel(focused).View(width, height)
}

// workTasksDetails renders the details region of the Tasks view: the details
// of the selected task, or its open file. Below 120 columns the details use
// the short form of the Compact board, with no blank first line and only the
// fields status, goal, verifier, and merge.
func (m *Model) workTasksDetails(width, height int, focused bool, suffix string) string {
	state := &m.tasks
	inner := max(width-4, 1)
	wide := m.width >= component.LongFormWidth
	title := "details"
	var lines []string
	switch item, ok := state.selected(); {
	case state.fileOpen:
		title = state.fileTitle
		text := state.fileContent
		if state.fileLoading {
			text = "loading…"
		}
		lines = strings.Split(ansi.Wrap(text, inner, ""), "\n")
	case !ok:
		lines = []string{"no task selected"}
	default:
		title = item.record.ID
		lines = taskDetailLines(state.data, item, inner, wide)
	}
	return taskDetailsPanel(title+suffix, focused, lines, width, height, &m.detailsOffset[tabWork])
}

// taskDetailsPanel draws lines in a panel from line *offset. Each line that
// fits shows. When lines are left below the view, the last body row is
// "↓ N more lines" (decision 2 of 2026-10-21). The scroll mark shows only
// when the lines do not fit.
func taskDetailsPanel(title string, focused bool, lines []string, width, height int, offset *int) string {
	body := max(height, component.PanelMinHeight) - 2
	*offset = min(max(*offset, 0), max(len(lines)-body, 0))
	end := min(*offset+body, len(lines))
	var vmark component.ScrollMark
	if len(lines) > body {
		vmark = component.Thumb(body, len(lines), body-1, *offset)
		if end < len(lines) {
			end--
		}
	}
	shown := append([]string(nil), lines[*offset:end]...)
	if rest := len(lines) - end; rest > 0 {
		shown = append(shown, dimStyle.Render(fmt.Sprintf("↓ %d more lines", rest)))
	}
	return component.Panel{Title: title, Focused: focused, Lines: shown, VMark: vmark}.View(width, height)
}

// taskDetailLines gives the details of one task at inner cells: the fields,
// the acceptance criteria of its brief, and its evidence.
func taskDetailLines(data tasksData, item taskItem, inner int, wide bool) []string {
	task := item.record
	field := func(label, value string, style lipgloss.Style) string {
		return component.DetailField(label, style.Render(fitWidth(value, inner-10)))
	}
	plain := lipgloss.NewStyle()
	goal := ""
	if item.goal != "" {
		goal = item.goal + " · wave " + item.wave
	}
	worker := ""
	if task.ActingAgent != "" {
		worker = task.ActingAgent
		if status := data.agents[task.ActingAgent]; status != "" {
			worker += " · " + status
		}
	}
	verifier := ""
	verification, _ := latestTaskVerification(task)
	if verification != nil {
		verifier = verification.Agent + " · " + verification.Result
	}
	branch := ""
	if task.Worktree != nil {
		branch = task.Worktree.Branch
	}
	merge := ""
	if task.Delivery != nil && task.Delivery.Commit != "" {
		merge = shortCommit(task.Delivery.Commit)
		if task.Delivery.Target != "" {
			merge += " on " + task.Delivery.Target
		}
	}
	status := field("status", taskDisplayStatus(task), taskStatusStyle(task))
	var lines []string
	if wide {
		lines = []string{
			"",
			field("title", task.Title, plain),
			status,
			field("goal", orDash(goal), plain),
			field("worker", orDash(worker), plain),
			field("verifier", orDash(verifier), plain),
			field("branch", orDash(branch), plain),
			field("merge", orDash(merge), plain),
			field("depends", orDash(strings.Join(task.DependsOn, ", ")), plain),
		}
	} else {
		lines = []string{status, field("goal", orDash(goal), plain), field("verifier", orDash(verifier), plain), field("merge", orDash(merge), plain)}
	}
	lines = append(lines, "", component.DetailSection("ACCEPTANCE"))
	mark := dimStyle.Render("○")
	if verification != nil && verification.Result == "pass" {
		mark = doneStyle.Render("✓")
	}
	criteria := markdownRows(item.brief, "AC")
	if len(criteria) == 0 {
		lines = append(lines, dimStyle.Render("no acceptance criteria"))
	}
	for _, row := range criteria {
		text := row[0]
		if len(row) > 1 {
			text += " " + row[1]
		}
		lines = append(lines, mark+" "+fitWidth(text, inner-2))
	}
	key := func(k string) string { return titleStyle.Render(k) }
	lines = append(lines, "", component.DetailSection("EVIDENCE"),
		key("b")+" brief.md    "+key("v")+" VERIFIED.md   "+key("d")+" deps graph", "")
	for _, row := range markdownRows(item.verified, "V") {
		if len(row) < 3 {
			continue
		}
		lines = append(lines, dimStyle.Render(row[0]+" "+padCell(row[1], verifiedCommandWidth)+row[2]))
	}
	return lines
}

// verifiedCommandWidth is the cells of the command column of an evidence
// line, with its gap.
const verifiedCommandWidth = 38

// markdownRows gives the cells of each row of a markdown table in text whose
// first cell starts with prefix and a digit.
func markdownRows(text, prefix string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for index := range cells {
			cells[index] = strings.TrimSpace(cells[index])
		}
		first := cells[0]
		if len(first) > len(prefix) && strings.HasPrefix(first, prefix) && first[len(prefix)] >= '0' && first[len(prefix)] <= '9' {
			rows = append(rows, cells)
		}
	}
	return rows
}

// fitWidth cuts text to width cells with …, or pads it with spaces.
func fitWidth(text string, width int) string {
	return padCell(ansi.Truncate(text, max(width, 0), "…"), width)
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

func shortCommit(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}

// taskDependencyLines gives the tasks that the selected task needs and the
// tasks that it blocks, with the status of each.
func taskDependencyLines(data tasksData, selected taskItem) []string {
	task := selected.record
	status := make(map[string]string, len(data.items))
	for _, item := range data.items {
		status[item.record.ID] = taskDisplayStatus(item.record)
	}
	lines := []string{component.DetailSection("NEEDS")}
	for _, dependency := range task.DependsOn {
		lines = append(lines, task.ID+" ← "+dependency+" · "+orDash(status[dependency]))
	}
	if len(task.DependsOn) == 0 {
		lines = append(lines, "—")
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
	lines = append(lines, "", component.DetailSection("BLOCKS"))
	for _, blocked := range blocks {
		lines = append(lines, task.ID+" → "+blocked+" · "+status[blocked])
	}
	if len(blocks) == 0 {
		lines = append(lines, "—")
	}
	return lines
}

// startTaskFileRead opens a file in the details of the Tasks view. The Goals
// view opens a goal design here too.
func (m *Model) startTaskFileRead(state *tasksState, path, title string) tea.Cmd {
	m.openTaskDetails()
	state.fileOpen = true
	state.fileLoading = true
	state.fileTitle = title
	state.fileContent = ""
	return func() tea.Msg {
		content, err := m.readScreenFile(path)
		return taskFileMsg{title: title, content: content, err: err}
	}
}

func (m *Model) acceptTaskFile(msg taskFileMsg) {
	m.tasks.fileLoading = false
	if errors.Is(msg.err, errDemoNoFile) {
		m.tasks.fileContent = demoNoFile
		m.demo.notice = demoNoFile
		return
	}
	if msg.err != nil {
		m.tasks.fileContent = "error: " + boundedError(msg.err.Error())
		return
	}
	m.tasks.fileTitle = msg.title
	m.tasks.fileContent = msg.content
}
