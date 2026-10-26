package goal

import (
	"log/slog"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// Log writes goal events with the goal component.
type Log struct {
	logger *slog.Logger
	closer func()
}

// OpenLog opens the goal log writer.
func OpenLog(home string) Log {
	writer, err := hivelog.NewWriter(home, "goal", time.Now)
	if err != nil {
		return Log{closer: func() {}}
	}
	return Log{logger: slog.New(writer), closer: func() { _ = writer.Close() }}
}

// Close releases the goal log writer.
func (l Log) Close() { l.closer() }

func (l Log) fields(command, projectID, goalID, agentID string) []any {
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

// Failure logs one failed goal command.
func (l Log) Failure(command, projectID, goalID, agentID, phase, message string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, goalID, agentID)
	fields = append(fields, "phase", phase, "error", message)
	l.logger.Error("goal command failed", fields...)
}

// Effect logs one goal command effect.
func (l Log) Effect(command, projectID, goalID, agentID, resource, path, action string) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, goalID, agentID)
	fields = append(fields, "resource", resource, "path", path, "action", action)
	l.logger.Debug("goal effect", fields...)
}

// Created logs one new goal.
func (l Log) Created(projectID, goalID, agentID string) {
	if l.logger == nil {
		return
	}
	l.logger.Debug("goal created", l.fields("goal create", projectID, goalID, agentID)...)
}

// Changed logs one goal or wave change.
func (l Log) Changed(projectID, goalID, agentID, action string, extra ...any) {
	if l.logger == nil {
		return
	}
	fields := l.fields("goal edit", projectID, goalID, agentID)
	fields = append(fields, "action", action)
	fields = append(fields, extra...)
	l.logger.Debug("goal changed", fields...)
}
