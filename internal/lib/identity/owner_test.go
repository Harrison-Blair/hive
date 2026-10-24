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
		Key: "owner-hive-g1", Name: "manager-hive", Role: "manager",
		Profile: records.ProfilePlanner, Owns: "hive", Project: "hive",
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
		Key: "owner-hive-g1", Name: "manager-hive", Role: "manager",
		Profile: records.ProfilePlanner, Owns: "hive", Project: "hive",
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
