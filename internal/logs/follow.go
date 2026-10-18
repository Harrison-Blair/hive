package logs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const followPollInterval = 500 * time.Millisecond

// Follow sends each new event after the offset. It polls the file every 500
// milliseconds, keeps a partial line until its newline arrives, and moves to
// the next daily file after it reads the rest of the old file.
func (r *Reader) Follow(ctx context.Context, from Offset) (<-chan Event, <-chan error) {
	events := make(chan Event)
	errs := make(chan error, 1)
	go func() {
		defer close(events)
		defer close(errs)

		path := from.Path
		if path == "" {
			path = r.pathFor(r.now().Local())
		}
		offset := from.Byte
		if offset < 0 {
			offset = 0
		}
		lineNumber := 0
		ticker := time.NewTicker(followPollInterval)
		defer ticker.Stop()

		for {
			nextOffset, nextLine, pending, err := r.pollFile(ctx, path, offset, lineNumber, events)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					errs <- nil
				} else {
					errs <- err
				}
				return
			}
			offset = nextOffset
			lineNumber = nextLine
			if !pending {
				nextPath := r.pathFor(r.now().Local())
				if nextPath != path {
					info, statErr := os.Lstat(nextPath)
					if statErr == nil && info.Mode().IsRegular() {
						path = nextPath
						offset = 0
						lineNumber = 0
						continue
					}
					if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
						errs <- statErr
						return
					}
				}
			}
			select {
			case <-ctx.Done():
				errs <- nil
				return
			case <-ticker.C:
			}
		}
	}()
	return events, errs
}

func (r *Reader) pollFile(ctx context.Context, path string, offset int64, lineNumber int, out chan<- Event) (int64, int, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return offset, lineNumber, false, nil
		}
		return offset, lineNumber, false, err
	}
	if !info.Mode().IsRegular() {
		return offset, lineNumber, false, fmt.Errorf("log is not a regular file: %s", path)
	}
	if info.Size() < offset {
		return offset, lineNumber, false, fmt.Errorf("log file truncated: %s", path)
	}
	if info.Size() == offset {
		return offset, lineNumber, false, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return offset, lineNumber, false, nil
		}
		return offset, lineNumber, false, err
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, lineNumber, false, err
	}

	reader := bufio.NewReaderSize(file, 64*1024)
	for {
		if err := ctx.Err(); err != nil {
			return offset, lineNumber, false, err
		}
		raw, total, complete, malformed, err := readBoundedLine(reader)
		if err != nil {
			return offset, lineNumber, false, err
		}
		if !complete {
			return offset, lineNumber, total > 0, nil
		}
		offset += int64(total)
		lineNumber++
		event := makeEvent(raw, malformed, path, lineNumber)
		select {
		case out <- event:
		case <-ctx.Done():
			return offset, lineNumber, false, ctx.Err()
		}
	}
}
