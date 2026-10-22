package agent

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func processInfoShellJSON(paneID string, shellPID int64, entries ...string) string {
	body := `{"pane_id":"` + paneID + `","shell_pid":` + jsonNumber(shellPID) + `,"foreground_process_group_id":null,"foreground_processes":[`
	for index, entry := range entries {
		if index > 0 {
			body += ","
		}
		body += entry
	}
	return body + `]}`
}

func TestFix01Inspect(t *testing.T) {
	sleep := startSleep(t)
	cases := []struct {
		name       string
		paneReply  func(*fakeHerdr)
		process    string
		wantStatus string
		wantError  string
	}{
		{
			name: "missing",
			paneReply: func(fake *fakeHerdr) {
				fake.failure("pane.get", "pane_not_found", "pane w1:p1 not found")
			},
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "empty",
			paneReply: func(fake *fakeHerdr) {
				fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    processInfoShellJSON("w1:p1", 999999),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "unrelated",
			paneReply: func(fake *fakeHerdr) {
				fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    processInfoShellJSON("w1:p1", 999999, processEntry(sleep.pid, "sleep", "sleep")),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "otherterm",
			paneReply: func(fake *fakeHerdr) {
				fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_other", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    processInfoShellJSON("w1:p1", sleep.pid, processEntry(sleep.pid, "bash", "bash")),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "shell",
			paneReply: func(fake *fakeHerdr) {
				fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    processInfoShellJSON("w1:p1", sleep.pid, processEntry(sleep.pid, "bash", "bash")),
			wantStatus: records.AgentExited,
			wantError:  "agent exited",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			record := liveRecord("worker-hive-1", "worker-hive-1")
			record.TerminalID = "term_a"
			writeRecord(t, home, record)
			fake := newFakeHerdr(t)
			fake.failure("agent.prompt", "agent_not_found", "agent target w1:p1 not found")
			test.paneReply(fake)
			if test.process != "" {
				fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+test.process+`}`)
			}
			envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x"})
			if envelope.Outcome != "failed" {
				t.Fatalf("outcome = %q; want failed (%#v)", envelope.Outcome, envelope)
			}
			if !strings.Contains(toString(envelope.Error), test.wantError) {
				t.Fatalf("error = %v; want %q", envelope.Error, test.wantError)
			}
			if got := readAgent(t, home, "worker-hive-1").Status; got != test.wantStatus {
				t.Fatalf("record status = %q; want %q", got, test.wantStatus)
			}
			if len(fake.recorded("pane.close")) != 0 {
				t.Fatalf("pane.close ran during inspection")
			}
		})
	}
}

func TestFix02SpawnRetainsStartSession(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1", "herdr:claude", "id", "start-session")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "different-session")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "different-session")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome == "ok" {
		t.Fatalf("envelope = %#v; want a refusal after a changed start session", envelope)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", record.Status)
	}
	if record.ProcessPID != 0 {
		t.Fatalf("record process pid = %d; want no later identity write", record.ProcessPID)
	}
	if !hasEffect(envelope.Effects, "agent", "started") {
		t.Fatalf("effects = %#v; want the started effect", envelope.Effects)
	}
}

func TestFix03StopRefusesChangedIdentity(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := liveRecord("worker-hive-1", "worker-hive-1")
	record.TerminalID = "term_a"
	record.WorkspaceID = "w1"
	record.ProcessPID = sleep.pid
	record.ProcessName = "claude"
	record.ProcessStart = sleep.start
	writeRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	changed := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "workspace.get" || changed {
			return
		}
		changed = true
		current := readAgent(t, home, "worker-hive-1")
		current.TerminalID = "term_other"
		current.ProcessStart = sleep.start + 1
		writeRecord(t, home, current)
	})
	_, envelope := Stop(context.Background(), home, readAgent(t, home, "worker-hive-1"))
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed (%#v)", envelope.Outcome, envelope)
	}
	if len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("pane.close ran after the checked identity changed")
	}
	stored := readAgent(t, home, "worker-hive-1")
	if stored.Status != records.AgentLive {
		t.Fatalf("record status = %q; want the changed live record", stored.Status)
	}
	if stored.TerminalID != "term_other" || stored.ProcessStart != sleep.start+1 {
		t.Fatalf("record identity = %q/%d; want the changed values", stored.TerminalID, stored.ProcessStart)
	}
}

func TestFix02WaitTimeoutBoundary(t *testing.T) {
	cases := []struct {
		name       string
		timeout    int
		wantOK     bool
		wantMillis string
	}{
		{name: "largest representable", timeout: 9223372031, wantOK: true, wantMillis: "9223372031000"},
		{name: "duration overflows", timeout: 9223372032, wantOK: false},
		{name: "max int overflows", timeout: math.MaxInt64, wantOK: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
			fake := newFakeHerdr(t)
			fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
			envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: "x", Wait: true, Timeout: test.timeout, TimeoutSet: true})
			requests := fake.recorded("agent.prompt")
			if !test.wantOK {
				if envelope.Outcome != "failed" {
					t.Fatalf("outcome = %q; want failed (%#v)", envelope.Outcome, envelope)
				}
				if len(requests) != 0 {
					t.Fatalf("agent.prompt requests = %d; want 0", len(requests))
				}
				if record := readAgent(t, home, "worker-hive-1"); record.Status != records.AgentLive {
					t.Fatalf("record status = %q; want unchanged live", record.Status)
				}
				return
			}
			if envelope.Outcome != "ok" {
				t.Fatalf("outcome = %q; want ok (%#v)", envelope.Outcome, envelope)
			}
			if len(requests) != 1 {
				t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
			}
			if !strings.Contains(requests[0].Params, `"timeout_ms":`+test.wantMillis) {
				t.Fatalf("params = %s; want timeout_ms %s", requests[0].Params, test.wantMillis)
			}
		})
	}
}
