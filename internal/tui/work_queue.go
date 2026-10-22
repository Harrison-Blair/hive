package tui

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// The columns of a Queue row: the status starts at this content cell, and a
// task ID takes at most nameWidth cells.
const (
	queueStatusAt   = 38
	queueNameWidth  = 30
	queueStatusSize = 10
)

var (
	queueBorderStyle = lipgloss.NewStyle().Foreground(component.Border)
	queueAccentStyle = lipgloss.NewStyle().Foreground(component.Accent)
	queuePlainStyle  = lipgloss.NewStyle()
)

// queueGroup is the tasks of one project under one grant. The rows are the
// dependency tree of the group: a task sits below the first task of the
// group that it depends on.
type queueGroup struct {
	project string
	grant   string
	rows    []queueTreeRow
}

type queueTreeRow struct {
	task  records.Task
	depth int
	stop  string
}

type queueScreen struct{ model *Model }

func (queueScreen) DataRequest() dataRequest { return goalsQueueDataRequest }

func (queueScreen) KeyHints() string {
	return "↑↓ move · enter open · s stops only · t task"
}

// View gives the list lines of the Queue view. The frame draws the Queue
// view with queueList and queueDetails.
func (s queueScreen) View(width, _ int) string {
	lines, _ := s.model.goalsQueue.queueLines()
	return strings.Join(lines, "\n")
}

func (s queueScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	state := &m.goalsQueue
	page := max(state.queueBody-1, 1)
	switch press.String() {
	case "up", "k":
		return m.queueMove(-1)
	case "down", "j":
		return m.queueMove(1)
	case "pgup":
		return m.queueMove(-page)
	case "pgdown":
		return m.queueMove(page)
	case "g":
		return m.queueMove(-len(state.data.tasks))
	case "G":
		return m.queueMove(len(state.data.tasks))
	case "s":
		state.showStops = !state.showStops
		state.keepQueueSelection()
		return m.logAction("queue stop filter", slog.Bool("stops_only", state.showStops))
	case "enter":
		// Decision 8: from 120 columns the details sit beside the list, and
		// enter gives them the focus. Below 120 columns the frame opens
		// the overlay before the key gets here.
		if state.selectedTaskID == "" {
			return nil
		}
		m.regionFocus[tabWork] = 1
		return m.logAction("focus region", slog.String("region", "details"))
	case "t":
		if state.selectedTaskID == "" {
			return nil
		}
		taskID := state.selectedTaskID
		m.regionFocus[tabWork] = 0
		m.selectScreen(tabWork, 2)
		m.tasks.selectedID = taskID
		m.tasks.scrollOffset = 0
		m.tasks.fileOpen = false
		m.tasks.fileLoading = false
		m.screenLoading = false
		return batch(m.logAction("queue open task", slog.String("task_id", taskID)), m.startCurrentScreenLoad())
	}
	return nil
}

func (m *Model) queueMove(delta int) tea.Cmd {
	if m.goalsQueue.moveQueueTask(delta) {
		return m.logAction("queue task selection", slog.String("task_id", m.goalsQueue.selectedTaskID))
	}
	return nil
}

// queueList draws the list region of the Queue view at width×height cells.
// It scrolls so that the selected task stays visible.
func (m *Model) queueList(width, height int, focused bool) string {
	state := &m.goalsQueue
	lines, selected := state.queueLines()
	body := max(height-2, 1)
	state.queueBody = body
	if selected >= 0 {
		state.queueOffset = min(state.queueOffset, selected)
		state.queueOffset = max(state.queueOffset, selected-body+1)
	}
	state.queueOffset = min(max(state.queueOffset, 0), max(len(lines)-body, 0))
	panel := component.Panel{
		Title:   state.queueTitle(),
		Focused: focused,
		Lines:   lines,
		VMark:   component.Thumb(body, len(lines), body, state.queueOffset),
	}
	return component.Viewport{Panel: panel, YOffset: state.queueOffset}.View(width, height)
}

