package herdr

import (
	"context"
	"fmt"
)

// WorktreeCreateParams is one worktree.create request.
type WorktreeCreateParams struct {
	Cwd    string `json:"cwd"`
	Branch string `json:"branch"`
	Base   string `json:"base"`
	Path   string `json:"path"`
	Label  string `json:"label"`
	Focus  bool   `json:"focus"`
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
