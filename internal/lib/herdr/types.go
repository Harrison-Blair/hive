package herdr

// AgentInfo is the detected agent in one pane.
type AgentInfo struct {
	PaneID         string            `json:"pane_id"`
	TerminalID     string            `json:"terminal_id"`
	WorkspaceID    string            `json:"workspace_id"`
	TabID          string            `json:"tab_id"`
	Name           string            `json:"name,omitempty"`
	Agent          string            `json:"agent,omitempty"`
	DisplayAgent   string            `json:"display_agent,omitempty"`
	AgentStatus    string            `json:"agent_status"`
	Focused        bool              `json:"focused"`
	Revision       uint64            `json:"revision"`
	Cwd            string            `json:"cwd,omitempty"`
	ForegroundCwd  string            `json:"foreground_cwd,omitempty"`
	Interactive    bool              `json:"interactive_ready,omitempty"`
	StateLabels    map[string]string `json:"state_labels,omitempty"`
	Tokens         map[string]string `json:"tokens,omitempty"`
	StateChangeSeq uint64            `json:"state_change_seq,omitempty"`
	AgentSession   *AgentSessionInfo `json:"agent_session,omitempty"`
}

// PaneInfo is one terminal area in a tab.
type PaneInfo struct {
	PaneID        string            `json:"pane_id"`
	TerminalID    string            `json:"terminal_id"`
	WorkspaceID   string            `json:"workspace_id"`
	TabID         string            `json:"tab_id"`
	Focused       bool              `json:"focused"`
	AgentStatus   string            `json:"agent_status"`
	Revision      uint64            `json:"revision"`
	Agent         string            `json:"agent,omitempty"`
	DisplayAgent  string            `json:"display_agent,omitempty"`
	Label         string            `json:"label,omitempty"`
	Title         string            `json:"title,omitempty"`
	Cwd           string            `json:"cwd,omitempty"`
	ForegroundCwd string            `json:"foreground_cwd,omitempty"`
	RestoreError  *string           `json:"restore_error,omitempty"`
	StateLabels   map[string]string `json:"state_labels,omitempty"`
	Tokens        map[string]string `json:"tokens,omitempty"`
	AgentSession  *AgentSessionInfo `json:"agent_session,omitempty"`
}

// TabInfo is one tab in a workspace.
type TabInfo struct {
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
	Number      uint   `json:"number"`
	Label       string `json:"label"`
	Focused     bool   `json:"focused"`
	PaneCount   uint   `json:"pane_count"`
	AgentStatus string `json:"agent_status"`
}

// WorkspaceInfo is the top-level container that holds tabs.
type WorkspaceInfo struct {
	WorkspaceID string                 `json:"workspace_id"`
	Number      uint                   `json:"number"`
	Label       string                 `json:"label"`
	Focused     bool                   `json:"focused"`
	PaneCount   uint                   `json:"pane_count"`
	TabCount    uint                   `json:"tab_count"`
	ActiveTabID string                 `json:"active_tab_id"`
	AgentStatus string                 `json:"agent_status"`
	Worktree    *WorkspaceWorktreeInfo `json:"worktree,omitempty"`
}

// WorktreeInfo is a Git worktree that herdr knows.
type WorktreeInfo struct {
	Path             string `json:"path"`
	Label            string `json:"label"`
	IsBare           bool   `json:"is_bare"`
	IsDetached       bool   `json:"is_detached"`
	IsPrunable       bool   `json:"is_prunable"`
	IsLinkedWorktree bool   `json:"is_linked_worktree"`
	Branch           string `json:"branch,omitempty"`
	OpenWorkspaceID  string `json:"open_workspace_id,omitempty"`
}
