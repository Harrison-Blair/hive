// Package list renders the task list of the parent task.List API.
package list

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Text renders a task list.
func Text(envelope output.Envelope) string {
	tasks, ok := envelope.Result.([]records.Task)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(tasks) == 0 {
		return "no tasks\n"
	}
	lines := make([]string, 0, len(tasks))
	for _, task := range tasks {
		line := fmt.Sprintf("%s\t%s\t%s\t%s",
			task.ID, task.Status, tasklib.OneLine(task.ActingAgent), tasklib.OneLine(task.Title))
		if task.NeedsUser != "" {
			line += " [needs user]"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}
