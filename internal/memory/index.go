package memory

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxMemoryFileSize = 1 << 20
	// IndexName is the file name of the memory index.
	IndexName = "MEMORY.md"
)

// memoryFileError keeps the path and the cause of one invalid memory file.
// It returns safe text for logs and exact text for the user.
type memoryFileError struct {
	path string
	err  error
}

func (e *memoryFileError) Error() string { return e.LogCategory() }

func (e *memoryFileError) Unwrap() error { return e.err }

// UserMessage returns the exact text for command output.
func (e *memoryFileError) UserMessage() string {
	return "memory file invalid: " + e.path + ": " + UserErrorText(e.err)
}

// LogCategory returns the safe text for one log event.
func (e *memoryFileError) LogCategory() string {
	return "memory file invalid: " + e.path + ": " + LogErrorText(e.err)
}

func invalidMemoryFile(path string, err error) error {
	return &memoryFileError{path: path, err: err}
}

// IndexEntry is one line of the memory index.
type IndexEntry struct {
	Name        string
	Description string
}

// readMemoryFile reads one memory file with the 1 MiB bound.
func readMemoryFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("symbolic link")
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxMemoryFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxMemoryFileSize {
		return nil, errors.New("file too large")
	}
	return data, nil
}

// ScanMemoryFiles reads and checks each memory file in dir, in name order.
func ScanMemoryFiles(dir string) ([]IndexEntry, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	found := make([]IndexEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == IndexName || !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := readMemoryFile(path)
		if err != nil {
			return nil, invalidMemoryFile(path, err)
		}
		base := strings.TrimSuffix(name, ".md")
		values, err := ParseMemory(data, base)
		if err != nil {
			return nil, invalidMemoryFile(path, err)
		}
		found = append(found, IndexEntry{Name: base, Description: values["description"]})
	}
	sort.Slice(found, func(left, right int) bool {
		return found[left].Name < found[right].Name
	})
	return found, nil
}

// BuildIndex renders the memory index from its entries.
func BuildIndex(entries []IndexEntry) []byte {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString("- [")
		builder.WriteString(entry.Name)
		builder.WriteString("](")
		builder.WriteString(entry.Name)
		builder.WriteString(".md): ")
		builder.WriteString(entry.Description)
		builder.WriteByte('\n')
	}
	return []byte(builder.String())
}

// UpsertEntry replaces the entry with the same name or appends the entry.
func UpsertEntry(entries []IndexEntry, entry IndexEntry) []IndexEntry {
	for index := range entries {
		if entries[index].Name == entry.Name {
			entries[index] = entry
			return entries
		}
	}
	return append(entries, entry)
}

// writeAtomic writes one Markdown file with one temporary file and a rename.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
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
	return os.Rename(temp, path)
}

// WriteAtomic writes one Markdown file with one temporary file and a rename.
func WriteAtomic(path string, data []byte) error {
	return writeAtomic(path, data)
}
