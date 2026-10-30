package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesTheHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "new", "home")
	openStore(t, home)
	for _, name := range []string{"schema.json"} {
		if info, err := os.Stat(filepath.Join(home, name)); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("%s: stat = %v, %v; want a file at the top of the home", name, info, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".hive")); !os.IsNotExist(err) {
		t.Fatalf(".hive: stat error = %v; want no .hive level in the home", err)
	}
	if _, err := os.Stat(filepath.Join(home, "state")); !os.IsNotExist(err) {
		t.Fatalf("state: stat error = %v; want no state directory in the home", err)
	}
}

func TestOpenAcceptsLinkedHome(t *testing.T) {
	target := t.TempDir()
	home := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(target, home); err != nil {
		t.Fatal(err)
	}
	openStore(t, home)
	if _, err := os.Stat(filepath.Join(target, "schema.json")); err != nil {
		t.Fatalf("schema in the link target: %v", err)
	}
}

func TestOpenRejectsWrongTypeHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(home, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if store, err := Open(home); err == nil {
		t.Fatalf("Open() = %v, nil; want an error for a regular file home", store)
	}
	data, err := os.ReadFile(home)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "not a directory" {
		t.Fatalf("Open() changed the wrong-type home to %q", data)
	}
}
