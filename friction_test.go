package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const frictionRoot = "reference/dogfood/friction"

func TestFrictionFilesMatchIndex(t *testing.T) {
	index, err := os.ReadFile(filepath.Join(frictionRoot, "FRICTION.md"))
	if err != nil {
		t.Fatalf("read friction index: %v", err)
	}
	rows := map[string]frictionRow{}
	for _, line := range strings.Split(string(index), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) != 3 {
			continue
		}
		issue := strings.TrimSpace(cells[0])
		if issue == "Issue" || strings.Trim(issue, "-:") == "" {
			continue
		}
		linkStart := strings.Index(issue, "](")
		if !strings.HasPrefix(issue, "[") || linkStart < 0 || !strings.HasSuffix(issue, ")") {
			t.Fatalf("invalid issue link in friction table: %q", issue)
		}
		title := issue[1:linkStart]
		path := filepath.ToSlash(filepath.Clean(issue[linkStart+2 : len(issue)-1]))
		if !strings.HasPrefix(path, "entries/") || !strings.HasSuffix(path, ".md") {
			t.Fatalf("invalid entry path in friction table: %q", path)
		}
		if _, exists := rows[path]; exists {
			t.Fatalf("duplicate table row for entry file: %s", path)
		}
		rows[path] = frictionRow{title: title, status: strings.TrimSpace(cells[1])}
	}

	entries := map[string]frictionEntry{}
	entryRoot := filepath.Join(frictionRoot, "entries")
	err = filepath.WalkDir(entryRoot, func(path string, item fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if item.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		relative, err := filepath.Rel(frictionRoot, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entry, err := frictionFrontmatter(string(content))
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.ToSlash(relative), err)
		}
		entries[filepath.ToSlash(relative)] = entry
		return nil
	})
	if err != nil {
		t.Fatalf("read friction entries: %v", err)
	}

	for path, row := range rows {
		entry, ok := entries[path]
		if !ok {
			t.Errorf("table row has no entry file: %s", path)
			continue
		}
		if row.title != entry.title {
			t.Errorf("title mismatch for %s: table has %q, frontmatter has %q", path, row.title, entry.title)
		}
		if row.status != entry.status {
			t.Errorf("status mismatch for %s: table has %q, frontmatter has %q", path, row.status, entry.status)
		}
		if entry.status == "Open" && strings.HasPrefix(path, "entries/resolved/") {
			t.Errorf("open entry is in entries/resolved/: %s", path)
		}
		if strings.HasPrefix(entry.status, "Resolved ") && strings.HasPrefix(path, "entries/") && !strings.HasPrefix(path, "entries/resolved/") {
			t.Errorf("resolved entry is in entries/: %s", path)
		}
	}
	for path := range entries {
		if _, ok := rows[path]; !ok {
			t.Errorf("entry file has no table row: %s", path)
		}
	}
}

type frictionRow struct {
	title  string
	status string
}

type frictionEntry struct {
	title  string
	status string
}

func frictionFrontmatter(content string) (frictionEntry, error) {
	if !strings.HasPrefix(content, "---\n") {
		return frictionEntry{}, fmt.Errorf("missing YAML frontmatter")
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return frictionEntry{}, fmt.Errorf("unterminated YAML frontmatter")
	}
	var entry frictionEntry
	for _, line := range strings.Split(rest[:end], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"`)
		switch strings.TrimSpace(key) {
		case "title":
			entry.title = value
		case "status":
			entry.status = value
		}
	}
	if entry.title == "" || (entry.status != "Open" && !strings.HasPrefix(entry.status, "Resolved ")) {
		return frictionEntry{}, fmt.Errorf("frontmatter must have a title and an Open or Resolved status")
	}
	return entry, nil
}
