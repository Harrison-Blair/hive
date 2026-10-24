package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// taskItem is one task of the Work board with the data that its popup
// shows. Brief and verified are the text of brief.md and VERIFIED.md, or ""
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

// tasksData is the data of the Work board: the goals in creation order, the
// tasks in ID order, the first project, and each agent by key and by name.
type tasksData struct {
	items   []taskItem
	goals   []records.Goal
	project string
	agents  map[string]records.Agent
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
	goals = append([]records.Goal(nil), goals...)
	sort.SliceStable(goals, func(i, j int) bool {
		if goals[i].CreatedAt != goals[j].CreatedAt {
			return goals[i].CreatedAt < goals[j].CreatedAt
		}
		return goals[i].ID < goals[j].ID
	})
	type place struct{ goal, wave string }
	places := make(map[string]place)
	for _, goal := range goals {
		for _, wave := range goal.Waves {
			for _, taskID := range wave.Tasks {
				places[taskID] = place{goal.ID, wave.ID}
			}
		}
	}
	data := tasksData{items: make([]taskItem, 0, len(tasks)), goals: goals, agents: make(map[string]records.Agent, 2*len(agents))}
	if len(projects) > 0 {
		data.project = projects[0].ID
	}
	for _, agent := range agents {
		data.agents[agent.Name] = agent
		data.agents[agent.Key] = agent
	}
	tasks = append([]records.Task(nil), tasks...)
	sort.Slice(tasks, func(i, j int) bool { return taskIDLess(tasks[i].ID, tasks[j].ID) })
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

// item gives the task with id.
func (d tasksData) item(id string) (taskItem, bool) {
	for _, item := range d.items {
		if item.record.ID == id {
			return item, true
		}
	}
	return taskItem{}, false
}

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

// latestTaskVerification gives the last verification of a task, or nil.
func latestTaskVerification(task records.Task) *records.Verification {
	if task.Verification != nil {
		return task.Verification
	}
	for index := len(task.VerificationHistory) - 1; index >= 0; index-- {
		if verification := task.VerificationHistory[index].Verification; verification != nil {
			return verification
		}
	}
	return nil
}
