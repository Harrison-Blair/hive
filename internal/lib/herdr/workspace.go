package herdr

import (
	"context"
	"fmt"
)

// WorkspaceWorktreeInfo is the typed worktree metadata of one workspace.
type WorkspaceWorktreeInfo struct {
	RepoKey          string `json:"repo_key"`
	RepoName         string `json:"repo_name"`
	RepoRoot         string `json:"repo_root"`
	CheckoutPath     string `json:"checkout_path"`
	IsLinkedWorktree bool   `json:"is_linked_worktree"`
}

// WorkspaceTarget names one workspace.
type WorkspaceTarget struct {
	WorkspaceID string `json:"workspace_id"`
}

// WorkspaceCloseParams names one workspace to close.
type WorkspaceCloseParams struct {
	WorkspaceID string `json:"workspace_id"`
	CloseGroup  bool   `json:"close_group"`
}

// WorkspaceGet reads one workspace.
func WorkspaceGet(ctx context.Context, workspaceID string, opts ...Option) (WorkspaceInfo, error) {
	var result struct {
		Type      string        `json:"type"`
		Workspace WorkspaceInfo `json:"workspace"`
	}
	if err := Call(ctx, "workspace.get", WorkspaceTarget{WorkspaceID: workspaceID}, &result, opts...); err != nil {
		return WorkspaceInfo{}, err
	}
	if result.Type != "workspace_info" {
		return WorkspaceInfo{}, fmt.Errorf("bad reply: result type %q, want workspace_info", result.Type)
	}
	return result.Workspace, nil
}

// WorkspaceList lists all workspaces.
func WorkspaceList(ctx context.Context, opts ...Option) ([]WorkspaceInfo, error) {
	var result struct {
		Type       string          `json:"type"`
		Workspaces []WorkspaceInfo `json:"workspaces"`
	}
	if err := Call(ctx, "workspace.list", struct{}{}, &result, opts...); err != nil {
		return nil, err
	}
	if result.Type != "workspace_list" {
		return nil, fmt.Errorf("bad reply: result type %q, want workspace_list", result.Type)
	}
	return result.Workspaces, nil
}

// WorkspaceClose closes one workspace.
func WorkspaceClose(ctx context.Context, workspaceID string, closeGroup bool, opts ...Option) error {
	var result struct {
		Type string `json:"type"`
	}
	params := WorkspaceCloseParams{WorkspaceID: workspaceID, CloseGroup: closeGroup}
	if err := Call(ctx, "workspace.close", params, &result, opts...); err != nil {
		return err
	}
	if result.Type != "ok" {
		return fmt.Errorf("bad reply: result type %q, want ok", result.Type)
	}
	return nil
}
