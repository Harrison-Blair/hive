package land

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func runGit(t *testing.T, dir string, args ...string) string {
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
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "config", "user.name", "Hive Test")
	runGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-q", "-m", "init")
	return repo
}

func newBare(t *testing.T) string {
	t.Helper()
	bare := filepath.Join(t.TempDir(), "origin.git")
	runGit(t, filepath.Dir(bare), "init", "-q", "--bare", bare)
	return bare
}

func addTaskBranch(t *testing.T, repo, branch, file, content string) string {
	t.Helper()
	runGit(t, repo, "checkout", "-q", "-b", branch, "main")
	if err := os.WriteFile(filepath.Join(repo, file), []byte(content), 0o644); err != nil {
		t.Fatalf("write branch file: %v", err)
	}
	runGit(t, repo, "add", file)
	runGit(t, repo, "commit", "-q", "-m", "work "+branch)
	runGit(t, repo, "checkout", "-q", "main")
	return strings.TrimSpace(runGit(t, repo, "rev-parse", branch))
}

func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	return home
}

func testProject(repo string) records.Project {
	return records.Project{ID: "repo-land", Path: repo, Base: "main", RegisteredAt: "2026-10-18T00:00:00Z"}
}

func testTask(id, branch string) records.Task {
	return records.Task{
		ID:           id,
		Project:      "repo-land",
		Title:        "task " + id,
		ActingAgent:  "worker-1",
		Status:       "verified",
		DependsOn:    []string{},
		Verification: &records.Verification{Agent: "verifier-1", Result: "pass", At: "2026-10-18T00:00:00Z"},
		Worktree: &records.Worktree{
			Path:   filepath.Join("/tmp", id),
			Branch: branch,
		},
		CreatedAt: "2026-10-18T00:00:00Z",
		UpdatedAt: "2026-10-18T00:00:00Z",
	}
}

func writeProject(t *testing.T, home string, project records.Project) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.Write("projects", project.ID, project); err != nil {
		t.Fatalf("write project: %v", err)
	}
}

func writeTask(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatalf("write task: %v", err)
	}
}

func loadTask(t *testing.T, home, taskID string) records.Task {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
		t.Fatalf("read task: %v", err)
	}
	return task
}

func writeBrief(t *testing.T, home, taskID, body string) {
	t.Helper()
	dir := filepath.Join(home, ".hive", "state", "briefs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir briefs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, taskID+".md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write brief: %v", err)
	}
}

func hasEffect(envelope output.Envelope, resource, action string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action {
			return true
		}
	}
	return false
}
