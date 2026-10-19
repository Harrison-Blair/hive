package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Hive Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Hive Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}

func newRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runTestGit(t, repo, "init", "-q", "-b", "main")
	runTestGit(t, repo, "config", "user.name", "Hive Test")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-q", "-m", "init")
	return repo
}

func TestGitxHelpersReadRealRepository(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	exists, err := RefExists(ctx, repo, "main")
	if err != nil || !exists {
		t.Fatalf("RefExists(main) = %v, %v, want true, nil", exists, err)
	}
	exists, err = RefExists(ctx, repo, "missing")
	if err != nil || exists {
		t.Fatalf("RefExists(missing) = %v, %v, want false, nil", exists, err)
	}

	exists, err = BranchExists(ctx, repo, "main")
	if err != nil || !exists {
		t.Fatalf("BranchExists(main) = %v, %v, want true, nil", exists, err)
	}
	exists, err = BranchExists(ctx, repo, "task/missing")
	if err != nil || exists {
		t.Fatalf("BranchExists(task/missing) = %v, %v, want false, nil", exists, err)
	}

	branch, err := CurrentBranch(ctx, repo)
	if err != nil || branch != "main" {
		t.Fatalf("CurrentBranch() = %q, %v, want main, nil", branch, err)
	}

	dirty, err := IsDirty(ctx, repo)
	if err != nil || dirty {
		t.Fatalf("IsDirty() = %v, %v, want false, nil", dirty, err)
	}
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write untracked: %v", err)
	}
	dirty, err = IsDirty(ctx, repo)
	if err != nil || !dirty {
		t.Fatalf("IsDirty() after untracked = %v, %v, want true, nil", dirty, err)
	}
	if err := os.Remove(filepath.Join(repo, "untracked.txt")); err != nil {
		t.Fatalf("remove untracked: %v", err)
	}

	runTestGit(t, repo, "branch", "task/merged")
	merged, err := IsMerged(ctx, repo, "main", "task/merged")
	if err != nil || !merged {
		t.Fatalf("IsMerged(task/merged) = %v, %v, want true, nil", merged, err)
	}
	runTestGit(t, repo, "checkout", "-q", "task/merged")
	if err := os.WriteFile(filepath.Join(repo, "branch.txt"), []byte("branch\n"), 0o644); err != nil {
		t.Fatalf("write branch file: %v", err)
	}
	runTestGit(t, repo, "add", "branch.txt")
	runTestGit(t, repo, "commit", "-q", "-m", "branch work")
	runTestGit(t, repo, "checkout", "-q", "main")
	merged, err = IsMerged(ctx, repo, "main", "task/merged")
	if err != nil || merged {
		t.Fatalf("IsMerged before merge = %v, %v, want false, nil", merged, err)
	}
	runTestGit(t, repo, "merge", "-q", "--no-ff", "-m", "merge branch", "task/merged")
	merged, err = IsMerged(ctx, repo, "main", "task/merged")
	if err != nil || !merged {
		t.Fatalf("IsMerged after merge = %v, %v, want true, nil", merged, err)
	}

	linked, err := IsLinkedWorktree(ctx, repo)
	if err != nil || linked {
		t.Fatalf("IsLinkedWorktree(primary) = %v, %v, want false, nil", linked, err)
	}
	primaryCommon, err := CommonDir(ctx, repo)
	if err != nil {
		t.Fatalf("CommonDir(primary): %v", err)
	}
	if !filepath.IsAbs(primaryCommon) {
		t.Fatalf("CommonDir(primary) = %q, want an absolute path", primaryCommon)
	}
	wantCommon, err := filepath.EvalSymlinks(filepath.Join(repo, ".git"))
	if err != nil {
		t.Fatalf("resolve primary .git: %v", err)
	}
	if primaryCommon != wantCommon {
		t.Fatalf("CommonDir(primary) = %q, want %q", primaryCommon, wantCommon)
	}

	worktree := filepath.Join(t.TempDir(), "linked")
	runTestGit(t, repo, "worktree", "add", "-q", "-b", "task/foo", worktree, "main")
	linked, err = IsLinkedWorktree(ctx, worktree)
	if err != nil || !linked {
		t.Fatalf("IsLinkedWorktree(linked) = %v, %v, want true, nil", linked, err)
	}
	linkedCommon, err := CommonDir(ctx, worktree)
	if err != nil {
		t.Fatalf("CommonDir(linked): %v", err)
	}
	if linkedCommon != primaryCommon {
		t.Fatalf("CommonDir(linked) = %q, want %q", linkedCommon, primaryCommon)
	}
	branch, err = CurrentBranch(ctx, worktree)
	if err != nil || branch != "task/foo" {
		t.Fatalf("CurrentBranch(linked) = %q, %v, want task/foo, nil", branch, err)
	}

	if err := os.WriteFile(filepath.Join(worktree, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	result, err := WorktreeRemove(ctx, repo, worktree, false)
	if err != nil {
		t.Fatalf("WorktreeRemove(no force): %v", err)
	}
	if result.ExitCode == 0 {
		t.Fatalf("WorktreeRemove(no force) exit = 0, want a Git failure")
	}
	if _, err := os.Lstat(worktree); err != nil {
		t.Fatalf("dirty worktree removed without force: %v", err)
	}
	result, err = WorktreeRemove(ctx, repo, worktree, true)
	if err != nil {
		t.Fatalf("WorktreeRemove(force): %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("WorktreeRemove(force) exit = %d: %s%s", result.ExitCode, result.Stdout, result.Stderr)
	}
	if _, err := os.Lstat(worktree); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists after remove: %v", err)
	}
	exists, err = BranchExists(ctx, repo, "task/foo")
	if err != nil || !exists {
		t.Fatalf("BranchExists(task/foo) after remove = %v, %v, want true, nil", exists, err)
	}
}

func TestFetchBranchTipReadsRemoteBranch(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	origin := filepath.Join(t.TempDir(), "origin.git")
	runTestGit(t, filepath.Dir(origin), "init", "--bare", "-q", origin)
	runTestGit(t, repo, "remote", "add", "origin", origin)
	runTestGit(t, repo, "push", "-u", "origin", "main")
	runTestGit(t, repo, "update-ref", "-d", "refs/remotes/origin/main")
	want := strings.TrimSpace(runTestGit(t, repo, "rev-parse", "HEAD"))

	got, exists, err := FetchBranchTip(ctx, repo, "origin", "main")
	if err != nil || !exists || got != want {
		t.Fatalf("FetchBranchTip(main) = %q, %v, %v; want %q, true, nil", got, exists, err, want)
	}

	got, exists, err = FetchBranchTip(ctx, repo, "origin", "missing")
	if err != nil || exists || got != "" {
		t.Fatalf("FetchBranchTip(missing) = %q, %v, %v; want empty, false, nil", got, exists, err)
	}
}

func TestGitxCommonDirThroughSymlink(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	link := filepath.Join(t.TempDir(), "repo-link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	common, err := CommonDir(ctx, link)
	if err != nil {
		t.Fatalf("CommonDir(symlink): %v", err)
	}
	want, err := filepath.EvalSymlinks(filepath.Join(repo, ".git"))
	if err != nil {
		t.Fatalf("resolve .git: %v", err)
	}
	if common != want {
		t.Fatalf("CommonDir(symlink) = %q, want %q", common, want)
	}
}
