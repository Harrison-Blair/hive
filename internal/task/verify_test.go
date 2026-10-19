package task

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func runVerifyGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Hive Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Hive Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}

func newVerifyGitRepo(t *testing.T, push, ahead, dirty, removeTracking bool) string {
	t.Helper()
	root := t.TempDir()
	repo := filepath.Join(root, "worktree")
	origin := filepath.Join(root, "origin.git")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runVerifyGit(t, root, "init", "--bare", "-q", origin)
	runVerifyGit(t, repo, "init", "-q", "-b", "task/verify")
	runVerifyGit(t, repo, "config", "user.name", "Hive Test")
	runVerifyGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runVerifyGit(t, repo, "add", "README.md")
	runVerifyGit(t, repo, "commit", "-q", "-m", "initial")
	runVerifyGit(t, repo, "remote", "add", "origin", origin)
	if push {
		runVerifyGit(t, repo, "push", "-u", "origin", "task/verify")
	}
	if ahead {
		if err := os.WriteFile(filepath.Join(repo, "ahead.txt"), []byte("ahead\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		runVerifyGit(t, repo, "add", "ahead.txt")
		runVerifyGit(t, repo, "commit", "-q", "-m", "ahead")
	}
	if dirty {
		if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("dirty\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if removeTracking {
		runVerifyGit(t, repo, "update-ref", "-d", "refs/remotes/origin/task/verify")
	}
	return repo
}

func snapshotTaskState(t *testing.T, home string) map[string][]byte {
	t.Helper()
	root := filepath.Join(home, ".hive", "state")
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relative] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func verifyWithWorktree(t *testing.T, repo string) (string, records.Task) {
	t.Helper()
	home, task := newSeed(t, "completed")
	task.Worktree = &records.Worktree{Path: repo, Branch: "task/verify"}
	seedTask(t, home, task)
	return home, task
}

func TestVerifyRefusesDirtyWorktreeWithoutChanges(t *testing.T) {
	repo := newVerifyGitRepo(t, true, false, true, false)
	home, task := verifyWithWorktree(t, repo)
	before := snapshotTaskState(t, home)

	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "failed" || envelope.Error != "worktree is dirty: commit and push before verify" {
		t.Fatalf("Verify() = %#v; want dirty worktree refusal", envelope)
	}
	if after := snapshotTaskState(t, home); !reflect.DeepEqual(after, before) {
		t.Fatalf("task state changed on dirty worktree refusal")
	}
}

func TestVerifyRefusesBranchWithoutRemoteWithoutChanges(t *testing.T) {
	repo := newVerifyGitRepo(t, false, false, false, false)
	home, task := verifyWithWorktree(t, repo)
	before := snapshotTaskState(t, home)

	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "failed" || envelope.Error != "branch not pushed: task/verify" {
		t.Fatalf("Verify() = %#v; want missing remote branch refusal", envelope)
	}
	if after := snapshotTaskState(t, home); !reflect.DeepEqual(after, before) {
		t.Fatalf("task state changed on missing remote branch refusal")
	}
}

func TestVerifyRefusesBranchAheadOfOriginWithoutChanges(t *testing.T) {
	repo := newVerifyGitRepo(t, true, true, false, false)
	home, task := verifyWithWorktree(t, repo)
	before := snapshotTaskState(t, home)

	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "failed" || envelope.Error != "branch not pushed: task/verify has commits not on origin" {
		t.Fatalf("Verify() = %#v; want unpushed commit refusal", envelope)
	}
	if after := snapshotTaskState(t, home); !reflect.DeepEqual(after, before) {
		t.Fatalf("task state changed on unpushed commit refusal")
	}
}

func TestVerifyAcceptsCleanPushedWorktree(t *testing.T) {
	repo := newVerifyGitRepo(t, true, false, false, true)
	home, task := verifyWithWorktree(t, repo)
	remoteTip := runVerifyGit(t, repo, "rev-parse", "HEAD")

	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Verify() = %#v; want clean pushed branch to pass", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "verified" || got.Verification == nil || got.Verification.Result != "pass" {
		t.Fatalf("task = %#v; want verified pass", got)
	}
	if got := runVerifyGit(t, repo, "rev-parse", "refs/remotes/origin/task/verify"); got != remoteTip {
		t.Fatalf("fetched origin tip = %q; want %q", got, remoteTip)
	}
	if _, err := gitx.Tip(context.Background(), repo, "FETCH_HEAD"); err != nil {
		t.Fatalf("verify did not fetch origin branch: %v", err)
	}
}
