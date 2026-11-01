package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveInDeletesTheRecordAndItsEmptyDirectory(t *testing.T) {
	home := t.TempDir()
	store := openStore(t, home)
	if err := store.WriteIn([]string{"worktrees", "k1"}, "worktree", map[string]string{"id": "k1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveIn([]string{"worktrees", "k1"}, "worktree"); err != nil {
		t.Fatalf("RemoveIn error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "worktrees", "k1")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("directory stat = %v; want it removed", err)
	}
	if err := store.RemoveIn([]string{"worktrees", "k1"}, "worktree"); err != nil {
		t.Fatalf("RemoveIn of a missing record error = %v; want nil", err)
	}
}

func TestRemoveInKeepsADirectoryWithOtherFiles(t *testing.T) {
	home := t.TempDir()
	store := openStore(t, home)
	if err := store.WriteIn([]string{"worktrees", "k1"}, "worktree", map[string]string{"id": "k1"}); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(home, "worktrees", "k1", "checkout")
	if err := os.Mkdir(other, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveIn([]string{"worktrees", "k1"}, "worktree"); err != nil {
		t.Fatalf("RemoveIn error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "worktrees", "k1", "worktree.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("record stat = %v; want it removed", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("other entry stat = %v; want it kept", err)
	}
}

func TestRemoveInRejectsASymlinkedCollection(t *testing.T) {
	home := t.TempDir()
	store := openStore(t, home)
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "worktree.json"), []byte(`{"id":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(home, "worktrees")); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveIn([]string{"worktrees"}, "worktree"); err == nil {
		t.Fatal("RemoveIn error = nil; want store path invalid")
	}
	if _, err := os.Stat(filepath.Join(target, "worktree.json")); err != nil {
		t.Fatalf("target record stat = %v; want it kept", err)
	}
}
