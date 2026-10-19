package task

import (
	"log/slog"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// operationLog writes task events with the task component.
type operationLog struct {
	logger *slog.Logger
	closer func()
}

// openOperationLog uses an injected logger or opens the task log writer.
func openOperationLog(home string, injected *slog.Logger) operationLog {
	if injected != nil {
		return operationLog{logger: injected, closer: func() {}}
	}
	writer, err := hivelog.NewWriter(home, "task", time.Now)
	if err != nil {
		return operationLog{closer: func() {}}
	}
	return operationLog{logger: slog.New(writer), closer: func() { _ = writer.Close() }}
}

// close releases the task log writer.
func (l operationLog) close() {
	l.closer()
}

// fields returns the identifiers that apply to one event.
func (l operationLog) fields(command, projectID, taskID, agentID string) []any {
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

// transition writes one completed state transition.
func (l operationLog) transition(command, projectID, taskID, agentID, from, to string, extra ...any) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, taskID, agentID)
	if from != "" {
		fields = append(fields, "from_status", from)
	}
	fields = append(fields, "to_status", to)
	fields = append(fields, extra...)
	l.logger.Debug("task transition", fields...)
}

// effect writes one completed write in the task prefix order.
func (l operationLog) effect(command, projectID, taskID, resource, path, action string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, taskID, "")
	fields = append(fields, "resource", resource, "path", path, "action", action)
	l.logger.Debug("task effect", fields...)
}

// failure writes one task error.
func (l operationLog) failure(command, projectID, taskID, agentID, phase, message string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, taskID, agentID)
	if phase != "" {
		fields = append(fields, "phase", phase)
	}
	fields = append(fields, "error", message)
	l.logger.Error("task error", fields...)
}
