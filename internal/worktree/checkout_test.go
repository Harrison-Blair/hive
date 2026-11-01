package worktree

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func git(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func ensure(t *testing.T, home, dir, branch, base string) (Checkout, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return Ensure(ctx, home, dir, branch, base, discard())
}

// linked makes a linked worktree of repo on branch side with one commit more
// than main, and returns its path and that commit.
func linked(t *testing.T, repo string) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "side")
	git(t, "-C", repo, "worktree", "add", "-q", "-b", "side", path)
	git(t, "-C", path, "commit", "-q", "--allow-empty", "-m", "side")
	return path, git(t, "-C", path, "rev-parse", "HEAD")
}

func TestEnsureCreatesANewBranchAtTheCurrentCommitOfTheDirectory(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	side, commit := linked(t, repo)
	checkout, err := ensure(t, home, side, "feat/x", "")
	if err != nil {
		t.Fatalf("Ensure error = %v", err)
	}
	if !checkout.Created || !checkout.Owned || checkout.Branch != "feat/x" {
		t.Fatalf("checkout = %+v; want created, owned, branch feat/x", checkout)
	}
	if !strings.HasPrefix(checkout.Path, filepath.Join(home, "worktrees")+string(filepath.Separator)) {
		t.Fatalf("path = %s; want it below %s/worktrees", checkout.Path, home)
	}
	if got := git(t, "-C", checkout.Path, "rev-parse", "HEAD"); got != commit {
		t.Fatalf("HEAD = %s; want the commit %s of the caller directory", got, commit)
	}
	if got := git(t, "-C", checkout.Path, "branch", "--show-current"); got != "feat/x" {
		t.Fatalf("branch = %s; want feat/x", got)
	}
}

func TestEnsureCreatesANewBranchAtTheBase(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	side, _ := linked(t, repo)
	checkout, err := ensure(t, home, side, "feat", "main")
	if err != nil {
		t.Fatalf("Ensure error = %v", err)
	}
	if got, want := git(t, "-C", checkout.Path, "rev-parse", "HEAD"), git(t, "-C", repo, "rev-parse", "main"); got != want {
		t.Fatalf("HEAD = %s; want main %s", got, want)
	}
}

