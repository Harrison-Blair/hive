package herdr

import (
	"context"
	"slices"
	"testing"
)

func TestPaneProcessInfoRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w2:p1","shell_pid":215879,"foreground_process_group_id":217209,"foreground_processes":[{"argv":["claude","--model","m"],"cmdline":"claude --model m","cwd":"/home/user/project","name":"claude","pid":217209}]}}`)
	info, err := PaneProcessInfo(context.Background(), "w2:p1")
	if err != nil {
		t.Fatalf("PaneProcessInfo error = %v", err)
	}
	if info.ForegroundProcessGroupID != 217209 || len(info.ForegroundProcesses) != 1 {
		t.Fatalf("info = %#v", info)
	}
	process := info.ForegroundProcesses[0]
	if process.PID != 217209 || process.Name != "claude" || !slices.Equal(process.Argv, []string{"claude", "--model", "m"}) {
		t.Fatalf("process = %#v", process)
	}
	requests := server.recorded("pane.process_info")
	if len(requests) != 1 || requests[0].Params != `{"pane_id":"w2:p1"}` {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestPaneProcessInfoRejectsWrongType(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.process_info", `{"type":"pane_info","process_info":{}}`)
	if _, err := PaneProcessInfo(context.Background(), "w2:p1"); err == nil {
		t.Fatalf("PaneProcessInfo error = nil; want a bad reply")
	}
}

func TestAgentWaitRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("agent.wait", `{"type":"agent_info","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"done","focused":false,"revision":4}}`)
	info, err := AgentWait(context.Background(), AgentWaitParams{Target: "w1:p1", Until: []string{"working"}, TimeoutMS: 500})
	if err != nil {
		t.Fatalf("AgentWait error = %v", err)
	}
	if info.AgentStatus != "done" {
		t.Fatalf("status = %q", info.AgentStatus)
	}
	requests := server.recorded("agent.wait")
	if want := `{"target":"w1:p1","until":["working"],"timeout_ms":500}`; len(requests) != 1 || requests[0].Params != want {
		t.Fatalf("requests = %#v; want params %s", requests, want)
	}
}
