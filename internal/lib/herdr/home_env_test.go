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

func TestPaneRunSendsTextAndEnter(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.send_input", `{"type":"ok"}`)
	if err := PaneRun(context.Background(), "w1:p2", "export HIVE_HOME='/tmp/h'"); err != nil {
		t.Fatalf("PaneRun(): %v", err)
	}
	requests := server.recorded("pane.send_input")
	if len(requests) != 1 || requests[0].Params != `{"pane_id":"w1:p2","text":"export HIVE_HOME='/tmp/h'","keys":["enter"]}` {
		t.Fatalf("pane.send_input params = %#v; want text and the enter key", requests)
	}
	server.scriptResult("pane.send_input", `{"type":"pane_info"}`)
	if err := PaneRun(context.Background(), "w1:p2", "true"); err == nil {
		t.Fatal("PaneRun() wrong type error = nil")
	}
}
