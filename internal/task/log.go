package task

import (
	"log/slog"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// Log writes task events with the task component.
type Log struct {
	Logger *slog.Logger
	closer func()
}

// OpenLog uses an injected logger or opens the task log writer.
func OpenLog(home string, injected *slog.Logger) Log {
	if injected != nil {
		return Log{Logger: injected, closer: func() {}}
	}
	writer, err := hivelog.NewWriter(home, "task", time.Now)
	if err != nil {
		return Log{closer: func() {}}
	}
	return Log{Logger: slog.New(writer), closer: func() { _ = writer.Close() }}
}

// Close releases the task log writer.
func (l Log) Close() {
	l.closer()
}

// Fields returns the identifiers that apply to one event.
func (l Log) Fields(command, projectID, taskID, agentID string) []any {
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	if agentID != "" {
		fields = append(fields, "agent_id", agentID)
	}
	return fields
}

// Transition writes one completed state transition.
func (l Log) Transition(command, projectID, taskID, agentID, from, to string, extra ...any) {
	if l.Logger == nil {
		return
	}
	fields := l.Fields(command, projectID, taskID, agentID)
	if from != "" {
		fields = append(fields, "from_status", from)
	}
	fields = append(fields, "to_status", to)
	fields = append(fields, extra...)
	l.Logger.Debug("task transition", fields...)
}

// Effect writes one completed write in the task prefix order.
func (l Log) Effect(command, projectID, taskID, resource, path, action string) {
	if l.Logger == nil {
		return
	}
	fields := l.Fields(command, projectID, taskID, "")
	fields = append(fields, "resource", resource, "path", path, "action", action)
	l.Logger.Debug("task effect", fields...)
}

// Failure writes one task error.
func (l Log) Failure(command, projectID, taskID, agentID, phase, message string) {
	if l.Logger == nil {
		return
	}
	fields := l.Fields(command, projectID, taskID, agentID)
	if phase != "" {
		fields = append(fields, "phase", phase)
	}
	fields = append(fields, "error", message)
	l.Logger.Error("task error", fields...)
}
