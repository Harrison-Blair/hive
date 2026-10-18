package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ListOptions holds the filters of one task list.
type ListOptions struct {
	Project string
	Status  string
	Ready   bool
}

// List returns task records in project and number order. A plain list takes no
// lock. The ready query uses one lock for the list and each dependency read.
func List(home string, options ListOptions) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	if !options.Ready {
		raw, err := registry.ListIn([]string{tasksCollection}, nil)
		if err != nil {
			return failedEnvelope("store", err.Error())
		}
		tasks, err := decodeTasks(raw)
		if err != nil {
			return failedEnvelope("store", err.Error())
		}
		return listEnvelope(tasks, options)
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		raw, err := locked.ListIn([]string{tasksCollection}, nil)
		if err != nil {
			envelope = failedEnvelope("store", err.Error())
			return nil
		}
		tasks, err := decodeTasks(raw)
		if err != nil {
			envelope = failedEnvelope("store", err.Error())
			return nil
		}
		candidates := filterTasks(tasks, options)
		ready := make([]records.Task, 0, len(candidates))
		for _, task := range candidates {
			if task.Status != "created" {
				continue
			}
			if dependenciesVerified(locked, task) {
				ready = append(ready, task)
			}
		}
		sortTasks(ready)
		envelope = okEnvelope(ready)
		return nil
	})
	if lockErr != nil {
		return failedEnvelope("store", lockErr.Error())
	}
	return envelope
}

// dependenciesVerified reports whether each dependency of the task is verified.
func dependenciesVerified(locked *store.LockedStore, task records.Task) bool {
	for _, dependencyID := range task.DependsOn {
		var dependency records.Task
		if err := locked.ReadIn([]string{tasksCollection}, dependencyID, &dependency); err != nil {
			return false
		}
		if dependency.Status != "verified" {
			return false
		}
	}
	return true
}

// decodeTasks decodes each task record.
func decodeTasks(raw []json.RawMessage) ([]records.Task, error) {
	tasks := make([]records.Task, 0, len(raw))
	for _, item := range raw {
		var task records.Task
		if err := json.Unmarshal(item, &task); err != nil {
			return nil, fmt.Errorf("record invalid: %v", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// filterTasks keeps the tasks that match the project and status filters.
func filterTasks(tasks []records.Task, options ListOptions) []records.Task {
	filtered := make([]records.Task, 0, len(tasks))
	for _, task := range tasks {
		if options.Project != "" && task.Project != options.Project {
			continue
		}
		if options.Status != "" && task.Status != options.Status {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered
}

// listEnvelope filters and sorts one lock-free list.
func listEnvelope(tasks []records.Task, options ListOptions) output.Envelope {
	filtered := filterTasks(tasks, options)
	sortTasks(filtered)
	return okEnvelope(filtered)
}

// sortTasks orders tasks by project, then by number, then by ID.
func sortTasks(tasks []records.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Project != tasks[j].Project {
			return tasks[i].Project < tasks[j].Project
		}
		left, right := taskNumber(tasks[i].ID), taskNumber(tasks[j].ID)
		if left != right {
			return left < right
		}
		return tasks[i].ID < tasks[j].ID
	})
}

// taskNumber returns the counter value in a task ID.
func taskNumber(id string) uint64 {
	prefix, _, ok := strings.Cut(id, "-")
	if !ok {
		return 0
	}
	number, err := strconv.ParseUint(prefix, 10, 64)
	if err != nil {
		return 0
	}
	return number
}

// okEnvelope builds one success outcome.
func okEnvelope(result any) output.Envelope {
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: result}
}
