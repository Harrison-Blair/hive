package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// newHome makes one fixture home with its marker.
func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

// registerProject writes one project record.
func registerProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: filepath.Join(home, "repo-"+id), Base: "main"}
	if err := registry.Write("projects", id, record); err != nil {
		t.Fatal(err)
	}
}
