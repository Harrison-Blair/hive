package tui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type goalsQueueData struct {
	goals    []records.Goal
	projects []records.Project
	tasks    []records.Task
}

type goalsQueueState struct {
	data             goalsQueueData
	selectedGoal     string
	selectedTaskID   string
	goalOffset       int
	designNotice     string
	showStops        bool
}

func loadGoalsQueue(home string) (goalsQueueData, error) {
	registry, err := store.Open(home)
	if err != nil {
		return goalsQueueData{}, err
	}
	goals, err := listRecordsIn[records.Goal](registry, []string{"goals"})
	if err != nil {
		return goalsQueueData{}, fmt.Errorf("goals: %w", err)
	}
	projects, err := listRecordsIn[records.Project](registry, []string{"projects"})
	if err != nil {
		return goalsQueueData{}, fmt.Errorf("projects: %w", err)
	}
	tasks, err := listRecordsIn[records.Task](registry, []string{"tasks"})
	if err != nil {
		return goalsQueueData{}, fmt.Errorf("tasks: %w", err)
	}
	return buildGoalsQueue(goals, projects, tasks), nil
}

func buildGoalsQueue(goals []records.Goal, projects []records.Project, tasks []records.Task) goalsQueueData {
	sort.Slice(goals, func(i, j int) bool {
		if left, right := goalRank(goals[i].Status), goalRank(goals[j].Status); left != right {
			return left < right
		}
		return goals[i].ID < goals[j].ID
	})
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
	sort.Slice(tasks, func(i, j int) bool { return taskIDLess(tasks[i].ID, tasks[j].ID) })
	return goalsQueueData{goals: goals, projects: projects, tasks: tasks}
}

// goalRank orders goals: active, then planned, then other goals, then done
// goals.
func goalRank(status string) int {
	switch status {
	case "active":
		return 0
	case "planned":
		return 1
	case "done":
		return 3
	}
	return 2
}

func listRecordsIn[T any](registry *store.Store, collection []string) ([]T, error) {
	raw, err := registry.ListIn(collection, nil)
	if err != nil {
		return nil, err
	}
	result := make([]T, 0, len(raw))
	for _, data := range raw {
		var record T
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %v", err)
		}
		result = append(result, record)
	}
	return result, nil
}

func (s *goalsQueueState) accept(data goalsQueueData) {
	s.data = data
	if !s.hasGoal(s.selectedGoal) {
		s.selectedGoal = ""
		for _, goal := range data.goals {
			if goal.Status == "active" {
				s.selectedGoal = goal.ID
				break
			}
		}
		if s.selectedGoal == "" && len(data.goals) > 0 {
			s.selectedGoal = data.goals[0].ID
		}
	}
	s.keepQueueSelection()
}

func (s goalsQueueState) hasGoal(id string) bool {
	for _, goal := range s.data.goals {
		if goal.ID == id {
			return true
		}
	}
	return false
}

func (s goalsQueueState) selectedWave() (*records.Goal, *records.GoalWave) {
	for i := range s.data.goals {
		goal := &s.data.goals[i]
		if goal.ID != s.selectedGoal {
			continue
		}
		for j := range goal.Waves {
			if goal.Waves[j].Status != "done" {
				return goal, &goal.Waves[j]
			}
		}
		return goal, nil
	}
	return nil, nil
}

func (s *goalsQueueState) moveGoal(delta int) bool {
	if len(s.data.goals) == 0 {
		return false
	}
	index := 0
	for i, goal := range s.data.goals {
		if goal.ID == s.selectedGoal {
			index = i
			break
		}
	}
	next := index + delta
	if next < 0 {
		next = 0
	}
	if next >= len(s.data.goals) {
		next = len(s.data.goals) - 1
	}
	if next == index {
		return false
	}
	s.selectedGoal = s.data.goals[next].ID
	return true
}

type queueScreen struct{ model *Model }

func (queueScreen) DataRequest() dataRequest { return goalsQueueDataRequest }

func (queueScreen) KeyHints() string {
	return "↑/↓ select task · enter open task · s stops only · order comes from the dependencies · 1–8 screens"
}

func (s queueScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.goalsQueue
	switch press.String() {
	case "up", "k":
		if state.moveQueueTask(-1) {
			return s.model.logAction("queue task selection", slog.String("task_id", state.selectedTaskID))
		}
	case "down", "j":
		if state.moveQueueTask(1) {
			return s.model.logAction("queue task selection", slog.String("task_id", state.selectedTaskID))
		}
	case "s":
		state.showStops = !state.showStops
		state.keepQueueSelection()
		return s.model.logAction("queue stop filter", slog.Bool("stops_only", state.showStops))
	case "enter":
		if state.selectedTaskID == "" {
			return nil
		}
		taskID := state.selectedTaskID
		s.model.screenIndex = 4
		s.model.tasks.tab = tasksAll
		s.model.tasks.projectIndex = 0
		s.model.tasks.selectedID = taskID
		s.model.tasks.scrollOffset = 0
		s.model.tasks.fileOpen = false
		s.model.tasks.fileLoading = false
		s.model.screenLoading = false
		return batch(s.model.logAction("queue open task", slog.String("task_id", taskID)), s.model.startCurrentScreenLoad())
	}
	return nil
}

