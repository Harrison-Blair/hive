package complete

import (
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// AC4: task complete copies the evidence folder of the worktree to the home.
// It copies no other file of the feature folder.
func TestCompleteCopiesEvidenceToHome(t *testing.T) {
	home, task := tasktest.NewSeed(t, "assigned")
	worktree := t.TempDir()
	task.Feature = "f"
	task.Worktree = &records.Worktree{Path: worktree, Branch: "task/" + task.ID}
	tasktest.SeedTask(t, home, task)
	evidence := ".hive/spec/f/tasks/1-seed/evidence/"
	tasktest.WriteFile(t, filepath.Join(worktree, filepath.FromSlash(evidence), "V1.md"), "red and green\n")
	tasktest.WriteFile(t, filepath.Join(worktree, filepath.FromSlash(evidence), "screens", "chat.png"), "png\n")
	tasktest.WriteFile(t, filepath.Join(worktree, ".hive", "spec", "f", "README.md"), "worktree README\n")

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	for name, want := range map[string]string{"V1.md": "red and green\n", "screens/chat.png": "png\n"} {
		relative := evidence + name
		if got, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != want {
			t.Errorf("home %s = %q, %v; want %q", relative, got, ok, want)
		}
		if !tasktest.HasTaskEffect(envelope, "evidence", "copied", relative) {
			t.Errorf("effects = %+v; want copied %s", envelope.Effects, relative)
		}
	}
	if _, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, ".hive", "spec", "f", "README.md")); ok {
		t.Fatalf("complete copied the feature README; want the evidence folder only")
	}
}

// AC4: a failed copy-out keeps the transition and reports a partial result.
func TestCompleteEvidenceCopyFailureIsPartial(t *testing.T) {
	home, task := tasktest.NewSeed(t, "assigned")
	worktree := t.TempDir()
	task.Feature = "f"
	task.Worktree = &records.Worktree{Path: worktree, Branch: "task/" + task.ID}
	tasktest.SeedTask(t, home, task)
	tasktest.WriteFile(t, filepath.Join(worktree, ".hive", "spec", "f", "tasks", "1-seed", "evidence", "V1.md"), "run\n")
	// A file at the home folder path blocks the copy.
	tasktest.WriteFile(t, filepath.Join(home, ".hive", "spec", "f", "tasks", "1-seed"), "not a folder")

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "partial" || envelope.Phase != "evidence" {
		t.Fatalf("Complete() = %#v; want partial evidence", envelope)
	}
	if got := tasktest.ReadTask(t, home, task.ID); got.Status != "completed" {
		t.Fatalf("status = %q; want completed", got.Status)
	}
}
