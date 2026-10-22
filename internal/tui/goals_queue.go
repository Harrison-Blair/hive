package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type goalsQueueData struct {
	goals    []records.Goal
	projects []records.Project
	tasks    []records.Task
	agents   []records.Agent
}

type goalsQueueState struct {
	data           goalsQueueData
	selectedGoal   string
	selectedTaskID string
	goalOffset     int
	designNotice   string
	showStops      bool
	queueOffset    int
	queueBody      int
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
	agents, err := listRecordsIn[records.Agent](registry, []string{"agents"})
	if err != nil {
		return goalsQueueData{}, fmt.Errorf("agents: %w", err)
	}
	return buildGoalsQueue(goals, projects, tasks, agents), nil
}

func buildGoalsQueue(goals []records.Goal, projects []records.Project, tasks []records.Task, agents []records.Agent) goalsQueueData {
	sort.Slice(goals, func(i, j int) bool {
		if left, right := goalRank(goals[i].Status), goalRank(goals[j].Status); left != right {
			return left < right
		}
		return goals[i].ID < goals[j].ID
	})
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
	sort.Slice(tasks, func(i, j int) bool { return taskIDLess(tasks[i].ID, tasks[j].ID) })
	return goalsQueueData{goals: goals, projects: projects, tasks: tasks, agents: agents}
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
