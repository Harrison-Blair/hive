package herdr

import (
	"context"
	"testing"
)

func TestWorktreeListRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("worktree.list", `{"type":"worktree_list","source":{"repo_key":"/tmp/demo/.git","repo_name":"demo","repo_root":"/tmp/demo","source_checkout_path":"/tmp/demo","source_workspace_id":"w7"},"worktrees":[]}`)
	listing, err := WorktreeList(context.Background(), WorktreeListParams{Cwd: "/tmp/demo"})
	if err != nil {
		t.Fatalf("WorktreeList error = %v", err)
	}
	if listing.Source.SourceWorkspaceID != "w7" || listing.Source.SourceCheckoutPath != "/tmp/demo" {
		t.Fatalf("listing = %#v", listing)
	}
	requests := server.recorded("worktree.list")
	if len(requests) != 1 {
		t.Fatalf("worktree.list requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"cwd":"/tmp/demo"}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
	server.scriptResult("worktree.list", `{"type":"wrong"}`)
	if _, err := WorktreeList(context.Background(), WorktreeListParams{Cwd: "/tmp/demo"}); err == nil {
		t.Fatalf("WorktreeList wrong type error = nil")
	}
}
