package herdr

import (
	"bytes"
	"context"
	"encoding/json"
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

type agentInfoWire struct {
	PaneID         *string           `json:"pane_id"`
	TerminalID     *string           `json:"terminal_id"`
	WorkspaceID    *string           `json:"workspace_id"`
	TabID          *string           `json:"tab_id"`
	Name           *string           `json:"name"`
	Agent          *string           `json:"agent"`
	DisplayAgent   *string           `json:"display_agent"`
	AgentStatus    *string           `json:"agent_status"`
	Focused        *bool             `json:"focused"`
	Revision       *uint64           `json:"revision"`
	Cwd            *string           `json:"cwd"`
	ForegroundCwd  *string           `json:"foreground_cwd"`
	Interactive    bool              `json:"interactive_ready"`
	LaunchPending  bool              `json:"launch_pending"`
	StateLabels    map[string]string `json:"state_labels"`
	Tokens         map[string]string `json:"tokens"`
	StateChangeSeq uint64            `json:"state_change_seq"`
	AgentSession   json.RawMessage   `json:"agent_session"`
}

// AgentStart starts one agent in one pane.
func AgentStart(ctx context.Context, params AgentStartParams, opts ...Option) (AgentStarted, error) {
	var result struct {
		Type  *string         `json:"type"`
		Agent json.RawMessage `json:"agent"`
		Argv  json.RawMessage `json:"argv"`
	}
	if err := Call(ctx, "agent.start", params, &result, opts...); err != nil {
		return AgentStarted{}, err
	}
	if result.Type == nil || *result.Type != "agent_started" {
		return AgentStarted{}, fmt.Errorf("bad reply: result type %q, want agent_started", stringValue(result.Type))
	}
	agent, err := decodeAgentInfo(result.Agent)
	if err != nil {
		return AgentStarted{}, fmt.Errorf("bad reply: %s", err)
	}
	argv, err := decodeStringArray(result.Argv, "agent_started argv")
	if err != nil {
		return AgentStarted{}, fmt.Errorf("bad reply: %s", err)
	}
	return AgentStarted{Type: *result.Type, Agent: agent, Argv: argv}, nil
}

// AgentWait waits for one settled agent status.
func AgentWait(ctx context.Context, params AgentWaitParams, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  *string         `json:"type"`
		Agent json.RawMessage `json:"agent"`
	}
	if err := Call(ctx, "agent.wait", params, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type == nil || *result.Type != "agent_info" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_info", stringValue(result.Type))
	}
	agent, err := decodeAgentInfo(result.Agent)
	if err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return agent, nil
}

// AgentPrompt sends one prompt to one agent.
func AgentPrompt(ctx context.Context, params AgentPromptParams, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  *string         `json:"type"`
		Agent json.RawMessage `json:"agent"`
	}
	if err := Call(ctx, "agent.prompt", params, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type == nil || *result.Type != "agent_prompted" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_prompted", stringValue(result.Type))
	}
	agent, err := decodeAgentInfo(result.Agent)
	if err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return agent, nil
}

// AgentFocus focuses one agent and marks its pane as seen.
func AgentFocus(ctx context.Context, target string, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  *string         `json:"type"`
		Agent json.RawMessage `json:"agent"`
	}
	if err := Call(ctx, "agent.focus", AgentTarget{Target: target}, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type == nil || *result.Type != "agent_info" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_info", stringValue(result.Type))
	}
	agent, err := decodeAgentInfo(result.Agent)
	if err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return agent, nil
}

// AgentGet reads one agent.
func AgentGet(ctx context.Context, target string, opts ...Option) (AgentInfo, error) {
	var result struct {
		Type  *string         `json:"type"`
		Agent json.RawMessage `json:"agent"`
	}
	if err := Call(ctx, "agent.get", AgentTarget{Target: target}, &result, opts...); err != nil {
		return AgentInfo{}, err
	}
	if result.Type == nil || *result.Type != "agent_info" {
		return AgentInfo{}, fmt.Errorf("bad reply: result type %q, want agent_info", stringValue(result.Type))
	}
	agent, err := decodeAgentInfo(result.Agent)
	if err != nil {
		return AgentInfo{}, fmt.Errorf("bad reply: %s", err)
	}
	return agent, nil
}