func TestEnsureReturnsTheExistingCheckoutOfABranch(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	side, _ := linked(t, repo)
	for _, test := range []struct{ branch, path string }{{"main", repo}, {"side", side}} {
		checkout, err := ensure(t, home, side, test.branch, "")
		if err != nil {
			t.Fatalf("Ensure(%s) error = %v", test.branch, err)
		}
		if checkout.Created || checkout.Owned || checkout.Path != test.path {
			t.Fatalf("Ensure(%s) = %+v; want the existing checkout %s, not created, not owned", test.branch, checkout, test.path)
		}
	}
	first, err := ensure(t, home, repo, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	again, err := ensure(t, home, side, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || !again.Owned || again.Path != first.Path {
		t.Fatalf("second Ensure = %+v; want the owned checkout %s, not created", again, first.Path)
	}
}

func TestEnsureChecksOutAnExistingBranchWithoutACheckout(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	side, commit := linked(t, repo)
	git(t, "-C", side, "branch", "kept")
	checkout, err := ensure(t, home, repo, "kept", "")
	if err != nil {
		t.Fatalf("Ensure error = %v", err)
	}
	if !checkout.Created || !checkout.Owned {
		t.Fatalf("checkout = %+v; want created and owned", checkout)
	}
	if got := git(t, "-C", checkout.Path, "rev-parse", "HEAD"); got != commit {
		t.Fatalf("HEAD = %s; want the tip %s of kept", got, commit)
	}
	if got := git(t, "-C", checkout.Path, "branch", "--show-current"); got != "kept" {
		t.Fatalf("branch = %s; want kept", got)
	}
}

func TestEnsureRejectsABaseForAnExistingBranch(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	git(t, "-C", repo, "branch", "kept")
	for _, branch := range []string{"main", "kept"} {
		if _, err := ensure(t, home, repo, branch, "main"); err == nil || !strings.Contains(err.Error(), "--base") {
			t.Fatalf("Ensure(%s, base main) error = %v; want a --base error", branch, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, "worktrees")); !os.IsNotExist(err) {
		t.Fatalf("worktrees stat = %v; want nothing made", err)
	}
}

func TestEnsureRejectsABadBranchAndADirectoryOutsideGit(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	if _, err := ensure(t, home, repo, "a..b", ""); err == nil || !strings.Contains(err.Error(), "a..b") {
		t.Fatalf("Ensure(a..b) error = %v; want an error that names the branch", err)
	}
	if _, err := ensure(t, home, t.TempDir(), "feat", ""); err == nil {
		t.Fatal("Ensure outside Git error = nil; want an error")
	}
	for _, base := range []string{"no-such-ref", "--force"} {
		if _, err := ensure(t, home, repo, "feat", base); err == nil {
			t.Fatalf("Ensure with base %s error = nil; want an error", base)
		}
	}
	if _, err := os.Stat(filepath.Join(home, "worktrees")); err == nil {
		entries, _ := os.ReadDir(filepath.Join(home, "worktrees"))
		if len(entries) != 0 {
			t.Fatalf("worktrees = %v; want no record left after the failures", entries)
		}
	}
}

// TestEnsureGivesEachRepositoryAndBranchItsOwnPath proves that two
// repositories with the same base name, and two branches with the same slug,
// get different checkouts.
func TestEnsureGivesEachRepositoryAndBranchItsOwnPath(t *testing.T) {
	home := agenttest.Home(t)
	one := filepath.Join(t.TempDir(), "app")
	two := filepath.Join(t.TempDir(), "app")
	for _, repo := range []string{one, two} {
		git(t, "init", "-q", "-b", "main", repo)
		git(t, "-C", repo, "commit", "-q", "--allow-empty", "-m", "init")
	}
	paths := map[string]bool{}
	for _, test := range []struct{ repo, branch string }{{one, "a/b"}, {one, "a-b"}, {two, "a/b"}} {
		checkout, err := ensure(t, home, test.repo, test.branch, "")
		if err != nil {
			t.Fatalf("Ensure(%s, %s) error = %v", test.repo, test.branch, err)
		}
		if paths[checkout.Path] {
			t.Fatalf("path %s given twice", checkout.Path)
		}
		paths[checkout.Path] = true
	}
}

// TestEnsureDoesNotOwnACheckoutWithoutItsRecord proves that ownership comes
// from the record that Ensure writes, not from the path below the home.
func TestEnsureDoesNotOwnACheckoutWithoutItsRecord(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	path := filepath.Join(home, "worktrees", "by-hand", "feat")
	git(t, "-C", repo, "worktree", "add", "-q", "-b", "feat", path)
	checkout, err := ensure(t, home, repo, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	if checkout.Owned || checkout.Path != path {
		t.Fatalf("checkout = %+v; want the checkout %s, not owned", checkout, path)
	}
}

// TestEnsureDoesNotOwnACheckoutAtAnotherPath proves that a record of a
// checkout that was removed by hand does not own a later checkout of the same
// branch at another path.
func TestEnsureDoesNotOwnACheckoutAtAnotherPath(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	first, err := ensure(t, home, repo, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	git(t, "-C", repo, "worktree", "remove", first.Path)
	other := filepath.Join(t.TempDir(), "feat")
	git(t, "-C", repo, "worktree", "add", "-q", other, "feat")
	checkout, err := ensure(t, home, repo, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	if checkout.Owned || checkout.Path != other {
		t.Fatalf("checkout = %+v; want %s, not owned", checkout, other)
	}
}

// TestEnsureAcceptsARepositoryNameThatStartsWithAnUnderscore proves that the
// key of the checkout is a valid store name for each repository name.
func TestEnsureAcceptsARepositoryNameThatStartsWithAnUnderscore(t *testing.T) {
	home := agenttest.Home(t)
	for _, name := range []string{"_app", "__", "-_x"} {
		repo := filepath.Join(t.TempDir(), name)
		git(t, "init", "-q", "-b", "main", repo)
		git(t, "-C", repo, "commit", "-q", "--allow-empty", "-m", "init")
		checkout, err := ensure(t, home, repo, "feat", "")
		if err != nil {
			t.Fatalf("Ensure in %s error = %v", name, err)
		}
		if !checkout.Owned || !checkout.Created {
			t.Fatalf("Ensure in %s = %+v; want created and owned", name, checkout)
		}
	}
}

// failHook makes the post-checkout hook of repo fail. Linked checkouts use
// the hooks of the common Git directory.
func failHook(t *testing.T, repo string) {
	t.Helper()
	hook := filepath.Join(repo, ".git", "hooks", "post-checkout")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestEnsureKeepsTheOwnershipOfACheckoutThatAFailedHookLeft proves that when
// git worktree add fails after it made the checkout, as with a failing
// post-checkout hook, Ensure reports the checkout and the new branch, and
// keeps the ownership record, so worktree remove can still remove it.
func TestEnsureKeepsTheOwnershipOfACheckoutThatAFailedHookLeft(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	failHook(t, repo)
	checkout, err := ensure(t, home, repo, "feat", "")
	if err == nil {
		t.Fatal("Ensure error = nil; want the hook failure")
	}
	if !checkout.Created || !checkout.Owned || !checkout.BranchCreated || checkout.Path == "" {
		t.Fatalf("checkout = %+v; want created, owned, branch created, with a path", checkout)
	}
	if got := git(t, "-C", checkout.Path, "branch", "--show-current"); got != "feat" {
		t.Fatalf("branch of %s = %s; want feat", checkout.Path, got)
	}
	found, ok, err := Lookup(context.Background(), home, repo, "feat", discard())
	if err != nil || !ok || !found.Owned || found.Path != checkout.Path {
		t.Fatalf("Lookup = %+v, %v, %v; want the owned checkout %s", found, ok, err, checkout.Path)
	}
}

// TestEnsureReportsABranchThatAFailedAddLeft proves that when git worktree
// add -b makes the branch and then fails before the checkout, Ensure reports
// the new branch and removes the ownership record.
func TestEnsureReportsABranchThatAFailedAddLeft(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	opened, err := openRepository(context.Background(), discard(), repo)
	if err != nil {
		t.Fatal(err)
	}
	key := keyOf(opened, "feat")
	// A file at the path of the checkout makes git worktree add fail.
	path := filepath.Join(home, Dir, key, leafOf("feat"))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "x"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	checkout, err := ensure(t, home, repo, "feat", "")
	if err == nil {
		t.Fatal("Ensure error = nil; want the add failure")
	}
	if checkout.Created || checkout.Owned || !checkout.BranchCreated {
		t.Fatalf("checkout = %+v; want not created, not owned, branch created", checkout)
	}
	if _, err := os.Stat(filepath.Join(home, Dir, key, "worktree.json")); !os.IsNotExist(err) {
		t.Fatalf("record stat = %v; want the record removed", err)
	}
}

// TestStaleGivesOnlyARecordWithoutItsCheckoutAndPath proves that Stale
// gives the record of a branch only when no checkout has the branch and
// nothing is at the path of the record.
func TestStaleGivesOnlyARecordWithoutItsCheckoutAndPath(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	checkout, err := ensure(t, home, repo, "feat", "")
	if err != nil {
		t.Fatal(err)
	}
	stale := func() bool {
		t.Helper()
		_, found, err := Stale(context.Background(), home, repo, "feat", discard())
		if err != nil {
			t.Fatal(err)
		}
		return found
	}
	if stale() {
		t.Fatal("Stale with the checkout = found; want not found")
	}
	// git worktree remove --force keeps the record, as a failed remove does.
	git(t, "-C", repo, "worktree", "remove", "--force", checkout.Path)
	if err := os.MkdirAll(checkout.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	if stale() {
		t.Fatal("Stale with a directory at the path = found; want not found")
	}
	if err := os.Remove(checkout.Path); err != nil {
		t.Fatal(err)
	}
	got, found, err := Stale(context.Background(), home, repo, "feat", discard())
	if err != nil || !found || got.Path != checkout.Path || !got.Owned {
		t.Fatalf("Stale = %+v, %v, %v; want the owned record of %s", got, found, err, checkout.Path)
	}
	if _, found, _ := Stale(context.Background(), home, repo, "other", discard()); found {
		t.Fatal("Stale(other) = found; want not found")
	}
}