func (s goalsQueueState) taskMap() map[string]records.Task {
	result := make(map[string]records.Task, len(s.data.tasks))
	for _, task := range s.data.tasks {
		result[task.ID] = task
	}
	return result
}

func taskLanded(task records.Task) bool {
	return task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed
}

func taskMark(task records.Task) string {
	if taskLanded(task) {
		return "✓"
	}
	if task.NeedsUser != "" {
		return "✋"
	}
	switch task.Status {
	case "assigned", "completed", "verified":
		return "▶"
	case "cancelled":
		return "✕"
	default:
		return "·"
	}
}

func isOpenTask(task records.Task) bool {
	if taskLanded(task) {
		return false
	}
	switch task.Status {
	case "created", "assigned", "completed", "verified":
		return true
	default:
		return false
	}
}

type queueRow struct {
	task    records.Task
	kind    string
	start   string
	state   string
	stopped bool
	waiting bool
}

func (s queueScreen) View(width, _ int) string {
	state := s.model.goalsQueue
	projects := state.queueProjects()
	cardWidth := width - 4
	if cardWidth < 1 {
		cardWidth = 1
	}
	blocks := make([]string, 0, len(projects))
	for _, project := range projects {
		allRows := state.projectQueue(project)
		rows := allRows
		if state.showStops {
			filtered := rows[:0]
			for _, row := range rows {
				if row.stopped || row.waiting {
					filtered = append(filtered, row)
				}
			}
			rows = filtered
		}
		goal := state.activeGoal(project)
		goalLabel, grant := "no active goal", "—"
		if goal != nil {
			goalLabel, grant = goal.Title, goal.Grant
		}
		lines := []string{titleStyle.Render(project + " · " + goalLabel + " · grant: " + grant), "#  task  starts when  state"}
		if len(rows) == 0 {
			emptyLabel := "(no open tasks)"
			if state.showStops && len(allRows) > 0 {
				emptyLabel = "(no tasks needing you)"
			}
			lines = append(lines, emptyLabel)
		}
		for index, row := range rows {
			position := row.kind
			if row.kind != "now" {
				position = strconv.Itoa(index - countQueueNow(rows) + 1)
			}
			line := fmt.Sprintf("%s  %s  %s  %s", position, row.task.ID, row.start, row.state)
			if row.task.ID == state.selectedTaskID {
				line = selectedStyle.Render("▶ " + line)
			}
			lines = append(lines, line)
		}
		blocks = append(blocks, panelStyle.Width(cardWidth).Render(strings.Join(lines, "\n")))
	}
	if len(blocks) == 0 {
		return "no open tasks"
	}
	return strings.Join(blocks, "\n\n")
}