// AgentList lists each live agent.
func AgentList(ctx context.Context, opts ...Option) ([]AgentInfo, error) {
	var result struct {
		Type   *string         `json:"type"`
		Agents json.RawMessage `json:"agents"`
	}
	if err := Call(ctx, "agent.list", struct{}{}, &result, opts...); err != nil {
		return nil, err
	}
	if result.Type == nil || *result.Type != "agent_list" {
		return nil, fmt.Errorf("bad reply: result type %q, want agent_list", stringValue(result.Type))
	}
	if isAbsentOrNull(result.Agents) {
		return nil, fmt.Errorf("bad reply: agent_list agents is missing")
	}
	var items []json.RawMessage
	if err := json.Unmarshal(result.Agents, &items); err != nil {
		return nil, fmt.Errorf("bad reply: agent_list agents: %s", err)
	}
	agents := make([]AgentInfo, 0, len(items))
	for index, item := range items {
		agent, err := decodeAgentInfo(item)
		if err != nil {
			return nil, fmt.Errorf("bad reply: agents[%d]: %s", index, err)
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func decodeAgentInfo(raw json.RawMessage) (AgentInfo, error) {
	if isAbsentOrNull(raw) {
		return AgentInfo{}, fmt.Errorf("agent is missing")
	}
	var wire agentInfoWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return AgentInfo{}, fmt.Errorf("agent: %s", err)
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
			return AgentInfo{}, fmt.Errorf("agent %s is missing or empty", field.name)
		}
	}
	if wire.Focused == nil {
		return AgentInfo{}, fmt.Errorf("agent focused is missing")
	}
	if wire.Revision == nil {
		return AgentInfo{}, fmt.Errorf("agent revision is missing")
	}
	if !validAgentStatus(*wire.AgentStatus) {
		return AgentInfo{}, fmt.Errorf("agent agent_status %q is not a protocol status", *wire.AgentStatus)
	}
	session, err := decodeOptionalSession(wire.AgentSession)
	if err != nil {
		return AgentInfo{}, err
	}
	info := AgentInfo{
		PaneID:         *wire.PaneID,
		TerminalID:     *wire.TerminalID,
		WorkspaceID:    *wire.WorkspaceID,
		TabID:          *wire.TabID,
		AgentStatus:    *wire.AgentStatus,
		Focused:        *wire.Focused,
		Revision:       *wire.Revision,
		Interactive:    wire.Interactive,
		LaunchPending:  wire.LaunchPending,
		StateLabels:    wire.StateLabels,
		Tokens:         wire.Tokens,
		StateChangeSeq: wire.StateChangeSeq,
		AgentSession:   session,
	}
	if wire.Name != nil {
		info.Name = *wire.Name
	}
	if wire.Agent != nil {
		info.Agent = *wire.Agent
	}
	if wire.DisplayAgent != nil {
		info.DisplayAgent = *wire.DisplayAgent
	}
	if wire.Cwd != nil {
		info.Cwd = *wire.Cwd
	}
	if wire.ForegroundCwd != nil {
		info.ForegroundCwd = *wire.ForegroundCwd
	}
	return info, nil
}

func decodeOptionalSession(raw json.RawMessage) (*AgentSessionInfo, error) {
	if isAbsentOrNull(raw) {
		return nil, nil
	}
	var session AgentSessionInfo
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("agent_session: %s", err)
	}
	if err := validateAgentSession(&session); err != nil {
		return nil, err
	}
	return &session, nil
}

func decodeStringArray(raw json.RawMessage, name string) ([]string, error) {
	if isAbsentOrNull(raw) {
		return nil, fmt.Errorf("%s is missing", name)
	}
	var items []*string
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("%s: %s", name, err)
	}
	values := make([]string, 0, len(items))
	for index, item := range items {
		if item == nil {
			return nil, fmt.Errorf("%s[%d] is null", name, index)
		}
		values = append(values, *item)
	}
	return values, nil
}

func validAgentStatus(status string) bool {
	switch status {
	case "idle", "working", "blocked", "done", "unknown":
		return true
	}
	return false
}

func isAbsentOrNull(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
