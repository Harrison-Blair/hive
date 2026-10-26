package worktree

import (
	"log/slog"
	"time"
	"unicode/utf8"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

const maxLogText = 4096

// OpenLogger returns the injected logger, or a logger on the daily log file
// of the home, and the function that closes it.
func OpenLogger(home string, injected *slog.Logger) (*slog.Logger, func()) {
	if injected != nil {
		return injected, func() {}
	}
	writer, err := hivelog.NewWriter(home, "worktree", time.Now)
	if err != nil {
		return nil, func() {}
	}
	return slog.New(writer), func() { _ = writer.Close() }
}

// LogEvent writes one debug event with the command, project, and task scope.
func LogEvent(logger *slog.Logger, message, command, projectID, taskID string, attrs ...any) {
	if logger == nil {
		return
	}
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	fields = append(fields, attrs...)
	logger.Debug(message, fields...)
}

// LogEffect writes one effect event.
func LogEffect(logger *slog.Logger, command, projectID, taskID, resource, path, action string) {
	LogEvent(logger, "worktree effect", command, projectID, taskID, "resource", resource, "path", path, "action", action)
}

// LogFailure writes one error event with a bounded error text.
func LogFailure(logger *slog.Logger, command, projectID, taskID, phase, message string) {
	if logger == nil {
		return
	}
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	if phase != "" {
		fields = append(fields, "phase", phase)
	}
	text, truncated := boundLogText(OneLine(message))
	fields = append(fields, "error", text)
	if truncated {
		fields = append(fields, "truncated", true)
	}
	logger.Error("worktree error", fields...)
}

func boundLogText(text string) (string, bool) {
	if len(text) <= maxLogText {
		return text, false
	}
	cut := maxLogText
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut], true
}
