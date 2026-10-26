package show

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

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
	envelope := Run(home, want.ID)
	if envelope.Outcome != "ok" || envelope.Result != want {
		t.Fatalf("Show() = %#v; want %#v", envelope, want)
	}
}

func TestRegisterShowRejectsInvalidIDBeforeStoreAccess(t *testing.T) {
	home := t.TempDir()
	envelope := Run(home, "bad_id")
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
	envelope := Run(home, "bad\nid")
	want := `invalid project id: "bad\nid": an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters`
	if envelope.Error != want {
		t.Fatalf("error = %v; want %q", envelope.Error, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("invalid ID caused store access: stat error = %v", err)
	}
}

func TestShowUnknownProject(t *testing.T) {
	envelope := Run(t.TempDir(), "unknown")
	if envelope.Error != "project not found: unknown" {
		t.Fatalf("error = %v; want project not found", envelope.Error)
	}
}
