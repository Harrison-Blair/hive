package tui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	goalScrollOffset int
	goalScrollHeight int
	goalViewWidth    int
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
	sort.Slice(goals, func(i, j int) bool {
		if goals[i].Status != goals[j].Status {
			return goals[i].Status == "active"
		}
		return goals[i].ID < goals[j].ID
	})
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
	sort.Slice(tasks, func(i, j int) bool { return taskIDLess(tasks[i].ID, tasks[j].ID) })
	return goalsQueueData{goals: goals, projects: projects, tasks: tasks}, nil
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
	s.keepGoalVisible()
	return true
}

func (s *goalsQueueState) keepGoalVisible() {
	if s.goalScrollHeight < 1 {
		return
	}
	width := s.goalViewWidth
	if width < 1 {
		width = 200
	}
	blocks, positions := s.goalBlocks(width)
	start, ok := positions[s.selectedGoal]
	if !ok {
		return
	}
	end := start + lipgloss.Height(blocks[goalBlockIndex(s.data.goals, s.selectedGoal)])
	if start < s.goalScrollOffset {
		s.goalScrollOffset = start
	} else if end > s.goalScrollOffset+s.goalScrollHeight {
		s.goalScrollOffset = end - s.goalScrollHeight
	}
	if s.goalScrollOffset < 0 {
		s.goalScrollOffset = 0
	}
}

func goalBlockIndex(goals []records.Goal, id string) int {
	for i, goal := range goals {
		if goal.ID == id {
			return i
		}
	}
	return 0
}

type goalsScreen struct{ model *Model }

type queueScreen struct{ model *Model }

func (goalsScreen) DataRequest() dataRequest { return goalsQueueDataRequest }
func (queueScreen) DataRequest() dataRequest { return goalsQueueDataRequest }

func (goalsScreen) KeyHints() string {
	return "↑/↓ select goal · t tasks · o design · ask in Chat · 1–8 screens"
}
func (queueScreen) KeyHints() string {
	return "↑/↓ select task · enter open task · s stops only · order comes from the dependencies · 1–8 screens"
}

