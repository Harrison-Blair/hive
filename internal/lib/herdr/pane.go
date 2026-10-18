package herdr

import (
	"context"
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

// PaneGet reads one pane.
func PaneGet(ctx context.Context, paneID string, opts ...Option) (PaneInfo, error) {
	var result struct {
		Type string   `json:"type"`
		Pane PaneInfo `json:"pane"`
	}
	if err := Call(ctx, "pane.get", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return PaneInfo{}, err
	}
	if result.Type != "pane_info" {
		return PaneInfo{}, fmt.Errorf("bad reply: result type %q, want pane_info", result.Type)
	}
	if err := validatePaneInfo(result.Pane); err != nil {
		return PaneInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return result.Pane, nil
}

// PaneClose closes one pane.
func PaneClose(ctx context.Context, paneID string, opts ...Option) error {
	var result struct {
		Type string `json:"type"`
	}
	if err := Call(ctx, "pane.close", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return err
	}
	if result.Type != "ok" {
		return fmt.Errorf("bad reply: result type %q, want ok", result.Type)
	}
	return nil
}

// PaneProcessInfoGet reads the process data of one pane.
func PaneProcessInfoGet(ctx context.Context, paneID string, opts ...Option) (PaneProcessInfo, error) {
	var result struct {
		Type        string          `json:"type"`
		ProcessInfo PaneProcessInfo `json:"process_info"`
	}
	if err := Call(ctx, "pane.process_info", PaneTarget{PaneID: paneID}, &result, opts...); err != nil {
		return PaneProcessInfo{}, err
	}
	if result.Type != "pane_process_info" {
		return PaneProcessInfo{}, fmt.Errorf("bad reply: result type %q, want pane_process_info", result.Type)
	}
	if result.ProcessInfo.PaneID == "" {
		return PaneProcessInfo{}, fmt.Errorf("bad reply: process_info pane_id is empty")
	}
	return result.ProcessInfo, nil
}

func validatePaneInfo(info PaneInfo) error {
	for _, field := range []struct {
		name  string
		value string
	}{
		{"pane_id", info.PaneID},
		{"terminal_id", info.TerminalID},
		{"workspace_id", info.WorkspaceID},
		{"tab_id", info.TabID},
		{"agent_status", info.AgentStatus},
	} {
		if field.value == "" {
			return fmt.Errorf("pane %s is empty", field.name)
		}
	}
	if info.AgentSession != nil {
		if err := validateAgentSession(info.AgentSession); err != nil {
			return err
		}
	}
	return nil
}
