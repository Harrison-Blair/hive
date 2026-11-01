package herdr

import (
	"context"
	"encoding/json"
	"fmt"
)

// WorktreeListParams is one worktree.list request.
type WorktreeListParams struct {
	Cwd string `json:"cwd,omitempty"`
}

// WorktreeOpenParams selects an existing checkout from its parent repository.
type WorktreeOpenParams struct {
	Cwd   string `json:"cwd"`
	Path  string `json:"path"`
	Focus bool   `json:"focus"`
}

// WorktreeOpened holds the workspace topology and whether it already existed.
type WorktreeOpened struct {
	Workspace   WorkspaceInfo `json:"workspace"`
	Tab         TabInfo       `json:"tab"`
	RootPane    PaneInfo      `json:"root_pane"`
	Worktree    WorktreeInfo  `json:"worktree"`
	AlreadyOpen bool          `json:"already_open"`
}

// WorktreeOpen opens or adopts a checkout into its repository workspace group.
func WorktreeOpen(ctx context.Context, params WorktreeOpenParams, opts ...Option) (WorktreeOpened, error) {
	var result struct {
		Type string `json:"type"`
		WorktreeOpened
		RootPane json.RawMessage `json:"root_pane"`
	}
	if err := Call(ctx, "worktree.open", params, &result, opts...); err != nil {
		return WorktreeOpened{}, err
	}
	if result.Type != "worktree_opened" {
		return WorktreeOpened{}, fmt.Errorf("bad reply: result type %q, want worktree_opened", result.Type)
	}
	if result.Workspace.WorkspaceID == "" {
		return WorktreeOpened{}, fmt.Errorf("bad reply: worktree_opened workspace_id is missing or empty")
	}
	if result.Tab.TabID == "" {
		return WorktreeOpened{}, fmt.Errorf("bad reply: worktree_opened tab_id is missing or empty")
	}
	pane, err := decodePaneInfo(result.RootPane)
	if err != nil {
		return WorktreeOpened{}, fmt.Errorf("bad reply: %s", err)
	}
	result.WorktreeOpened.RootPane = pane
	return result.WorktreeOpened, nil
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
