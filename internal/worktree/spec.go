package worktree

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

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

// WriteSpecCopy writes the copy-in record of one task.
func WriteSpecCopy(registry *store.Store, record SpecCopy) error {
	return registry.WriteIn([]string{specCopyCollection}, record.Task, record)
}
