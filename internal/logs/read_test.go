package logs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestReader(t *testing.T, now time.Time) (*Reader, string) {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	reader := NewReader(home)
	reader.now = func() time.Time { return now }
	return reader, home
}

func eventLine(t *testing.T, when time.Time, fields map[string]any) string {
	t.Helper()
	event := map[string]any{"time": when.Format(time.RFC3339Nano)}
	for key, value := range fields {
		event[key] = value
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func dailyPath(home string, day time.Time) string {
	return filepath.Join(home, ".hive", "logs", day.Format("2006-01-02")+".jsonl")
}

func writeDaily(t *testing.T, home string, day time.Time, lines ...string) {
	t.Helper()
	data := strings.Join(lines, "\n")
	if len(lines) != 0 {
		data += "\n"
	}
	if err := os.WriteFile(dailyPath(home, day), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendDaily(t *testing.T, home string, day time.Time, text string) {
	t.Helper()
	file, err := os.OpenFile(dailyPath(home, day), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func localNoon(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, time.Local)
}

func TestReadReadsDailyFilesInDateOrder(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	first := today.AddDate(0, 0, -2)
	second := today.AddDate(0, 0, -1)
	writeDaily(t, home, first, eventLine(t, first.Add(-11*time.Hour), map[string]any{"msg": "one"}))
	writeDaily(t, home, second, eventLine(t, second.Add(-11*time.Hour), map[string]any{"msg": "two"}))
	writeDaily(t, home, today, eventLine(t, today.Add(-11*time.Hour), map[string]any{"msg": "three"}))

	events, offset, err := reader.Read(context.Background(), startOfDay(first))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"one", "two", "three"}
	if len(events) != len(want) {
		t.Fatalf("events = %d; want %d", len(events), len(want))
	}
	for index, message := range want {
		if got := eventMessage(events[index]); got != message {
			t.Errorf("event %d message = %q; want %q", index, got, message)
		}
	}
	if offset.Path != dailyPath(home, today) || offset.Byte != fileSize(t, dailyPath(home, today)) {
		t.Errorf("offset = %#v; want end of %s", offset, dailyPath(home, today))
	}
}

func TestReadSkipsMissingDailyFile(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	first := today.AddDate(0, 0, -2)
	writeDaily(t, home, first, eventLine(t, first, map[string]any{"msg": "one"}))
	writeDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "three"}))

	events, _, err := reader.Read(context.Background(), first)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || eventMessage(events[0]) != "one" || eventMessage(events[1]) != "three" {
		t.Fatalf("events = %#v; want one then three", events)
	}
}

func TestReadLineCap(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	long := strings.Repeat("x", maxLineBytes+1)
	valid := eventLine(t, today, map[string]any{"msg": "after long line"})
	writeDaily(t, home, today, long, valid)

	events, _, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d; want 2", len(events))
	}
	if !events[0].Malformed() {
		t.Fatalf("event 0 is not malformed")
	}
	if len(events[0].Raw) != 200 || !strings.HasPrefix(string(events[0].Raw), "xxxx") {
		t.Errorf("malformed prefix length = %d; want the first 200 bytes", len(events[0].Raw))
	}
	if rendered := MalformedText(events[0]); !strings.HasPrefix(rendered, "? xxxx") {
		t.Errorf("malformed text = %q; want ? and the raw prefix", rendered)
	}
	if events[1].Malformed() || eventMessage(events[1]) != "after long line" {
		t.Errorf("event after the long line = %#v; want valid", events[1])
	}
}

func TestReadAcceptsLineAtCap(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	head := `{"time":"` + today.Format(time.RFC3339Nano) + `","msg":"`
	tail := `"}`
	padding := maxLineBytes - len(head) - len(tail)
	if padding < 0 {
		t.Fatal("test line head is too long")
	}
	line := head + strings.Repeat("x", padding) + tail
	if len(line) != maxLineBytes {
		t.Fatalf("test line length = %d; want %d", len(line), maxLineBytes)
	}
	writeDaily(t, home, today, line, eventLine(t, today, map[string]any{"msg": "after cap"}))

	events, _, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Malformed() || events[1].Malformed() {
		t.Fatalf("events = %#v; want two valid lines at and after the cap", events)
	}
}

func TestReadMarksInvalidJSONMalformed(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	writeDaily(t, home, today, "not json", eventLine(t, today, map[string]any{"msg": "valid"}))

	events, _, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || !events[0].Malformed() || string(events[0].Raw) != "not json" {
		t.Fatalf("events = %#v; want malformed not json then valid", events)
	}
	if events[1].Malformed() || eventMessage(events[1]) != "valid" {
		t.Errorf("event 1 = %#v; want valid", events[1])
	}
}

func TestReadRejectsNonRegularFile(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	path := dailyPath(home, today)
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := reader.Read(context.Background(), time.Time{}); err == nil {
		t.Fatal("Read() error = nil; want an error for a directory")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "target.jsonl")
	writeDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "target"}))
	if err := os.Rename(dailyPath(home, today), target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := reader.Read(context.Background(), time.Time{}); err == nil {
		t.Fatal("Read() error = nil; want an error for a symbolic link")
	}
}

func TestSinceCutoffKeepsAtAndAfter(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	cutoff := today.Add(-2 * time.Hour)
	writeDaily(t, home, today,
		eventLine(t, cutoff.Add(-time.Second), map[string]any{"msg": "before"}),
		eventLine(t, cutoff, map[string]any{"msg": "at"}),
		eventLine(t, cutoff.Add(time.Second), map[string]any{"msg": "after"}),
	)

	events, _, err := reader.Read(context.Background(), cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || eventMessage(events[0]) != "at" || eventMessage(events[1]) != "after" {
		t.Fatalf("events = %#v; want at then after", events)
	}
}

func TestSincePreviousDayReadsTwoFiles(t *testing.T) {
	today := localNoon(time.Now())
	previous := today.AddDate(0, 0, -1)
	reader, home := newTestReader(t, today)
	cutoff := previous.Add(time.Hour)
	writeDaily(t, home, previous,
		eventLine(t, cutoff.Add(-time.Minute), map[string]any{"msg": "old"}),
		eventLine(t, cutoff.Add(time.Minute), map[string]any{"msg": "previous after"}),
	)
	writeDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "today"}))

	events, _, err := reader.Read(context.Background(), cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || eventMessage(events[0]) != "previous after" || eventMessage(events[1]) != "today" {
		t.Fatalf("events = %#v; want previous after then today", events)
	}
}

func eventMessage(event Event) string {
	message, _ := event.Fields["msg"].(string)
	return message
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}
