package tui

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// The columns of the goals list and of the waves table of the details.
const (
	goalIDWidth     = 24
	goalStatusWidth = 10
)

type goalsScreen struct{ model *Model }

func (goalsScreen) DataRequest() dataRequest { return goalsQueueDataRequest }

func (goalsScreen) KeyHints() string {
	return "↑/↓ select goal · t tasks · o design · ask in Chat · 1–8 screens"
}

func (s goalsScreen) View(width, height int) string {
	return s.model.goalsList(width, height, false)
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
	case "g":
		moved = state.moveGoal(-len(state.data.goals))
	case "G":
		moved = state.moveGoal(len(state.data.goals))
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
		// Demo mode reads no live file. The fixture text or the demo notice
		// takes the place of this check.
		if s.model.demo == nil {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				state.designNotice = "no design for this wave"
				return s.model.logAction("goal design unavailable", slog.String("goal_id", goal.ID), slog.String("wave_id", wave.ID), slog.String("feature", feature))
			}
		}
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

// goalsList draws the list of goals at exactly width×height cells. The
// selected goal stays visible.
func (m *Model) goalsList(width, height int, focused bool) string {
	state := &m.goalsQueue
	goals := state.data.goals
	visible := max(height, component.PanelMinHeight) - 3
	selected := goalIndex(goals, state.selectedGoal)
	off := min(state.goalOffset, max(selected, 0))
	off = max(off, selected-visible+1)
	off = min(max(off, 0), max(len(goals)-visible, 0))
	state.goalOffset = off
	lines := []string{"  " + dimStyle.Render(padCell("GOAL", goalIDWidth)+padCell("STATUS", goalStatusWidth)+"WAVES")}
	if len(goals) == 0 {
		lines = append(lines, "  "+dimStyle.Render("no goals"))
	}
	for index := off; index < min(off+max(visible, 1), len(goals)); index++ {
		lines = append(lines, goalRow(goals[index], index == selected))
	}
	mark := component.Thumb(visible, len(goals), visible, off)
	if mark.Size > 0 {
		mark.Start++ // the track starts below the header row
	}
	title := fmt.Sprintf("goals · %d", len(goals))
	return component.Panel{Title: title, Focused: focused, Lines: lines, VMark: mark}.View(width, height)
}

// goalRow gives one row of the goals list: the marker, the ID, the status,
// and the done waves of all waves.
func goalRow(goal records.Goal, selected bool) string {
	marker, id := "  ", padCell(cutCell(goal.ID, goalIDWidth-1), goalIDWidth)
	if selected {
		marker, id = titleStyle.Render("▶")+" ", titleStyle.Render(id)
	}
	done := 0
	for _, wave := range goal.Waves {
		if wave.Status == "done" {
			done++
		}
	}
	status := goalStatus(goal.Status)
	return marker + id + goalStatusStyle(status).Render(padCell(cutCell(status, goalStatusWidth-1), goalStatusWidth)) + dimStyle.Render(fmt.Sprintf("%d/%d", done, len(goal.Waves)))
}

// goalStatus gives the status word of a goal or a wave. The word is the
// plain-text signal of the state, and the color only adds to it.
func goalStatus(status string) string {
	if status == "" {
		return "unknown"
	}
	return status
}

func goalStatusStyle(status string) lipgloss.Style {
	switch status {
	case "active":
		return workingStyle
	case "planned":
		return infoStyle
	case "done":
		return doneStyle
	}
	return lipgloss.NewStyle()
}

// goalDetails gives the title and the lines of the details of the selected
// goal: its fields, its waves, its tasks, its stops, and its action keys.
func (m *Model) goalDetails(width int) (string, []string) {
	state := &m.goalsQueue
	index := goalIndex(state.data.goals, state.selectedGoal)
	if index < 0 {
		return "details", []string{"no goal selected"}
	}
	goal := state.data.goals[index]
	tasks := state.taskMap()
	status := goalStatus(goal.Status)
	lines := []string{
		"",
		component.DetailField("title", goal.Title),
		component.DetailField("status", goalStatusStyle(status).Render(padCell(status, width-14))),
		component.DetailField("project", goal.Project),
		component.DetailField("grant", goal.Grant),
		component.DetailField("manager", "manager-"+goal.Project),
		"",
		component.DetailSection("WAVES"),
		dimStyle.Render(padCell("WAVE", 5) + padCell("TITLE", 27) + padCell("DONE WHEN", 11) + padCell("LANDED", 8) + "STATUS"),
	}
	var taskLines, stops []string
	for _, wave := range goal.Waves {
		landed := 0
		for _, id := range wave.Tasks {
			task, ok := tasks[id]
			if ok && taskLanded(task) {
				landed++
			}
			taskLines = append(taskLines, padCell(dimStyle.Render(cutCell(wave.ID, 3)), 4)+padCell(cutCell(id, 29), 30)+dimStyle.Render(goalTaskState(task, ok)))
			if ok && task.NeedsUser != "" {
				stops = append(stops, id+"  "+task.NeedsUser)
			}
		}
		if wave.NeedsUser != "" {
			stops = append(stops, "wave "+wave.ID+"  "+wave.NeedsUser)
		}
		lines = append(lines, titleStyle.Render(padCell(cutCell(wave.ID, 4), 4))+" "+
			padCell(cutCell(wave.Title, 26), 27)+padCell(cutCell(wave.DoneWhen, 10), 11)+
			padCell(fmt.Sprintf("%d/%d", landed, len(wave.Tasks)), 8)+waveState(wave.Status))
	}
	if len(goal.Waves) == 0 {
		lines = append(lines, dimStyle.Render("no waves"))
	}
	lines = append(lines, "", component.DetailSection("TASKS"))
	if len(taskLines) == 0 {
		taskLines = []string{dimStyle.Render("no tasks")}
	}
	lines = append(lines, taskLines...)
	lines = append(lines, "", component.DetailSection("NEEDS YOU"))
	if len(stops) == 0 {
		stops = []string{dimStyle.Render("nothing")}
	}
	lines = append(lines, stops...)
	lines = append(lines, "", titleStyle.Render("o")+" goal design   "+titleStyle.Render("t")+" open task in Tasks")
	return goal.ID, lines
}

// waveState gives the glyph and the word of a wave status.
func waveState(status string) string {
	status = goalStatus(status)
	if status == "done" {
		return doneStyle.Render("✓") + " " + doneStyle.Render(status)
	}
	return dimStyle.Render("○ " + status)
}

// goalTaskState gives the delivery of a landed task, or the status of
// another task.
func goalTaskState(task records.Task, ok bool) string {
	switch {
	case !ok:
		return "missing"
	case taskLanded(task):
		return task.Delivery.Mode + " " + task.Delivery.Commit[:min(8, len(task.Delivery.Commit))]
	}
	return taskMark(task) + " " + taskStatusLabel(task.Status)
}

func goalIndex(goals []records.Goal, id string) int {
	for index, goal := range goals {
		if goal.ID == id {
			return index
		}
	}
	return -1
}

// cutCell cuts text to width cells with ….
func cutCell(text string, width int) string { return ansi.Truncate(text, max(width, 0), "…") }
