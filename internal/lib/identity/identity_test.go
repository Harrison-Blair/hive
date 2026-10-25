package identity

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func writeAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func writeClaim(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

func resolveHeld(t *testing.T, home, agentFlag, paneID string) (records.Agent, error) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	var inner error
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		record, inner = ResolveHeld(locked, agentFlag, paneID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return record, inner
}

func liveWorker(key, name, pane string, generation uint64) records.Agent {
	return records.Agent{
		Key: key, Name: name, Role: "worker", Project: "one", Task: "1-seed",
		Status: records.AgentLive, PaneID: pane, Generation: generation,
	}
}

func TestResolveExplicitFlagReturnsTheNamedLiveRecord(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, liveWorker("worker-1", "worker-1", "w1:p1", 1))
	for _, flag := range []string{"worker-1"} {
		record, err := resolveHeld(t, home, flag, "")
		if err != nil || record.Key != "worker-1" {
			t.Fatalf("ResolveHeld(%q) = %#v, %v; want worker-1", flag, record, err)
		}
	}
	if _, err := resolveHeld(t, home, "nope", ""); err == nil || err.Error() != "agent not found: nope" {
		t.Fatalf("ResolveHeld(nope) error = %v; want agent not found", err)
	}
}

func TestResolveFlagRejectsStaleRecord(t *testing.T) {
	home := t.TempDir()
	record := liveWorker("worker-1", "worker-1", "w1:p1", 1)
	record.Status = records.AgentStale
	writeAgent(t, home, record)
	if _, err := resolveHeld(t, home, "worker-1", ""); err == nil || err.Error() != "agent not found: worker-1" {
		t.Fatalf("ResolveHeld(stale) error = %v; want agent not found", err)
	}
}

func TestResolvePaneUsesPaneAndPaneHistory(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, liveWorker("worker-1", "worker-1", "w1:p1", 1))
	moved := liveWorker("worker-2", "worker-2", "w2:p2", 1)
	moved.PaneHistory = []string{"w1:p9"}
	writeAgent(t, home, moved)
	if record, err := resolveHeld(t, home, "", "w1:p1"); err != nil || record.Key != "worker-1" {
		t.Fatalf("ResolveHeld(pane) = %#v, %v; want worker-1", record, err)
	}
	if record, err := resolveHeld(t, home, "", "w1:p9"); err != nil || record.Key != "worker-2" {
		t.Fatalf("ResolveHeld(history) = %#v, %v; want worker-2", record, err)
	}
}

func TestResolvePaneRejectsStaleAndUnknown(t *testing.T) {
	home := t.TempDir()
	record := liveWorker("worker-1", "worker-1", "w1:p1", 1)
	record.Status = records.AgentStale
	writeAgent(t, home, record)
	if _, err := resolveHeld(t, home, "", "w1:p1"); !errors.Is(err, ErrCallerUnknown) {
		t.Fatalf("ResolveHeld(stale) error = %v; want caller unknown", err)
	}
	if _, err := resolveHeld(t, home, "", ""); !errors.Is(err, ErrCallerUnknown) {
		t.Fatalf("ResolveHeld(empty pane) error = %v; want caller unknown", err)
	}
}

func TestResolveAmbiguousPaneSortsNames(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, liveWorker("worker-b", "worker-b", "w1:p1", 1))
	writeAgent(t, home, liveWorker("worker-a", "worker-a", "w1:p1", 1))
	_, err := resolveHeld(t, home, "", "w1:p1")
	if err == nil || err.Error() != "caller ambiguous: worker-a, worker-b" {
		t.Fatalf("ResolveHeld(ambiguous) error = %v; want sorted names", err)
	}
}

func TestResolveSingletonGenerationAndClaimKey(t *testing.T) {
	t.Run("old generation", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{Key: "manager-one-g1", Name: "manager-one", Role: "manager", Owns: "one", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
		writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one-g2", Generation: 2})
		if _, err := resolveHeld(t, home, "", "w1:p1"); !errors.Is(err, ErrCallerUnknown) {
			t.Fatalf("ResolveHeld(old generation) error = %v; want caller unknown", err)
		}
	})
	t.Run("cross claim", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{Key: "manager-two-g1", Name: "manager-two", Role: "manager", Owns: "two", Project: "two", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
		writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-two-g1", Generation: 1})
		if _, err := resolveHeld(t, home, "", "w1:p1"); !errors.Is(err, ErrCallerUnknown) {
			t.Fatalf("ResolveHeld(cross claim) error = %v; want caller unknown", err)
		}
		if _, err := resolveHeld(t, home, "manager-two", ""); err == nil || err.Error() != "agent not found: manager-two" {
			t.Fatalf("ResolveHeld(cross claim flag) error = %v; want agent not found", err)
		}
	})
	t.Run("valid claim", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{Key: "manager-one-g2", Name: "manager-one", Role: "manager", Owns: "one", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 2})
		writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one-g2", Generation: 2})
		byPane, err := resolveHeld(t, home, "", "w1:p1")
		if err != nil || byPane.Key != "manager-one-g2" {
			t.Fatalf("ResolveHeld(pane) = %#v, %v; want manager-one-g2", byPane, err)
		}
		byFlag, err := resolveHeld(t, home, "manager-one", "")
		if err != nil || byFlag.Key != byPane.Key {
			t.Fatalf("ResolveHeld(flag) = %#v, %v; want the pane record", byFlag, err)
		}
	})
}