func (s goalsScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.goalsQueue
	state.designNotice = ""
	var moved bool
	switch press.String() {
	case "up", "k":
		moved = state.moveGoal(-1)
	case "down", "j":
		moved = state.moveGoal(1)
	case "t":
		goal, wave := state.selectedWave()
		if goal == nil || wave == nil || len(wave.Tasks) == 0 {
			return nil
		}
		taskID := wave.Tasks[0]
		s.model.screenIndex = 4
		s.model.tasks.tab = tasksAll
		s.model.tasks.projectIndex = 0
		s.model.tasks.selectedID = taskID
		s.model.tasks.scrollOffset = 0
		s.model.tasks.fileOpen = false
		s.model.tasks.fileLoading = false
		s.model.screenLoading = false
		return batch(s.model.logAction("goal open wave tasks", slog.String("goal_id", goal.ID), slog.String("wave_id", wave.ID), slog.String("task_id", taskID)), s.model.startCurrentScreenLoad())
	case "o":
		goal, wave := state.selectedWave()
		if goal == nil || wave == nil {
			state.designNotice = "no design for this wave"
			return s.model.logAction("goal design unavailable")
		}
		tasks := state.taskMap()
		feature := ""
		for _, id := range wave.Tasks {
			if task, ok := tasks[id]; ok && task.Feature != "" {
				feature = task.Feature
				break
			}
		}
		if feature == "" {
			state.designNotice = "no design for this wave"
			return s.model.logAction("goal design unavailable", slog.String("goal_id", goal.ID), slog.String("wave_id", wave.ID))
		}
		path := filepath.Join(s.model.home, ".hive", "spec", feature, "README.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			state.designNotice = "no design for this wave"
			return s.model.logAction("goal design unavailable", slog.String("goal_id", goal.ID), slog.String("wave_id", wave.ID), slog.String("feature", feature))
		}
		state.designNotice = ""
		s.model.screenIndex = 4
		s.model.tasks.tab = tasksAll
		s.model.tasks.fileOpen = false
		s.model.tasks.fileLoading = false
		s.model.screenLoading = false
		read := s.model.startTaskFileRead(&s.model.tasks, path, "README.md")
		return batch(s.model.logAction("goal design open", slog.String("goal_id", goal.ID), slog.String("wave_id", wave.ID), slog.String("feature", feature)), read)
	}
	if moved {
		return s.model.logAction("goal selection", slog.String("goal_id", state.selectedGoal))
	}
	return nil
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

func (s goalsScreen) View(width, height int) string {
	state := &s.model.goalsQueue
	blocks, positions := state.goalBlocks(width)
	state.goalViewWidth = width
	lines := strings.Split(strings.Join(blocks, "\n\n"), "\n")
	state.goalScrollHeight = height - 2
	if state.goalScrollHeight < 1 {
		state.goalScrollHeight = 1
	}
	if start, ok := positions[state.selectedGoal]; ok {
		if start < state.goalScrollOffset {
			state.goalScrollOffset = start
		}
		end := start + lipgloss.Height(blocks[goalBlockIndex(state.data.goals, state.selectedGoal)])
		if end > state.goalScrollOffset+state.goalScrollHeight {
			state.goalScrollOffset = end - state.goalScrollHeight
		}
	}
	if state.goalScrollOffset < 0 {
		state.goalScrollOffset = 0
	}
	if state.goalScrollOffset >= len(lines) {
		state.goalScrollOffset = 0
	}
	end := min(len(lines), state.goalScrollOffset+state.goalScrollHeight)
	content := strings.Join(lines[state.goalScrollOffset:end], "\n")
	if state.designNotice != "" {
		return state.designNotice + "\n" + content
	}
	return content
}

func (s goalsQueueState) goalBlocks(width int) ([]string, map[string]int) {
	cardWidth := width - 4
	if cardWidth < 1 {
		cardWidth = 1
	}
	blocks := make([]string, 0, len(s.data.goals)+1)
	positions := make(map[string]int, len(s.data.goals))
	line := 0
	for _, goal := range s.data.goals {
		positions[goal.ID] = line
		lines := goalLines(goal, s.taskMap(), goal.ID == s.selectedGoal)
		block := panelStyle.Width(cardWidth).Render(strings.Join(lines, "\n"))
		if goal.Status != "active" {
			block = dimStyle.Render(block)
		}
		blocks = append(blocks, block)
		line += lipgloss.Height(block) + 1
	}
	unassigned := s.unassignedTasks()
	lines := []string{titleStyle.Render("unassigned")}
	if len(unassigned) == 0 {
		lines = append(lines, "(none)")
	} else {
		for _, project := range sortedTaskProjects(unassigned) {
			ids := make([]string, 0)
			for _, task := range unassigned {
				if task.Project == project {
					ids = append(ids, taskMark(task)+" "+task.ID)
				}
			}
			lines = append(lines, project+": "+strings.Join(ids, "  "))
		}
	}
	blocks = append(blocks, panelStyle.Width(cardWidth).Render(strings.Join(lines, "\n")))
	return blocks, positions
}

func goalLines(goal records.Goal, tasks map[string]records.Task, selected bool) []string {
	current := "—"
	for _, wave := range goal.Waves {
		if wave.Status != "done" {
			current = doneWhenLabel(wave.DoneWhen)
			break
		}
	}
	prefix := "  "
	if selected {
		prefix = selectedStyle.Render("▶ ")
	}
	goalStatus := goal.Status
	if goalStatus == "" {
		goalStatus = "unknown"
	}
	lines := []string{titleStyle.Render(prefix + goal.Project + " · " + goal.Title + " · " + goalStatus + " · done when: " + current)}
	for _, wave := range goal.Waves {
		landed := 0
		for _, id := range wave.Tasks {
			if task, ok := tasks[id]; ok && taskLanded(task) {
				landed++
			}
		}
		segments := []string{doneStyle.Render("dispatch ✓")}
		waveStatus := wave.Status
		if waveStatus == "" {
			waveStatus = "unknown"
		}
		progressMark := "▶"
		taskStyle := workingStyle
		if landed == len(wave.Tasks) {
			progressMark, taskStyle = "✓", doneStyle
		}
		segments = append(segments, taskStyle.Render(fmt.Sprintf("tasks %d/%d %s", landed, len(wave.Tasks), progressMark)))
		points := []struct{ key, label string }{
			{key: "local", label: "local implementation"},
			{key: "dev", label: "landed on dev"},
			{key: "pr", label: "PR open"},
			{key: "main", label: "merged to main"},
			{key: "release", label: "release published"},
		}
		currentIndex := -1
		for i, point := range points {
			if point.key == wave.DoneWhen {
				currentIndex = i
				break
			}
		}
		if currentIndex < 0 {
			segments = append(segments, selectedStyle.Render("◎ "+doneWhenLabel(wave.DoneWhen)))
		} else {
			for i := currentIndex; i < len(points); i++ {
				point := points[i]
				text := point.label
				if i == currentIndex {
					text = selectedStyle.Render("◎ " + text)
				} else {
					text = dimStyle.Render(text)
				}
				segments = append(segments, text)
			}
		}
		stageLine := strings.Join(segments, doneStyle.Render(" ─▶ "))
		lines = append(lines, titleStyle.Render("wave "+wave.ID+" · "+wave.Title+" · "+waveStatus), stageLine)
		if wave.NeedsUser != "" {
			lines = append(lines, "✋ "+wave.NeedsUser)
		}
		if len(wave.Tasks) == 0 {
			lines = append(lines, "(no tasks)")
		} else {
			marks := make([]string, 0, len(wave.Tasks))
			for _, id := range wave.Tasks {
				if task, ok := tasks[id]; ok {
					marks = append(marks, taskMark(task)+" "+id)
				} else {
					marks = append(marks, "· "+id)
				}
			}
			lines = append(lines, strings.Join(marks, "  "))
		}
	}
	return lines
}

func doneWhenLabel(point string) string {
	switch point {
	case "local":
		return "local implementation"
	case "dev":
		return "landed on dev"
	case "pr":
		return "PR open"
	case "main":
		return "merged to main"
	case "release":
		return "release published"
	default:
		return point
	}
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

func (s goalsQueueState) unassignedTasks() []records.Task {
	assigned := make(map[string]bool)
	for _, goal := range s.data.goals {
		for _, wave := range goal.Waves {
			for _, id := range wave.Tasks {
				assigned[id] = true
			}
		}
	}
	result := make([]records.Task, 0)
	for _, task := range s.data.tasks {
		if isOpenTask(task) && !assigned[task.ID] {
			result = append(result, task)
		}
	}
	return result
}

func sortedTaskProjects(tasks []records.Task) []string {
	set := make(map[string]bool)
	for _, task := range tasks {
		set[task.Project] = true
	}
	projects := make([]string, 0, len(set))
	for project := range set {
		projects = append(projects, project)
	}
	sort.Strings(projects)
	return projects
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
