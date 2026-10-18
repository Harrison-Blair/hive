package memory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	scopelib "github.com/Harrison-Blair/hive/internal/lib/scope"
	storelib "github.com/Harrison-Blair/hive/internal/lib/store"
)

var memoryNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// WriteOptions holds the values for one memory write.
type WriteOptions struct {
	Scope      string
	Name       string
	Input      io.Reader
	Deadline   time.Duration
	Logger     *slog.Logger
	IndexWrite func(path string, data []byte) error
}

type phaseError struct {
	phase string
	err   error
}

func (e *phaseError) Error() string { return e.err.Error() }

func (e *phaseError) Unwrap() error { return e.err }

// Write writes one memory and then the memory index.
func Write(home string, options WriteOptions) output.Envelope {
	resolved, err := scopelib.Resolve(home, options.Scope)
	if err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "scope", err)
	}
	options.Logger = logWithProject(options.Logger, resolved.ProjectID)
	if !memoryNamePattern.MatchString(options.Name) {
		return failEnvelope(options.Logger, "memory write", options.Scope, "validate", fmt.Errorf("invalid memory name: %s", displayValue(options.Name)))
	}
	content, err := scopelib.ReadInput(options.Input, 1<<20, options.Deadline)
	if err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "input", err)
	}
	values, err := parseMemory(content, options.Name)
	if err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "frontmatter", err)
	}
	memoryDir := filepath.Join(resolved.Dir, "memory")
	memoryPath := filepath.Join(memoryDir, options.Name+".md")
	indexPath := filepath.Join(memoryDir, indexName)
	if err := scopelib.CheckDir(home, memoryDir); err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "path", err)
	}
	if err := scopelib.CheckFile(home, memoryPath); err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "path", err)
	}
	if err := scopelib.CheckFile(home, indexPath); err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "path", err)
	}
	return writeMemory(home, resolved, options, memoryDir, memoryPath, indexPath, content, values["description"])
}

func writeMemory(home string, resolved scopelib.Scope, options WriteOptions, memoryDir, memoryPath, indexPath string, content []byte, description string) output.Envelope {
	store, err := storelib.Open(home)
	if err != nil {
		return failEnvelope(options.Logger, "memory write", options.Scope, "store", err)
	}
	var result output.Envelope
	err = store.WithLock(context.Background(), func(*storelib.LockedStore) error {
		if err := scopelib.CheckDir(home, memoryDir); err != nil {
			return &phaseError{phase: "path", err: err}
		}
		if err := scopelib.CheckFile(home, memoryPath); err != nil {
			return &phaseError{phase: "path", err: err}
		}
		if err := scopelib.CheckFile(home, indexPath); err != nil {
			return &phaseError{phase: "path", err: err}
		}
		if err := os.MkdirAll(memoryDir, 0o755); err != nil {
			return &phaseError{phase: "memory", err: err}
		}
		entries, err := scanMemoryFiles(memoryDir)
		if err != nil {
			return &phaseError{phase: "scan", err: err}
		}
		action := "created"
		info, err := os.Lstat(memoryPath)
		switch {
		case err == nil:
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return &phaseError{phase: "path", err: fmt.Errorf("path invalid: %s", memoryPath)}
			}
			action = "updated"
		case !errors.Is(err, os.ErrNotExist):
			return &phaseError{phase: "path", err: err}
		}
		if err := writeAtomic(memoryPath, content); err != nil {
			return &phaseError{phase: "memory", err: err}
		}
		memoryEffect := output.Effect{
			Resource: "memory",
			Path:     relativePath(home, memoryPath),
			Action:   action,
		}
		logEffect(options.Logger, "memory write", options.Scope, memoryEffect)
		entries = upsertEntry(entries, indexEntry{Name: options.Name, Description: description})
		sort.Slice(entries, func(left, right int) bool {
			return entries[left].Name < entries[right].Name
		})
		writeIndex := writeAtomic
		if options.IndexWrite != nil {
			writeIndex = options.IndexWrite
		}
		if err := writeIndex(indexPath, buildIndex(entries)); err != nil {
			logFailure(options.Logger, "memory write", options.Scope, "index", err)
			result = output.Envelope{
				Outcome: "partial",
				Phase:   "index",
				Effects: []output.Effect{memoryEffect},
				Error:   displayValue(err.Error()),
				Result:  nil,
			}
			return nil
		}
		indexEffect := output.Effect{
			Resource: "index",
			Path:     relativePath(home, indexPath),
			Action:   "updated",
		}
		logEffect(options.Logger, "memory write", options.Scope, indexEffect)
		result = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: []output.Effect{memoryEffect, indexEffect},
			Error:   nil,
			Result:  nil,
		}
		return nil
	})
	if err != nil {
		phase := "lock"
		var failure *phaseError
		if errors.As(err, &failure) {
			phase = failure.phase
			err = failure.err
		}
		return failEnvelope(options.Logger, "memory write", options.Scope, phase, err)
	}
	return result
}

// WriteText renders the effects of one memory write.
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

func failed(phase, message string) output.Envelope {
	return output.Envelope{
		Outcome: "failed",
		Phase:   phase,
		Effects: []output.Effect{},
		Error:   displayValue(message),
		Result:  nil,
	}
}

func failEnvelope(logger *slog.Logger, command, scope, phase string, err error) output.Envelope {
	logFailure(logger, command, scope, phase, err)
	return failed(phase, userErrorText(err))
}

func logEffect(logger *slog.Logger, command, scope string, effect output.Effect) {
	if logger == nil {
		return
	}
	logger.Debug("effect",
		"command", command,
		"scope", scope,
		"resource", effect.Resource,
		"path", effect.Path,
		"action", effect.Action,
	)
}

type userMessaged interface {
	UserMessage() string
}

type logCategorized interface {
	LogCategory() string
}

func userErrorText(err error) string {
	var messaged userMessaged
	if errors.As(err, &messaged) {
		return messaged.UserMessage()
	}
	return err.Error()
}

func logErrorText(err error) string {
	var categorized logCategorized
	if errors.As(err, &categorized) {
		return categorized.LogCategory()
	}
	return displayValue(err.Error())
}

func logWithProject(logger *slog.Logger, projectID string) *slog.Logger {
	if logger == nil || projectID == "" {
		return logger
	}
	return logger.With("project_id", projectID)
}

func logFailure(logger *slog.Logger, command, scope, phase string, err error) {
	if logger == nil {
		return
	}
	logger.Debug("error",
		"command", command,
		"scope", scope,
		"phase", phase,
		"error", logErrorText(err),
	)
}
