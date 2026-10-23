package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func tickUntilCount(t *testing.T, m *Model, want int) *Model {
	t.Helper()
	for len(m.events) < want {
		cmd := m.startCurrentScreenLoad()
		if cmd == nil {
			t.Fatalf("screen data load did not start with %d events; want %d", len(m.events), want)
		}
		m, _ = raise(t, m, cmd())
	}
	return m
}

func eventMessages(m *Model) []string {
	var result []string
	for _, event := range m.events {
		if message, ok := event.Fields["msg"].(string); ok && !strings.HasPrefix(message, "tui ") {
			result = append(result, message)
		}
	}
	return result
}

func countLogLines(t *testing.T, home string) int {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := 0
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			t.Fatal(err)
		}
		lines += strings.Count(string(data), "\n")
	}
	return lines
}

func exitedPID(t *testing.T) int64 {
	t.Helper()
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Skipf("cannot start the true command: %v", err)
	}
	return int64(cmd.Process.Pid)
}

// TestLogRefreshKeepsFileOrder checks that a background reload adds each new
// line once and keeps the order from the file.
func TestLogRefreshKeepsFileOrder(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "initial"))
	m := newTestLogsModel(t, home, clock)

	appendEvent(t, path, eventJSON(t, day, "INFO", "append A"))
	appendEvent(t, path, eventJSON(t, day, "INFO", "append B"))
	m = loadLogs(t, m)

	got := eventMessages(m)
	want := []string{"initial", "append A", "append B"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("log order=%q; want %q", got, want)
	}
	view := m.View().Content
	if strings.Index(view, "append A") > strings.Index(view, "append B") {
		t.Fatalf("rendered order reversed A then B:\n%s", view)
	}
}

var errFixtureOutput = errors.New("fix fixture output failure")

type fixtureWriter struct {
	short bool
	calls int
}

func (w *fixtureWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.short {
		return 0, nil
	}
	return 0, errFixtureOutput
}

func TestRunHomeReportsOutputFailure(t *testing.T) {
	for _, short := range []bool{false, true} {
		t.Run(fmt.Sprintf("short=%v", short), func(t *testing.T) {
			home := newTestHome(t)
			writer := &fixtureWriter{short: short}
			done := make(chan error, 1)
			go func() { done <- RunHome(home, strings.NewReader("q"), writer) }()
			select {
			case err := <-done:
				if err == nil {
					t.Fatalf("RunHome returned nil for short=%v; want an output error", short)
				}
				if writer.calls == 0 {
					t.Fatal("the TUI made no output write")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("RunHome did not settle after an output failure")
			}
		})
	}
}

type fixtureInputError struct{}

func (fixtureInputError) Read([]byte) (int, error) {
	return 0, errors.New("fix fixture input failure")
}

func TestRunHomeReportsInputFailure(t *testing.T) {
	home := newTestHome(t)
	done := make(chan error, 1)
	go func() { done <- RunHome(home, fixtureInputError{}, io.Discard) }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "fix fixture input failure") {
			t.Fatalf("RunHome input error=%v; want the fixture input failure", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunHome did not settle after an input failure")
	}
}

func TestRunHomeReportsFileWriterFailure(t *testing.T) {
	file, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
	if err != nil {
		t.Skipf("/dev/full is not available: %v", err)
	}
	defer func() { _ = file.Close() }()
	home := newTestHome(t)
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader("q"), file) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("RunHome returned nil for a full file writer; want an output error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunHome did not settle after a file writer failure")
	}
}

func TestRunHomeFiniteEOFStops(t *testing.T) {
	home := newTestHome(t)
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader(""), io.Discard) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome on a finite empty input: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("RunHome did not stop on finite input EOF")
	}
}

func TestHeaderRejectsOutOfRangePIDs(t *testing.T) {
	self := int64(os.Getpid())
	selfStart, err := process.Start(self)
	if err != nil {
		t.Fatal(err)
	}
	exited := exitedPID(t)
	cases := []struct {
		name string
		pid  int64
		want string
	}{
		{"self", self, "● supervisor"},
		{"zero", 0, "○ supervisor"},
		{"negative", -1, "○ supervisor"},
		{"exited", exited, "○ supervisor"},
		{"plus 2^32", self + (1 << 32), "○ supervisor"},
		{"max int32 plus one", int64(math.MaxInt32) + 1, "○ supervisor"},
		{"max int64", math.MaxInt64, "○ supervisor"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			home := newTestHome(t)
			pidStart := uint64(0)
			if item.pid == self {
				pidStart = selfStart
			}
			writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
				Key: "supervisor", Agent: "supervisor", Generation: 1, PID: item.pid, PIDStart: pidStart,
			})
			clock := fixedClock(t, time.Date(2026, 5, 2, 12, 0, 0, 0, time.Local))
			m := newTestModel(t, home, clock)
			m = loadHeader(t, m)
			requireView(t, m, item.want)
		})
	}
}

