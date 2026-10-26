package verify

import (
	"path/filepath"
	"testing"

	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// AC4: task verify copies VERIFIED.md to the home. An ignored evidence file
// does not make the worktree dirty.
func TestVerifyCopiesVerifiedToHome(t *testing.T) {
	repo := newVerifyGitRepo(t, true, false, false, true)
	tasktest.WriteFile(t, filepath.Join(repo, ".gitignore"), ".hive/spec/\n")
	runVerifyGit(t, repo, "add", ".gitignore")
	runVerifyGit(t, repo, "commit", "-q", "-m", "ignore the spec")
	runVerifyGit(t, repo, "push", "-q", "origin", "task/verify")
	home, task := verifyWithWorktree(t, repo)
	task.Feature = "f"
	tasktest.SeedTask(t, home, task)
	evidence := ".hive/spec/f/tasks/1-seed/evidence/"
	tasktest.WriteFile(t, filepath.Join(repo, filepath.FromSlash(evidence), "VERIFIED.md"), "pass\n")
	tasktest.WriteFile(t, filepath.Join(repo, filepath.FromSlash(evidence), "V1.md"), "worker run\n")

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Verify() = %#v; want ok with ignored evidence files", envelope)
	}
	relative := evidence + "VERIFIED.md"
	if got, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != "pass\n" {
		t.Fatalf("home VERIFIED.md = %q, %v; want the verifier file", got, ok)
	}
	if !tasktest.HasTaskEffect(envelope, "evidence", "copied", relative) {
		t.Fatalf("effects = %+v; want copied %s", envelope.Effects, relative)
	}
	if _, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(evidence), "V1.md")); ok {
		t.Fatalf("verify copied V1.md; want VERIFIED.md only")
	}
}
