package memory

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxMemoryFileSize = 1 << 20
	indexName         = "MEMORY.md"
)

type indexEntry struct {
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

func scanMemoryFiles(dir string) ([]indexEntry, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	found := make([]indexEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == indexName || !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := readMemoryFile(path)
		if err != nil {
			return nil, fmt.Errorf("memory file invalid: %s: %v", path, err)
		}
		base := strings.TrimSuffix(name, ".md")
		values, err := parseMemory(data, base)
		if err != nil {
			return nil, fmt.Errorf("memory file invalid: %s: %v", path, err)
		}
		found = append(found, indexEntry{Name: base, Description: values["description"]})
	}
	sort.Slice(found, func(left, right int) bool {
		return found[left].Name < found[right].Name
	})
	return found, nil
}

func buildIndex(entries []indexEntry) []byte {
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

func upsertEntry(entries []indexEntry, entry indexEntry) []indexEntry {
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
