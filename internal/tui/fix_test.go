package tui

import (
	"bytes"
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
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// startInitCommands runs each command that Init returns, as the program would.
func startInitCommands(t *testing.T, m *Model) {
	t.Helper()
	cmd := m.Init()
	if cmd == nil {
		return
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init returned one command; the fixture needs the tick and follow commands")
	}
	for _, sub := range batch {
		sub := sub
		go func() { _ = sub() }()
	}
}

func tickUntilCount(t *testing.T, m *Model, want int) *Model {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for len(m.events) < want {
		if time.Now().After(deadline) {
			t.Fatalf("events=%d; want %d before the timeout", len(m.events), want)
		}
		m, _ = raise(t, m, tickMsg{})
		time.Sleep(20 * time.Millisecond)
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

// TestInitDoesNotConsumeFollowerEvents starts the actual Init commands. A
// competing follower command would take append A, so the tick would show only
// append B. The one ordered path must show A before B.
func TestInitDoesNotConsumeFollowerEvents(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	path := dailyPath(home, day)
	writeEvents(t, path, eventJSON(t, day, "INFO", "initial"))
	m := newTestModel(t, home, clock)

	startInitCommands(t, m)
	appendEvent(t, path, eventJSON(t, day, "INFO", "append A"))
	appendEvent(t, path, eventJSON(t, day, "INFO", "append B"))
	time.Sleep(700 * time.Millisecond)
	m = tickUntilCount(t, m, 3)

	got := eventMessages(m)
	want := []string{"initial", "append A", "append B"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("follower order=%q; want %q", got, want)
	}
	view := m.View().Content
	if strings.Index(view, "append A") > strings.Index(view, "append B") {
		t.Fatalf("rendered order reversed A then B:\n%s", view)
	}
}

var fixtureOutputError = errors.New("fix fixture output failure")

type fixtureWriter struct {
	short bool
	calls int
}

func (w *fixtureWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.short {
		return 0, nil
	}
	return 0, fixtureOutputError
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

func TestRunHomeFiniteEOFStops(t *testing.T) {
	home := newTestHome(t)
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader(""), io.Discard) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome on a finite empty input: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunHome did not stop on finite input EOF")
	}
}

func TestHeaderRejectsOutOfRangePIDs(t *testing.T) {
	self := int64(os.Getpid())
	exited := exitedPID(t)
	cases := []struct {
		name string
		pid  int64
		want string
	}{
		{"self", self, "supervisor: live"},
		{"zero", 0, "supervisor: down"},
		{"negative", -1, "supervisor: down"},
		{"exited", exited, "supervisor: down"},
		{"plus 2^32", self + (1 << 32), "supervisor: down"},
		{"max int32 plus one", int64(math.MaxInt32) + 1, "supervisor: down"},
		{"max int64", math.MaxInt64, "supervisor: down"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			home := newTestHome(t)
			writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
				Key: "supervisor", Agent: "supervisor", Generation: 1, PID: item.pid,
			})
			clock := fixedClock(t, time.Date(2026, 5, 2, 12, 0, 0, 0, time.Local))
			m := newTestModel(t, home, clock)
			m, _ = raise(t, m, tickMsg{})
			requireView(t, m, item.want)
		})
	}
}

func TestHeaderPersistentErrorDoesNotLogEachTick(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, home string)
	}{
		{
			name: "agents",
			setup: func(t *testing.T, home string) {
				path := filepath.Join(home, ".hive", "state", "agents")
				if err := os.WriteFile(path, []byte("not a directory\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "inbox",
			setup: func(t *testing.T, home string) {
				path := filepath.Join(home, ".hive", "state", "inbox")
				if err := os.WriteFile(path, []byte("not a directory\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			home := newTestHome(t)
			if err := os.MkdirAll(filepath.Join(home, ".hive", "state"), 0o755); err != nil {
				t.Fatal(err)
			}
			item.setup(t, home)
			clock := fixedClock(t, time.Date(2026, 5, 3, 12, 0, 0, 0, time.Local))
			m := newTestModel(t, home, clock)
			requireView(t, m, "error:")
			before := countLogLines(t, home)
			if before == 0 {
				t.Fatal("the first header error wrote no log event")
			}
			for index := 0; index < 4; index++ {
				m, _ = raise(t, m, tickMsg{})
			}
			after := countLogLines(t, home)
			if after != before {
				t.Fatalf("log events grew from %d to %d over unchanged error ticks", before, after)
			}
			if _, cmd := raise(t, m, key('q', "q")); cmd == nil {
				t.Fatal("quit does not work in the persistent error state")
			} else {
				quitMessage(t, cmd)
			}
		})
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
