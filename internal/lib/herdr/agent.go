package herdr

import (
	"context"
	"fmt"
)

// CodeAgentNotRunning reports that the agent process left the pane while a
// wait was active. errors.go keeps the other shared codes.
const CodeAgentNotRunning = "agent_not_running"

// AgentSessionInfo is the native session reference of one agent.
type AgentSessionInfo struct {
	Source string `json:"source"`
	Agent  string `json:"agent"`
	Kind   string `json:"kind"`
	Value  string `json:"value"`
}

// AgentStartParams is one agent.start request.
type AgentStartParams struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	PaneID    string   `json:"pane_id"`
	Args      []string `json:"args"`
	TimeoutMS uint64   `json:"timeout_ms"`
}

// AgentStarted is one agent.start reply.
type AgentStarted struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
	Argv  []string  `json:"argv"`
}

// AgentTarget names one agent by name or pane ID.
type AgentTarget struct {
	Target string `json:"target"`
}

// AgentWaitParams is one agent.wait request.
type AgentWaitParams struct {
	Target    string   `json:"target"`
	Until     []string `json:"until"`
	TimeoutMS uint64   `json:"timeout_ms"`
}

// AgentPromptWaitOptions is the optional wait object of agent.prompt.
type AgentPromptWaitOptions struct {
	Until     []string `json:"until"`
	TimeoutMS uint64   `json:"timeout_ms"`
}

// AgentPromptParams is one agent.prompt request.
type AgentPromptParams struct {
	Target string                  `json:"target"`
	Text   string                  `json:"text"`
	Wait   *AgentPromptWaitOptions `json:"wait,omitempty"`
}

// AgentStart starts one agent in one pane.
func AgentStart(ctx context.Context, params AgentStartParams, opts ...Option) (AgentStarted, error) {
	var result AgentStarted
	if err := Call(ctx, "agent.start", params, &result, opts...); err != nil {
		return AgentStarted{}, err
	}
	if result.Type != "agent_started" {
		return AgentStarted{}, fmt.Errorf("bad reply: result type %q, want agent_started", result.Type)
	}
	if err := validateAgentInfo(result.Agent); err != nil {
		return AgentStarted{}, fmt.Errorf("bad reply: %s", err)
	}
	if len(result.Argv) == 0 {
		return AgentStarted{}, fmt.Errorf("bad reply: agent_started argv is empty")
	}
	return result, nil
}

// AgentWait waits for one settled agent status.
func AgentWait(ctx context.Context, params AgentWaitParams, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  string    `json:"type"`
		Agent AgentInfo `json:"agent"`
	}
	if err := Call(ctx, "agent.wait", params, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type != "agent_info" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_info", result.Type)
	}
	if err := validateAgentInfo(result.Agent); err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return result.Agent, nil
}

// AgentPrompt sends one prompt to one agent.
func AgentPrompt(ctx context.Context, params AgentPromptParams, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  string    `json:"type"`
		Agent AgentInfo `json:"agent"`
	}
	if err := Call(ctx, "agent.prompt", params, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type != "agent_prompted" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_prompted", result.Type)
	}
	if err := validateAgentInfo(result.Agent); err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return result.Agent, nil
}

// AgentGet reads one agent.
func AgentGet(ctx context.Context, target string, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  string    `json:"type"`
		Agent AgentInfo `json:"agent"`
	}
	if err := Call(ctx, "agent.get", AgentTarget{Target: target}, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type != "agent_info" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_info", result.Type)
	}
	if err := validateAgentInfo(result.Agent); err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return result.Agent, nil
}

// AgentList lists each live agent.
func AgentList(ctx context.Context, opts ...Option) ([]AgentInfo, error) {
	var result struct {
		Type   string      `json:"type"`
		Agents []AgentInfo `json:"agents"`
	}
	if err := Call(ctx, "agent.list", struct{}{}, &result, opts...); err != nil {
		return nil, err
	}
	if result.Type != "agent_list" {
		return nil, fmt.Errorf("bad reply: result type %q, want agent_list", result.Type)
	}
	for index, agent := range result.Agents {
		if err := validateAgentInfo(agent); err != nil {
			return nil, fmt.Errorf("bad reply: agents[%d]: %s", index, err)
		}
	}
	return result.Agents, nil
}

func validateAgentInfo(info AgentInfo) error {
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
			return fmt.Errorf("agent %s is empty", field.name)
		}
	}
	if info.AgentSession != nil {
		if err := validateAgentSession(info.AgentSession); err != nil {
			return err
		}
	}
	return nil
}

func validateAgentSession(session *AgentSessionInfo) error {
	if session.Source == "" || session.Agent == "" || session.Kind == "" || session.Value == "" {
		return fmt.Errorf("agent_session has an empty required field")
	}
	if session.Kind != "id" && session.Kind != "path" {
		return fmt.Errorf("agent_session kind %q is not id or path", session.Kind)
	}
	return nil
}
