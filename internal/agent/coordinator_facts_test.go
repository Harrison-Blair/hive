package agent

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// TestCoordinatorStartRecordHasFacts proves AC7: a new coordinator record has
// kind coordinator and the role mirror coordinator.
func TestCoordinatorStartRecordHasFacts(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptCoordinatorStart(fake, "w1:p1", "w1:t1", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "working", "coordinator")+`}`)

	if err := startTestCoordinator(home, "w1:p1"); err != nil {
		t.Fatalf("start coordinator: %v", err)
	}
	record := readAgentFields(t, home, "coordinator-g1")
	if record["kind"] != "coordinator" || record["profile"] != "coordinator" || record["role"] != "coordinator" {
		t.Fatalf("record = %#v; want kind, profile, and role coordinator", record)
	}
}

// TestCoordinatorClaimReadsKind proves AC7: the claim code reads kind, not
// role, when it finds the next generation. The seeded record has no role.
func TestCoordinatorClaimReadsKind(t *testing.T) {
	home := testHome(t)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, "coordinator-g5", map[string]any{
		"key": "coordinator-g5", "name": "coordinator", "kind": "coordinator",
		"profile": "coordinator", "status": records.AgentStale, "generation": 5,
	}); err != nil {
		t.Fatal(err)
	}
	launch, err := ClaimCoordinatorStart(context.Background(), home, nil)
	if err != nil {
		t.Fatalf("ClaimCoordinatorStart() = %v; want the starting claim", err)
	}
	claim, ok := readCoordinatorClaim(t, home)
	if !ok || claim.Generation != 6 {
		t.Fatalf("claim = %#v, present %t; want generation 6 after the kind record", claim, ok)
	}
	if err := launch.Release(); err != nil {
		t.Fatalf("Release() = %v", err)
	}
}
