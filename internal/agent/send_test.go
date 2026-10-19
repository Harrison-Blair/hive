package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func liveRecord(key, name string) records.Agent {
	return records.Agent{
		Key:     key,
		Name:    name,
		Role:    "worker",
		Project: "hive",
		Task:    "12-spawn-worker",
		Harness: "claude",
		PaneID:  "w1:p1",
		Status:  records.AgentLive,
		Spawned: true,
	}
}

func TestSendNoWaitShape(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "done", "worker-hive-1")+`}`)
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "Reply with pong."})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(SendResult)
	if !ok {
		t.Fatalf("result = %T; want SendResult", envelope.Result)
	}
	if result.Status != "done" || result.Agent.Name != "worker-hive-1" {
		t.Fatalf("result = %#v", result)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
	}
	if !strings.Contains(requests[0].Params, `"text":"⬢ Reply with pong."`) {
		t.Fatalf("params = %s; want the hive marker before the message", requests[0].Params)
	}
	want := `{"target":"w1:p1","text":"⬢ Reply with pong."}`
	if requests[0].Params != want {
		t.Fatalf("params = %s; want %s", requests[0].Params, want)
	}
}

func TestSendDoesNotDuplicateHiveMarker(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	for range 2 {
		fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "done", "worker-hive-1")+`}`)
	}
	for index, text := range []string{"Reply with pong.", "⬢ Reply with pong."} {
		if envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: text}); envelope.Outcome != "ok" {
			t.Fatalf("envelope %d = %#v; want ok", index+1, envelope)
		}
		requests := fake.recorded("agent.prompt")
		want := `{"target":"w1:p1","text":"⬢ Reply with pong."}`
		if len(requests) != index+1 || requests[index].Params != want {
			t.Fatalf("params = %#v; want one marker prefix", requests)
		}
	}
}

func TestSendWaitShapes(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	if envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true}); envelope.Outcome != "ok" {
		t.Fatalf("default wait envelope = %#v", envelope)
	}
	if envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true, Timeout: 120, TimeoutSet: true}); envelope.Outcome != "ok" {
		t.Fatalf("explicit wait envelope = %#v", envelope)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 2 {
		t.Fatalf("requests = %d; want 2", len(requests))
	}
	wantDefault := `{"target":"w1:p1","text":"⬢ x","wait":{"until":["idle","blocked","done"],"timeout_ms":600000}}`
	if requests[0].Params != wantDefault {
		t.Fatalf("default params = %s; want %s", requests[0].Params, wantDefault)
	}
	wantExplicit := `{"target":"w1:p1","text":"⬢ x","wait":{"until":["idle","blocked","done"],"timeout_ms":120000}}`
	if requests[1].Params != wantExplicit {
		t.Fatalf("explicit params = %s; want %s", requests[1].Params, wantExplicit)
	}
}

func TestSendErrorMapping(t *testing.T) {
	cases := []struct {
		name       string
		code       string
		message    string
		outcome    string
		wantText   string
		wantStatus string
	}{
		{name: "blocked", code: "agent_blocked", message: "agent is blocked", outcome: "failed", wantText: "agent_blocked", wantStatus: records.AgentLive},
		{name: "stalled", code: "agent_prompt_stalled", message: "no working state", outcome: "failed", wantText: "input", wantStatus: records.AgentLive},
		{name: "not ready", code: "agent_not_ready", message: "not an active named agent", outcome: "failed", wantText: "agent_not_ready", wantStatus: records.AgentLive},
		{name: "timeout", code: "timeout", message: "timed out", outcome: "partial", wantText: "timeout", wantStatus: records.AgentLive},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
			fake := newFakeHerdr(t)
			fake.failure("agent.prompt", test.code, test.message)
			envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true})
			if envelope.Outcome != test.outcome {
				t.Fatalf("outcome = %q; want %q (%#v)", envelope.Outcome, test.outcome, envelope)
			}
			if !strings.Contains(strings.ToLower(toString(envelope.Error)), strings.ToLower(test.wantText)) {
				t.Fatalf("error = %v; want %q", envelope.Error, test.wantText)
			}
			record := readAgent(t, home, "worker-hive-1")
			if record.Status != test.wantStatus {
				t.Fatalf("record status = %q; want %q", record.Status, test.wantStatus)
			}
			if test.code == "agent_blocked" || test.code == "agent_prompt_stalled" || test.code == "agent_not_ready" {
				if len(fake.recorded("pane.get")) != 0 {
					t.Fatalf("pane.get ran for %s", test.code)
				}
			}
		})
	}
}

func TestSendAgentNotFoundInspectsExited(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := liveRecord("worker-hive-1", "worker-hive-1")
	record.TerminalID = "term_a"
	writeRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.failure("agent.prompt", "agent_not_found", "agent target w1:p1 not found")
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoShellJSON("w1:p1", sleep.pid, processEntry(sleep.pid, "bash", "bash"))+`}`)
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x"})
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "agent exited") {
		t.Fatalf("envelope = %#v", envelope)
	}
	stored := readAgent(t, home, "worker-hive-1")
	if stored.Status != records.AgentExited {
		t.Fatalf("record status = %q; want exited", stored.Status)
	}
}

func TestSendNonLiveRecord(t *testing.T) {
	home := testHome(t)
	record := liveRecord("worker-hive-1", "worker-hive-1")
	record.Status = records.AgentStopped
	writeRecord(t, home, record)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x"})
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "agent is stopped") {
		t.Fatalf("envelope = %#v", envelope)
	}
	if len(fake.recorded("agent.prompt")) != 0 {
		t.Fatalf("agent.prompt ran for a stopped record")
	}
}

func TestSendTransportErrorIsPartial(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.drop("agent.prompt")
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true})
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if len(fake.recorded("agent.prompt")) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(fake.recorded("agent.prompt")))
	}
}

func TestSendNoRetry(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.failure("agent.prompt", "agent_prompt_stalled", "stalled")
	_ = Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true})
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1 and no retry", got)
	}
}

func TestSendLogsTextBody(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	text := "T10SENDBODYSENTINEL"
	if envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: text}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	logText := readLog(t, home)
	if !strings.Contains(logText, text) {
		t.Fatalf("log does not contain the prompt body")
	}
}

func TestSendEmptyText(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: ""})
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if len(fake.recorded("agent.prompt")) != 0 {
		t.Fatalf("agent.prompt ran for empty text")
	}
}
