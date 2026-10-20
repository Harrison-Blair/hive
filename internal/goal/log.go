package goal

import (
	"log/slog"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

type operationLog struct {
	logger *slog.Logger
	closer func()
}

func openOperationLog(home string) operationLog {
	writer, err := hivelog.NewWriter(home, "goal", time.Now)
	if err != nil {
		return operationLog{closer: func() {}}
	}
	return operationLog{logger: slog.New(writer), closer: func() { _ = writer.Close() }}
}

func (l operationLog) close() { l.closer() }

func (l operationLog) fields(command, projectID, goalID, agentID string) []any {
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if goalID != "" {
		fields = append(fields, "goal_id", goalID)
	}
	if agentID != "" {
		fields = append(fields, "agent_id", agentID)
	}
	return fields
}

func (l operationLog) failure(command, projectID, goalID, agentID, phase, message string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, goalID, agentID)
	fields = append(fields, "phase", phase, "error", message)
	l.logger.Error("goal command failed", fields...)
}

func (l operationLog) effect(command, projectID, goalID, agentID, resource, path, action string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, goalID, agentID)
	fields = append(fields, "resource", resource, "path", path, "action", action)
	l.logger.Debug("goal effect", fields...)
}

func (l operationLog) created(projectID, goalID, agentID string) {
	if l.logger == nil {
		return
	}
	l.logger.Debug("goal created", l.fields("goal create", projectID, goalID, agentID)...)
}

func (l operationLog) changed(projectID, goalID, agentID, action string, extra ...any) {
	if l.logger == nil {
		return
	}
	fields := l.fields("goal edit", projectID, goalID, agentID)
	fields = append(fields, "action", action)
	fields = append(fields, extra...)
	l.logger.Debug("goal changed", fields...)
}
