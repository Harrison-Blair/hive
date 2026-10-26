package resolve

import (
	"context"
	"fmt"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func liveAgentInfo(name, paneID, terminalID, workspaceID, tabID, kind, cwd string) string {
	return `{"type":"agent_info","agent":{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"idle","focused":false,"revision":2,"name":"` + name + `","agent":"` + kind + `","cwd":"` + cwd + `"}}`
}

func TestResolve(t *testing.T) {
	t.Run("exited sets exited", func(t *testing.T) {
		home := agenttest.NewHome(t)
		agenttest.WriteRecord(t, home, agenttest.ResolveSeedRecord())
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentExited})
		if envelope.Outcome != "ok" || envelope.Phase != "complete" {
			t.Fatalf("Run() = %+v, want ok complete", envelope)
		}
		if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentExited {
			t.Fatalf("status = %q, want exited", got.Status)
		}
	})

	t.Run("live matches and updates the ids", func(t *testing.T) {
		home := agenttest.NewHome(t)
		agenttest.WriteRecord(t, home, agenttest.ResolveSeedRecord())
		fake := agenttest.NewFakeHerdr(t)
		fake.Result("agent.get", liveAgentInfo("worker-3", "w2:p5", "term_a", "w2", "w2:t3", "claude", "/tmp/old"))
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "ok" || envelope.Phase != "complete" {
			t.Fatalf("Run() = %+v, want ok complete", envelope)
		}
		got := agenttest.ReadAgent(t, home, "worker-3")
		if got.Status != records.AgentLive {
			t.Fatalf("status = %q, want live", got.Status)
		}
		if got.PaneID != "w2:p5" || got.WorkspaceID != "w2" || got.TabID != "w2:t3" || got.TerminalID != "term_a" {
			t.Fatalf("ids = %q/%q/%q/%q, want the live ids", got.PaneID, got.WorkspaceID, got.TabID, got.TerminalID)
		}
	})

	t.Run("live without a name targets the pane", func(t *testing.T) {
		home := agenttest.NewHome(t)
		record := agenttest.ResolveSeedRecord()
		record.Name = ""
		agenttest.WriteRecord(t, home, record)
		fake := agenttest.NewFakeHerdr(t)
		fake.Result("agent.get", liveAgentInfo("", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/old"))
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "ok" {
			t.Fatalf("Run() = %+v, want ok", envelope)
		}
		requests := fake.Recorded("agent.get")
		if len(requests) != 1 || !strings.Contains(requests[0].Params, `"target":"w1:p1"`) {
			t.Fatalf("agent.get params = %+v, want the pane target", requests)
		}
	})

	t.Run("live matches the process identity", func(t *testing.T) {
		home := agenttest.NewHome(t)
		sleep := agenttest.StartSleep(t)
		record := agenttest.ResolveSeedRecord()
		record.ProcessPID = sleep.PID
		record.ProcessStart = sleep.Start
		agenttest.WriteRecord(t, home, record)
		fake := agenttest.NewFakeHerdr(t)
		fake.Result("agent.get", liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/old"))
		fake.Result("pane.process_info", `{"type":"pane_process_info","process_info":`+agenttest.ProcessInfoJSON("w1:p1", agenttest.ProcessEntry(sleep.PID, "claude", "claude"))+`}`)
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "ok" {
			t.Fatalf("Run() = %+v, want ok", envelope)
		}
		if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentLive {
			t.Fatalf("status = %q, want live", got.Status)
		}
	})

	t.Run("live without process fields skips the process read", func(t *testing.T) {
		home := agenttest.NewHome(t)
		agenttest.WriteRecord(t, home, agenttest.ResolveSeedRecord())
		fake := agenttest.NewFakeHerdr(t)
		fake.Result("agent.get", liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/old"))
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "ok" {
			t.Fatalf("Run() = %+v, want ok", envelope)
		}
		if requests := fake.Recorded("pane.process_info"); len(requests) != 0 {
			t.Fatalf("pane.process_info = %+v, want no process read", requests)
		}
	})

	cases := []struct {
		name       string
		mutate     func(*records.Agent)
		live       string
		process    string
		wantErr    string
		wantStatus string
	}{
		{
			name:    "different kind",
			live:    liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "codex", "/tmp/old"),
			wantErr: "agent identity differs: kind",
		},
		{
			name:    "different cwd",
			live:    liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/other"),
			wantErr: "agent identity differs: cwd",
		},
		{
			name:    "different terminal_id",
			live:    liveAgentInfo("worker-3", "w1:p1", "term_b", "w1", "w1:t1", "claude", "/tmp/old"),
			wantErr: "agent identity differs: terminal_id",
		},
		{
			name:       "different process_pid",
			mutate:     func(record *records.Agent) { record.ProcessPID = 424242 },
			live:       liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/old"),
			process:    `{"type":"pane_process_info","process_info":` + agenttest.ProcessInfoJSON("w1:p1", agenttest.ProcessEntry(999999, "claude", "claude")) + `}`,
			wantErr:    "agent identity differs: process_pid",
			wantStatus: records.AgentUnresolved,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := agenttest.NewHome(t)
			record := agenttest.ResolveSeedRecord()
			if tc.mutate != nil {
				tc.mutate(&record)
			}
			agenttest.WriteRecord(t, home, record)
			fake := agenttest.NewFakeHerdr(t)
			fake.Result("agent.get", tc.live)
			if tc.process != "" {
				fake.Result("pane.process_info", tc.process)
			}
			envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != tc.wantErr {
				t.Fatalf("Run() = %+v, want %q", envelope, tc.wantErr)
			}
			got := agenttest.ReadAgent(t, home, "worker-3")
			if got.Status != records.AgentUnresolved || got.PaneID != record.PaneID {
				t.Fatalf("record = %+v, want the unchanged record", got)
			}
		})
	}

	t.Run("different process_start", func(t *testing.T) {
		home := agenttest.NewHome(t)
		sleep := agenttest.StartSleep(t)
		record := agenttest.ResolveSeedRecord()
		record.ProcessPID = sleep.PID
		record.ProcessStart = sleep.Start + 1
		agenttest.WriteRecord(t, home, record)
		fake := agenttest.NewFakeHerdr(t)
		fake.Result("agent.get", liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", "/tmp/old"))
		fake.Result("pane.process_info", `{"type":"pane_process_info","process_info":`+agenttest.ProcessInfoJSON("w1:p1", agenttest.ProcessEntry(sleep.PID, "claude", "claude"))+`}`)
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "agent identity differs: process_start" {
			t.Fatalf("Run() = %+v, want the process_start difference", envelope)
		}
		if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentUnresolved {
			t.Fatalf("status = %q, want unresolved", got.Status)
		}
	})

	t.Run("agent_not_found", func(t *testing.T) {
		home := agenttest.NewHome(t)
		agenttest.WriteRecord(t, home, agenttest.ResolveSeedRecord())
		fake := agenttest.NewFakeHerdr(t)
		fake.Failure("agent.get", "agent_not_found", "agent target worker-3 not found")
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "no live agent named worker-3" {
			t.Fatalf("Run() = %+v, want the agent_not_found message", envelope)
		}
		if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentUnresolved {
			t.Fatalf("status = %q, want unresolved", got.Status)
		}
	})

	t.Run("timeout keeps the record", func(t *testing.T) {
		home := agenttest.NewHome(t)
		agenttest.WriteRecord(t, home, agenttest.ResolveSeedRecord())
		fake := agenttest.NewFakeHerdr(t)
		fake.Failure("agent.get", "timeout", "request timed out")
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "timeout: request timed out" {
			t.Fatalf("Run() = %+v, want the herdr code and message", envelope)
		}
		if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentUnresolved {
			t.Fatalf("status = %q, want unresolved", got.Status)
		}
	})

	t.Run("not unresolved", func(t *testing.T) {
		home := agenttest.NewHome(t)
		record := agenttest.ResolveSeedRecord()
		record.Status = records.AgentLive
		agenttest.WriteRecord(t, home, record)
		fake := agenttest.NewFakeHerdr(t)
		envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
		if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "agent is live" {
			t.Fatalf("Run() = %+v, want agent is live", envelope)
		}
		if requests := fake.Recorded("agent.get"); len(requests) != 0 {
			t.Fatalf("agent.get = %+v, want no request", requests)
		}
	})
}
