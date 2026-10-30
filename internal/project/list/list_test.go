package list

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestListSortsProjectsByID(t *testing.T) {
	home := t.TempDir()
	for _, id := range []string{"zeta", "alpha"} {
		dir := filepath.Join(home, "projects", id)
		if err := os.MkdirAll(filepath.Join(dir, "worktrees"), 0o755); err != nil {
			t.Fatal(err)
		}
		record := `{"id":"` + id + `","path":"/repo/` + id + `","base":"main","registered_at":"now"}`
		if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(record), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	envelope := List(home)
	projects, ok := envelope.Result.([]records.Project)
	if envelope.Outcome != "ok" || !ok || len(projects) != 2 || projects[0].ID != "alpha" || projects[1].ID != "zeta" {
		t.Fatalf("List() = %#v; want alpha then zeta", envelope)
	}
}

func TestListEmptyProjectsIsAnEmptyArray(t *testing.T) {
	envelope := List(t.TempDir())
	projects, ok := envelope.Result.([]records.Project)
	if envelope.Outcome != "ok" || !ok || projects == nil || len(projects) != 0 {
		t.Fatalf("List() = %#v; want non-nil empty slice", envelope)
	}
}
