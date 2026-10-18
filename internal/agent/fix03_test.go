package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func fix03RecordPath(home, key string) string {
	return filepath.Join(home, ".hive", "state", "agents", key+".json")
}

// TestSpawnKeepsStartSessionBeforePaneRead proves AC1 of the fix03 brief. The
// start reply carries session s. The durable record must hold s before the
// pane.get request leaves hive.
func TestSpawnKeepsStartSessionBeforePaneRead(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+fix02AgentJSON("unknown", "s")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+fix02PaneJSON("s")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	seen := make(chan string, 1)
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "pane.get" {
			return
		}
		registry, err := store.Open(home)
		if err != nil {
			seen <- "store open: " + err.Error()
			return
		}
		var record records.Agent
		if err := registry.ReadIn([]string{"agents"}, "t10-coord-g1", &record); err != nil {
			seen <- "record read: " + err.Error()
			return
		}
		seen <- record.AgentSessionRef
	})
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome = %q; want ok (%#v)", envelope.Outcome, envelope)
	}
	if got := <-seen; got != fix02FirstRef {
		t.Fatalf("durable session before pane.get = %q; want %q", got, fix02FirstRef)
	}
	if got := readAgent(t, home, "t10-coord-g1").AgentSessionRef; got != fix02FirstRef {
		t.Fatalf("durable session after spawn = %q; want %q", got, fix02FirstRef)
	}
}

// TestSpawnKeepsPaneTopologyOnSessionRefusal proves AC2 of the fix03 brief. A
// pane reply whose session is absent or differs keeps the earliest session and
// the known topology, and the result stays unresolved.
func TestSpawnKeepsPaneTopologyOnSessionRefusal(t *testing.T) {
	cases := []struct {
		name string
		pane string
	}{
		{name: "absent", pane: ""},
		{name: "differs", pane: "other"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
			fake.result("agent.start", `{"type":"agent_started","agent":`+fix02AgentJSON("unknown", "s")+`,"argv":["claude"]}`)
			fake.result("pane.get", `{"type":"pane_info","pane":`+fix02PaneJSON(test.pane)+`}`)
			fake.result("agent.wait", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
			fake.result("agent.get", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
			envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
			if envelope.Outcome != "partial" {
				t.Fatalf("outcome = %q; want partial (%#v)", envelope.Outcome, envelope)
			}
			checkFix03Topology(t, readAgent(t, home, "t10-coord-g1"))
			result, ok := envelope.Result.(SpawnResult)
			if !ok {
				t.Fatalf("result type = %T; want SpawnResult", envelope.Result)
			}
			checkFix03Topology(t, result.Agent)
			if !hasEffect(envelope.Effects, "agent", "started") {
				t.Errorf("effects = %#v; want the started effect", envelope.Effects)
			}
			if len(fake.recorded("pane.close")) != 0 {
				t.Errorf("pane.close ran %d times; want none", len(fake.recorded("pane.close")))
			}
		})
	}
}

func checkFix03Topology(t *testing.T, record records.Agent) {
	t.Helper()
	if record.Status != records.AgentUnresolved {
		t.Errorf("record status = %q; want unresolved", record.Status)
	}
	if record.AgentSessionRef != fix02FirstRef {
		t.Errorf("record session = %q; want %q", record.AgentSessionRef, fix02FirstRef)
	}
	if record.TerminalID != "term_a" || record.WorkspaceID != "w1" || record.TabID != "w1:t1" || record.Cwd != "/tmp/fixture" {
		t.Errorf("record topology = terminal %q workspace %q tab %q cwd %q; want term_a w1 w1:t1 /tmp/fixture",
			record.TerminalID, record.WorkspaceID, record.TabID, record.Cwd)
	}
}

// TestSpawnWaitTimeoutWithMatchingSessionKeepsLive proves AC3 of the fix03
// brief. A wait timeout is an error, not a typed reply. Matching start, pane,
// and get sessions plus proved process ownership keep partial with a live record.
func TestSpawnWaitTimeoutWithMatchingSessionKeepsLive(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+fix02AgentJSON("unknown", "s")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+fix02PaneJSON("s")+`}`)
	fake.failure("agent.wait", "timeout", "timed out waiting for agent status")
	fake.result("agent.get", `{"type":"agent_info","agent":`+fix02AgentJSON("working", "s")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial (%#v)", envelope.Outcome, envelope)
	}
	if !strings.Contains(toString(envelope.Error), "agent not ready after 120 s") {
		t.Fatalf("error = %v; want the readiness timeout", envelope.Error)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentLive {
		t.Fatalf("record status = %q; want live", record.Status)
	}
	if record.AgentSessionRef != fix02FirstRef {
		t.Errorf("record session = %q; want %q", record.AgentSessionRef, fix02FirstRef)
	}
	if record.ProcessPID != sleep.pid || record.ProcessStart != sleep.start {
		t.Errorf("record process = pid %d start %d; want %d %d", record.ProcessPID, record.ProcessStart, sleep.pid, sleep.start)
	}
}

// TestSpawnSessionStoreFailureKeepsTruth proves AC4 of the fix03 brief. A
// store failure on the earliest start-session write returns partial with the
// actual store error, sends no pane read, and claims no unwritten session.
func TestSpawnSessionStoreFailureKeepsTruth(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+fix02AgentJSON("unknown", "s")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+fix02PaneJSON("s")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", "s")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	path := fix03RecordPath(home, "t10-coord-g1")
	broken := []byte("{not a record}\n")
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "agent.start" {
			return
		}
		_ = os.WriteFile(path, broken, 0o644)
	})
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial (%#v)", envelope.Outcome, envelope)
	}
	text := toString(envelope.Error)
	if !strings.Contains(text, "record invalid") {
		t.Fatalf("error = %v; want the actual store error", envelope.Error)
	}
	if strings.Contains(text, "agent session") {
		t.Fatalf("error = %v; must not claim a session observation", envelope.Error)
	}
	result, ok := envelope.Result.(SpawnResult)
	if !ok {
		t.Fatalf("result type = %T; want SpawnResult", envelope.Result)
	}
	if result.Agent.AgentSessionRef != "" {
		t.Fatalf("result session = %q; must not claim an unwritten session", result.Agent.AgentSessionRef)
	}
	if data, err := os.ReadFile(path); err != nil {
		t.Fatalf("read record: %v", err)
	} else if string(data) != string(broken) {
		t.Fatalf("record bytes = %q; want no session write", data)
	}
	if !hasEffect(envelope.Effects, "agent", "started") {
		t.Errorf("effects = %#v; want the started effect", envelope.Effects)
	}
	if got := len(fake.recorded("pane.get")); got != 0 {
		t.Fatalf("pane.get ran %d times; the session write must fail before the read", got)
	}
}
