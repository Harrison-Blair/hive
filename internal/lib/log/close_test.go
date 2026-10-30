package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClosedWriterRejectsALateEvent checks that Close ends the writer. A late
// event must not open the daily file again and must not create a new one.
func TestClosedWriterRejectsALateEvent(t *testing.T) {
	home := t.TempDir()
	clock := func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local) }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	logger.Debug("before close")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "logs", "2026-01-02.jsonl")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	record := slog.NewRecord(clock(), slog.LevelDebug, "after close", 0)
	if err := writer.Handle(context.Background(), record); err == nil {
		t.Fatal("the closed writer accepted a late event")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) || strings.Contains(string(after), "after close") {
		t.Fatalf("a late event changed the log file:\n%s", after)
	}
	if entries, err := os.ReadDir(filepath.Join(home, "logs")); err != nil || len(entries) != 1 {
		t.Fatalf("log files = %v, %v; want the one before close", entries, err)
	}
}