func TestResolveFlagAndPaneUseTheSameClaimCheck(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, records.Agent{Key: "manager-one-g1", Name: "manager-one", Role: "manager", Owns: "one", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
	writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one-g1", Generation: 2})
	for _, route := range []struct{ flag, pane string }{{"manager-one", ""}, {"", "w1:p1"}} {
		record, err := resolveHeld(t, home, route.flag, route.pane)
		if err == nil || record.Key != "" {
			t.Fatalf("ResolveHeld(%q, %q) = %#v, %v; want the same claim refusal", route.flag, route.pane, record, err)
		}
	}
}

func TestResolveDoesNotReadTheEnvironment(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, liveWorker("worker-1", "worker-1", "w1:p1", 1))
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	if _, err := resolveHeld(t, home, "", ""); !errors.Is(err, ErrCallerUnknown) {
		t.Fatalf("ResolveHeld(empty pane) error = %v; want caller unknown", err)
	}
}

// renameRecord moves one state record to a different file name.
func renameRecord(t *testing.T, home, collection, oldName, newName string) {
	t.Helper()
	dir := filepath.Join(home, ".hive", "state", collection)
	if err := os.Rename(filepath.Join(dir, oldName+".json"), filepath.Join(dir, newName+".json")); err != nil {
		t.Fatal(err)
	}
}

func managerFixture(t *testing.T, home string) {
	t.Helper()
	writeAgent(t, home, records.Agent{
		Key: "manager-one-g7", Name: "manager-one", Role: "manager", Owns: "one", Project: "one",
		Status: records.AgentLive, PaneID: "w1:p1", Generation: 7,
	})
	writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one-g7", Generation: 7})
}

func TestResolveRejectsSelectedClaimAlias(t *testing.T) {
	t.Run("other project holder", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{
			Key: "manager-two-g7", Name: "manager-two", Role: "manager", Owns: "two", Project: "two",
			Status: records.AgentLive, PaneID: "w2:p2", Generation: 7,
		})
		writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-two-g7", Generation: 7})
		writeClaim(t, home, records.Claim{Key: "owner-two", Agent: "manager-two-g7", Generation: 7})
		if _, err := resolveHeld(t, home, "owner-one", ""); err == nil || err.Error() != "agent not found: owner-one" {
			t.Fatalf("ResolveHeld(manager-one) error = %v; want agent not found", err)
		}
		if record, err := resolveHeld(t, home, "", "w2:p2"); err != nil || record.Key != "manager-two-g7" {
			t.Fatalf("ResolveHeld(manager-two pane) = %#v, %v; want the legitimate holder", record, err)
		}
	})
	t.Run("worker holder", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, liveWorker("worker-key", "worker-name", "w1:p1", 1))
		writeClaim(t, home, records.Claim{Key: "owner-one", Agent: "worker-key", Generation: 1})
		if _, err := resolveHeld(t, home, "owner-one", ""); err == nil || err.Error() != "agent not found: owner-one" {
			t.Fatalf("ResolveHeld(manager-one) error = %v; want agent not found", err)
		}
	})
}

func TestResolveRejectsPathKeyMismatch(t *testing.T) {
	routes := []struct {
		name  string
		agent string
		pane  string
	}{
		{"claim key", "owner-one", ""},
		{"holder key", "manager-one-g7", ""},
		{"pane", "", "w1:p1"},
		{"history", "", "w1:p0"},
	}
	mismatches := []struct {
		name       string
		collection string
		oldName    string
		newName    string
	}{
		{"holder", "agents", "manager-one-g7", "wrong-holder"},
		{"claim", "claims", "owner-one", "wrong-claim"},
	}
	for _, mismatch := range mismatches {
		for _, route := range routes {
			t.Run(mismatch.name+"/"+route.name, func(t *testing.T) {
				home := t.TempDir()
				managerFixture(t, home)
				if route.name == "history" {
					writeAgent(t, home, records.Agent{
						Key: "manager-one-g7", Name: "manager-one", Role: "manager", Owns: "one", Project: "one",
						Status: records.AgentLive, PaneID: "w1:p1", PaneHistory: []string{"w1:p0"}, Generation: 7,
					})
				}
				renameRecord(t, home, mismatch.collection, mismatch.oldName, mismatch.newName)
				if _, err := resolveHeld(t, home, route.agent, route.pane); err == nil {
					t.Fatalf("ResolveHeld(%q, %q) accepted a %s path mismatch", route.agent, route.pane, mismatch.name)
				}
			})
		}
	}
}

func TestResolveThroughWholeHomeAlias(t *testing.T) {
	home := t.TempDir()
	managerFixture(t, home)
	alias := filepath.Join(t.TempDir(), "home-alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	record, err := resolveHeld(t, alias, "manager-one", "")
	if err != nil || record.Key != "manager-one-g7" {
		t.Fatalf("ResolveHeld(alias) = %#v, %v; want the manager record", record, err)
	}
}
