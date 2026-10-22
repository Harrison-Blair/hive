package herdr

import (
	"context"
	"testing"
)

func TestTabRenameRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("tab.rename", `{"type":"tab_info","tab":{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"worker-1","focused":false,"pane_count":1,"agent_status":"unknown"}}`)
	if err := TabRename(context.Background(), "w1:t2", "worker-1"); err != nil {
		t.Fatalf("TabRename error = %v", err)
	}
	requests := server.recorded("tab.rename")
	if len(requests) != 1 {
		t.Fatalf("tab.rename requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"tab_id":"w1:t2","label":"worker-1"}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
}
