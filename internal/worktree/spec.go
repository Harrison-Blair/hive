package worktree

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// specCopyCollection is the state collection of the copy-in records.
const specCopyCollection = "spec-copies"

// SpecCopy is the copy-in record of one task. Land compares each spec file
// with its hash at copy-in.
type SpecCopy struct {
	Task     string     `json:"task"`
	Feature  string     `json:"feature"`
	Files    []SpecFile `json:"files"`
	CopiedAt string     `json:"copied_at"`
}

// SpecFile is one copied file: its repository-relative path and its SHA-256
// hash in hex.
type SpecFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// SpecCopyRelPath returns the home-relative path of one copy-in record.
func SpecCopyRelPath(taskID string) string {
	return ".hive/state/" + specCopyCollection + "/" + taskID + ".json"
}

// ReadSpecCopy reads the copy-in record of one task.
func ReadSpecCopy(home, taskID string) (SpecCopy, bool, error) {
	registry, err := store.Open(home)
	if err != nil {
		return SpecCopy{}, false, err
	}
	var record SpecCopy
	if err := registry.ReadIn([]string{specCopyCollection}, taskID, &record); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SpecCopy{}, false, nil
		}
		return SpecCopy{}, false, err
	}
	return record, true, nil
}

// HashFile returns the SHA-256 hash of one file in hex.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// featureSpecRelDir returns the repository-relative folder of one feature.
func featureSpecRelDir(feature string) string {
	return ".hive/spec/" + feature + "/"
}

// copySpecIn copies the feature folder of the home into the worktree. It
// skips each task folder other than the folder of the task, and each file that
// is not a regular file. A home without the folder copies nothing.
func copySpecIn(home, worktreePath string, task records.Task) (SpecCopy, error) {
	record := SpecCopy{Task: task.ID, Feature: task.Feature, Files: []SpecFile{}, CopiedAt: now()}
	relDir := featureSpecRelDir(task.Feature)
	source := filepath.Join(home, filepath.FromSlash(relDir))
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		inner, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		inner = filepath.ToSlash(inner)
		if entry.IsDir() {
			parent, name, nested := strings.Cut(inner, "/")
			if nested && parent == "tasks" && !strings.Contains(name, "/") && name != task.ID {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative := relDir + inner
		destination := filepath.Join(worktreePath, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		record.Files = append(record.Files, SpecFile{Path: relative, SHA256: hex.EncodeToString(sum[:])})
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return record, fmt.Errorf("copy spec in: %s", oneLine(err.Error()))
	}
	return record, nil
}

// writeSpecCopy writes the copy-in record of one task.
func writeSpecCopy(registry *store.Store, record SpecCopy) error {
	return registry.WriteIn([]string{specCopyCollection}, record.Task, record)
}

// DesignCopy gives the counts of the design copy-in of one worktree create.
type DesignCopy struct {
	Features int `json:"features"`
	Files    int `json:"files"`
}

// copyDesignIn copies the design/ folder of each feature of the home other
// than the feature of the task into the worktree. It copies each regular file
// below each folder and no other file. It writes no hash record, because land
// copies back only the feature of the task. A home without .hive/spec/ copies
// nothing.
func copyDesignIn(home, worktreePath string, task records.Task) (DesignCopy, error) {
	counts := DesignCopy{}
	features, err := os.ReadDir(filepath.Join(home, ".hive", "spec"))
	if errors.Is(err, os.ErrNotExist) {
		return counts, nil
	}
	if err != nil {
		return counts, fmt.Errorf("copy design in: %s", oneLine(err.Error()))
	}
	for _, feature := range features {
		if !feature.IsDir() || feature.Name() == task.Feature {
			continue
		}
		relDir := featureSpecRelDir(feature.Name()) + "design/"
		source := filepath.Join(home, filepath.FromSlash(relDir))
		if info, err := os.Stat(source); err != nil || !info.IsDir() {
			continue
		}
		counts.Features++
		err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			inner, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			destination := filepath.Join(worktreePath, filepath.FromSlash(relDir), inner)
			if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(destination, data, 0o644); err != nil {
				return err
			}
			counts.Files++
			return nil
		})
		if err != nil {
			return counts, fmt.Errorf("copy design in: %s", oneLine(err.Error()))
		}
	}
	return counts, nil
}
