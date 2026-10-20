package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

// StoredRecord is one JSON record that WalkRecords returns.
type StoredRecord struct {
	Collection []string
	Name       string
	Data       json.RawMessage
}

// Open creates the state directory if it does not exist. It rejects a static
// symbolic link or a wrong type at the .hive ancestor or the state root.
func Open(home string) (*Store, error) {
	store := &Store{home: home}
	if err := store.checkHive(); err != nil {
		return nil, err
	}
	if err := store.checkRoot(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(store.stateDir(), 0o755); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) stateDir() string {
	return filepath.Join(s.home, ".hive", "state")
}

func (s *Store) lockPath() string {
	return filepath.Join(s.stateDir(), "lock")
}

// Read reads one record without taking the mutation lock.
func (s *Store) Read(kind, id string, target any) error {
	return s.readUnlocked(kind, id, target)
}

func (s *Store) readUnlocked(kind, id string, target any) error {
	path, err := s.recordPath(kind, id)
	if err != nil {
		return err
	}
	if err := s.checkRead(path, []string{kind}); err != nil {
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
	data, err := marshalRecord(value)
	if err != nil {
		return err
	}
	return s.withLock(context.Background(), func(locked *LockedStore) error {
		return locked.store.writeData(kind, id, data)
	})
}

// Update runs one read-change-write operation under the store lock.
func (s *Store) Update(kind, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return s.withLock(context.Background(), func(locked *LockedStore) error {
		return locked.Update(kind, id, fn)
	})
}

// List reads all valid store segments of one kind in ID order.
func (s *Store) List(kind string) ([]json.RawMessage, error) {
	if err := validateSegment("kind", kind); err != nil {
		return nil, err
	}
	return s.listUnlocked(kind)
}

func (s *Store) listUnlocked(kind string) ([]json.RawMessage, error) {
	dir := filepath.Join(s.stateDir(), kind)
	if err := s.checkList(dir, []string{kind}); err != nil {
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
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if err := validateSegment("id", id); err != nil {
			continue
		}
		data, err := readRecordFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		found = append(found, json.RawMessage(data))
	}
	return found, nil
}

// Next increments and returns a named counter. The first value is 1.
func (s *Store) Next(name string) (uint64, error) {
	var next uint64
	err := s.withLock(context.Background(), func(locked *LockedStore) error {
		var err error
		next, err = locked.Next(name)
		return err
	})
	return next, err
}

func (s *Store) nextUnlocked(name string) (uint64, error) {
	if err := validateSegment("counter", name); err != nil {
		return 0, err
	}
	var next uint64
	err := s.updateUnlocked("counters", name, func(raw json.RawMessage) (json.RawMessage, error) {
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

// ReadIn reads one record below a nested collection without taking the lock.
func (s *Store) ReadIn(collection []string, id string, target any) error {
	return s.readInUnlocked(collection, id, target)
}

func (s *Store) readInUnlocked(collection []string, id string, target any) error {
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
	return s.withLock(context.Background(), func(locked *LockedStore) error {
		return locked.store.writeInData(collection, id, data)
	})
}

// UpdateIn runs one read-change-write operation below a nested collection.
func (s *Store) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return s.withLock(context.Background(), func(locked *LockedStore) error {
		return locked.UpdateIn(collection, id, fn)
	})
}

// ListIn reads the immediate directory of one nested collection without a lock.
func (s *Store) ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error) {
	return s.listInUnlocked(collection, include)
}

func (s *Store) listInUnlocked(collection []string, include func(filename string) bool) ([]json.RawMessage, error) {
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
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if err := validateSegment("id", id); err != nil {
			continue
		}
		if include != nil && !include(name) {
			continue
		}
		data, err := readRecordFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		found = append(found, json.RawMessage(data))
	}
	return found, nil
}

// WalkRecords walks each record below the state root under the store lock.
func (s *Store) WalkRecords(ctx context.Context, fn func(StoredRecord) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.WithLock(ctx, func(locked *LockedStore) error {
		return locked.walkRecords(ctx, fn)
	})
}

func (l *LockedStore) walkRecords(ctx context.Context, fn func(StoredRecord) error) error {
	if err := l.available(); err != nil {
		return err
	}
	root := l.store.stateDir()
	if err := l.store.checkRoot(); err != nil {
		return err
	}
	nodes, err := collectWalkNodes(root, nil, true)
	if err != nil {
		return err
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].rel < nodes[j].rel
	})
	for _, node := range nodes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if node.link {
			if strings.HasSuffix(node.name, ".json") {
				return fmt.Errorf("record invalid: %s", node.path)
			}
			return fmt.Errorf("store path invalid: %s", node.path)
		}
		if node.special {
			if strings.HasSuffix(node.name, ".json") {
				return fmt.Errorf("record invalid: %s", node.path)
			}
			return fmt.Errorf("store path invalid: %s", node.path)
		}
		if node.dir || !strings.HasSuffix(node.name, ".json") {
			continue
		}
		data, err := readRecordFile(node.path)
		if err != nil {
			return err
		}
		record := StoredRecord{
			Collection: append([]string(nil), node.collection...),
			Name:       node.name,
			Data:       data,
		}
		if err := fn(record); err != nil {
			return err
		}
	}
	return nil
}

