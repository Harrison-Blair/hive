package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const maxLineBytes = 1 << 20

// Event is one line from a daily log file. A line that is too long or is not a
// JSON object is malformed. A malformed event has no fields.
type Event struct {
	Raw    []byte
	Fields map[string]any

	path      string
	line      int
	malformed bool
}

// Malformed reports whether the line is not a valid event.
func (e Event) Malformed() bool { return e.malformed }

// Path returns the daily file of the event.
func (e Event) Path() string { return e.path }

// Line returns the one-based line number in the daily file.
func (e Event) Line() int { return e.line }

// Offset is the end of a read in one daily file.
type Offset struct {
	Path string
	Byte int64

	cutoff time.Time
	line   int
}

// Reader reads the daily log files of one home.
type Reader struct {
	home string
	now  func() time.Time
}

// NewReader makes a reader for one home.
func NewReader(home string) *Reader {
	return &Reader{home: home, now: time.Now}
}

// Read returns the events at or after the cutoff and the end offset. A zero
// cutoff reads today and keeps each event. It reads each daily file from the
// cutoff date to today in date order and skips the dates without a file.
func (r *Reader) Read(ctx context.Context, since time.Time) ([]Event, Offset, error) {
	today := startOfDay(r.now().Local())
	first := today
	cutoff := time.Time{}
	if !since.IsZero() {
		cutoff = since
		first = startOfDay(since.Local())
	}
	offset := Offset{Path: r.pathFor(today), cutoff: cutoff}
	var events []Event
	for day := first; !day.After(today); day = day.AddDate(0, 0, 1) {
		if err := ctx.Err(); err != nil {
			return events, offset, err
		}
		path := r.pathFor(day)
		fileEvents, consumed, lines, exists, err := r.readFile(ctx, path, cutoff)
		if err != nil {
			return events, offset, err
		}
		events = append(events, fileEvents...)
		if day.Equal(today) {
			offset.Path = path
			offset.line = lines
			if exists {
				offset.Byte = consumed
			}
		}
	}
	return events, offset, nil
}

func (r *Reader) readFile(ctx context.Context, path string, cutoff time.Time) ([]Event, int64, int, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, 0, false, nil
		}
		return nil, 0, 0, true, err
	}
	if !info.Mode().IsRegular() {
		return nil, 0, 0, true, fmt.Errorf("log is not a regular file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, true, err
	}
	defer func() { _ = file.Close() }()

	reader := bufio.NewReaderSize(file, 64*1024)
	var events []Event
	var consumed int64
	lineNumber := 0
	for {
		if err := ctx.Err(); err != nil {
			return events, consumed, lineNumber, true, err
		}
		raw, total, complete, malformed, err := readBoundedLine(reader)
		if err != nil {
			return events, consumed, lineNumber, true, err
		}
		if !complete {
			break
		}
		consumed += int64(total)
		lineNumber++
		event := makeEvent(raw, malformed, path, lineNumber)
		if !eventAfterCutoff(event, cutoff) {
			continue
		}
		events = append(events, event)
	}
	return events, consumed, lineNumber, true, nil
}

// eventAfterCutoff reports whether an event is at or after the cutoff. A zero
// cutoff keeps each event. A malformed event has no time and is always kept.
func eventAfterCutoff(event Event, cutoff time.Time) bool {
	if cutoff.IsZero() || event.Malformed() {
		return true
	}
	when, ok := eventTime(event.Fields)
	return ok && !when.Before(cutoff)
}

func makeEvent(raw []byte, malformed bool, path string, line int) Event {
	event := Event{Raw: raw, path: path, line: line, malformed: malformed}
	if malformed {
		return event
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		event.malformed = true
		return event
	}
	event.Fields = fields
	return event
}

// readBoundedLine reads one line with a 1 MiB cap. A longer line is drained
// without a copy and returns only its first 200 bytes. An incomplete final line
// returns complete false.
func readBoundedLine(reader *bufio.Reader) ([]byte, int, bool, bool, error) {
	var line []byte
	prefix := make([]byte, 0, 200)
	total := 0
	malformed := false
	for {
		chunk, err := reader.ReadSlice('\n')
		total += len(chunk)
		if len(prefix) < 200 {
			need := 200 - len(prefix)
			if need > len(chunk) {
				need = len(chunk)
			}
			prefix = append(prefix, chunk[:need]...)
		}
		if !malformed {
			content := total
			if err == nil {
				content--
			}
			if content > maxLineBytes {
				malformed = true
				line = nil
			} else {
				line = append(line, chunk...)
			}
		}
		if err == nil {
			if malformed {
				return prefix, total, true, true, nil
			}
			if len(line) != 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) != 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			return line, total, true, false, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil, total, false, malformed, nil
		}
		return nil, total, false, malformed, err
	}
}

func (r *Reader) pathFor(day time.Time) string {
	return filepath.Join(r.home, ".hive", "logs", day.Format("2006-01-02")+".jsonl")
}

func startOfDay(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
}
