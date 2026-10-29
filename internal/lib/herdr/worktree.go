package herdr

import (
	"context"
	"fmt"
)

// WorktreeListParams is one worktree.list request.
type WorktreeListParams struct {
	Cwd string `json:"cwd,omitempty"`
}

// WorktreeSourceInfo identifies the repository and checkout for one listing.
type WorktreeSourceInfo struct {
	RepoKey            string `json:"repo_key"`
	RepoName           string `json:"repo_name"`
	RepoRoot           string `json:"repo_root"`
	SourceCheckoutPath string `json:"source_checkout_path"`
	SourceWorkspaceID  string `json:"source_workspace_id,omitempty"`
}

// WorktreeListResult is one worktree.list reply.
type WorktreeListResult struct {
	Type      string             `json:"type"`
	Source    WorktreeSourceInfo `json:"source"`
	Worktrees []WorktreeInfo     `json:"worktrees"`
}

// WorktreeList lists the worktrees of the repository that owns cwd.
func WorktreeList(ctx context.Context, params WorktreeListParams, opts ...Option) (WorktreeListResult, error) {
	var result WorktreeListResult
	if err := Call(ctx, "worktree.list", params, &result, opts...); err != nil {
		return WorktreeListResult{}, err
	}
	if result.Type != "worktree_list" {
		return WorktreeListResult{}, fmt.Errorf("bad reply: result type %q, want worktree_list", result.Type)
	}
	return result, nil
}

// WorktreeCreateParams is one worktree.create request.
type WorktreeCreateParams struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
	Cwd         string `json:"cwd,omitempty"`
	Branch      string `json:"branch"`
	Base        string `json:"base"`
	Path        string `json:"path"`
	Label       string `json:"label,omitempty"`
	Focus       bool   `json:"focus"`
}

// WorktreeCreateResult is one worktree.create reply.
type WorktreeCreateResult struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
	Tab       TabInfo       `json:"tab"`
	RootPane  PaneInfo      `json:"root_pane"`
	Worktree  WorktreeInfo  `json:"worktree"`
}

// WorktreeCreate creates one linked worktree and opens its workspace.
func WorktreeCreate(ctx context.Context, params WorktreeCreateParams, opts ...Option) (WorktreeCreateResult, error) {
	var result WorktreeCreateResult
	if err := Call(ctx, "worktree.create", params, &result, opts...); err != nil {
		return WorktreeCreateResult{}, err
	}
	if result.Type != "worktree_created" {
		return WorktreeCreateResult{}, fmt.Errorf("bad reply: result type %q, want worktree_created", result.Type)
	}
	return result, nil
}
