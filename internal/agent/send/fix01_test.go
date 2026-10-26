package send

import (
	"context"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"math"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestFix01Inspect(t *testing.T) {
	sleep := agenttest.StartSleep(t)
	cases := []struct {
		name       string
		paneReply  func(*agenttest.FakeHerdr)
		process    string
		wantStatus string
		wantError  string
	}{
		{
			name: "missing",
			paneReply: func(fake *agenttest.FakeHerdr) {
				fake.Failure("pane.get", "pane_not_found", "pane w1:p1 not found")
			},
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "empty",
			paneReply: func(fake *agenttest.FakeHerdr) {
				fake.Result("pane.get", `{"type":"pane_info","pane":`+agenttest.PaneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    agenttest.ProcessInfoShellJSON("w1:p1", 999999),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "unrelated",
			paneReply: func(fake *agenttest.FakeHerdr) {
				fake.Result("pane.get", `{"type":"pane_info","pane":`+agenttest.PaneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    agenttest.ProcessInfoShellJSON("w1:p1", 999999, agenttest.ProcessEntry(sleep.PID, "sleep", "sleep")),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "otherterm",
			paneReply: func(fake *agenttest.FakeHerdr) {
				fake.Result("pane.get", `{"type":"pane_info","pane":`+agenttest.PaneInfoJSON("w1:p1", "term_other", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    agenttest.ProcessInfoShellJSON("w1:p1", sleep.PID, agenttest.ProcessEntry(sleep.PID, "bash", "bash")),
			wantStatus: records.AgentUnresolved,
			wantError:  "agent unresolved",
		},
		{
			name: "shell",
			paneReply: func(fake *agenttest.FakeHerdr) {
				fake.Result("pane.get", `{"type":"pane_info","pane":`+agenttest.PaneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			},
			process:    agenttest.ProcessInfoShellJSON("w1:p1", sleep.PID, agenttest.ProcessEntry(sleep.PID, "bash", "bash")),
			wantStatus: records.AgentExited,
			wantError:  "agent exited",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := agenttest.NewHome(t)
			record := agenttest.LiveRecord("worker-hive-1", "worker-hive-1")
			record.TerminalID = "term_a"
			agenttest.WriteRecord(t, home, record)
			fake := agenttest.NewFakeHerdr(t)
			fake.Failure("agent.prompt", "agent_not_found", "agent target w1:p1 not found")
			test.paneReply(fake)
			if test.process != "" {
				fake.Result("pane.process_info", `{"type":"pane_process_info","process_info":`+test.process+`}`)
			}
			envelope := Run(context.Background(), home, Options{Agent: "worker-hive-1", Text: "x"})
			if envelope.Outcome != "failed" {
				t.Fatalf("outcome = %q; want failed (%#v)", envelope.Outcome, envelope)
			}
			if !strings.Contains(agenttest.ToString(envelope.Error), test.wantError) {
				t.Fatalf("error = %v; want %q", envelope.Error, test.wantError)
			}
			if got := agenttest.ReadAgent(t, home, "worker-hive-1").Status; got != test.wantStatus {
				t.Fatalf("record status = %q; want %q", got, test.wantStatus)
			}
			if len(fake.Recorded("pane.close")) != 0 {
				t.Fatalf("pane.close ran during inspection")
			}
		})
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
			home := agenttest.NewHome(t)
			agenttest.WriteRecord(t, home, agenttest.LiveRecord("worker-hive-1", "worker-hive-1"))
			fake := agenttest.NewFakeHerdr(t)
			fake.Result("agent.prompt", `{"type":"agent_prompted","agent":`+agenttest.AgentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
			envelope := Run(context.Background(), home, Options{Agent: "worker-hive-1", Text: "x", Wait: true, Timeout: test.timeout, TimeoutSet: true})
			requests := fake.Recorded("agent.prompt")
			if !test.wantOK {
				if envelope.Outcome != "failed" {
					t.Fatalf("outcome = %q; want failed (%#v)", envelope.Outcome, envelope)
				}
				if len(requests) != 0 {
					t.Fatalf("agent.prompt requests = %d; want 0", len(requests))
				}
				if record := agenttest.ReadAgent(t, home, "worker-hive-1"); record.Status != records.AgentLive {
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
