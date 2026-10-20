package herdr

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestAgentWaitRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.wait", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":2,"name":"coordinator-g1","agent":"claude","agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"s1"}}}`)
	info, err := AgentWait(context.Background(), AgentWaitParams{Target: "w1:p1", Until: []string{"idle", "blocked", "done"}, TimeoutMS: 120000}, WithDeadline(125*time.Second))
	if err != nil {
		t.Fatalf("AgentWait error = %v", err)
	}
	if info.AgentStatus != "idle" {
		t.Fatalf("AgentStatus = %q", info.AgentStatus)
	}
	if info.AgentSession == nil || info.AgentSession.Value != "s1" {
		t.Fatalf("AgentSession = %#v", info.AgentSession)
	}
	requests := server.recorded("agent.wait")
	if got, want := requests[0].Params, `{"target":"w1:p1","until":["idle","blocked","done"],"timeout_ms":120000}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
}

func TestAgentWaitRejectsMalformedAgentInfo(t *testing.T) {
	cases := []string{
		`{"type":"agent_info","agent":null}`,
		`{"type":"agent_info"}`,
		`{"type":"agent_info","agent":{"pane_id":"","terminal_id":"","agent_status":""}}`,
		`{"type":"wait_matched","event":{}}`,
		`null`,
	}
	for _, reply := range cases {
		t.Run(reply, func(t *testing.T) {
			server := newFakeServer(t)
			server.scriptResult("agent.wait", reply)
			_, err := AgentWait(context.Background(), AgentWaitParams{Target: "w1:p1"})
			if err == nil {
				t.Fatalf("AgentWait error = nil; want an error")
			}
		})
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

func TestAgentFocusRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.focus", `{"type":"agent_info","agent":{"pane_id":"w1:pA","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":true,"revision":7}}`)
	info, err := AgentFocus(context.Background(), "w1:pA")
	if err != nil {
		t.Fatalf("AgentFocus error = %v", err)
	}
	if info.PaneID != "w1:pA" || !info.Focused {
		t.Fatalf("focused agent = %+v", info)
	}
	requests := server.recorded("agent.focus")
	if len(requests) != 1 || requests[0].Params != `{"target":"w1:pA"}` {
		t.Fatalf("agent.focus requests = %+v; want target w1:pA", requests)
	}
}

func TestAgentGetAndList(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.get", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":5,"name":"n"}}`)
	server.scriptResult("agent.list", `{"type":"agent_list","agents":[{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":5}]}`)
	info, err := AgentGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("AgentGet error = %v", err)
	}
	if info.Name != "n" {
		t.Fatalf("Name = %q", info.Name)
	}
	agents, err := AgentList(context.Background())
	if err != nil {
		t.Fatalf("AgentList error = %v", err)
	}
	if len(agents) != 1 || agents[0].PaneID != "w1:p1" {
		t.Fatalf("agents = %#v", agents)
	}
	getRequests := server.recorded("agent.get")
	if len(getRequests) != 1 {
		t.Fatalf("agent.get requests = %d; want 1", len(getRequests))
	}
	if got, want := getRequests[0].Params, `{"target":"w1:p1"}`; got != want {
		t.Fatalf("agent.get params = %s; want %s", got, want)
	}
	listRequests := server.recorded("agent.list")
	if len(listRequests) != 1 {
		t.Fatalf("agent.list requests = %d; want 1", len(listRequests))
	}
	if got, want := listRequests[0].Params, `{}`; got != want {
		t.Fatalf("agent.list params = %s; want %s", got, want)
	}
}
