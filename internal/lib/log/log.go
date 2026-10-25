// Package log writes structured JSON events to daily files in the hive home.
package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const maxBodyBytes = 4096

var activeRequestID atomic.Pointer[string]

// NewRequestID makes a random request ID with eight lowercase hex characters.
func NewRequestID() (string, error) {
	var random [4]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return "r-" + hex.EncodeToString(random[:]), nil
}

// SetRequestID sets the request ID for log events until the returned restore
// function runs.
func SetRequestID(id string) func() {
	previous := activeRequestID.Load()
	current := id
	activeRequestID.Store(&current)
	return func() { activeRequestID.Store(previous) }
}

var errClosed = errors.New("log: writer is closed")

type outputState struct {
	mu     sync.Mutex
	home   string
	file   *os.File
	date   string
	closed bool
	clock  func() time.Time
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
	gate := homeGate(home)
	gate.RLock()
	defer gate.RUnlock()
	if homeClosed(home) {
		return nil, errClosed
	}
	date := clock().Local().Format("2006-01-02")
	if err := checkStaticLogPath(home, date); err != nil {
		return nil, err
	}
	logsDir := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, err
	}
	return &Writer{
		state:     &outputState{home: home, clock: clock},
		component: component,
	}, nil
}

// homeGates guards each home. CloseHome takes the write lock, so it waits for
// each write in progress. A writer created after CloseHome writes no event and
// creates no log file.
var homeGates sync.Map

// closedHomes records each home that CloseHome settled.
var closedHomes sync.Map

func homeGate(home string) *sync.RWMutex {
	gate, _ := homeGates.LoadOrStore(home, &sync.RWMutex{})
	return gate.(*sync.RWMutex)
}

func homeClosed(home string) bool {
	_, closed := closedHomes.Load(home)
	return closed
}

// CloseHome stops each log writer of home and waits for each write in
// progress. A writer created after CloseHome fails, so a late command cannot
// recreate the log file.
func CloseHome(home string) {
	gate := homeGate(home)
	gate.Lock()
	closedHomes.Store(home, struct{}{})
	gate.Unlock()
}

// Body returns a body attribute. The writer caps it at 4096 bytes.
func Body(text string) slog.Attr {
	return slog.String("body", text)
}

func (w *Writer) Enabled(context.Context, slog.Level) bool { return true }

func (w *Writer) Handle(_ context.Context, record slog.Record) error {
	gate := homeGate(w.state.home)
	gate.RLock()
	defer gate.RUnlock()
	if homeClosed(w.state.home) {
		return errClosed
	}
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	if w.state.closed {
		return errClosed
	}

	now := w.state.clock().Local()
	date := now.Format("2006-01-02")
	if w.state.file == nil || w.state.date != date {
		if err := checkStaticLogPath(w.state.home, date); err != nil {
			return err
		}
		path := filepath.Join(w.state.home, ".hive", "logs", date+".jsonl")
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		previous := w.state.file
		w.state.file = file
		w.state.date = date
		if previous != nil {
			if err := previous.Close(); err != nil {
				return err
			}
		}
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
	if request := activeRequestID.Load(); request != nil && *request != "" {
		event["request_id"] = *request
	}
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
	w.state.closed = true
	if w.state.file == nil {
		return nil
	}
	err := w.state.file.Close()
	w.state.file = nil
	w.state.date = ""
	return err
}

func checkStaticLogPath(home, date string) error {
	hiveDir := filepath.Join(home, ".hive")
	if err := checkNativeDirectory(hiveDir); err != nil {
		return err
	}
	logsDir := filepath.Join(hiveDir, "logs")
	if err := checkNativeDirectory(logsDir); err != nil {
		return err
	}
	return checkRegularOrMissing(filepath.Join(logsDir, date+".jsonl"))
}

func checkNativeDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("log directory is not a native directory: %s", path)
	}
	return nil
}

func checkRegularOrMissing(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("log path is not a regular file: %s", path)
	}
	return nil
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
		if body, ok := value.(string); ok {
			if capped, cut := capBody(body); cut {
				value = capped
				*truncated = true
			}
		}
	}
	key := attr.Key
	if len(groups) != 0 {
		key = fmt.Sprintf("%s.%s", strings.Join(groups, "."), attr.Key)
	}
	event[key] = value
}

// capBody returns the largest complete-character prefix of text that stays
// within maxBodyBytes after the encoding/json replacement of each invalid byte
// with U+FFFD. It returns the text unchanged when the decoded sequence fits.
func capBody(text string) (string, bool) {
	decoded := 0
	end := 0
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		width := size
		if r == utf8.RuneError && size == 1 {
			width = len("\uFFFD")
		}
		if decoded+width > maxBodyBytes {
			break
		}
		decoded += width
		end += size
	}
	if end == len(text) {
		return text, false
	}
	var builder strings.Builder
	builder.Grow(decoded)
	for index := 0; index < end; {
		r, size := utf8.DecodeRuneInString(text[index:end])
		if r == utf8.RuneError && size == 1 {
			builder.WriteString("\uFFFD")
		} else {
			builder.WriteString(text[index : index+size])
		}
		index += size
	}
	return builder.String(), true
}
