package herdr

import (
	"context"
	"strings"
	"testing"
)

func TestWorktreeOpenRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("worktree.open", `{"type":"worktree_opened","workspace":{"workspace_id":"w2"},"tab":{"tab_id":"w2:t1"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_a","workspace_id":"w2","tab_id":"w2:t1","focused":false,"agent_status":"unknown","revision":1},"worktree":{"path":"/tmp/feat","is_linked_worktree":true},"already_open":true}`)
	opened, err := WorktreeOpen(context.Background(), WorktreeOpenParams{Cwd: "/tmp/repo", Path: "/tmp/feat"})
	if err != nil {
		t.Fatal(err)
	}
	if !opened.AlreadyOpen || opened.Workspace.WorkspaceID != "w2" || opened.Tab.TabID != "w2:t1" || opened.RootPane.PaneID != "w2:p1" || opened.Worktree.Path != "/tmp/feat" {
		t.Fatalf("opened = %+v", opened)
	}
	requests := server.recorded("worktree.open")
	if len(requests) != 1 || requests[0].Params != `{"cwd":"/tmp/repo","path":"/tmp/feat","focus":false}` {
		t.Fatalf("requests = %+v; want explicit source and path, no label, trust, or env", requests)
	}
}

func TestWorktreeOpenRejectsBadReplies(t *testing.T) {
	for _, test := range []struct{ reply, want string }{
		{`{"type":"wrong"}`, "worktree_opened"},
		{`{"type":"worktree_opened"}`, "workspace"},
		{`{"type":"worktree_opened","workspace":{"workspace_id":"w2"}}`, "tab"},
		{`{"type":"worktree_opened","workspace":{"workspace_id":"w2"},"tab":{"tab_id":"w2:t1"}}`, "pane"},
	} {
		t.Run(test.want, func(t *testing.T) {
			server := newFakeServer(t)
			server.scriptResult("worktree.open", test.reply)
			_, err := WorktreeOpen(context.Background(), WorktreeOpenParams{Cwd: "/tmp/repo", Path: "/tmp/feat"})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v; want %s", err, test.want)
			}
		})
	}
}
