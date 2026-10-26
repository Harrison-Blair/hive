// Package write writes the Markdown handoff of one role.
package write

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	scopelib "github.com/Harrison-Blair/hive/internal/lib/scope"
	memorylib "github.com/Harrison-Blair/hive/internal/memory"
)

// WriteOptions holds the values for one handoff write.
type WriteOptions struct {
	Scope    string
	Input    io.Reader
	Deadline time.Duration
	Logger   *slog.Logger
}

// Write replaces one handoff file.
func Write(home string, options WriteOptions) output.Envelope {
	resolved, err := scopelib.Resolve(home, options.Scope)
	if err != nil {
		return failEnvelope(options.Logger, options.Scope, "scope", err)
	}
	options.Logger = logWithProject(options.Logger, resolved.ProjectID)
	content, err := scopelib.ReadInput(options.Input, 1<<20, options.Deadline)
	if err != nil {
		return failEnvelope(options.Logger, options.Scope, "input", err)
	}
	if len(content) == 0 {
		return failEnvelope(options.Logger, options.Scope, "input", errors.New("empty input"))
	}
	path := filepath.Join(resolved.Dir, "handoff.md")
	if err := scopelib.CheckFile(home, path); err != nil {
		return failEnvelope(options.Logger, options.Scope, "path", err)
	}
	if err := memorylib.WriteAtomic(path, content); err != nil {
		return failEnvelope(options.Logger, options.Scope, "handoff", err)
	}
	effect := output.Effect{
		Resource: "handoff",
		Path:     relativePath(home, path),
		Action:   "updated",
	}
	logEffect(options.Logger, options.Scope, effect)
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{effect},
		Error:   nil,
		Result:  nil,
	}
}

// WriteText renders the effect of one handoff write.
func WriteText(envelope output.Envelope) string {
	lines := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func relativePath(home, path string) string {
	relative, err := filepath.Rel(home, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func failEnvelope(logger *slog.Logger, scope, phase string, err error) output.Envelope {
	message := oneLine(err.Error())
	logFailure(logger, scope, phase, message)
	return output.Envelope{
		Outcome: "failed",
		Phase:   phase,
		Effects: []output.Effect{},
		Error:   message,
		Result:  nil,
	}
}

func logEffect(logger *slog.Logger, scope string, effect output.Effect) {
	if logger == nil {
		return
	}
	logger.Debug("effect",
		"command", "handoff write",
		"scope", scope,
		"resource", effect.Resource,
		"path", effect.Path,
		"action", effect.Action,
	)
}

func logWithProject(logger *slog.Logger, projectID string) *slog.Logger {
	if logger == nil || projectID == "" {
		return logger
	}
	return logger.With("project_id", projectID)
}

func logFailure(logger *slog.Logger, scope, phase, message string) {
	if logger == nil {
		return
	}
	logger.Debug("error",
		"command", "handoff write",
		"scope", scope,
		"phase", phase,
		"error", message,
	)
}

func oneLine(value string) string {
	if !utf8.ValidString(value) {
		return fmt.Sprintf("%q", value)
	}
	for _, char := range value {
		if unicode.IsControl(char) {
			return fmt.Sprintf("%q", value)
		}
	}
	return value
}