func (s goalsQueueState) visibleQueueRows() []queueRow {
	rows := make([]queueRow, 0)
	for _, project := range s.queueProjects() {
		for _, row := range s.projectQueue(project) {
			if !s.showStops || row.stopped || row.waiting {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func (s *goalsQueueState) keepQueueSelection() {
	rows := s.visibleQueueRows()
	for _, row := range rows {
		if row.task.ID == s.selectedTaskID {
			return
		}
	}
	s.selectedTaskID = ""
	if len(rows) > 0 {
		s.selectedTaskID = rows[0].task.ID
	}
}

func (s *goalsQueueState) moveQueueTask(delta int) bool {
	rows := s.visibleQueueRows()
	if len(rows) == 0 {
		return false
	}
	index := 0
	for i, row := range rows {
		if row.task.ID == s.selectedTaskID {
			index = i
			break
		}
	}
	next := index + delta
	if next < 0 {
		next = 0
	}
	if next >= len(rows) {
		next = len(rows) - 1
	}
	if next == index {
		return false
	}
	s.selectedTaskID = rows[next].task.ID
	return true
}

func queueTaskCounts(tasks []records.Task) (running, stopped, waiting int) {
	taskMap := make(map[string]records.Task, len(tasks))
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	for _, task := range tasks {
		if !isOpenTask(task) {
			continue
		}
		if task.NeedsUser != "" {
			stopped++
		} else if dependencyHasStop(task, taskMap, make(map[string]bool)) {
			waiting++
		} else {
			running++
		}
	}
	return running, stopped, waiting
}

func (s goalsQueueState) queueProjects() []string {
	set := make(map[string]bool)
	for _, project := range s.data.projects {
		set[project.ID] = true
	}
	for _, task := range s.data.tasks {
		if isOpenTask(task) {
			set[task.Project] = true
		}
	}
	projects := make([]string, 0, len(set))
	for project := range set {
		projects = append(projects, project)
	}
	sort.Strings(projects)
	return projects
}

func (s goalsQueueState) activeGoal(project string) *records.Goal {
	for i := range s.data.goals {
		if s.data.goals[i].Project == project && s.data.goals[i].Status == "active" {
			return &s.data.goals[i]
		}
	}
	return nil
}

func (s goalsQueueState) projectQueue(project string) []queueRow {
	taskMap := s.taskMap()
	all := make([]records.Task, 0)
	for _, task := range s.data.tasks {
		if task.Project == project && isOpenTask(task) {
			all = append(all, task)
		}
	}
	now := make([]records.Task, 0)
	created := make([]records.Task, 0)
	for _, task := range all {
		if task.Status == "created" {
			created = append(created, task)
		} else {
			now = append(now, task)
		}
	}
	sort.Slice(now, func(i, j int) bool { return taskIDLess(now[i].ID, now[j].ID) })
	created = dependencyOrder(created)
	rows := make([]queueRow, 0, len(all))
	for _, task := range now {
		state := "▶ " + taskStatusLabel(task.Status)
		start := "—"
		stopped := task.NeedsUser != ""
		waiting := !stopped && dependencyHasStop(task, taskMap, make(map[string]bool))
		if stopped {
			state, start = "✋ you", taskStart(task, taskMap)
		} else if waiting {
			state, start = "┄ waits", "you"
		}
		rows = append(rows, queueRow{task: task, kind: "now", start: start, state: state, stopped: stopped, waiting: waiting})
	}
	for _, task := range created {
		stopped := task.NeedsUser != ""
		waiting := !stopped && dependencyHasStop(task, taskMap, make(map[string]bool))
		start := taskStart(task, taskMap)
		state := "◆ auto"
		if stopped {
			state = "✋ you"
		} else if waiting {
			state = "┄ waits"
		}
		rows = append(rows, queueRow{task: task, kind: "created", start: start, state: state, stopped: stopped, waiting: waiting})
	}
	return rows
}

func countQueueNow(rows []queueRow) int {
	count := 0
	for _, row := range rows {
		if row.kind == "now" {
			count++
		}
	}
	return count
}

func taskStatusLabel(status string) string {
	switch status {
	case "assigned":
		return "assigned"
	case "completed":
		return "verify"
	case "verified":
		return "verified"
	default:
		return status
	}
}

func taskStart(task records.Task, tasks map[string]records.Task) string {
	deps := make([]string, 0, len(task.DependsOn))
	for _, id := range task.DependsOn {
		deps = append(deps, id+" lands")
	}
	start := "now"
	if len(deps) == 1 {
		start = deps[0]
	} else if len(deps) > 1 {
		start = strings.Join(deps[:len(deps)-1], ", ") + " and " + deps[len(deps)-1]
	}
	if task.NeedsUser != "" {
		if start == "now" {
			return "you"
		}
		return start + ", then you"
	}
	if dependencyHasStop(task, tasks, make(map[string]bool)) {
		return "you"
	}
	return start
}

func dependencyHasStop(task records.Task, tasks map[string]records.Task, seen map[string]bool) bool {
	if seen[task.ID] {
		return false
	}
	seen[task.ID] = true
	for _, id := range task.DependsOn {
		dependency, ok := tasks[id]
		if !ok {
			continue
		}
		if dependency.NeedsUser != "" || dependencyHasStop(dependency, tasks, seen) {
			return true
		}
	}
	return false
}

func dependencyOrder(created []records.Task) []records.Task {
	remaining := make(map[string]records.Task, len(created))
	for _, task := range created {
		remaining[task.ID] = task
	}
	ordered := make([]records.Task, 0, len(created))
	for len(remaining) > 0 {
		ready := make([]records.Task, 0)
		for _, task := range remaining {
			blocked := false
			for _, id := range task.DependsOn {
				if _, exists := remaining[id]; exists {
					blocked = true
					break
				}
			}
			if !blocked {
				ready = append(ready, task)
			}
		}
		if len(ready) == 0 {
			for _, task := range remaining {
				ready = append(ready, task)
			}
		}
		sort.Slice(ready, func(i, j int) bool { return taskIDLess(ready[i].ID, ready[j].ID) })
		for _, task := range ready {
			ordered = append(ordered, task)
			delete(remaining, task.ID)
		}
	}
	return ordered
}

func taskIDLess(left, right string) bool {
	leftPrefix, _, _ := strings.Cut(left, "-")
	rightPrefix, _, _ := strings.Cut(right, "-")
	leftNumber, leftErr := strconv.ParseUint(leftPrefix, 10, 64)
	rightNumber, rightErr := strconv.ParseUint(rightPrefix, 10, 64)
	if leftErr == nil && rightErr == nil && leftNumber != rightNumber {
		return leftNumber < rightNumber
	}
	return left < right
}
