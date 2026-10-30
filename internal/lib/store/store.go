package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxRecordSize    = 1 << 20
	maxComponents    = 8
	maxComponentSize = 200
)

var componentPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type Store struct {
	home string
}

// Open creates the home if it does not exist and checks its schema.
func Open(home string) (*Store, error) {
	store := &Store{home: home}
	if err := os.MkdirAll(store.stateDir(), 0o755); err != nil {
		return nil, err
	}
	if err := store.checkSchema(); err != nil {
		return nil, err
	}
	return store, nil
}

// stateDir is the root of the records, the lock, and the schema: the home.
func (s *Store) stateDir() string {
	return s.home
}

func (s *Store) lockPath() string {
	return filepath.Join(s.stateDir(), "lock")
}

// ReadIn reads one record below a nested collection without taking the lock.
func (s *Store) ReadIn(collection []string, id string, target any) error {
	path, err := s.recordPathIn(collection, id)
	if err != nil {
		return err
	}
	if err := s.checkRead(path, collection); err != nil {
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

// WriteIn writes one record below a nested collection under the store lock.
func (s *Store) WriteIn(collection []string, id string, value any) error {
	data, err := marshalRecord(value)
	if err != nil {
		return err
	}
	return s.withLock(defaultLockWait, func() error {
		return s.writeInData(collection, id, data)
	})
}

// ListEach reads the record id in each immediate subdirectory of collection,
// in name order, without a lock. It skips a subdirectory without the record.
func (s *Store) ListEach(collection []string, id string) ([]json.RawMessage, error) {
	dir, err := s.collectionPath(collection)
	if err != nil {
		return nil, err
	}
	if err := s.checkList(dir, collection); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []json.RawMessage{}, nil
	}
	if err != nil {
		return nil, err
	}
	found := []json.RawMessage{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() || validateComponent("collection", name) != nil {
			continue
		}
		sub := append(append([]string{}, collection...), name)
		path, err := s.recordPathIn(sub, id)
		if err != nil {
			return nil, err
		}
		if err := s.checkRead(path, sub); err != nil {
			return nil, err
		}
		data, err := readRecordFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		found = append(found, json.RawMessage(data))
	}
	return found, nil
}

func (s *Store) writeInData(collection []string, id string, data []byte) error {
	path, err := s.recordPathIn(collection, id)
	if err != nil {
		return err
	}
	return s.writePath(path, collection, data)
}

func (s *Store) writePath(path string, collection []string, data []byte) error {
	if !validObject(data) {
		return fmt.Errorf("record invalid: %s", path)
	}
	if err := s.checkRead(path, collection); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeRecordFile(path, data)
}

func (s *Store) recordPathIn(collection []string, id string) (string, error) {
	dir, err := s.collectionPath(collection)
	if err != nil {
		return "", err
	}
	if err := validateComponent("id", id); err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".json"), nil
}

func (s *Store) collectionPath(collection []string) (string, error) {
	if err := validateCollection(collection); err != nil {
		return "", err
	}
	parts := append([]string{s.stateDir()}, collection...)
	return filepath.Join(parts...), nil
}

func (s *Store) checkRead(path string, collection []string) error {
	if err := s.checkLock(); err != nil {
		return err
	}
	if err := checkDirs(s.stateDir(), collection); err != nil {
		return err
	}
	return checkLeaf(path)
}

func (s *Store) checkList(dir string, collection []string) error {
	if err := s.checkLock(); err != nil {
		return err
	}
	return checkDirs(s.stateDir(), collection)
}

func (s *Store) checkLock() error {
	path := s.lockPath()
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("store path invalid: %s", path)
	}
	return nil
}

func checkDirs(base string, collection []string) error {
	current := base
	for _, name := range collection {
		current = filepath.Join(current, name)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("store path invalid: %s", current)
		}
	}
	return nil
}

func checkLeaf(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("record invalid: %s", path)
	}
	return nil
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
	if len(data) > maxRecordSize {
		return fmt.Errorf("record too large: %s", path)
	}
	if !validObject(data) {
		return fmt.Errorf("record invalid: %s", path)
	}
	return atomicWrite(path, data)
}

func marshalRecord(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func validObject(data []byte) bool {
	if len(bytes.TrimSpace(data)) == 0 || !json.Valid(data) {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(data, &object) == nil && object != nil
}

func validateCollection(collection []string) error {
	if len(collection) < 1 || len(collection) > maxComponents {
		return fmt.Errorf("invalid store collection: %q", strings.Join(collection, "/"))
	}
	for _, name := range collection {
		if err := validateComponent("collection", name); err != nil {
			return err
		}
	}
	return nil
}

func validateComponent(kind, value string) error {
	if value == "" || len(value) > maxComponentSize || !componentPattern.MatchString(value) {
		return fmt.Errorf("invalid store %s: %q", kind, value)
	}
	return nil
}
