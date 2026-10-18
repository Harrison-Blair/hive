package herdr

import (
	"context"
	"strings"
	"testing"
)

func TestWorkspaceCreateRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("workspace.create", `{"type":"workspace_created","workspace":{"workspace_id":"w6","number":1,"label":"hive:demo","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w6:t1","agent_status":"unknown"},"tab":{"tab_id":"w6:t1","workspace_id":"w6","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w6:p1","terminal_id":"term_a","workspace_id":"w6","tab_id":"w6:t1","focused":false,"agent_status":"unknown","revision":0,"cwd":"/tmp/demo"}}`)
	created, err := WorkspaceCreate(context.Background(), WorkspaceCreateParams{Cwd: "/tmp/demo", Label: "hive:demo", Focus: false})
	if err != nil {
		t.Fatalf("WorkspaceCreate error = %v", err)
	}
	if created.Workspace.WorkspaceID != "w6" || created.Tab.TabID != "w6:t1" || created.RootPane.PaneID != "w6:p1" || created.RootPane.Cwd != "/tmp/demo" {
		t.Fatalf("created = %#v", created)
	}
	requests := server.recorded("workspace.create")
	if len(requests) != 1 {
		t.Fatalf("workspace.create requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"cwd":"/tmp/demo","label":"hive:demo","focus":false}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
	server.scriptResult("workspace.create", `{"type":"workspace_info"}`)
	if _, err := WorkspaceCreate(context.Background(), WorkspaceCreateParams{}); err == nil {
		t.Fatalf("WorkspaceCreate wrong type error = nil")
	}
}

func TestPaneListRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.list", `{"type":"pane_list","panes":[{"pane_id":"w6:p1","terminal_id":"term_a","workspace_id":"w6","tab_id":"w6:t1","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp/demo"},{"pane_id":"w6:p2","terminal_id":"term_b","workspace_id":"w6","tab_id":"w6:t2","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp/demo"}]}`)
	panes, err := PaneList(context.Background(), "w6")
	if err != nil {
		t.Fatalf("PaneList error = %v", err)
	}
	if len(panes) != 2 || panes[0].PaneID != "w6:p1" || panes[1].TabID != "w6:t2" {
		t.Fatalf("panes = %#v", panes)
	}
	requests := server.recorded("pane.list")
	if len(requests) != 1 || requests[0].Params != `{"workspace_id":"w6"}` {
		t.Fatalf("pane.list requests = %#v", requests)
	}
	server.scriptResult("pane.list", `{"type":"pane_list","panes":null}`)
	if _, err := PaneList(context.Background(), "w6"); err == nil {
		t.Fatalf("PaneList null panes error = nil")
	}
}

func TestTabListRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("tab.list", `{"type":"tab_list","tabs":[{"tab_id":"w6:t2","workspace_id":"w6","number":2,"label":"2","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w6:t1","workspace_id":"w6","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"}]}`)
	tabs, err := TabList(context.Background(), "w6")
	if err != nil {
		t.Fatalf("TabList error = %v", err)
	}
	if len(tabs) != 2 || tabs[0].TabID != "w6:t2" || tabs[0].Number != 2 {
		t.Fatalf("tabs = %#v", tabs)
	}
	requests := server.recorded("tab.list")
	if len(requests) != 1 || requests[0].Params != `{"workspace_id":"w6"}` {
		t.Fatalf("tab.list requests = %#v", requests)
	}
	server.scriptResult("tab.list", `{"type":"tab_list","tabs":[{"tab_id":"","workspace_id":"w6","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"}]}`)
	if _, err := TabList(context.Background(), "w6"); err == nil {
		t.Fatalf("TabList empty tab_id error = nil")
	}
}

func TestPaneGetRestoreError(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":0}}`)
	pane, err := PaneGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("PaneGet absent restore_error error = %v", err)
	}
	if pane.RestoreError != nil {
		t.Fatalf("RestoreError = %#v; want nil", pane.RestoreError)
	}
	server.scriptResult("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":0,"restore_error":null}}`)
	pane, err = PaneGet(context.Background(), "w1:p1")
	if err != nil || pane.RestoreError != nil {
		t.Fatalf("PaneGet null restore_error = %#v, %v", pane, err)
	}
	server.scriptResult("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":0,"restore_error":"Saved directory is unavailable."}}`)
	pane, err = PaneGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("PaneGet present restore_error error = %v", err)
	}
	if pane.RestoreError == nil || *pane.RestoreError != "Saved directory is unavailable." {
		t.Fatalf("RestoreError = %#v", pane.RestoreError)
	}
	server.scriptResult("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":0,"restore_error":7}}`)
	if _, err := PaneGet(context.Background(), "w1:p1"); err == nil {
		t.Fatalf("PaneGet wrong-type restore_error error = nil")
	} else if !strings.Contains(err.Error(), "restore_error") {
		t.Fatalf("PaneGet wrong-type restore_error error = %v; want restore_error", err)
	}
}
