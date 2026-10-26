package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestEditCopiesFeatureBriefToActiveWorktree(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	worktree := t.TempDir()
	task := records.Task{
		ID:        "1-seed",
		Project:   "one",
		Title:     "Seed",
		Feature:   "feature-a",
		Status:    "assigned",
		DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-01T00:00:00Z",
		Worktree:  &records.Worktree{Path: worktree},
	}
	seedTask(t, home, task)
	seedAgent(t, home, records.Agent{
		Key: "manager-one", Name: "manager-one", Profile: records.ProfilePlanner, Owns: "one", Project: "one",
		Status: records.AgentLive, PaneID: "w1:manager", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "owner-one", records.Claim{
		Key: "owner-one", Agent: "manager-one", Generation: 1,
	}); err != nil {
		t.Fatal(err)
	}
	stateBrief := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	if err := os.MkdirAll(filepath.Dir(stateBrief), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stateBrief, []byte("old brief"), 0o644); err != nil {
		t.Fatal(err)
	}

	const updated = "updated brief\n"
	envelope := Edit(home, EditOptions{
		Task: task.ID, Agent: "manager-one", BriefSet: true, BriefPath: "-", Input: strings.NewReader(updated),
	})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Edit() = %+v, want ok", envelope)
	}
	featureBrief := filepath.Join(worktree, ".hive", "spec", "feature-a", "tasks", task.ID, "brief.md")
	if got, err := os.ReadFile(featureBrief); err != nil || string(got) != updated {
		t.Fatalf("worktree brief = %q, %v; want %q", got, err, updated)
	}
	if len(envelope.Effects) != 3 || envelope.Effects[2].Resource != "brief" || envelope.Effects[2].Action != "copied" || envelope.Effects[2].Path != featureBrief {
		t.Fatalf("effects = %+v, want the updated state brief, the updated home feature brief, and the copied worktree brief", envelope.Effects)
	}
}
