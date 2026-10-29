package herdr

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestAgentStartRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.start", `{"type":"agent_started","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"unknown","focused":false,"revision":1,"name":"coordinator-g1","agent":"claude"},"argv":["claude","--model","m"]}`)
	started, err := AgentStart(context.Background(), AgentStartParams{Name: "coordinator-g1", Kind: "claude", PaneID: "w1:p1", Args: []string{"--model", "m"}, TimeoutMS: 120000})
	if err != nil {
		t.Fatalf("AgentStart error = %v", err)
	}
	if started.Type != "agent_started" {
		t.Fatalf("Type = %q", started.Type)
	}
	if started.Agent.PaneID != "w1:p1" || started.Agent.Name != "coordinator-g1" {
		t.Fatalf("Agent = %#v", started.Agent)
	}
	if !reflect.DeepEqual(started.Argv, []string{"claude", "--model", "m"}) {
		t.Fatalf("Argv = %#v", started.Argv)
	}
	requests := server.recorded("agent.start")
	if len(requests) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(requests))
	}
	want := `{"name":"coordinator-g1","kind":"claude","pane_id":"w1:p1","args":["--model","m"],"timeout_ms":120000}`
	if requests[0].Params != want {
		t.Fatalf("params = %s; want %s", requests[0].Params, want)
	}
}

func TestAgentStartTypedError(t *testing.T) {
	server := newFakeServer(t)
	server.scriptError("agent.start", "invalid_agent_name", "bad name")
	_, err := AgentStart(context.Background(), AgentStartParams{Name: "x", Kind: "claude", PaneID: "w1:p1"})
	var herdrError *Error
	if !errors.As(err, &herdrError) {
		t.Fatalf("error = %v; want *herdr.Error", err)
	}
	if herdrError.Code != "invalid_agent_name" {
		t.Fatalf("code = %q", herdrError.Code)
	}
}

func TestAgentPromptShapes(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.prompt", `{"type":"agent_prompted","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":3}}`)
	server.scriptResult("agent.prompt", `{"type":"agent_prompted","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"done","focused":false,"revision":4}}`)
	if _, err := AgentPrompt(context.Background(), AgentPromptParams{Target: "w1:p1", Text: "hello"}); err != nil {
		t.Fatalf("AgentPrompt without wait error = %v", err)
	}
	wait := &AgentPromptWaitOptions{Until: []string{"idle", "blocked", "done"}, TimeoutMS: 600000}
	if _, err := AgentPrompt(context.Background(), AgentPromptParams{Target: "w1:p1", Text: "hello", Wait: wait}); err != nil {
		t.Fatalf("AgentPrompt with wait error = %v", err)
	}
	requests := server.recorded("agent.prompt")
	if len(requests) != 2 {
		t.Fatalf("agent.prompt requests = %d; want 2", len(requests))
	}
	if got, want := requests[0].Params, `{"target":"w1:p1","text":"hello"}`; got != want {
		t.Fatalf("no-wait params = %s; want %s", got, want)
	}
	if strings.Contains(requests[0].Params, "wait") {
		t.Fatalf("no-wait params = %s; must not contain wait", requests[0].Params)
	}
	if got, want := requests[1].Params, `{"target":"w1:p1","text":"hello","wait":{"until":["idle","blocked","done"],"timeout_ms":600000}}`; got != want {
		t.Fatalf("wait params = %s; want %s", got, want)
	}
}

func TestAgentGetPreservesReadinessFields(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.get", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":5,"launch_pending":true,"interactive_ready":false}}`)
	server.scriptResult("agent.get", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":6,"launch_pending":false,"interactive_ready":true}}`)
	for index, want := range []map[string]any{
		{"launch_pending": true},
		{"interactive_ready": true},
	} {
		info, err := AgentGet(context.Background(), "w1:p1")
		if err != nil {
			t.Fatalf("AgentGet(%d) error = %v", index, err)
		}
		data, err := json.Marshal(info)
		if err != nil {
			t.Fatalf("marshal AgentInfo: %v", err)
		}
		var fields map[string]any
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatalf("decode AgentInfo: %v", err)
		}
		for key, value := range want {
			if fields[key] != value {
				t.Fatalf("AgentInfo readiness fields = %#v; want %s=%v", fields, key, value)
			}
		}
	}
}

func TestAgentGet(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.get", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":5,"name":"n"}}`)
	info, err := AgentGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("AgentGet error = %v", err)
	}
	if info.Name != "n" {
		t.Fatalf("Name = %q", info.Name)
	}
	getRequests := server.recorded("agent.get")
	if len(getRequests) != 1 {
		t.Fatalf("agent.get requests = %d; want 1", len(getRequests))
	}
	if got, want := getRequests[0].Params, `{"target":"w1:p1"}`; got != want {
		t.Fatalf("agent.get params = %s; want %s", got, want)
	}
}
