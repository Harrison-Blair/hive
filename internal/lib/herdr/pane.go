package herdr

import (
	"context"
	"encoding/json"
	"fmt"
)

// PaneTarget names one pane.
type PaneTarget struct {
	PaneID string `json:"pane_id"`
}

type paneInfoWire struct {
	PaneID        *string           `json:"pane_id"`
	TerminalID    *string           `json:"terminal_id"`
	WorkspaceID   *string           `json:"workspace_id"`
	TabID         *string           `json:"tab_id"`
	Focused       *bool             `json:"focused"`
	AgentStatus   *string           `json:"agent_status"`
	Revision      *uint64           `json:"revision"`
	Agent         *string           `json:"agent"`
	DisplayAgent  *string           `json:"display_agent"`
	Label         *string           `json:"label"`
	Title         *string           `json:"title"`
	Cwd           *string           `json:"cwd"`
	ForegroundCwd *string           `json:"foreground_cwd"`
	RestoreError  *string           `json:"restore_error"`
	StateLabels   map[string]string `json:"state_labels"`
	Tokens        map[string]string `json:"tokens"`
	AgentSession  json.RawMessage   `json:"agent_session"`
}

// PaneGet reads one pane.
func PaneGet(ctx context.Context, paneID string, opts ...Option) (PaneInfo, error) {
	var result struct {
		Type *string         `json:"type"`
		Pane json.RawMessage `json:"pane"`
	}
	if err := Call(ctx, "pane.get", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return PaneInfo{}, err
	}
	if result.Type == nil || *result.Type != "pane_info" {
		return PaneInfo{}, fmt.Errorf("bad reply: result type %q, want pane_info", stringValue(result.Type))
	}
	pane, err := decodePaneInfo(result.Pane)
	if err != nil {
		return PaneInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return pane, nil
}

// PaneList lists the panes of one workspace, or of each workspace when
// workspaceID is empty.
func PaneList(ctx context.Context, workspaceID string, opts ...Option) ([]PaneInfo, error) {
	params := struct {
		WorkspaceID string `json:"workspace_id,omitempty"`
	}{workspaceID}
	var result struct {
		Type  *string           `json:"type"`
		Panes []json.RawMessage `json:"panes"`
	}
	if err := Call(ctx, "pane.list", params, &result, opts...); err != nil {
		return nil, err
	}
	if result.Type == nil || *result.Type != "pane_list" {
		return nil, fmt.Errorf("bad reply: result type %q, want pane_list", stringValue(result.Type))
	}
	panes := make([]PaneInfo, 0, len(result.Panes))
	for _, raw := range result.Panes {
		pane, err := decodePaneInfo(raw)
		if err != nil {
			return nil, fmt.Errorf("bad reply: %s", err)
		}
		panes = append(panes, pane)
	}
	return panes, nil
}

// PaneClose closes one pane.
func PaneClose(ctx context.Context, paneID string, opts ...Option) error {
	var result struct {
		Type *string `json:"type"`
	}
	if err := Call(ctx, "pane.close", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return err
	}
	if result.Type == nil || *result.Type != "ok" {
		return fmt.Errorf("bad reply: result type %q, want ok", stringValue(result.Type))
	}
	return nil
}

func decodePaneInfo(raw json.RawMessage) (PaneInfo, error) {
	if isAbsentOrNull(raw) {
		return PaneInfo{}, fmt.Errorf("pane is missing")
	}
	var wire paneInfoWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return PaneInfo{}, fmt.Errorf("pane: %s", err)
	}
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"pane_id", wire.PaneID},
		{"terminal_id", wire.TerminalID},
		{"workspace_id", wire.WorkspaceID},
		{"tab_id", wire.TabID},
		{"agent_status", wire.AgentStatus},
	} {
		if field.value == nil || *field.value == "" {
			return PaneInfo{}, fmt.Errorf("pane %s is missing or empty", field.name)
		}
	}
	if wire.Focused == nil {
		return PaneInfo{}, fmt.Errorf("pane focused is missing")
	}
	if wire.Revision == nil {
		return PaneInfo{}, fmt.Errorf("pane revision is missing")
	}
	if !validAgentStatus(*wire.AgentStatus) {
		return PaneInfo{}, fmt.Errorf("pane agent_status %q is not a protocol status", *wire.AgentStatus)
	}
	session, err := decodeOptionalSession(wire.AgentSession)
	if err != nil {
		return PaneInfo{}, err
	}
	info := PaneInfo{
		PaneID:       *wire.PaneID,
		TerminalID:   *wire.TerminalID,
		WorkspaceID:  *wire.WorkspaceID,
		TabID:        *wire.TabID,
		Focused:      *wire.Focused,
		AgentStatus:  *wire.AgentStatus,
		Revision:     *wire.Revision,
		StateLabels:  wire.StateLabels,
		Tokens:       wire.Tokens,
		AgentSession: session,
	}
	if wire.Agent != nil {
		info.Agent = *wire.Agent
	}
	if wire.DisplayAgent != nil {
		info.DisplayAgent = *wire.DisplayAgent
	}
	if wire.Label != nil {
		info.Label = *wire.Label
	}
	if wire.Title != nil {
		info.Title = *wire.Title
	}
	if wire.Cwd != nil {
		info.Cwd = *wire.Cwd
	}
	if wire.ForegroundCwd != nil {
		info.ForegroundCwd = *wire.ForegroundCwd
	}
	info.RestoreError = wire.RestoreError
	return info, nil
}

// ProcessInfo is the shell and the foreground processes of one pane.
type ProcessInfo struct {
	PaneID                   string    `json:"pane_id"`
	ShellPID                 int       `json:"shell_pid"`
	ForegroundProcessGroupID int       `json:"foreground_process_group_id"`
	ForegroundProcesses      []Process `json:"foreground_processes"`
}

// Process is one process of the foreground process group of a pane.
type Process struct {
	PID  int      `json:"pid"`
	Name string   `json:"name"`
	Argv []string `json:"argv"`
}

// PaneProcessInfo reads the shell and the foreground processes of one pane.
func PaneProcessInfo(ctx context.Context, paneID string, opts ...Option) (ProcessInfo, error) {
	var result struct {
		Type        *string     `json:"type"`
		ProcessInfo ProcessInfo `json:"process_info"`
	}
	if err := Call(ctx, "pane.process_info", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return ProcessInfo{}, err
	}
	if result.Type == nil || *result.Type != "pane_process_info" {
		return ProcessInfo{}, fmt.Errorf("bad reply: result type %q, want pane_process_info", stringValue(result.Type))
	}
	return result.ProcessInfo, nil
}
