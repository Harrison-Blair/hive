package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestSupervisorBinaryCheckWarnsWhenOlderThanInstalledHive(t *testing.T) {
	home := newHome(t)
	oldTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	hiveTime := oldTime.Add(time.Hour)
	executable, _ := startFakeSupervisor(t, home, oldTime)
	installFakeHive(t, hiveTime)

	check := findCheck(t, Run(context.Background(), home, nil), "supervisor binary")
	if check.Status != StatusWarn {
		t.Fatalf("supervisor binary = %+v, want warn", check)
	}
	for _, want := range []string{oldTime.UTC().Format(time.RFC3339), hiveTime.UTC().Format(time.RFC3339), executable} {
		if !strings.Contains(check.Detail, want) {
			t.Fatalf("detail = %q, want %q", check.Detail, want)
		}
	}
}

func TestSupervisorBinaryCheckWarnsWhenExecutableIsDeleted(t *testing.T) {
	home := newHome(t)
	oldTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	executable, _ := startFakeSupervisor(t, home, oldTime)
	if err := os.Remove(executable); err != nil {
		t.Fatal(err)
	}
	hiveTime := time.Now().Truncate(time.Second)
	installFakeHive(t, hiveTime)

	check := findCheck(t, Run(context.Background(), home, nil), "supervisor binary")
	if check.Status != StatusWarn {
		t.Fatalf("supervisor binary = %+v, want deleted executable warning", check)
	}
	for _, want := range []string{"deleted", oldTime.UTC().Format(time.RFC3339), hiveTime.UTC().Format(time.RFC3339)} {
		if !strings.Contains(check.Detail, want) {
			t.Fatalf("detail = %q, want %q", check.Detail, want)
		}
	}
}

func TestStaleWorktreesCheckNamesLandedAndCancelledTasks(t *testing.T) {
	home := newHome(t)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range []records.Task{
		{ID: "1-landed", Project: "demo", Status: "complete", Worktree: &records.Worktree{Path: makeTaskWorktree(t, home, "demo", "1-landed")}, Delivery: &records.Delivery{Commit: "abc123", Pushed: true}},
		{ID: "2-cancelled", Project: "demo", Status: "cancelled", Worktree: &records.Worktree{Path: makeTaskWorktree(t, home, "demo", "2-cancelled")}},
		{ID: "3-active", Project: "demo", Status: "active", Worktree: &records.Worktree{Path: makeTaskWorktree(t, home, "demo", "3-active")}},
	} {
		if err := registry.Write("tasks", task.ID, task); err != nil {
			t.Fatal(err)
		}
	}

	check := findCheck(t, Run(context.Background(), home, nil), "stale worktrees")
	if check.Status != StatusWarn {
		t.Fatalf("stale worktrees = %+v, want warn", check)
	}
	for _, want := range []string{"demo/1-landed", "demo/2-cancelled"} {
		if !strings.Contains(check.Detail, want) {
			t.Fatalf("detail = %q, want %q", check.Detail, want)
		}
	}
	if strings.Contains(check.Detail, "3-active") {
		t.Fatalf("detail = %q, must omit active worktree", check.Detail)
	}
}

func TestUnpushedTaskBranchesCheckNamesCommitsAheadOfUpstream(t *testing.T) {
	home := newHome(t)
	worktree := makeTaskWorktree(t, home, "demo", "1-unpushed")
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitTest(t, "", "init", "--bare", "--quiet", remote)
	runGitTest(t, worktree, "init", "--quiet", "-b", "task/1-unpushed")
	runGitTest(t, worktree, "config", "user.email", "worker@example.test")
	runGitTest(t, worktree, "config", "user.name", "Worker")
	writeGitFile(t, worktree, "first.txt", "first\n")
	runGitTest(t, worktree, "add", "first.txt")
	runGitTest(t, worktree, "commit", "--quiet", "-m", "first")
	runGitTest(t, worktree, "remote", "add", "origin", remote)
	runGitTest(t, worktree, "push", "--quiet", "-u", "origin", "HEAD")
	writeGitFile(t, worktree, "second.txt", "second\n")
	runGitTest(t, worktree, "add", "second.txt")
	runGitTest(t, worktree, "commit", "--quiet", "-m", "second")

	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("tasks", "1-unpushed", records.Task{
		ID: "1-unpushed", Project: "demo", Status: "active", Worktree: &records.Worktree{Path: worktree, Branch: "task/1-unpushed"},
	}); err != nil {
		t.Fatal(err)
	}

	check := findCheck(t, Run(context.Background(), home, nil), "unpushed task branches")
	if check.Status != StatusWarn || !strings.Contains(check.Detail, "demo/1-unpushed") || !strings.Contains(check.Detail, "1 commit") {
		t.Fatalf("unpushed task branches = %+v, want one commit ahead", check)
	}
}

func startFakeSupervisor(t *testing.T, home string, modified time.Time) (string, string) {
	t.Helper()
	source, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(t.TempDir(), "supervisor")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, data, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(executable, modified, modified); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestFakeSupervisorProcess$")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = command.Process.Kill()
		_ = command.Wait()
	})
	pidStart, err := process.Start(int64(command.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 1, PID: int64(command.Process.Pid), PIDStart: pidStart,
	}); err != nil {
		t.Fatal(err)
	}
	return executable, filepath.Join("/proc", fmt.Sprint(command.Process.Pid), "exe")
}

func TestFakeSupervisorProcess(t *testing.T) {
	if filepath.Base(os.Args[0]) != "supervisor" {
		return
	}
	for {
		time.Sleep(time.Hour)
	}
}

func installFakeHive(t *testing.T, modified time.Time) string {
	t.Helper()
	directory := t.TempDir()
	executable := filepath.Join(directory, "hive")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(executable, modified, modified); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	return executable
}

func makeTaskWorktree(t *testing.T, home, project, task string) string {
	t.Helper()
	path := filepath.Join(home, "projects", project, "worktrees", task)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func runGitTest(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	if directory != "" {
		command.Dir = directory
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func writeGitFile(t *testing.T, directory, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