// queueDetails gives the title and the lines of the details of the selected
// task, for a details panel of width cells.
func (m *Model) queueDetails(width int) (string, []string) {
	state := m.goalsQueue
	task, ok := state.queueSelected()
	if !ok {
		return "details", []string{"", dimStyle.Render("no task selected")}
	}
	valueWidth := max(width-4-10, 1)
	field := func(label, value string, style lipgloss.Style) string {
		return component.DetailField(label, style.Render(fitText(value, valueWidth)))
	}
	status, statusStyle := queueStatus(task)
	goal, grant := "none", "none"
	if g, wave := state.taskGoal(task.ID); g != nil {
		goal = g.ID + " · wave " + wave.ID
	}
	if text := state.taskGrant(task); text != "" {
		grant = text
	}
	worker := "none"
	if task.ActingAgent != "" {
		worker = task.ActingAgent
		if agent, ok := state.agent(task.ActingAgent); ok {
			worker += " · " + agent.Status
		}
	}
	stop := state.stopReason(task)
	workerStyle := queuePlainStyle
	if stop != "" {
		workerStyle = queueAccentStyle
	}
	depends := "none"
	if len(task.DependsOn) > 0 {
		depends = strings.Join(task.DependsOn, ", ")
	}
	lines := []string{
		"",
		field("title", task.Title, queuePlainStyle),
		field("status", status, statusStyle),
		field("goal", goal, queuePlainStyle),
		field("grant", grant, queuePlainStyle),
		field("worker", worker, workerStyle),
		field("depends", depends, queuePlainStyle),
		"",
	}
	if stop != "" {
		explain := stop
		if task.NeedsUser == "" {
			explain = "The worker record is live, but its pane is gone."
		}
		lines = append(lines,
			component.DetailSection("STOP"),
			explain,
			"The task cannot move until the user decides.",
			"",
			component.DetailSection("NEXT STEP"),
			dimStyle.Render("Ask the coordinator in 2 Chat."),
			"",
		)
	}
	filter := "show stops only"
	if state.showStops {
		filter = "show all tasks"
	}
	lines = append(lines, titleStyle.Render("t")+" open task in Tasks  "+titleStyle.Render("s")+" "+filter)
	return task.ID, lines
}

// queueTitle gives the title of the list: the project when there is one,
// the task count, and the stop count.
func (s goalsQueueState) queueTitle() string {
	parts := []string{"queue"}
	if projects := s.queueProjectIDs(); len(projects) == 1 {
		parts = append(parts, projects[0])
	}
	stops := 0
	for _, task := range s.data.tasks {
		if s.stopReason(task) != "" {
			stops++
		}
	}
	return strings.Join(append(parts, plural(len(s.data.tasks), "task"), plural(stops, "stop")), " · ")
}

func plural(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

// queueLines gives the body lines of the list and the index of the line of
// the selected task, or -1.
func (s goalsQueueState) queueLines() ([]string, int) {
	groups := s.queueGroups()
	lines := []string{""}
	selected := -1
	if len(groups) == 0 {
		empty := "no tasks"
		if s.showStops {
			empty = "no stops"
		}
		return append(lines, dimStyle.Render(empty)), selected
	}
	multi := len(s.queueProjectIDs()) > 1
	for index, group := range groups {
		if index > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, dimStyle.Render(group.header(multi)), "")
		for _, row := range group.rows {
			if row.task.ID == s.selectedTaskID {
				selected = len(lines)
			}
			lines = append(lines, row.line(row.task.ID == s.selectedTaskID))
			if row.stop != "" {
				lines = append(lines, strings.Repeat(" ", row.nameAt())+queueAccentStyle.Render("! stop: "+row.stop))
			}
		}
	}
	return lines, selected
}

func (g queueGroup) header(multi bool) string {
	text := "NO GRANT"
	switch {
	case g.rows == nil:
		text = "no tasks"
	case g.grant != "":
		text = "GRANT  " + g.grant
	}
	if multi || g.rows == nil {
		text = g.project + " · " + text
	}
	return text
}

