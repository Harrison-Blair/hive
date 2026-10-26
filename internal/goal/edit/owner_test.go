package edit

import (
	"os"
	"path/filepath"
	"testing"

	goallib "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// goalOwnerHome makes a home with a project, an owner, a coordinator, one
// builder that owns nothing, and one active goal.
func goalOwnerHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	project := records.Project{ID: "demo", Path: t.TempDir(), Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"projects"}, project.ID, project); err != nil {
		t.Fatal(err)
	}
	agents := []records.Agent{
		{Key: "owner-demo-1", Name: "owner-demo", Profile: records.ProfilePlanner, Owns: "demo", Project: "demo", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1},
		{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, Status: records.AgentLive, PaneID: "w0:p1", Generation: 1},
		{Key: "builder-demo-1", Name: "builder-demo", Profile: records.ProfileBuilder, Project: "demo", Task: "1-task", Status: records.AgentLive, PaneID: "w2:p1", Generation: 1},
	}
	for _, record := range agents {
		if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
			t.Fatal(err)
		}
	}
	claims := []records.Claim{
		{Key: "owner-demo", Agent: "owner-demo-1", Generation: 1},
		{Key: "coordinator", Agent: "coordinator-g1", Generation: 1},
	}
	for _, claim := range claims {
		if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
			t.Fatal(err)
		}
	}
	goal := records.Goal{
		ID: "g1-demo", Project: "demo", Title: "Demo", Grant: "grant line",
		Status: "active", Waves: []records.GoalWave{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{goallib.GoalsCollection}, goal.ID, goal); err != nil {
		t.Fatal(err)
	}
	return home
}

// TestGoalEditNeedsTheProjectOwnerOrCoordinator proves the D7 refusal text in
// the goal package.
func TestGoalEditNeedsTheProjectOwnerOrCoordinator(t *testing.T) {
	home := goalOwnerHome(t)
	const goalID = "g1-demo"
	readStatus := func() string {
		t.Helper()
		registry, err := store.Open(home)
		if err != nil {
			t.Fatal(err)
		}
		var record records.Goal
		if err := registry.ReadIn([]string{goallib.GoalsCollection}, goalID, &record); err != nil {
			t.Fatal(err)
		}
		return record.Status
	}
	reset := func() {
		t.Helper()
		registry, err := store.Open(home)
		if err != nil {
			t.Fatal(err)
		}
		var record records.Goal
		if err := registry.ReadIn([]string{goallib.GoalsCollection}, goalID, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = "active"
		if err := registry.WriteIn([]string{goallib.GoalsCollection}, goalID, record); err != nil {
			t.Fatal(err)
		}
	}

	for _, agent := range []string{"owner-demo", "coordinator"} {
		envelope := Run(home, Options{Goal: goalID, Agent: agent, Status: "done", StatusSet: true})
		if envelope.Outcome != "ok" {
			t.Fatalf("Edit(%s) = %#v; want ok", agent, envelope)
		}
		if got := readStatus(); got != "done" {
			t.Fatalf("goal status = %q; want done", got)
		}
		reset()
	}

	refused := Run(home, Options{Goal: goalID, Agent: "builder-demo", Status: "done", StatusSet: true})
	if refused.Outcome != "failed" || refused.Error != "not the owner of project demo or the coordinator: builder-demo" {
		t.Fatalf("Edit(builder-demo) = %#v; want the exact owner refusal", refused)
	}
	if got := readStatus(); got != "active" {
		t.Fatalf("goal status = %q; want active after a refusal", got)
	}
}
