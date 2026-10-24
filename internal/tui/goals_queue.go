package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func taskLanded(task records.Task) bool {
	return task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed
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

func plural(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

func fitText(text string, width int) string {
	text = ansi.Truncate(text, width, "…")
	return text + strings.Repeat(" ", max(width-ansi.StringWidth(text), 0))
}