// nameAt gives the content cell where the task ID of the row starts.
func (r queueTreeRow) nameAt() int {
	if r.depth == 0 {
		return 2
	}
	return 2*r.depth + 3
}

// line draws one task row: the tree mark, the task ID, and the status.
func (r queueTreeRow) line(selected bool) string {
	lead, rest := "  ", ""
	if r.depth > 0 {
		rest = strings.Repeat(" ", 2*r.depth-2) + queueBorderStyle.Render("└▶") + " "
	}
	nameStyle := queuePlainStyle
	if selected {
		lead = titleStyle.Render("▶") + " "
		nameStyle = titleStyle
	}
	width := min(queueNameWidth, max(queueStatusAt-1-r.nameAt(), 1))
	gap := strings.Repeat(" ", max(queueStatusAt-r.nameAt()-width, 0))
	status, statusStyle := queueStatus(r.task)
	return lead + rest + nameStyle.Render(fitText(r.task.ID, width)) + gap + statusStyle.Render(fitText(status, queueStatusSize))
}

// queueStatus gives the status word of a task and its style.
func queueStatus(task records.Task) (string, lipgloss.Style) {
	if taskLanded(task) {
		return "landed", doneStyle
	}
	return taskStatusLabel(task.Status), infoStyle
}

// fitText cuts text to width cells with …, or pads it with spaces to width
// cells.
func fitText(text string, width int) string {
	text = ansi.Truncate(text, width, "…")
	return text + strings.Repeat(" ", max(width-ansi.StringWidth(text), 0))
}

// queueProjectIDs gives each registered project and each project of a task,
// in order.
func (s goalsQueueState) queueProjectIDs() []string {
	set := make(map[string]bool)
	for _, project := range s.data.projects {
		set[project.ID] = true
	}
	for _, task := range s.data.tasks {
		set[task.Project] = true
	}
	projects := make([]string, 0, len(set))
	for project := range set {
		projects = append(projects, project)
	}
	sort.Strings(projects)
	return projects
}

// queueGroups gives the groups of the list in order: by project, then each
// grant in order, then the tasks with no grant. With stops only, a group
// keeps only its stopped tasks, each at depth 0, and a project with no
// tasks has no group.
func (s goalsQueueState) queueGroups() []queueGroup {
	byKey := make(map[[2]string][]records.Task)
	for _, task := range s.data.tasks {
		key := [2]string{task.Project, s.taskGrant(task)}
		byKey[key] = append(byKey[key], task)
	}
	keys := make([][2]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, right := keys[i], keys[j]
		if left[0] != right[0] {
			return left[0] < right[0]
		}
		if (left[1] == "") != (right[1] == "") {
			return right[1] == ""
		}
		return left[1] < right[1]
	})
	var groups []queueGroup
	for _, project := range s.queueProjectIDs() {
		found := false
		for _, key := range keys {
			if key[0] != project {
				continue
			}
			found = true
			rows := s.queueTree(byKey[key])
			if s.showStops {
				stopped := make([]queueTreeRow, 0)
				for _, row := range rows {
					if row.stop != "" {
						row.depth = 0
						stopped = append(stopped, row)
					}
				}
				if len(stopped) == 0 {
					continue
				}
				rows = stopped
			}
			groups = append(groups, queueGroup{project: project, grant: key[1], rows: rows})
		}
		if !found && !s.showStops {
			groups = append(groups, queueGroup{project: project})
		}
	}
	return groups
}

