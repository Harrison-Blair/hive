package chat

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestRule1NamesTheLiveAgent proves the exact D8 rule 1 text.
func TestRule1NamesTheLiveAgent(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedManager(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p2", Text: "hello"})
	assertSendDenied(t, home, fake, envelope, "chat send is for the user only: only the user can send chat; rule 1 (caller pane belongs to the live agent manager-hive)")
}

// TestRule3NamesTheLiveAgent proves the exact D8 rule 3 text.
func TestRule3NamesTheLiveAgent(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("w2")
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w2:p2", Text: "hello"})
	assertSendDenied(t, home, fake, envelope, "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent worker-one)")
}

// TestRule3UsesTheFactsNotTheRoleMirror proves that rule 3 refuses a live
// agent with no role mirror, because the rule reads the facts.
func TestRule3UsesTheFactsNotTheRoleMirror(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedAgent(t, home, records.Agent{
		Key: "general-one", Name: "general-one", Profile: records.ProfileGeneral,
		PaneID: "w3:p1", WorkspaceID: "w3", Status: records.AgentLive, Generation: 1,
	})
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("w3")
	err := ValidateCaller(context.Background(), home, "", "w3:p2", "")
	want := "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent general-one)"
	if err == nil || err.Error() != want {
		t.Fatalf("ValidateCaller() error = %v; want %q", err, want)
	}
}

// TestRule3StillAllowsTheCoordinatorWorkspace proves that the coordinator
// workspace stays the user workspace for a live agent that is not the
// coordinator.
func TestRule3StillAllowsTheCoordinatorWorkspace(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedManager(t, home)
	fake := newFakeHerdr(t)
	if err := ValidateCaller(context.Background(), home, "", "w1:p3", ""); err != nil {
		t.Fatalf("ValidateCaller() error = %v; want the user pane allowed", err)
	}
	if got := len(fake.recorded("pane.get")); got != 1 {
		t.Fatalf("pane.get calls = %d; want one workspace check", got)
	}
}
