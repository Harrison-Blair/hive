package identity

import (
	"context"
	"errors"
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
	moved.PaneHistory = []string{"w1:p9", "w1:p1"}
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
		writeAgent(t, home, records.Agent{Key: "manager-one-g1", Name: "manager-one", Role: "manager", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
		writeClaim(t, home, records.Claim{Key: "manager-one", Agent: "manager-one-g2", Generation: 2})
		if _, err := resolveHeld(t, home, "", "w1:p1"); !errors.Is(err, ErrCallerUnknown) {
			t.Fatalf("ResolveHeld(old generation) error = %v; want caller unknown", err)
		}
	})
	t.Run("cross claim", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{Key: "manager-two-g1", Name: "manager-two", Role: "manager", Project: "two", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
		writeClaim(t, home, records.Claim{Key: "manager-one", Agent: "manager-two-g1", Generation: 1})
		if _, err := resolveHeld(t, home, "", "w1:p1"); !errors.Is(err, ErrCallerUnknown) {
			t.Fatalf("ResolveHeld(cross claim) error = %v; want caller unknown", err)
		}
		if _, err := resolveHeld(t, home, "manager-two", ""); err == nil || err.Error() != "agent not found: manager-two" {
			t.Fatalf("ResolveHeld(cross claim flag) error = %v; want agent not found", err)
		}
	})
	t.Run("valid claim", func(t *testing.T) {
		home := t.TempDir()
		writeAgent(t, home, records.Agent{Key: "manager-one-g2", Name: "manager-one", Role: "manager", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 2})
		writeClaim(t, home, records.Claim{Key: "manager-one", Agent: "manager-one-g2", Generation: 2})
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
	writeAgent(t, home, records.Agent{Key: "manager-one-g1", Name: "manager-one", Role: "manager", Project: "one", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
	writeClaim(t, home, records.Claim{Key: "manager-one", Agent: "manager-one-g1", Generation: 2})
	for _, route := range []struct{ flag, pane string }{{"manager-one", ""}, {"", "w1:p1"}} {
		if _, err := resolveHeld(t, home, route.flag, route.pane); err == nil || err.Error() != "agent not found: manager-one" {
			t.Fatalf("ResolveHeld(%q, %q) error = %v; want the same refusal", route.flag, route.pane, err)
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
