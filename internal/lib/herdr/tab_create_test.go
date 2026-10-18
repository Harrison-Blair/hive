package herdr

import (
	"context"
	"testing"
)

func TestTabCreateRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("tab.create", `{"type":"tab_created","tab":{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"2","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w1:p2","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t2","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp/demo"}}`)
	created, err := TabCreate(context.Background(), TabCreateParams{WorkspaceID: "w1", Cwd: "/tmp/demo", Focus: false})
	if err != nil {
		t.Fatalf("TabCreate error = %v", err)
	}
	if created.Tab.TabID != "w1:t2" || created.RootPane.PaneID != "w1:p2" || created.RootPane.Cwd != "/tmp/demo" {
		t.Fatalf("created = %#v", created)
	}
	requests := server.recorded("tab.create")
	if len(requests) != 1 {
		t.Fatalf("tab.create requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"workspace_id":"w1","cwd":"/tmp/demo","focus":false}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
	server.scriptResult("tab.create", `{"type":"workspace_info"}`)
	if _, err := TabCreate(context.Background(), TabCreateParams{}); err == nil {
		t.Fatalf("TabCreate wrong type error = nil")
	}
}
