package edit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// AC3: task edit --brief writes the brief to the home and to the worktree.
func TestEditBriefWritesHomeFeatureBrief(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	worktree := t.TempDir()
	task := records.Task{
		ID: "1-seed", Project: "one", Title: "Seed", Feature: "feature-a",
		Status: "assigned", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		Worktree: &records.Worktree{Path: worktree},
	}
	tasktest.SeedTask(t, home, task)
	tasktest.SeedAgent(t, home, records.Agent{
		Key: "manager-one", Name: "manager-one", Profile: records.ProfilePlanner, Owns: "one", Project: "one",
		Status: records.AgentLive, PaneID: "w1:manager", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	seedClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one", Generation: 1})
	tasktest.WriteFile(t, filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"), "old brief")

	const updated = "updated brief\n"
	envelope := Run(home, Options{Task: task.ID, Agent: "manager-one", BriefSet: true, BriefPath: "-", Input: strings.NewReader(updated)})
	if envelope.Outcome != "ok" {
		t.Fatalf("Edit() = %+v; want ok", envelope)
	}
	relative := ".hive/spec/feature-a/tasks/1-seed/brief.md"
	if got, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != updated {
		t.Fatalf("home feature brief = %q, %v; want %q", got, ok, updated)
	}
	if got, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(worktree, filepath.FromSlash(relative))); !ok || got != updated {
		t.Fatalf("worktree feature brief = %q, %v; want %q", got, ok, updated)
	}
	if !tasktest.HasTaskEffect(envelope, "brief", "updated", relative) {
		t.Fatalf("effects = %+v; want the updated home feature brief", envelope.Effects)
	}
}

func seedClaim(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}
