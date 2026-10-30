package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListEachReadsTheNamedRecordOfEachSubdirectory(t *testing.T) {
	store := openStore(t, t.TempDir())
	for _, id := range []string{"beta", "alpha"} {
		if err := store.WriteIn([]string{"projects", id}, "project", map[string]string{"id": id}); err != nil {
			t.Fatal(err)
		}
	}
	// A directory without the record, a stray file, and an invalid name are skipped.
	if err := os.MkdirAll(filepath.Join(store.home, "projects", "empty", "worktrees"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.home, "projects", "notes.json"), []byte(`{"id":"notes"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(store.home, "projects", "Bad"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.home, "projects", "Bad", "project.json"), []byte(`{"id":"Bad"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := store.ListEach([]string{"projects"}, "project")
	if err != nil {
		t.Fatalf("ListEach() error = %v", err)
	}
	var got []string
	for _, raw := range found {
		got = append(got, strings.Join(strings.Fields(string(raw)), ""))
	}
	if want := []string{`{"id":"alpha"}`, `{"id":"beta"}`}; strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("ListEach() = %v; want %v", got, want)
	}
}

func TestListEachMissingCollectionIsEmpty(t *testing.T) {
	store := openStore(t, t.TempDir())
	found, err := store.ListEach([]string{"projects"}, "project")
	if err != nil || found == nil || len(found) != 0 {
		t.Fatalf("ListEach() = %v, %v; want an empty list", found, err)
	}
}

func TestListEachRejectsLinkedSubdirectory(t *testing.T) {
	store := openStore(t, t.TempDir())
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "project.json"), []byte(`{"id":"linked"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(store.home, "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(store.home, "projects", "linked")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ListEach([]string{"projects"}, "project"); err == nil || !strings.Contains(err.Error(), "store path invalid") {
		t.Fatalf("ListEach() error = %v; want store path invalid", err)
	}
}