func TestHeaderErrorDoesNotStopShell(t *testing.T) {
	home := newTestHome(t)
	if err := os.MkdirAll(filepath.Join(home, ".hive", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".hive", "state", "agents")
	if err := os.WriteFile(path, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	clock := fixedClock(t, time.Date(2026, 5, 3, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, home, clock)
	m = loadHeader(t, m)
	requireView(t, m, "error:")
	m, _ = raise(t, m, ctrl('g'))
	requireView(t, m, "╭─ project agents ", "error:")
	m = loadHeader(t, m)
	requireView(t, m, "error:")
	_, cmd := raise(t, m, key('q', "q"))
	if cmd != nil {
		t.Fatal("q returned a command while a header error was visible")
	}
}

func TestRunHomeFiniteFDPipeStops(t *testing.T) {
	home := newTestHome(t)
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = readEnd.Close() }()
	done := make(chan error, 1)
	go func() { done <- RunHome(home, readEnd, io.Discard) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome on a finite pipe: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("RunHome did not stop on a finite pipe EOF")
	}
}

func TestRunHomeFiniteFDFileStops(t *testing.T) {
	home := newTestHome(t)
	file, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	done := make(chan error, 1)
	go func() { done <- RunHome(home, file, io.Discard) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome on a finite regular file: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("RunHome did not stop on a finite regular file EOF")
	}
}

func TestActionLogRefusalShowsError(t *testing.T) {
	cases := []struct {
		name        string
		obstruction string
		rotation    bool
		want        string
	}{
		{"first-open/daily-link", "daily-link", false, "not a regular file"},
		{"first-open/daily-fifo", "daily-fifo", false, "not a regular file"},
		{"first-open/linked-log-dir", "linked-log-dir", false, "not a native directory"},
		{"rotation/daily-link", "daily-link", true, "not a regular file"},
		{"rotation/daily-fifo", "daily-fifo", true, "not a regular file"},
		{"rotation/linked-log-dir", "linked-log-dir", true, "not a native directory"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			home := newTestHome(t)
			day := time.Date(2043, 7, 9, 12, 0, 0, 0, time.Local)
			clock := fixedClock(t, day)
			if item.rotation {
				writeEvents(t, dailyPath(home, day), eventJSON(t, day, "INFO", "retained old line"))
			}
			m := newTestLogsModel(t, home, clock)
			if item.rotation {
				var cmd tea.Cmd
				m, cmd = raise(t, m, key('f', "f"))
				if cmd != nil {
					_ = cmd()
				}
				clock.set(day.AddDate(0, 0, 1))
			}
			external := t.TempDir()
			sentinel := filepath.Join(external, "sentinel")
			original := []byte("owned sentinel must stay exact\n")
			if err := os.WriteFile(sentinel, original, 0o600); err != nil {
				t.Fatal(err)
			}
			path := dailyPath(home, clock.now())
			switch item.obstruction {
			case "daily-link":
				if err := os.Symlink(sentinel, path); err != nil {
					t.Fatal(err)
				}
			case "daily-fifo":
				if err := exec.Command("mkfifo", path).Run(); err != nil {
					t.Skipf("mkfifo is not available: %v", err)
				}
			case "linked-log-dir":
				dir := filepath.Join(home, ".hive", "logs")
				if err := os.Rename(dir, dir+"-owned-prior"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(external, dir); err != nil {
					t.Fatal(err)
				}
			}
			start := time.Now()
			var actionCmd tea.Cmd
			m, actionCmd = raise(t, m, key('f', "f"))
			if actionCmd != nil {
				m, _ = raise(t, m, actionCmd())
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Fatalf("the action blocked for %v on a static obstruction", elapsed)
			}
			requireView(t, m, "error:", item.want)
			rejectView(t, m, "no events today")
			data, err := os.ReadFile(sentinel)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, original) {
				t.Fatalf("the external target changed from %d to %d bytes", len(original), len(data))
			}
			entries, err := os.ReadDir(external)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("the external directory changed to %d entries", len(entries))
			}
		})
	}
}

func TestHealthyActionLogIsWrittenOnce(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2043, 7, 10, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	m := newTestLogsModel(t, home, clock)
	before := countLogLines(t, home)
	_, cmd := raise(t, m, key('f', "f"))
	if cmd != nil {
		_ = cmd()
	}
	after := countLogLines(t, home)
	if after != before+1 {
		t.Fatalf("healthy action log lines went from %d to %d; want one event", before, after)
	}
	data, err := os.ReadFile(dailyPath(home, day))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatal(err)
	}
	if event["msg"] != "tui action" || event["action"] != "follow" || event["component"] != "tui" {
		t.Fatalf("healthy action event fields = %v", event)
	}
}

func TestNewRejectsLinkedDailyLogWithoutWrite(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 4, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	if err := os.MkdirAll(filepath.Join(home, ".hive", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target.jsonl")
	original := []byte("0123456789abcdef\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, dailyPath(home, day)); err != nil {
		t.Fatal(err)
	}
	m := newTestModel(t, home, clock)
	requireView(t, m, "error:", "not a regular file")
	rejectView(t, m, "no events today")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, original) {
		t.Fatalf("the rejected daily link changed its target from %d to %d bytes", len(original), len(data))
	}
}

func TestNewRejectsDailyFIFOWithoutBlocking(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 5, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	if err := os.MkdirAll(filepath.Join(home, ".hive", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := dailyPath(home, day)
	if err := exec.Command("mkfifo", path).Run(); err != nil {
		t.Skipf("mkfifo is not available: %v", err)
	}
	start := time.Now()
	m := newTestModel(t, home, clock)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("New blocked for %v on a rejected FIFO", elapsed)
	}
	requireView(t, m, "error:", "not a regular file")
}

func TestNewRejectsLinkedLogDirectoryWithoutWrite(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(home, ".hive", "logs")); err != nil {
		t.Fatal(err)
	}
	m := newTestModel(t, home, clock)
	requireView(t, m, "error:", "not a native directory")
	rejectView(t, m, "no events today")
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("the rejected log directory gained %d entries", len(entries))
	}
}
