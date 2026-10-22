package land

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func seedOne(t *testing.T, repo string) (string, records.Task) {
	t.Helper()
	home := newHome(t)
	writeProject(t, home, testProject(repo))
	task := testTask("1-a", "task/1-a")
	writeTask(t, home, task)
	return home, task
}

func TestPreconditions(t *testing.T) {
	t.Run("a task that is not verified", func(t *testing.T) {
		repo := newRepo(t)
		home, task := seedOne(t, repo)
		task.Status = "created"
		writeTask(t, home, task)
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task is created: land needs verified" {
			t.Fatalf("Land() = %+v, want failed with the status message", envelope)
		}
	})

	t.Run("a task without a branch", func(t *testing.T) {
		repo := newRepo(t)
		home, task := seedOne(t, repo)
		task.Worktree = nil
		writeTask(t, home, task)
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task has no worktree branch" {
			t.Fatalf("Land() = %+v, want failed with the branch message", envelope)
		}
	})

	t.Run("a base that is not the current branch", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home, _ := seedOne(t, repo)
		runGit(t, repo, "checkout", "-q", "-b", "other")
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "base branch not checked out: main" {
			t.Fatalf("Land() = %+v, want failed with the base message", envelope)
		}
	})

	t.Run("a dirty repository", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home, _ := seedOne(t, repo)
		if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("dirty\n"), 0o644); err != nil {
			t.Fatalf("write dirty file: %v", err)
		}
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "repository is dirty" {
			t.Fatalf("Land() = %+v, want failed with the dirty message", envelope)
		}
	})

	t.Run("a retry with another mode", func(t *testing.T) {
		repo := newRepo(t)
		home, task := seedOne(t, repo)
		task.Delivery = &records.Delivery{Mode: "merge", Commit: "deadbeef"}
		writeTask(t, home, task)
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if envelope.Outcome != "failed" || envelope.Error != "task delivered in mode merge" {
			t.Fatalf("Land() = %+v, want failed with the delivered-mode message", envelope)
		}
	})

	t.Run("two landings on one repository run one after the other", func(t *testing.T) {
		repo := newRepo(t)
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		addTaskBranch(t, repo, "task/2-b", "b.txt", "b\n")
		writeTask(t, home, testTask("1-a", "task/1-a"))
		writeTask(t, home, testTask("2-b", "task/2-b"))

		envelopes := make(chan output.Envelope, 2)
		var group sync.WaitGroup
		for _, id := range []string{"1-a", "2-b"} {
			group.Add(1)
			go func(taskID string) {
				defer group.Done()
				envelopes <- Land(context.Background(), home, Options{Task: taskID, Mode: "merge", Grant: "test"})
			}(id)
		}
		group.Wait()
		close(envelopes)

		commits := map[string]string{}
		for envelope := range envelopes {
			if envelope.Outcome != "ok" {
				t.Fatalf("Land() = %+v, want ok", envelope)
			}
			result, ok := envelope.Result.(Result)
			if !ok {
				t.Fatalf("result = %#v, want a land result", envelope.Result)
			}
			commits[result.Task.ID] = result.Commit
		}
		if commits["1-a"] == "" || commits["2-b"] == "" || commits["1-a"] == commits["2-b"] {
			t.Fatalf("commits = %v, want one distinct commit for each task", commits)
		}
		for _, id := range []string{"1-a", "2-b"} {
			record := loadTask(t, home, id)
			if record.Delivery == nil || record.Delivery.Commit != commits[id] {
				t.Fatalf("record %s delivery = %+v, want commit %s", id, record.Delivery, commits[id])
			}
		}
	})
}

func TestReopenedTaskRequiresFreshPassVerificationBeforeLand(t *testing.T) {
	t.Run("assigned before a new verification", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home, task := seedOne(t, repo)
		task.Status = "assigned"
		task.Verification = nil
		writeTask(t, home, task)

		envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task is assigned: land needs verified" {
			t.Fatalf("Land() = %+v; want the assigned status error", envelope)
		}
		task.Status = "completed"
		writeTask(t, home, task)
		envelope = Land(context.Background(), home, Options{Task: task.ID, Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task is completed: land needs verified" {
			t.Fatalf("Land() after completion without a new verification = %+v; want the completed status error", envelope)
		}

		task.Status = "verified"
		task.Verification = &records.Verification{Agent: "verifier-new", Result: "pass", At: "2026-10-19T00:00:00Z"}
		writeTask(t, home, task)
		envelope = Land(context.Background(), home, Options{Task: task.ID, Mode: "merge"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Land() after a new pass = %+v; want ok", envelope)
		}
	})

	t.Run("verified without a check", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home, task := seedOne(t, repo)
		task.Verification = nil
		writeTask(t, home, task)

		envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task verification is not a pass" {
			t.Fatalf("Land() = %+v; want a missing verification error", envelope)
		}
	})

	t.Run("verified after a failed check", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home, task := seedOne(t, repo)
		task.Verification.Result = "fail"
		writeTask(t, home, task)

		envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge"})
		if envelope.Outcome != "failed" || envelope.Error != "task verification is not a pass" {
			t.Fatalf("Land() = %+v; want a failed verification error", envelope)
		}
	})
}
