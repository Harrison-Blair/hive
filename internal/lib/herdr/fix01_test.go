package herdr

import (
	"context"
	"strings"
	"testing"
)

const validAgentInfo = `{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":0,"extra":"allowed"}`

func TestFix01AgentWireValidation(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		reply   string
		call    func() error
		wantErr string
	}{
		{
			name:   "get missing focused",
			method: "agent.get",
			reply:  `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","revision":1}}`,
			call:   func() error { _, err := AgentGet(context.Background(), "w1:p1"); return err },
		},
		{
			name:   "get null revision",
			method: "agent.get",
			reply:  `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":null}}`,
			call:   func() error { _, err := AgentGet(context.Background(), "w1:p1"); return err },
		},
		{
			name:   "get invalid status",
			method: "agent.get",
			reply:  `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"settled","focused":false,"revision":1}}`,
			call:   func() error { _, err := AgentGet(context.Background(), "w1:p1"); return err },
		},
		{
			name:   "list missing agents",
			method: "agent.list",
			reply:  `{"type":"agent_list"}`,
			call:   func() error { _, err := AgentList(context.Background()); return err },
		},
		{
			name:   "list null agents",
			method: "agent.list",
			reply:  `{"type":"agent_list","agents":null}`,
			call:   func() error { _, err := AgentList(context.Background()); return err },
		},
		{
			name:   "start null argv item",
			method: "agent.start",
			reply:  `{"type":"agent_started","agent":` + validAgentInfo + `,"argv":[null]}`,
			call: func() error {
				_, err := AgentStart(context.Background(), AgentStartParams{Name: "n", Kind: "claude", PaneID: "w1:p1"})
				return err
			},
		},
		{
			name:   "start missing argv",
			method: "agent.start",
			reply:  `{"type":"agent_started","agent":` + validAgentInfo + `}`,
			call: func() error {
				_, err := AgentStart(context.Background(), AgentStartParams{Name: "n", Kind: "claude", PaneID: "w1:p1"})
				return err
			},
		},
		{
			name:   "pane missing focused",
			method: "pane.get",
			reply:  `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"unknown","revision":1}}`,
			call:   func() error { _, err := PaneGet(context.Background(), "w1:p1"); return err },
		},
		{
			name:   "process entry missing pid",
			method: "pane.process_info",
			reply:  `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","foreground_processes":[{"name":"bash"}]}}`,
			call: func() error {
				_, err := PaneProcessInfoGet(context.Background(), "w1:p1")
				return err
			},
		},
		{
			name:   "process entry null pid",
			method: "pane.process_info",
			reply:  `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","foreground_processes":[{"pid":null,"name":"bash"}]}}`,
			call: func() error {
				_, err := PaneProcessInfoGet(context.Background(), "w1:p1")
				return err
			},
		},
		{
			name:   "process entry null name",
			method: "pane.process_info",
			reply:  `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","foreground_processes":[{"pid":1,"name":null}]}}`,
			call: func() error {
				_, err := PaneProcessInfoGet(context.Background(), "w1:p1")
				return err
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			server := newFakeServer(t)
			server.scriptResult(test.method, test.reply)
			err := test.call()
			if err == nil {
				t.Fatalf("error = nil; want a refusal")
			}
			if !strings.Contains(err.Error(), "bad reply") {
				t.Fatalf("error = %v; want a bad reply error", err)
			}
		})
	}
}

func TestFix01AgentWireAcceptsValidReplies(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.get", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"unknown","focused":false,"revision":0,"name":null,"agent":null,"cwd":null,"agent_session":null,"extra":{"any":true}}}`)
	info, err := AgentGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("AgentGet error = %v", err)
	}
	if info.PaneID != "w1:p1" || info.Revision != 0 || info.Focused {
		t.Fatalf("info = %#v; want a valid zero revision and false focus", info)
	}
	server.scriptResult("agent.list", `{"type":"agent_list","agents":[]}`)
	agents, err := AgentList(context.Background())
	if err != nil {
		t.Fatalf("AgentList error = %v", err)
	}
	if len(agents) != 0 {
		t.Fatalf("agents = %#v; want the valid empty list", agents)
	}
	server.scriptResult("agent.start", `{"type":"agent_started","agent":`+validAgentInfo+`,"argv":[],"extra":1}`)
	started, err := AgentStart(context.Background(), AgentStartParams{Name: "n", Kind: "claude", PaneID: "w1:p1"})
	if err != nil {
		t.Fatalf("AgentStart error = %v", err)
	}
	if len(started.Argv) != 0 {
		t.Fatalf("argv = %#v; want the valid empty array", started.Argv)
	}
	server.scriptResult("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","shell_pid":null,"foreground_processes":[],"tty":null}}`)
	if _, err := PaneProcessInfoGet(context.Background(), "w1:p1"); err != nil {
		t.Fatalf("PaneProcessInfoGet error = %v", err)
	}
}
