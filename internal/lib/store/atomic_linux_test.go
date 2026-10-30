package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

type inotifyEvent struct {
	name string
	mask uint32
}

func TestAtomicWriteCreatesDotPrefixedTemporaryFile(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Close(fd) })
	if _, err := syscall.InotifyAddWatch(fd, dir, syscall.IN_CREATE|syscall.IN_MOVED_TO); err != nil {
		t.Fatal(err)
	}
	if err := syscall.SetNonblock(fd, true); err != nil {
		t.Fatal(err)
	}
	if err := store.Write("tasks", "atomic", map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	events, err := readInotifyEvents(fd)
	if err != nil {
		t.Fatal(err)
	}
	var temporary string
	for _, event := range events {
		if event.mask&syscall.IN_CREATE == 0 || event.name == "" || event.name == "atomic.json" {
			continue
		}
		temporary = event.name
		break
	}
	if temporary == "" {
		t.Fatalf("events = %#v; want a create event for a temporary file", events)
	}
	if !strings.HasPrefix(temporary, ".") {
		t.Fatalf("temporary file %q does not start with a dot", temporary)
	}
	if _, err := os.Stat(filepath.Join(dir, temporary)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary file %q remains: stat error = %v", temporary, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "atomic.json" {
		t.Fatalf("record directory entries = %v; want only atomic.json", entries)
	}
}

func readInotifyEvents(fd int) ([]inotifyEvent, error) {
	var events []inotifyEvent
	buffer := make([]byte, 4096)
	for {
		count, err := syscall.Read(fd, buffer)
		if err != nil {
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				return events, nil
			}
			return nil, err
		}
		if count == 0 {
			return events, nil
		}
		data := buffer[:count]
		for len(data) >= syscall.SizeofInotifyEvent {
			mask := binary.NativeEndian.Uint32(data[4:8])
			nameLength := int(binary.NativeEndian.Uint32(data[12:16]))
			data = data[syscall.SizeofInotifyEvent:]
			if nameLength <= 0 {
				continue
			}
			if nameLength > len(data) {
				return nil, errors.New("short inotify event")
			}
			name := string(bytes.TrimRight(data[:nameLength], "\x00"))
			data = data[nameLength:]
			events = append(events, inotifyEvent{name: name, mask: mask})
		}
	}
}
