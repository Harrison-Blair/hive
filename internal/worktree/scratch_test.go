package worktree

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func scratch(t *testing.T, home, dir string) (string, bool, error) {
	t.Helper()
	return Scratch(context.Background(), home, dir, discard())
}

// TestScratchOfALinkedCheckoutUsesTheRepositoryName checks that a linked
// checkout gets the scratch place of its repository, named after the
// directory of the repository and not after the checkout directory.
func TestScratchOfALinkedCheckoutUsesTheRepositoryName(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "My Repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, "-C", repo, "init", "-q", "-b", "main")
	git(t, "-C", repo, "commit", "-q", "--allow-empty", "-m", "init")
	checkout := filepath.Join(t.TempDir(), "other-name")
	git(t, "-C", repo, "worktree", "add", "-q", "-b", "side", checkout)
	home := t.TempDir()

	path, created, err := scratch(t, home, checkout)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "tmp", "my-repo")
	if path != want || !created {
		t.Fatalf("scratch = %q, created %v; want %q, created true", path, created, want)
	}
	if info, err := os.Stat(want); err != nil || !info.IsDir() {
		t.Fatalf("scratch place %s: %v; want a directory", want, err)
	}
	path, created, err = scratch(t, home, repo)
	if err != nil || path != want || created {
		t.Fatalf("second scratch = %q, created %v, %v; want %q, created false", path, created, err, want)
	}
}

// TestScratchOutsideGitUsesTheDirectoryName checks that a directory outside a
// Git repository gives the slug of its own base name.
func TestScratchOutsideGitUsesTheDirectoryName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Plain Dir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	path, created, err := scratch(t, home, dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "tmp", "plain-dir"); path != want || !created {
		t.Fatalf("scratch = %q, created %v; want %q, created true", path, created, want)
	}
}

// TestScratchFailsWhenItCannotMakeThePlace checks the error when the tmp
// directory of the home is a file.
func TestScratchFailsWhenItCannotMakeThePlace(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "tmp"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "plain")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if path, _, err := scratch(t, home, dir); err == nil {
		t.Fatalf("scratch = %q; want an error", path)
	}
}