// queueTree orders the tasks of one group as a dependency tree. A task in a
// dependency cycle starts its own tree.
func (s goalsQueueState) queueTree(tasks []records.Task) []queueTreeRow {
	sort.Slice(tasks, func(i, j int) bool { return taskIDLess(tasks[i].ID, tasks[j].ID) })
	inGroup := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		inGroup[task.ID] = true
	}
	children := make(map[string][]records.Task)
	var roots []records.Task
	for _, task := range tasks {
		parent := ""
		for _, id := range task.DependsOn {
			if id != task.ID && inGroup[id] {
				parent = id
				break
			}
		}
		if parent == "" {
			roots = append(roots, task)
		} else {
			children[parent] = append(children[parent], task)
		}
	}
	rows := make([]queueTreeRow, 0, len(tasks))
	seen := make(map[string]bool, len(tasks))
	var walk func(task records.Task, depth int)
	walk = func(task records.Task, depth int) {
		if seen[task.ID] {
			return
		}
		seen[task.ID] = true
		rows = append(rows, queueTreeRow{task: task, depth: depth, stop: s.stopReason(task)})
		for _, child := range children[task.ID] {
			walk(child, depth+1)
		}
	}
	for _, task := range roots {
		walk(task, 0)
	}
	for _, task := range tasks {
		walk(task, 0)
	}
	return rows
}

// stopReason gives the reason that an open task cannot move without the
// user, or "".
func (s goalsQueueState) stopReason(task records.Task) string {
	if !isOpenTask(task) {
		return ""
	}
	if task.NeedsUser != "" {
		return task.NeedsUser
	}
	if agent, ok := s.agent(task.ActingAgent); ok && agent.Status == records.AgentUnresolved {
		return "worker unresolved"
	}
	return ""
}

// agent gives the newest agent record with the name or the key.
func (s goalsQueueState) agent(name string) (records.Agent, bool) {
	var found records.Agent
	ok := false
	for _, agent := range s.data.agents {
		if name == "" || (agent.Name != name && agent.Key != name) {
			continue
		}
		if !ok || agent.Generation > found.Generation {
			found, ok = agent, true
		}
	}
	return found, ok
}

// taskGoal gives the goal and the wave that hold the task.
func (s goalsQueueState) taskGoal(id string) (*records.Goal, *records.GoalWave) {
	for i := range s.data.goals {
		goal := &s.data.goals[i]
		for j := range goal.Waves {
			for _, taskID := range goal.Waves[j].Tasks {
				if taskID == id {
					return goal, &goal.Waves[j]
				}
			}
		}
	}
	return nil, nil
}

// taskGrant gives the grant of the goal of the task, then the grant of its
// delivery.
func (s goalsQueueState) taskGrant(task records.Task) string {
	if goal, _ := s.taskGoal(task.ID); goal != nil && goal.Grant != "" {
		return goal.Grant
	}
	if task.Delivery != nil {
		return task.Delivery.Grant
	}
	return ""
}

// queueOrder gives the tasks of the list in the order of its rows.
func (s goalsQueueState) queueOrder() []records.Task {
	var tasks []records.Task
	for _, group := range s.queueGroups() {
		for _, row := range group.rows {
			tasks = append(tasks, row.task)
		}
	}
	return tasks
}

func (s goalsQueueState) queueSelected() (records.Task, bool) {
	for _, task := range s.queueOrder() {
		if task.ID == s.selectedTaskID {
			return task, true
		}
	}
	return records.Task{}, false
}

// keepQueueSelection keeps the selected task when the list shows it, and
// selects the first task when it does not.
func (s *goalsQueueState) keepQueueSelection() {
	tasks := s.queueOrder()
	if _, ok := s.queueSelected(); ok {
		return
	}
	s.selectedTaskID = ""
	if len(tasks) > 0 {
		s.selectedTaskID = tasks[0].ID
	}
}

func (s *goalsQueueState) moveQueueTask(delta int) bool {
	tasks := s.queueOrder()
	if len(tasks) == 0 {
		return false
	}
	index := 0
	for i, task := range tasks {
		if task.ID == s.selectedTaskID {
			index = i
			break
		}
	}
	next := min(max(index+delta, 0), len(tasks)-1)
	if next == index {
		return false
	}
	s.selectedTaskID = tasks[next].ID
	return true
}
