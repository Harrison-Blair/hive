package herdr

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPaneGetRequestAndReply(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp/x","agent_session":{"source":"herdr:claude","agent":"claude","kind":"path","value":"/tmp/s"}}}`)
	pane, err := PaneGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("PaneGet error = %v", err)
	}
	if pane.Cwd != "/tmp/x" {
		t.Fatalf("Cwd = %q", pane.Cwd)
	}
	if pane.AgentSession == nil || pane.AgentSession.Kind != "path" {
		t.Fatalf("AgentSession = %#v", pane.AgentSession)
	}
	requests := server.recorded("pane.get")
	if len(requests) != 1 {
		t.Fatalf("pane.get requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"pane_id":"w1:p1"}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
}

func TestPaneGetRejectsWrongTypeAndMissingPane(t *testing.T) {
	cases := []string{
		`{"type":"pane_info","pane":null}`,
		`{"type":"workspace_info","pane":{}}`,
		`{"type":"pane_info","pane":{"pane_id":"","terminal_id":"","workspace_id":"","tab_id":"","agent_status":""}}`,
	}
	for _, reply := range cases {
		server := newFakeServer(t)
		server.scriptResult("pane.get", reply)
		if _, err := PaneGet(context.Background(), "w1:p1"); err == nil {
			t.Fatalf("PaneGet(%s) error = nil; want an error", reply)
		}
	}
}

func TestPaneClose(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.close", `{"type":"ok"}`)
	if err := PaneClose(context.Background(), "w1:p1"); err != nil {
		t.Fatalf("PaneClose error = %v", err)
	}
	requests := server.recorded("pane.close")
	if len(requests) != 1 {
		t.Fatalf("pane.close requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"pane_id":"w1:p1"}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
	server.scriptResult("pane.close", `{"type":"pane_info"}`)
	if err := PaneClose(context.Background(), "w1:p1"); err == nil {
		t.Fatalf("PaneClose wrong type error = nil")
	}
	server.scriptError("pane.close", "pane_not_found", "pane w1:p1 not found")
	err := PaneClose(context.Background(), "w1:p1")
	var herdrError *Error
	if !errors.As(err, &herdrError) || herdrError.Code != "pane_not_found" {
		t.Fatalf("PaneClose error = %v; want pane_not_found", err)
	}
}

func TestPaneProcessInfoShape(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","shell_pid":null,"foreground_process_group_id":null,"foreground_processes":[{"pid":42,"name":"claude","argv":["claude"],"cmdline":"claude","cwd":"/tmp/x"}]}}`)
	info, err := PaneProcessInfoGet(context.Background(), "w1:p1")
	if err != nil {
		t.Fatalf("PaneProcessInfoGet error = %v", err)
	}
	if info.PaneID != "w1:p1" {
		t.Fatalf("PaneID = %q", info.PaneID)
	}
	if info.ShellPID != nil || info.ForegroundProcessGroupID != nil {
		t.Fatalf("nullable fields = %#v; want nil", info)
	}
	if len(info.ForegroundProcesses) != 1 || info.ForegroundProcesses[0].PID != 42 || info.ForegroundProcesses[0].Name != "claude" {
		t.Fatalf("processes = %#v", info.ForegroundProcesses)
	}
	if !reflect.DeepEqual(info.ForegroundProcesses[0].Argv, []string{"claude"}) {
		t.Fatalf("argv = %#v", info.ForegroundProcesses[0].Argv)
	}
	requests := server.recorded("pane.process_info")
	if len(requests) != 1 {
		t.Fatalf("pane.process_info requests = %d; want 1", len(requests))
	}
	if got, want := requests[0].Params, `{"pane_id":"w1:p1"}`; got != want {
		t.Fatalf("params = %s; want %s", got, want)
	}
}

func TestPaneProcessInfoRejectsWrongType(t *testing.T) {
	for _, reply := range []string{`{"type":"pane_info","process_info":{}}`, `{"type":"pane_process_info","process_info":null}`, `{"type":"pane_process_info","process_info":{"pane_id":""}}`} {
		server := newFakeServer(t)
		server.scriptResult("pane.process_info", reply)
		if _, err := PaneProcessInfoGet(context.Background(), "w1:p1"); err == nil {
			t.Fatalf("PaneProcessInfoGet(%s) error = nil; want an error", reply)
		}
	}
}
