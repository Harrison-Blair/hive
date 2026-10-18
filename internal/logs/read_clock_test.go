package logs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewReaderWithClockUsesFixedClock(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, 1, 2, 12, 0, 0, 0, time.Local)
	path := filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl")
	line := `{"time":"2026-01-02T11:00:00Z","level":"INFO","msg":"fixed clock"}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	reader := NewReaderWithClock(home, func() time.Time { return fixed })
	events, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || eventMessage(events[0]) != "fixed clock" {
		t.Fatalf("events = %#v; want one fixed clock event", events)
	}
	if offset.Path != path || offset.Byte != int64(len(line)) {
		t.Fatalf("offset = %#v; want end of %s", offset, path)
	}
}
