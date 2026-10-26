package agent

import (
	"context"
	"encoding/json"
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

func liveRecord(key, name string) records.Agent {
	return records.Agent{
		Key:     key,
		Name:    name,
		Profile: records.ProfileBuilder,
		Project: "hive",
		Task:    "12-spawn-worker",
		Harness: "claude",
		PaneID:  "w1:p1",
		Status:  records.AgentLive,
		Spawned: true,
	}
}
