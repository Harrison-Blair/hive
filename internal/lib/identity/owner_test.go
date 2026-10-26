package identity

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestResolveOwnerClaimFromFacts proves AC6: an owner with the claim
// owner-hive is a valid caller by --agent and by its pane.
func TestResolveOwnerClaimFromFacts(t *testing.T) {
	home := t.TempDir()
	owner := records.Agent{
		Key: "owner-hive-g1", Name: "manager-hive", Profile: records.ProfilePlanner,
		Owns: "hive", Project: "hive",
		Status: records.AgentLive, PaneID: "w1:p1", Generation: 1,
	}
	writeAgent(t, home, owner)
	writeClaim(t, home, records.Claim{Key: "owner-hive", Agent: owner.Key, Generation: 1})

	for _, test := range []struct{ name, flag, pane string }{
		{name: "claim key", flag: "owner-hive"},
		{name: "agent name", flag: "manager-hive"},
		{name: "pane", pane: "w1:p1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			record, err := resolveHeld(t, home, test.flag, test.pane)
			if err != nil || record.Key != owner.Key {
				t.Fatalf("ResolveHeld(%q, %q) = %#v, %v; want %s", test.flag, test.pane, record, err, owner.Key)
			}
		})
	}
}

// TestResolveRefusesAFactClaimWithoutTheClaimRecord shows the fact claim is
// required.
func TestResolveRefusesAFactClaimWithoutTheClaimRecord(t *testing.T) {
	home := t.TempDir()
	owner := records.Agent{
		Key: "owner-hive-g1", Name: "manager-hive", Profile: records.ProfilePlanner,
		Owns: "hive", Project: "hive",
		Status: records.AgentLive, PaneID: "w1:p1", Generation: 1,
	}
	writeAgent(t, home, owner)
	writeClaim(t, home, records.Claim{Key: "manager-hive", Agent: owner.Key, Generation: 1})
	if _, err := resolveHeld(t, home, "owner-hive", ""); err == nil || err.Error() != "agent not found: owner-hive" {
		t.Fatalf("ResolveHeld(owner-hive) error = %v; want agent not found", err)
	}
	if _, err := resolveHeld(t, home, "", "w1:p1"); err == nil || err.Error() != ErrCallerUnknown.Error() {
		t.Fatalf("ResolveHeld(pane) error = %v; want caller unknown", err)
	}
}

// TestResolveLegacyBuiltinClaim proves that a live built-in record without
// its facts resolves by the coordinator claim that names it. The TUI tests
// write this old shape, and migration version 2 writes the facts of a real
// home before any command reads a record.
func TestResolveLegacyBuiltinClaim(t *testing.T) {
	home := t.TempDir()
	record := records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator,
		PaneID: "w1:p1", Status: records.AgentLive, Generation: 1,
	}
	writeAgent(t, home, record)
	writeClaim(t, home, records.Claim{Key: records.AgentKindCoordinator, Agent: record.Key, Generation: 1})

	got, err := resolveHeld(t, home, records.AgentKindCoordinator, "")
	if err != nil || got.Key != record.Key {
		t.Fatalf("ResolveHeld(coordinator) = %#v, %v; want %s", got, err, record.Key)
	}
	byPane, err := resolveHeld(t, home, "", "w1:p1")
	if err != nil || byPane.Key != record.Key {
		t.Fatalf("ResolveHeld(pane) = %#v, %v; want %s", byPane, err, record.Key)
	}
}

// TestResolveRefusesALegacyProjectClaim shows the legacy exception covers
// only the built-in singleton claims. A record without facts cannot take an
// owner claim.
func TestResolveRefusesALegacyProjectClaim(t *testing.T) {
	home := t.TempDir()
	record := records.Agent{
		Key: "worker-g1", Name: "worker", Profile: records.ProfileBuilder,
		PaneID: "w1:p1", Status: records.AgentLive, Generation: 1,
	}
	writeAgent(t, home, record)
	writeClaim(t, home, records.Claim{Key: "owner-hive", Agent: record.Key, Generation: 1})

	if _, err := resolveHeld(t, home, "owner-hive", ""); err == nil || err.Error() != "agent not found: owner-hive" {
		t.Fatalf("ResolveHeld(owner-hive) error = %v; want agent not found", err)
	}
}
