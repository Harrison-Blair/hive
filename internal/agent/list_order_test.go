package agent

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestListOrder(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, records.Agent{
		Key: "a-live", Name: "a-live", Role: "worker", Project: "one",
		Status: records.AgentLive, PaneID: "w1:p1", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	writeRecord(t, home, records.Agent{
		Key: "b-unresolved", Name: "b-unresolved", Role: "worker", Project: "one",
		Status: records.AgentUnresolved, PaneID: "w1:p2", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})

	envelope := List(home, ListOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("List() = %+v, want ok", envelope)
	}
	list, ok := envelope.Result.([]records.Agent)
	if !ok || len(list) != 2 {
		t.Fatalf("result = %+v, want two records", envelope.Result)
	}
	if list[0].Key != "b-unresolved" || list[1].Key != "a-live" {
		t.Fatalf("order = %q, %q; want the unresolved record first", list[0].Key, list[1].Key)
	}
}