type walkNode struct {
	path       string
	rel        string
	name       string
	collection []string
	dir        bool
	link       bool
	special    bool
}

func collectWalkNodes(dir string, collection []string, root bool) ([]walkNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	nodes := []walkNode{}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		relParts := append(append([]string(nil), collection...), name)
		node := walkNode{
			path:       path,
			rel:        strings.Join(relParts, "/"),
			name:       name,
			collection: append([]string(nil), collection...),
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			node.link = true
			nodes = append(nodes, node)
		case info.IsDir():
			node.dir = true
			nodes = append(nodes, node)
			children, err := collectWalkNodes(path, relParts, false)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, children...)
		case info.Mode().IsRegular() && root && name == "lock":
			continue
		case info.Mode().IsRegular() && strings.HasPrefix(name, ".") && strings.Contains(name, ".tmp-"):
			continue
		case info.Mode().IsRegular():
			nodes = append(nodes, node)
		default:
			node.special = true
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (s *Store) writeData(kind, id string, data []byte) error {
	path, err := s.recordPath(kind, id)
	if err != nil {
		return err
	}
	return s.writePath(path, []string{kind}, data)
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

func (s *Store) updateUnlocked(kind, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	path, err := s.recordPath(kind, id)
	if err != nil {
		return err
	}
	return s.updatePath(path, []string{kind}, fn)
}

func (s *Store) updateInUnlocked(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	path, err := s.recordPathIn(collection, id)
	if err != nil {
		return err
	}
	return s.updatePath(path, collection, fn)
}

func (s *Store) updatePath(path string, collection []string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	if err := s.checkRead(path, collection); err != nil {
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
	return s.writePath(path, collection, next)
}

func (s *Store) recordPath(kind, id string) (string, error) {
	if err := validateSegment("kind", kind); err != nil {
		return "", err
	}
	if err := validateSegment("id", id); err != nil {
		return "", err
	}
	return filepath.Join(s.stateDir(), kind, id+".json"), nil
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
	if err := s.checkHive(); err != nil {
		return err
	}
	if err := s.checkRoot(); err != nil {
		return err
	}
	if err := s.checkLock(); err != nil {
		return err
	}
	if err := checkDirs(s.stateDir(), collection); err != nil {
		return err
	}
	return checkLeaf(path)
}

func (s *Store) checkList(dir string, collection []string) error {
	if err := s.checkHive(); err != nil {
		return err
	}
	if err := s.checkRoot(); err != nil {
		return err
	}
	if err := s.checkLock(); err != nil {
		return err
	}
	return checkDirs(s.stateDir(), collection)
}

func (s *Store) checkHive() error {
	path := filepath.Join(s.home, ".hive")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("store path invalid: %s", path)
	}
	return nil
}

func (s *Store) checkRoot() error {
	root := s.stateDir()
	info, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("store path invalid: %s", root)
	}
	return nil
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

// validateSegment keeps the legacy name rule for Read, Write, Update, and Next.
func validateSegment(kind, value string) error {
	if value == "" || strings.Contains(value, "/") || strings.Contains(value, "..") {
		return fmt.Errorf("invalid store %s: %q", kind, value)
	}
	return nil
}
