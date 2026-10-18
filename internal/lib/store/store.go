package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const maxRecordSize = 1 << 20

type Store struct {
	home string
}

// Open creates the state directory if it does not exist.
func Open(home string) (*Store, error) {
	stateDir := filepath.Join(home, ".hive", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, err
	}
	return &Store{home: home}, nil
}

// Read reads one record without taking the mutation lock.
func (s *Store) Read(kind, id string, target any) error {
	path, err := s.recordPath(kind, id)
	if err != nil {
		return err
	}
	data, err := readRecordFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("record invalid: %s", path)
	}
	return nil
}

// Write replaces one record while holding the store lock.
func (s *Store) Write(kind, id string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return s.withLock(func() error {
		return s.write(kind, id, data)
	})
}

// Update runs one read-change-write operation under the store lock.
func (s *Store) Update(kind, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return s.withLock(func() error {
		path, err := s.recordPath(kind, id)
		if err != nil {
			return err
		}
		current, err := readRecordFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if errors.Is(err, os.ErrNotExist) {
			current = nil
		}
		next, err := fn(current)
		if err != nil {
			return err
		}
		if !validObject(next) {
			return fmt.Errorf("record invalid: %s", path)
		}
		return s.write(kind, id, next)
	})
}

// List reads all valid-ID records of one kind in ID order.
func (s *Store) List(kind string) ([]json.RawMessage, error) {
	if err := validateSegment("kind", kind); err != nil {
		return nil, err
	}
	dir := filepath.Join(s.home, ".hive", "state", kind)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []json.RawMessage{}, nil
	}
	if err != nil {
		return nil, err
	}
	var recordsFound []json.RawMessage
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if !records.ValidID(id) {
			continue
		}
		data, err := readRecordFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		recordsFound = append(recordsFound, json.RawMessage(data))
	}
	if recordsFound == nil {
		recordsFound = []json.RawMessage{}
	}
	return recordsFound, nil
}

// Next increments and returns a named counter. The first value is 1.
func (s *Store) Next(name string) (uint64, error) {
	if err := validateSegment("counter", name); err != nil {
		return 0, err
	}
	var next uint64
	err := s.Update("counters", name, func(raw json.RawMessage) (json.RawMessage, error) {
		var counter struct {
			Value uint64 `json:"value"`
		}
		if len(raw) != 0 {
			if err := json.Unmarshal(raw, &counter); err != nil {
				return nil, err
			}
		}
		counter.Value++
		next = counter.Value
		return json.Marshal(counter)
	})
	return next, err
}

func (s *Store) write(kind, id string, data []byte) error {
	path, err := s.recordPath(kind, id)
	if err != nil {
		return err
	}
	if !validObject(data) {
		return fmt.Errorf("record invalid: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeRecordFile(path, data)
}

func (s *Store) recordPath(kind, id string) (string, error) {
	if err := validateSegment("kind", kind); err != nil {
		return "", err
	}
	if err := validateSegment("id", id); err != nil {
		return "", err
	}
	return filepath.Join(s.home, ".hive", "state", kind, id+".json"), nil
}

func (s *Store) withLock(fn func() error) error {
	lockPath := filepath.Join(s.home, ".hive", "state", "lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = lock.Close()
			return err
		}
		if !time.Now().Before(deadline) {
			_ = lock.Close()
			return ErrLockTimeout
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer func() { _ = lock.Close() }()
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) //nolint:errcheck
	return fn()
}

func readRecordFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("record invalid: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxRecordSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxRecordSize {
		return nil, fmt.Errorf("record too large: %s", path)
	}
	if !validObject(data) {
		return nil, fmt.Errorf("record invalid: %s", path)
	}
	return data, nil
}

func writeRecordFile(path string, data []byte) error {
	if !validObject(data) {
		return fmt.Errorf("record invalid: %s", path)
	}
	return atomicWrite(path, data)
}

func validObject(data []byte) bool {
	if len(bytes.TrimSpace(data)) == 0 || !json.Valid(data) {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(data, &object) == nil && object != nil
}

func validateSegment(kind, value string) error {
	if value == "" || strings.Contains(value, "/") || strings.Contains(value, "..") {
		return fmt.Errorf("invalid store %s: %q", kind, value)
	}
	return nil
}
