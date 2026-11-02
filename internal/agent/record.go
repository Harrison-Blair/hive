package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ErrNotFound reports that no record matches the name or the pane.
var ErrNotFound = errors.New("agent record not found")

// Record is the saved fact set of one agent.
type Record struct {
	Name      string    `json:"name"`
	Pane      string    `json:"pane"`
	Parent    string    `json:"parent"`
	Harness   string    `json:"harness"`
	Model     string    `json:"model"`
	Cwd       string    `json:"cwd"`
	Worktree  string    `json:"worktree"`
	StartedAt time.Time `json:"started_at"`
}

// MaxNameLen is the longest agent name in bytes.
const MaxNameLen = 64

// ValidateName returns an error unless name is 1 to 64 characters of [a-z0-9_-].
func ValidateName(name string) error {
	if name == "" {
		return errors.New("agent name is empty")
	}
	if len(name) > MaxNameLen {
		return fmt.Errorf("agent name %q is longer than %d characters", name, MaxNameLen)
	}
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return fmt.Errorf("agent name %q has a character outside [a-z0-9_-]", name)
		}
	}
	return nil
}

func dir(home string) string {
	return filepath.Join(home, "agents")
}

func path(home, name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}
	return filepath.Join(dir(home), name+".json"), nil
}

// Save writes the record atomically. It creates the directory when needed.
func Save(home string, record Record) error {
	target, err := path(home, record.Name)
	if err != nil {
		return err
	}
	record.StartedAt = record.StartedAt.UTC()
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir(home), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir(home), "."+record.Name+".json.tmp-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, target)
}

// Load reads the record of name. A missing record gives ErrNotFound.
func Load(home, name string) (Record, error) {
	target, err := path(home, name)
	if err != nil {
		return Record{}, err
	}
	return read(target)
}

func read(file string) (Record, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return Record{}, fmt.Errorf("%w: %s", ErrNotFound, strings.TrimSuffix(filepath.Base(file), ".json"))
	}
	if err != nil {
		return Record{}, err
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("read agent record %s: %w", file, err)
	}
	return record, nil
}

// List returns each record sorted by name. It skips files without the .json
// suffix and hidden files.
func List(home string) ([]Record, error) {
	entries, err := os.ReadDir(dir(home))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var records []Record
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".json") {
			continue
		}
		record, err := read(filepath.Join(dir(home), name))
		if errors.Is(err, ErrNotFound) {
			// Another process removed the record after ReadDir.
			continue
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return records, nil
}

// Remove deletes the record of name. A missing record is not an error.
func Remove(home, name string) error {
	target, err := path(home, name)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// removeWait bounds the wait of RemoveIfSame for the layout lock. It is short,
// because a spawn holds that lock until its harness starts, and a list or a
// graph refresh must not stall that long.
const removeWait = 2 * time.Second

// RemoveIfSame deletes the record of want.Name when the file still holds the
// same agent: the same pane and start time. It holds the layout lock, which
// each writer of records holds, so a record that a spawn or a bare hive saved
// for the name in between stays, also in the same pane. It reports whether it
// deleted the record. A missing record is not an error.
func RemoveIfSame(home string, want Record) (bool, error) {
	target, err := path(home, want.Name)
	if err != nil {
		return false, err
	}
	records, err := store.Open(home)
	if err != nil {
		return false, err
	}
	removed := false
	err = records.Hold("layout", removeWait, func() error {
		record, err := read(target)
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		if err != nil || record.Pane != want.Pane || !record.StartedAt.Equal(want.StartedAt) {
			return err
		}
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		removed = true
		return nil
	})
	return removed, err
}

// FindByPane returns the record whose pane is pane, or ErrNotFound.
func FindByPane(home, pane string) (Record, error) {
	records, err := List(home)
	if err != nil {
		return Record{}, err
	}
	for _, record := range records {
		if record.Pane == pane {
			return record, nil
		}
	}
	return Record{}, fmt.Errorf("%w: pane %s", ErrNotFound, pane)
}
