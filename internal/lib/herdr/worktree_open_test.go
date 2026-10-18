package herdr

import (
	"context"
	"testing"
)

func TestWorktreeOpenRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("worktree.open", `{"type":"worktree_opened","workspace":{"workspace_id":"w2","number":2,"label":"1-worker-demo","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_a","workspace_id":"w2","tab_id":"w2:t1","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp/wt"},"worktree":{"path":"/tmp/wt","branch":"task/1-worker-demo","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"open_workspace_id":"w2","label":"repo"},"already_open":false}`)
	opened, err := WorktreeOpen(context.Background(), WorktreeOpenParams{Cwd: "/tmp/repo", Path: "/tmp/wt", Label: "1-worker-demo", Focus: false})
	if err != nil {
		t.Fatalf("WorktreeOpen error = %v", err)
	}
	if opened.Workspace.WorkspaceID != "w2" || opened.Tab.TabID != "w2:t1" || opened.RootPane.PaneID != "w2:p1" {
		t.Fatalf("opened = %#v", opened)
	}
	if opened.AlreadyOpen || opened.Worktree.Branch != "task/1-worker-demo" {
		t.Fatalf("opened = %#v", opened)
	}
	requests := server.recorded("worktree.open")
	if len(requests) != 1 {
		t.Fatalf("worktree.open requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"cwd":"/tmp/repo","path":"/tmp/wt","label":"1-worker-demo","focus":false}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
	server.scriptResult("worktree.open", `{"type":"worktree_created"}`)
	if _, err := WorktreeOpen(context.Background(), WorktreeOpenParams{}); err == nil {
		t.Fatalf("WorktreeOpen wrong type error = nil")
	}
}
