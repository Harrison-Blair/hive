package herdr

import (
	"context"
	"encoding/json"
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

// WorkspaceCreateParams is one workspace.create request.
type WorkspaceCreateParams struct {
	Cwd   string `json:"cwd,omitempty"`
	Label string `json:"label,omitempty"`
	Focus bool   `json:"focus"`
}

// WorkspaceCreated is the new workspace, its root tab, and its root pane.
type WorkspaceCreated struct {
	Workspace WorkspaceInfo `json:"workspace"`
	Tab       TabInfo       `json:"tab"`
	RootPane  PaneInfo      `json:"root_pane"`
}

// WorkspaceCreate makes one workspace with a root tab and a root pane.
func WorkspaceCreate(ctx context.Context, params WorkspaceCreateParams, opts ...Option) (WorkspaceCreated, error) {
	var result struct {
		Type      *string         `json:"type"`
		Workspace json.RawMessage `json:"workspace"`
		Tab       json.RawMessage `json:"tab"`
		RootPane  json.RawMessage `json:"root_pane"`
	}
	if err := Call(ctx, "workspace.create", params, &result, opts...); err != nil {
		return WorkspaceCreated{}, err
	}
	if result.Type == nil || *result.Type != "workspace_created" {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: result type %q, want workspace_created", stringValue(result.Type))
	}
	if isAbsentOrNull(result.Workspace) {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: workspace_created workspace is missing")
	}
	var workspace WorkspaceInfo
	if err := json.Unmarshal(result.Workspace, &workspace); err != nil {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: workspace: %s", err)
	}
	if workspace.WorkspaceID == "" {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: workspace_created workspace_id is empty")
	}
	if isAbsentOrNull(result.Tab) {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: workspace_created tab is missing")
	}
	var tab TabInfo
	if err := json.Unmarshal(result.Tab, &tab); err != nil {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: tab: %s", err)
	}
	pane, err := decodePaneInfo(result.RootPane)
	if err != nil {
		return WorkspaceCreated{}, fmt.Errorf("bad reply: %s", err)
	}
	return WorkspaceCreated{Workspace: workspace, Tab: tab, RootPane: pane}, nil
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
