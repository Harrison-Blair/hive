package project

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestListSortsProjectsByID(t *testing.T) {
	home := t.TempDir()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"zeta", "alpha"} {
		if err := registry.Write("projects", id, records.Project{ID: id, Path: "/repo/" + id, Base: "main", RegisteredAt: "now"}); err != nil {
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
