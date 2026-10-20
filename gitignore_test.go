package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// skipWithoutGit skips a test that needs git and a work tree.
func skipWithoutGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		t.Skip("not inside a git work tree")
	}
}

// TestTrackedFilesAreNotIgnored checks that the allowlist .gitignore allows
// every tracked file, so a new file at the same path is not silently ignored.
func TestTrackedFilesAreNotIgnored(t *testing.T) {
	skipWithoutGit(t)
	out, err := exec.Command("git", "ls-files", "--cached", "--ignored", "--exclude-standard").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	if ignored := strings.TrimSpace(string(out)); ignored != "" {
		t.Errorf(".gitignore ignores tracked files; add the narrowest allow rule for each:\n%s", ignored)
	}
}

func TestGitHubWorkflowIsNotIgnored(t *testing.T) {
	skipWithoutGit(t)
	err := exec.Command("git", "check-ignore", "--no-index", "-q", ".github/workflows/ci.yml").Run()
	if err == nil {
		t.Fatal("git ignores .github/workflows/ci.yml")
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("git check-ignore: %v", err)
	}
}

// TestHomeMarkerIsTracked checks that Git does not ignore the home marker,
// which every clone needs so that hive never creates a home.
func TestHomeMarkerIsTracked(t *testing.T) {
	skipWithoutGit(t)
	// check-ignore exits 1 when the path is not ignored.
	err := exec.Command("git", "check-ignore", "-q", ".hive/README.md").Run()
	if err == nil {
		t.Fatal("git ignores .hive/README.md; the home marker must be tracked")
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("git check-ignore: %v", err)
	}
}

// TestPersonalPathsAreIgnored checks that Git ignores each personal and runtime
// path of the home, so a push of this public repository never carries them.
func TestFeatureSpecPathsAreNotIgnored(t *testing.T) {
	skipWithoutGit(t)
	for _, path := range []string{
		".hive/spec/x/README.md",
		".hive/spec/x/tasks/1-y/evidence/V1.md",
	} {
		if err := exec.Command("git", "check-ignore", "-q", path).Run(); err == nil {
			t.Errorf("git ignores %s", path)
		} else {
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Errorf("git check-ignore %s: %v", path, err)
			}
		}
	}
}

func TestPersonalPathsAreIgnored(t *testing.T) {
	skipWithoutGit(t)
	for _, path := range []string{
		"projects/p1/handoff.md",
		"projects/p1/memory/fact.md",
		"projects/p1/worktrees/t1/main.go",
		"memory/fact.md",
		"overrides/preferences.md",
		"plans/plan.md",
		"handoff.md",
		".hive/state/tasks.json",
		".hive/logs/2026-10-17.jsonl",
		".hive/tmp/scratch.md",
	} {
		// check-ignore exits 0 when the path is ignored and 1 when it is not.
		if err := exec.Command("git", "check-ignore", "-q", path).Run(); err != nil {
			t.Errorf("git does not ignore %s", path)
		}
	}
}

// TestSourcePathsUnderIgnoredNames checks that the ignore rules for the home
// directories match only at the root of the home. A source package that has
// the name of one of them, such as cmd/memory, must stay tracked.
func TestSourcePathsUnderIgnoredNames(t *testing.T) {
	skipWithoutGit(t)
	for _, path := range []string{
		"cmd/memory/memory.go",
		"internal/memory/memory.go",
		"cmd/plans/x.go",
		"internal/overrides/x.go",
		"internal/projects/x.go",
	} {
		// check-ignore exits 1 when the path is not ignored and 0 when it is.
		err := exec.Command("git", "check-ignore", "-q", path).Run()
		if err == nil {
			t.Errorf("git ignores %s; anchor the rule for its directory name to the root", path)
			continue
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			t.Errorf("git check-ignore %s: %v", path, err)
		}
	}
}
