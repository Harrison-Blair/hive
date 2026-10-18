package log

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBodyCapsAt4096BytesAndMarksTruncated(t *testing.T) {
	home := t.TempDir()
	clock := func() time.Time { return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.Local) }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	logger.Debug("body event", Body(strings.Repeat("x", 5000)))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	event := readOneEvent(t, filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl"))
	body, ok := event["body"].(string)
	if !ok || len(body) != 4096 || event["truncated"] != true {
		t.Fatalf("body length = %d, truncated = %v; want 4096 and true", len(body), event["truncated"])
	}
	eventTime, ok := event["time"].(string)
	if !ok || !strings.HasPrefix(eventTime, "2026-01-02T03:04:05") {
		t.Errorf("event time = %v; want injected local time", event["time"])
	}
	if event["level"] != "DEBUG" || event["msg"] != "body event" || event["component"] != "cli" {
		t.Errorf("event fields = %#v", event)
	}
	info, err := os.Stat(filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl"))
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("log file mode = %v, %v; want 0644", info, err)
	}
}

func TestBodyAtLimitDoesNotMarkTruncated(t *testing.T) {
	home := t.TempDir()
	writer, err := NewWriter(home, "cli", func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local) })
	if err != nil {
		t.Fatal(err)
	}
	slog.New(writer).Debug("exact", Body(strings.Repeat("x", 4096)))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	event := readOneEvent(t, filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl"))
	if len(event["body"].(string)) != 4096 {
		t.Fatalf("body length = %d; want 4096", len(event["body"].(string)))
	}
	if _, ok := event["truncated"]; ok {
		t.Errorf("event has truncated field for exact-limit body: %#v", event)
	}
}

func TestConcurrentLogEventsAreWholeJSONLines(t *testing.T) {
	home := t.TempDir()
	writer, err := NewWriter(home, "cli", func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local) })
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	const writers, eventsPerWriter = 8, 100
	var wg sync.WaitGroup
	for worker := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := range eventsPerWriter {
				logger.Debug("parallel", "worker", worker, "event", event)
			}
		}()
	}
	wg.Wait()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		var event map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("event line %d is invalid JSON: %v", count+1, err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != writers*eventsPerWriter {
		t.Fatalf("event lines = %d; want %d", count, writers*eventsPerWriter)
	}
}

func TestOpenAppendsToDailyFile(t *testing.T) {
	home := t.TempDir()
	logs := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	date := time.Now().Local().Format("2006-01-02")
	path := filepath.Join(logs, date+".jsonl")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	clock := time.Now()
	writer, err := NewWriter(home, "cli", func() time.Time { return clock })
	if err != nil {
		t.Fatal(err)
	}
	slog.New(writer).Debug("appended")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "existing\n") || strings.Count(string(data), "\n") != 2 {
		t.Fatalf("daily file = %q; want the old line and one appended event", data)
	}
}

func TestWriterRotatesAtLocalMidnight(t *testing.T) {
	home := t.TempDir()
	current := time.Date(2026, time.January, 2, 23, 59, 59, 0, time.Local)
	writer, err := NewWriter(home, "cli", func() time.Time { return current })
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	logger.Debug("before midnight")
	current = time.Date(2026, time.January, 3, 0, 0, 1, 0, time.Local)
	logger.Debug("after midnight")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	for date, message := range map[string]string{
		"2026-01-02": "before midnight",
		"2026-01-03": "after midnight",
	} {
		event := readOneEvent(t, filepath.Join(home, ".hive", "logs", date+".jsonl"))
		if event["msg"] != message {
			t.Errorf("%s event message = %v; want %q", date, event["msg"], message)
		}
	}
}

func readOneEvent(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("log file is not one JSON event: %v; data %q", err, data)
	}
	return event
}

func TestLogFileNamesIncludeTheLocalDate(t *testing.T) {
	home := t.TempDir()
	current := time.Date(2026, 1, 2, 0, 0, 0, 0, time.FixedZone("example", 2*60*60))
	writer, err := NewWriter(home, "cli", func() time.Time { return current })
	if err != nil {
		t.Fatal(err)
	}
	slog.New(writer).Debug("local date")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	localDate := current.Local().Format("2006-01-02")
	if _, err := os.Stat(filepath.Join(home, ".hive", "logs", localDate+".jsonl")); err != nil {
		t.Fatalf("no log for local date %s: %v", localDate, err)
	}
}
