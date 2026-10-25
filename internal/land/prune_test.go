package land

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestLandPrunesStoppedActingAgent(t *testing.T) {
	repo := newRepo(t)
	bare := newBare(t)
	runGit(t, repo, "remote", "add", "origin", bare)
	runGit(t, repo, "push", "-q", "-u", "origin", "main")
	addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
	home := newHome(t)
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask("1-a", "task/1-a"))
	writeLandAgent(t, home, records.Agent{
		Key: "worker-1", Name: "worker-1", Project: "repo-land", Task: "1-a",
		Status: records.AgentStopped, StartedAt: "2026-01-01T00:00:00Z",
	})

	envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "demo"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Land() = %#v; want ok", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", "worker-1.json")); !os.IsNotExist(err) {
		t.Fatalf("stopped acting agent still exists: %v", err)
	}
}
