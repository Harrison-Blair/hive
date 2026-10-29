package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRejectsLinkedHiveAncestor(t *testing.T) {
	home := t.TempDir()
	target := t.TempDir()
	hivePath := filepath.Join(home, ".hive")
	if err := os.Symlink(target, hivePath); err != nil {
		t.Fatal(err)
	}
	store, err := Open(home)
	if err == nil || err.Error() != "store path invalid: "+hivePath {
		t.Fatalf("Open() = %v, %v; want store path invalid for %s", store, err, hivePath)
	}
	entries, readErr := os.ReadDir(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("Open() created external state: %v", entries)
	}
}

func TestOpenRejectsWrongTypeHiveAncestor(t *testing.T) {
	home := t.TempDir()
	hivePath := filepath.Join(home, ".hive")
	if err := os.WriteFile(hivePath, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := Open(home)
	if err == nil || err.Error() != "store path invalid: "+hivePath {
		t.Fatalf("Open() = %v, %v; want store path invalid for %s", store, err, hivePath)
	}
	data, readErr := os.ReadFile(hivePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "not a directory" {
		t.Fatalf("Open() changed the wrong-type ancestor to %q", data)
	}
}

func TestOpenRejectsLinkedStateRoot(t *testing.T) {
	home := t.TempDir()
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(home, ".hive", "state")
	if err := os.Symlink(target, statePath); err != nil {
		t.Fatal(err)
	}
	store, err := Open(home)
	if err == nil || err.Error() != "store path invalid: "+statePath {
		t.Fatalf("Open() = %v, %v; want store path invalid for %s", store, err, statePath)
	}
	entries, readErr := os.ReadDir(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("Open() created external state: %v", entries)
	}
}

func TestOpenRejectsWrongTypeStateRoot(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(home, ".hive", "state")
	if err := os.WriteFile(statePath, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := Open(home)
	if err == nil || err.Error() != "store path invalid: "+statePath {
		t.Fatalf("Open() = %v, %v; want store path invalid for %s", store, err, statePath)
	}
	data, readErr := os.ReadFile(statePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "not a directory" {
		t.Fatalf("Open() changed the wrong-type state root to %q", data)
	}
}

func TestOperationsRecheckLinkedHiveAncestor(t *testing.T) {
	home := t.TempDir()
	store := openStore(t, home)
	if err := os.RemoveAll(filepath.Join(home, ".hive")); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	hivePath := filepath.Join(home, ".hive")
	if err := os.Symlink(target, hivePath); err != nil {
		t.Fatal(err)
	}
	want := "store path invalid: " + hivePath
	checks := []struct {
		name string
		run  func() error
	}{
		{"ReadIn", func() error {
			return store.ReadIn([]string{"tasks"}, "one", &map[string]any{})
		}},
		{"WriteIn", func() error {
			return store.WriteIn([]string{"tasks"}, "one", map[string]string{"id": "one"})
		}},
		{"ListIn", func() error {
			_, err := store.ListIn([]string{"tasks"}, nil)
			return err
		}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			err := check.run()
			if err == nil || err.Error() != want {
				t.Fatalf("%s() error = %v; want %q", check.name, err, want)
			}
		})
	}
	entries, readErr := os.ReadDir(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("operations created external state: %v", entries)
	}
}
