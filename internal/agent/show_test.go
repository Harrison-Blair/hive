package agent

import (
	"errors"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestShowResolvesClaimBeforeWorkerRecordKey(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", Name: "leader", NameSet: true, PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	worker := records.Agent{Key: "coordinator", Name: "coordinator", Role: "worker", Status: records.AgentLive, Generation: 1, PaneID: "w1:p2"}
	if err := st.WriteIn([]string{"agents"}, "coordinator", worker); err != nil {
		t.Fatal(err)
	}
	envelope := Show(home, "coordinator")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the claim holder", envelope)
	}
	record, ok := envelope.Result.(records.Agent)
	if !ok || record.Key != "leader-g1" {
		t.Fatalf("Show() = %#v; want leader-g1 rather than the worker alias", envelope.Result)
	}
}

func TestShowDoesNotResolveUnclaimedSingletonOrphan(t *testing.T) {
	for _, name := range []string{"coordinator", "leader"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			seam := &seamWriter{writeCollection: "claims", writeErr: errors.New("seam claim failure")}
			options := RegisterOptions{Role: "coordinator", Name: name, NameSet: true, PaneID: "orphan-pane"}
			envelope := runRegisterLocked(t, home, seam, options)
			if envelope.Outcome != "partial" || envelope.Phase != "claim" {
				t.Fatalf("runRegisterLocked() = %#v; want partial claim", envelope)
			}
			if got := Show(home, name); got.Outcome == "ok" {
				t.Fatalf("Show(%q) = %#v; want no role holder without a claim", name, got)
			}
			key := name + "-g1"
			if got := Show(home, key); got.Outcome != "ok" {
				t.Fatalf("Show(%q) = %#v; want direct generation-record inspection", key, got)
			}
		})
	}
}

func TestShowResolvesCustomSingletonNameThroughItsClaim(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", Name: "leader", NameSet: true, PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	envelope := Show(home, "leader")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the custom singleton holder", envelope)
	}
	record, ok := envelope.Result.(records.Agent)
	if !ok || record.Key != "leader-g1" {
		t.Fatalf("Show() = %#v; want leader-g1", envelope.Result)
	}
}
