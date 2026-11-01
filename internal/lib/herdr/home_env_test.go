package herdr

import (
	"context"
	"testing"
)

func TestWorkspaceCreatePassesEnvironment(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("workspace.create", `{"type":"workspace_created","workspace":{"workspace_id":"w2","number":2,"label":"demo","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_a","workspace_id":"w2","tab_id":"w2:t1","focused":false,"agent_status":"unknown","revision":1}}`)
	if _, err := WorkspaceCreate(context.Background(), WorkspaceCreateParams{Cwd: "/tmp/demo", Env: map[string]string{"HIVE_HOME": "/tmp/hive-home"}}); err != nil {
		t.Fatalf("WorkspaceCreate(): %v", err)
	}
	requests := server.recorded("workspace.create")
	if len(requests) != 1 || requests[0].Params != `{"cwd":"/tmp/demo","focus":false,"env":{"HIVE_HOME":"/tmp/hive-home"}}` {
		t.Fatalf("workspace.create params = %#v; want the environment", requests)
	}
}
