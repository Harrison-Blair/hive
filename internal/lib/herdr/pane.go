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

// PaneProcessInfoProcess is one process in the foreground group of a pane.
type PaneProcessInfoProcess struct {
	PID     uint32   `json:"pid"`
	Name    string   `json:"name"`
	Argv0   *string  `json:"argv0"`
	Argv    []string `json:"argv"`
	Cmdline string   `json:"cmdline"`
	Cwd     string   `json:"cwd"`
}

// PaneProcessInfo is the process data of one pane.
type PaneProcessInfo struct {
	PaneID                   string                   `json:"pane_id"`
	ShellPID                 *uint32                  `json:"shell_pid"`
	TTY                      *string                  `json:"tty"`
	ForegroundProcessGroupID *uint32                  `json:"foreground_process_group_id"`
	ForegroundProcesses      []PaneProcessInfoProcess `json:"foreground_processes"`
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

// PaneProcessInfoGet reads the process data of one pane.
func PaneProcessInfoGet(ctx context.Context, paneID string, opts ...Option) (PaneProcessInfo, error) {
	var result struct {
		Type        *string         `json:"type"`
		ProcessInfo json.RawMessage `json:"process_info"`
	}
	if err := Call(ctx, "pane.process_info", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return PaneProcessInfo{}, err
	}
	if result.Type == nil || *result.Type != "pane_process_info" {
		return PaneProcessInfo{}, fmt.Errorf("bad reply: result type %q, want pane_process_info", stringValue(result.Type))
	}
	info, err := decodeProcessInfo(result.ProcessInfo)
	if err != nil {
		return PaneProcessInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return info, nil
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
	return info, nil
}

func decodeProcessInfo(raw json.RawMessage) (PaneProcessInfo, error) {
	if isAbsentOrNull(raw) {
		return PaneProcessInfo{}, fmt.Errorf("process_info is missing")
	}
	var wire struct {
		PaneID                   *string         `json:"pane_id"`
		ShellPID                 *uint32         `json:"shell_pid"`
		TTY                      *string         `json:"tty"`
		ForegroundProcessGroupID *uint32         `json:"foreground_process_group_id"`
		ForegroundProcesses      json.RawMessage `json:"foreground_processes"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return PaneProcessInfo{}, fmt.Errorf("process_info: %s", err)
	}
	if wire.PaneID == nil || *wire.PaneID == "" {
		return PaneProcessInfo{}, fmt.Errorf("process_info pane_id is missing or empty")
	}
	info := PaneProcessInfo{
		PaneID:                   *wire.PaneID,
		ShellPID:                 wire.ShellPID,
		TTY:                      wire.TTY,
		ForegroundProcessGroupID: wire.ForegroundProcessGroupID,
	}
	if isAbsentOrNull(wire.ForegroundProcesses) {
		if len(wire.ForegroundProcesses) > 0 {
			return PaneProcessInfo{}, fmt.Errorf("process_info foreground_processes is null")
		}
		return info, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(wire.ForegroundProcesses, &items); err != nil {
		return PaneProcessInfo{}, fmt.Errorf("process_info foreground_processes: %s", err)
	}
	info.ForegroundProcesses = make([]PaneProcessInfoProcess, 0, len(items))
	for index, item := range items {
		entry, err := decodeProcessEntry(item)
		if err != nil {
			return PaneProcessInfo{}, fmt.Errorf("foreground_processes[%d]: %s", index, err)
		}
		info.ForegroundProcesses = append(info.ForegroundProcesses, entry)
	}
	return info, nil
}

func decodeProcessEntry(raw json.RawMessage) (PaneProcessInfoProcess, error) {
	if isAbsentOrNull(raw) {
		return PaneProcessInfoProcess{}, fmt.Errorf("entry is missing")
	}
	var wire struct {
		PID     *uint32         `json:"pid"`
		Name    *string         `json:"name"`
		Argv0   *string         `json:"argv0"`
		Argv    json.RawMessage `json:"argv"`
		Cmdline *string         `json:"cmdline"`
		Cwd     *string         `json:"cwd"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return PaneProcessInfoProcess{}, err
	}
	if wire.PID == nil {
		return PaneProcessInfoProcess{}, fmt.Errorf("pid is missing")
	}
	if wire.Name == nil {
		return PaneProcessInfoProcess{}, fmt.Errorf("name is missing")
	}
	entry := PaneProcessInfoProcess{PID: *wire.PID, Name: *wire.Name, Argv0: wire.Argv0}
	if wire.Cmdline != nil {
		entry.Cmdline = *wire.Cmdline
	}
	if wire.Cwd != nil {
		entry.Cwd = *wire.Cwd
	}
	if len(wire.Argv) > 0 && !isAbsentOrNull(wire.Argv) {
		var items []*string
		if err := json.Unmarshal(wire.Argv, &items); err != nil {
			return PaneProcessInfoProcess{}, fmt.Errorf("argv: %s", err)
		}
		entry.Argv = make([]string, 0, len(items))
		for index, item := range items {
			if item == nil {
				return PaneProcessInfoProcess{}, fmt.Errorf("argv[%d] is null", index)
			}
			entry.Argv = append(entry.Argv, *item)
		}
	}
	return entry, nil
}
