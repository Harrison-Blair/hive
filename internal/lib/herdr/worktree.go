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

// WorktreeOpenParams is one worktree.open request. The request gives exactly
// one of Path and Branch.
type WorktreeOpenParams struct {
	Cwd   string `json:"cwd,omitempty"`
	Path  string `json:"path,omitempty"`
	Label string `json:"label,omitempty"`
	Focus bool   `json:"focus"`
}

// WorktreeOpenResult is one worktree.open reply.
type WorktreeOpenResult struct {
	Type        string        `json:"type"`
	Workspace   WorkspaceInfo `json:"workspace"`
	Tab         TabInfo       `json:"tab"`
	RootPane    PaneInfo      `json:"root_pane"`
	Worktree    WorktreeInfo  `json:"worktree"`
	AlreadyOpen bool          `json:"already_open"`
}

// WorktreeOpen opens one existing worktree as a workspace.
func WorktreeOpen(ctx context.Context, params WorktreeOpenParams, opts ...Option) (WorktreeOpenResult, error) {
	var result WorktreeOpenResult
	if err := Call(ctx, "worktree.open", params, &result, opts...); err != nil {
		return WorktreeOpenResult{}, err
	}
	if result.Type != "worktree_opened" {
		return WorktreeOpenResult{}, fmt.Errorf("bad reply: result type %q, want worktree_opened", result.Type)
	}
	return result, nil
}
