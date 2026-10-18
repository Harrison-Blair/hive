package project

import (
	"os"
	"path/filepath"
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

func TestShowReturnsOneProject(t *testing.T) {
	home := t.TempDir()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	want := records.Project{ID: "alpha", Path: "/repo/alpha", Base: "dev", RegisteredAt: "2026-01-02T03:04:05Z"}
	if err := registry.Write("projects", want.ID, want); err != nil {
		t.Fatal(err)
	}
	envelope := Show(home, want.ID)
	if envelope.Outcome != "ok" || envelope.Result != want {
		t.Fatalf("Show() = %#v; want %#v", envelope, want)
	}
}

func TestRegisterShowRejectsInvalidIDBeforeStoreAccess(t *testing.T) {
	home := t.TempDir()
	envelope := Show(home, "bad_id")
	want := "invalid project id: bad_id: an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters"
	if envelope.Error != want {
		t.Fatalf("error = %v; want %q", envelope.Error, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("invalid ID caused store access: stat error = %v", err)
	}
}

func TestRegisterShowEscapesControlCharactersInInvalidID(t *testing.T) {
	home := t.TempDir()
	envelope := Show(home, "bad\nid")
	want := `invalid project id: "bad\nid": an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters`
	if envelope.Error != want {
		t.Fatalf("error = %v; want %q", envelope.Error, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("invalid ID caused store access: stat error = %v", err)
	}
}

func TestShowUnknownProject(t *testing.T) {
	envelope := Show(t.TempDir(), "unknown")
	if envelope.Error != "project not found: unknown" {
		t.Fatalf("error = %v; want project not found", envelope.Error)
	}
}
