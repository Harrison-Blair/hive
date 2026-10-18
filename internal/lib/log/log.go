// Package log writes structured JSON events to daily files in the hive home.
package log

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxBodyBytes = 4096

type outputState struct {
	mu    sync.Mutex
	home  string
	file  *os.File
	date  string
	clock func() time.Time
}

// Writer is a slog handler that rotates files when the local date changes.
type Writer struct {
	state     *outputState
	component string
	attrs     []slog.Attr
	groups    []string
}

// Open returns a logger that writes events to the current daily file.
func Open(home, component string) (*slog.Logger, error) {
	writer, err := NewWriter(home, component, time.Now)
	if err != nil {
		return nil, err
	}
	return slog.New(writer), nil
}

// NewWriter creates a log writer with an explicit clock dependency.
func NewWriter(home, component string, clock func() time.Time) (*Writer, error) {
	logsDir := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, err
	}
	return &Writer{
		state:     &outputState{home: home, clock: clock},
		component: component,
	}, nil
}

// Body returns a body attribute. The writer caps it at 4096 bytes.
func Body(text string) slog.Attr {
	return slog.String("body", text)
}

func (w *Writer) Enabled(context.Context, slog.Level) bool { return true }

func (w *Writer) Handle(_ context.Context, record slog.Record) error {
	w.state.mu.Lock()
	defer w.state.mu.Unlock()

	now := w.state.clock().Local()
	date := now.Format("2006-01-02")
	if w.state.file == nil || w.state.date != date {
		if w.state.file != nil {
			if err := w.state.file.Close(); err != nil {
				return err
			}
		}
		path := filepath.Join(w.state.home, ".hive", "logs", date+".jsonl")
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		w.state.file = file
		w.state.date = date
	}

	event := map[string]any{
		"time":      now.Format(time.RFC3339Nano),
		"level":     strings.ToUpper(record.Level.String()),
		"msg":       record.Message,
		"component": w.component,
	}
	truncated := false
	for _, attr := range w.attrs {
		addAttr(event, w.groups, attr, &truncated)
	}
	record.Attrs(func(attr slog.Attr) bool {
		addAttr(event, w.groups, attr, &truncated)
		return true
	})
	if truncated {
		event["truncated"] = true
	}
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	_, err = w.state.file.Write(line)
	return err
}

func (w *Writer) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *w
	clone.attrs = append(append([]slog.Attr(nil), w.attrs...), attrs...)
	return &clone
}

func (w *Writer) WithGroup(name string) slog.Handler {
	clone := *w
	clone.groups = append(append([]string(nil), w.groups...), name)
	return &clone
}

// Close closes the current daily file.
func (w *Writer) Close() error {
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	if w.state.file == nil {
		return nil
	}
	err := w.state.file.Close()
	w.state.file = nil
	w.state.date = ""
	return err
}

func addAttr(event map[string]any, groups []string, attr slog.Attr, truncated *bool) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	if attr.Value.Kind() == slog.KindGroup {
		group := append(append([]string(nil), groups...), attr.Key)
		for _, child := range attr.Value.Group() {
			addAttr(event, group, child, truncated)
		}
		return
	}
	value := attr.Value.Any()
	if attr.Key == "body" {
		if body, ok := value.(string); ok && len(body) > maxBodyBytes {
			value = body[:maxBodyBytes]
			*truncated = true
		}
	}
	key := attr.Key
	if len(groups) != 0 {
		key = fmt.Sprintf("%s.%s", strings.Join(groups, "."), attr.Key)
	}
	event[key] = value
}
