package land

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func installFakeGH(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	callsFile := filepath.Join(dir, "calls.txt")
	bodyFile := filepath.Join(dir, "body.txt")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$GH_CALLS_FILE"
prev=""
for arg in "$@"; do
  if [ "$prev" = "--body-file" ]; then
    cat "$arg" > "$GH_BODY_FILE"
  fi
  prev="$arg"
done
if [ "$1" = "pr" ] && [ "$2" = "view" ]; then
  printf '%s\n' "$GH_VIEW_OUTPUT"
  exit "${GH_VIEW_EXIT:-0}"
fi
printf '%s\n' "$GH_CREATE_OUTPUT"
exit "${GH_CREATE_EXIT:-0}"
`
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatalf("write gh: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_CALLS_FILE", callsFile)
	t.Setenv("GH_BODY_FILE", bodyFile)
	return callsFile, bodyFile
}

func assertNoBodyFiles(t *testing.T, home string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(home, ".hive", "tmp", "land-*"))
	if err != nil {
		t.Fatalf("glob body files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("body files = %v, want the file removed", matches)
	}
}

func TestPR(t *testing.T) {
	t.Run("pushes, calls gh with the body, and records the URL", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		tip := addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		writeBrief(t, home, "1-a", "land demo\n\nSecond paragraph.\n")
		callsFile, bodyFile := installFakeGH(t)
		t.Setenv("GH_CREATE_OUTPUT", "https://example.com/pr/create")
		t.Setenv("GH_VIEW_OUTPUT", `{"url":"https://example.com/pr/view"}`)

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr", Evidence: "https://evidence/1", Draft: true})
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || result.PRURL != "https://example.com/pr/create" || !result.Pushed {
			t.Fatalf("result = %#v, want the URL and a push", envelope.Result)
		}
		if got := strings.TrimSpace(runGit(t, bare, "rev-parse", "task/1-a")); got != tip {
			t.Fatalf("origin branch = %s, want %s", got, tip)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.PRURL != "https://example.com/pr/create" || !record.Delivery.Pushed ||
			record.Delivery.Commit != tip || record.Delivery.Evidence != "https://evidence/1" {
			t.Fatalf("delivery = %+v, want the pull request delivery", record.Delivery)
		}
		body, err := os.ReadFile(bodyFile)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		want := "land demo\n\nEvidence: https://evidence/1"
		if string(body) != want {
			t.Fatalf("body = %q, want %q", body, want)
		}
		calls, err := os.ReadFile(callsFile)
		if err != nil {
			t.Fatalf("read calls: %v", err)
		}
		for _, part := range []string{"pr create", "--base main", "--head task/1-a", "--title task 1-a", "--draft"} {
			if !strings.Contains(string(calls), part) {
				t.Fatalf("gh args = %q, want %q", calls, part)
			}
		}
		assertNoBodyFiles(t, home)
	})

	t.Run("a gh failure gives partial with the push as an effect", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		writeBrief(t, home, "1-a", "land demo\n")
		installFakeGH(t)
		t.Setenv("GH_CREATE_OUTPUT", "gh create failed")
		t.Setenv("GH_CREATE_EXIT", "1")

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if envelope.Outcome != "partial" || envelope.Phase != "pull request" {
			t.Fatalf("Land() = %+v, want partial pull request", envelope)
		}
		if !hasEffect(envelope, "push", "pushed") {
			t.Fatalf("effects = %+v, want the push effect", envelope.Effects)
		}
		result, ok := envelope.Result.(Result)
		if !ok || !result.Pushed || result.PRURL != "" {
			t.Fatalf("result = %#v, want a push and no URL", envelope.Result)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || !record.Delivery.Pushed || record.Delivery.PRURL != "" {
			t.Fatalf("delivery = %+v, want pushed true and no URL", record.Delivery)
		}
		assertNoBodyFiles(t, home)
	})

	t.Run("a retry reads an existing pull request", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		tip := addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		task := testTask("1-a", "task/1-a")
		task.Delivery = &records.Delivery{Mode: "pr", Target: "main", Commit: tip, Pushed: true}
		writeTask(t, home, task)
		callsFile, _ := installFakeGH(t)
		t.Setenv("GH_VIEW_OUTPUT", `{"url":"https://example.com/pr/view"}`)
		t.Setenv("GH_CREATE_OUTPUT", "https://example.com/pr/create")

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		result, ok := envelope.Result.(Result)
		if !ok || result.PRURL != "https://example.com/pr/view" {
			t.Fatalf("result = %#v, want the viewed URL", envelope.Result)
		}
		calls, err := os.ReadFile(callsFile)
		if err != nil {
			t.Fatalf("read calls: %v", err)
		}
		if !strings.Contains(string(calls), "pr view") {
			t.Fatalf("gh args = %q, want a pr view", calls)
		}
		if strings.Contains(string(calls), "pr create") {
			t.Fatalf("gh args = %q, want no pr create", calls)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.PRURL != "https://example.com/pr/view" {
			t.Fatalf("delivery = %+v, want the viewed URL", record.Delivery)
		}
		assertNoBodyFiles(t, home)
	})

	t.Run("the body holds the first paragraph", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		writeBrief(t, home, "1-a", "first paragraph line one\nfirst paragraph line two\n\nsecond paragraph\n")
		_, bodyFile := installFakeGH(t)
		t.Setenv("GH_CREATE_OUTPUT", "https://example.com/pr/create")
		t.Setenv("GH_VIEW_OUTPUT", `{"url":"https://example.com/pr/view"}`)

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Land() = %+v, want ok", envelope)
		}
		body, err := os.ReadFile(bodyFile)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		want := "first paragraph line one\nfirst paragraph line two"
		if string(body) != want {
			t.Fatalf("body = %q, want %q", body, want)
		}
	})

	t.Run("a later failure keeps the push effect and the record", func(t *testing.T) {
		repo := newRepo(t)
		bare := newBare(t)
		runGit(t, repo, "remote", "add", "origin", bare)
		tip := addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))

		envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if envelope.Outcome != "partial" || envelope.Phase != "brief" {
			t.Fatalf("Land() = %+v, want partial brief", envelope)
		}
		if !hasEffect(envelope, "push", "pushed") {
			t.Fatalf("effects = %+v, want the push effect", envelope.Effects)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || !record.Delivery.Pushed || record.Delivery.Commit != tip {
			t.Fatalf("delivery = %+v, want the pushed commit %s", record.Delivery, tip)
		}
	})

	t.Run("a retry pushes the recorded commit and keeps it", func(t *testing.T) {
		repo := newRepo(t)
		runGit(t, repo, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
		tip := addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
		home := newHome(t)
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask("1-a", "task/1-a"))
		writeBrief(t, home, "1-a", "land demo\n")
		installFakeGH(t)
		t.Setenv("GH_VIEW_OUTPUT", `{"url":"https://example.com/pr/view"}`)
		t.Setenv("GH_VIEW_EXIT", "1")
		t.Setenv("GH_CREATE_OUTPUT", "https://example.com/pr/create")

		first := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if first.Outcome != "partial" {
			t.Fatalf("Land() = %+v, want partial", first)
		}
		firstResult, ok := first.Result.(Result)
		if !ok || firstResult.Commit != tip {
			t.Fatalf("result = %#v, want the recorded commit %s", first.Result, tip)
		}

		runGit(t, repo, "checkout", "-q", "task/1-a")
		if err := os.WriteFile(filepath.Join(repo, "b.txt"), []byte("b\n"), 0o644); err != nil {
			t.Fatalf("write branch file: %v", err)
		}
		runGit(t, repo, "add", "b.txt")
		runGit(t, repo, "commit", "-q", "-m", "second")
		runGit(t, repo, "checkout", "-q", "main")
		if moved := strings.TrimSpace(runGit(t, repo, "rev-parse", "task/1-a")); moved == tip {
			t.Fatalf("branch tip = %s, want a moved branch", moved)
		}

		bare := newBare(t)
		runGit(t, repo, "remote", "set-url", "origin", bare)
		retry := Land(context.Background(), home, Options{Task: "1-a", Mode: "pr"})
		if retry.Outcome != "ok" {
			t.Fatalf("retry = %+v, want ok", retry)
		}
		retryResult, ok := retry.Result.(Result)
		if !ok || retryResult.Commit != tip {
			t.Fatalf("retry result = %#v, want the kept commit %s", retry.Result, tip)
		}
		if got := strings.TrimSpace(runGit(t, bare, "rev-parse", "task/1-a")); got != tip {
			t.Fatalf("origin branch = %s, want the recorded commit %s", got, tip)
		}
		record := loadTask(t, home, "1-a")
		if record.Delivery == nil || record.Delivery.Commit != tip {
			t.Fatalf("delivery = %+v, want the kept commit %s", record.Delivery, tip)
		}
	})
}
