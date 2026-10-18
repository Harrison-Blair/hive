package land

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	t.Run("merge without fast-forward pushes the base", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		runGit(t, repo, "push", "-q", "-u", "origin", "main")
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "demo"})
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || result.Commit == "" {
			t.Fatalf("result = %#v, want a merge commit", envelope.Result)
		}
		parents := strings.Fields(strings.TrimSpace(runGit(t, repo, "rev-list", "--parents", "-n", "1", result.Commit)))
		if len(parents) != 3 {
			t.Fatalf("merge commit parents = %v, want two parents", parents)
		}
		if got := strings.TrimSpace(runGit(t, bare, "rev-parse", "main")); got != result.Commit {
			t.Fatalf("origin main = %s, want the merge commit %s", got, result.Commit)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.Mode != "merge" || record.Delivery.Target != "main" ||
			record.Delivery.Commit != result.Commit || !record.Delivery.Pushed || record.Delivery.Grant != "demo" {
			t.Fatalf("delivery = %+v, want the merge delivery", record.Delivery)
		}
	})

	t.Run("a repository without origin lands without a push", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || result.Pushed || !strings.Contains(result.Note, "no origin") {
			t.Fatalf("result = %#v, want an unpushed merge with a note", envelope.Result)
		}
		if lines := strings.Count(strings.TrimSpace(runGit(t, repo, "log", "--oneline", "-n", "3")), "\n") + 1; lines != 3 {
			t.Fatalf("log lines = %d, want the merge, work, and init commits", lines)
		}
	})

	t.Run("a conflicting branch aborts and leaves the repository clean", func(t *testing.T) {
		repo := newRepo(t)
		runGit(t, repo, "checkout", "-q", "-b", "task/1-a", "main")
		if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("branch\n"), 0o644); err != nil {
			t.Fatalf("write branch file: %v", err)
		}
		runGit(t, repo, "add", "README.md")
		runGit(t, repo, "commit", "-q", "-m", "branch change")
		runGit(t, repo, "checkout", "-q", "main")
		if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("main\n"), 0o644); err != nil {
			t.Fatalf("write main file: %v", err)
		}
		runGit(t, repo, "add", "README.md")
		runGit(t, repo, "commit", "-q", "-m", "main change")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" {
			t.Fatalf("Land() = %+v, want failed", envelope)
		}
		message := fmt.Sprint(envelope.Error)
		if !strings.Contains(message, "merge conflict") || !strings.Contains(message, "README.md") {
			t.Fatalf("error = %q, want the conflict files", message)
		}
		if status := strings.TrimSpace(runGit(t, repo, "status", "--porcelain")); status != "" {
			t.Fatalf("status = %q, want a clean repository", status)
		}
	})

	t.Run("a bad remote gives partial and a retry pushes", func(t *testing.T) {
		repo := newRepo(t)
		runGit(t, repo, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "partial" {
			t.Fatalf("Land() = %+v, want partial", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || result.Pushed || result.Commit == "" {
			t.Fatalf("result = %#v, want an unpushed merge with a commit", envelope.Result)
		}
		if !hasEffect(envelope, "merge", "merged") {
			t.Fatalf("effects = %+v, want the merge effect", envelope.Effects)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.Commit != result.Commit || record.Delivery.Pushed {
			t.Fatalf("delivery = %+v, want the commit and pushed false", record.Delivery)
		}

		bare := newBare(t)
		runGit(t, repo, "remote", "set-url", "origin", bare)
		retry := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if retry.Outcome != "ok" {
			t.Fatalf("retry = %+v, want ok", retry)
		}
		retryResult, ok := retry.Result.(Result)
		if !ok || !retryResult.Pushed || retryResult.Commit != result.Commit {
			t.Fatalf("retry result = %#v, want the kept commit and a push", retry.Result)
		}
		record = loadTask(t, home, "1-a")
		if record.Delivery == nil || !record.Delivery.Pushed {
			t.Fatalf("delivery = %+v, want pushed true", record.Delivery)
		}
	})

	t.Run("a pre-merge hook that exits with 1 fails clean", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		hook := filepath.Join(repo, ".git", "hooks", "pre-merge-commit")
		if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
			t.Fatalf("mkdir hooks: %v", err)
		}
		if err := os.WriteFile(hook, []byte("#!/bin/sh\necho hook-says-no\nexit 1\n"), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" {
			t.Fatalf("Land() = %+v, want failed", envelope)
		}
		if !strings.Contains(fmt.Sprint(envelope.Error), "hook-says-no") {
			t.Fatalf("error = %q, want the hook output", envelope.Error)
		}
		if status := strings.TrimSpace(runGit(t, repo, "status", "--porcelain")); status != "" {
			t.Fatalf("status = %q, want a clean repository", status)
		}
	})

	t.Run("a merge past the deadline fails", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		hook := filepath.Join(repo, ".git", "hooks", "pre-merge-commit")
		if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
			t.Fatalf("mkdir hooks: %v", err)
		}
		if err := os.WriteFile(hook, []byte("#!/bin/sh\nsleep 5\n"), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		started := time.Now()
		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", MergeTimeout: 200 * time.Millisecond})
		if envelope.Outcome != "failed" {
			t.Fatalf("Land() = %+v, want failed", envelope)
		}
		if !strings.Contains(fmt.Sprint(envelope.Error), "timeout") {
			t.Fatalf("error = %q, want the timeout", envelope.Error)
		}
		if elapsed := time.Since(started); elapsed > 3*time.Second {
			t.Fatalf("Land() returned after %v, want the merge deadline", elapsed)
		}
	})

	t.Run("an expired caller deadline still cleans the repository", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		hook := filepath.Join(repo, ".git", "hooks", "pre-merge-commit")
		if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
			t.Fatalf("mkdir hooks: %v", err)
		}
		if err := os.WriteFile(hook, []byte("#!/bin/sh\nsleep 5\n"), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		envelope := Land(ctx, home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "failed" {
			t.Fatalf("Land() = %+v, want failed", envelope)
		}
		if status := strings.TrimSpace(runGit(t, repo, "status", "--porcelain")); status != "" {
			t.Fatalf("Land() = %+v, status = %q, want a clean repository", envelope, status)
		}
	})

	t.Run("a pre-push hook sees the recorded merge commit", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		runGit(t, repo, "push", "-q", "-u", "origin", "main")
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		record := filepath.Join(home, ".hive", "state", "tasks", "1-a.json")
		script := "#!/bin/sh\n" +
			"if ! grep -q '\"commit\":\"[0-9a-f]' \"" + record + "\"; then\n" +
			"  echo delivery-not-recorded-before-push >&2\n" +
			"  exit 1\n" +
			"fi\n"
		hook := filepath.Join(repo, ".git", "hooks", "pre-push")
		if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
			t.Fatalf("mkdir hooks: %v", err)
		}
		if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || !result.Pushed {
			t.Fatalf("result = %#v, want a push", envelope.Result)
		}
	})

	t.Run("a retry without citations keeps the recorded grant and evidence", func(t *testing.T) {
		repo := newRepo(t)
		runGit(t, repo, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		first := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "grant-1", Evidence: "ev-1"})
		if first.Outcome != "partial" {
			t.Fatalf("Land() = %+v, want partial", first)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.Grant != "grant-1" || record.Delivery.Evidence != "ev-1" {
			t.Fatalf("delivery = %+v, want the recorded citations", record.Delivery)
		}

		bare := newBare(t)
		runGit(t, repo, "remote", "set-url", "origin", bare)
		retry := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
		if retry.Outcome != "ok" {
			t.Fatalf("retry = %+v, want ok", retry)
		}
		record = loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.Grant != "grant-1" || record.Delivery.Evidence != "ev-1" {
			t.Fatalf("delivery = %+v, want the kept citations", record.Delivery)
		}
	})

	t.Run("a delivery record failure keeps the merge effect", func(t *testing.T) {
		repo := newRepo(t)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		huge := strings.Repeat("\x01", 100000)

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: huge, Evidence: huge})
		if envelope.Outcome != "failed" || envelope.Phase != "record" {
			t.Fatalf("Land() = %+v, want a failed record", envelope)
		}
		if !hasEffect(envelope, "merge", "merged") {
			t.Fatalf("effects = %+v, want the merge effect", envelope.Effects)
		}
	})
}
