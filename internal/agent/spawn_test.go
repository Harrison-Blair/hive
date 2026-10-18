package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func baseSpawnOptions(name string) SpawnOptions {
	return SpawnOptions{
		Role:    "coordinator",
		Harness: "claude",
		Model:   "m",
		Name:    name,
		NameSet: true,
		PaneID:  "w1:p1",
		PaneSet: true,
	}
}

func scriptSpawnSuccess(t *testing.T, fake *fakeHerdr, home string, sleep sleepProcess, waitStatus string) {
	t.Helper()
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude","--model","m"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", waitStatus, "t10-coord-g1")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", waitStatus, "t10-coord-g1")+`}`)
}

func TestSpawnIdleRecordsIdentity(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "idle")
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(SpawnResult)
	if !ok {
		t.Fatalf("result type = %T; want SpawnResult", envelope.Result)
	}
	record := result.Agent
	if record.Key != "t10-coord-g1" || record.Name != "t10-coord" {
		t.Fatalf("record key/name = %q/%q; want t10-coord-g1/t10-coord", record.Key, record.Name)
	}
	if record.Status != records.AgentLive || !record.Spawned {
		t.Fatalf("record status/spawned = %q/%t; want live/true", record.Status, record.Spawned)
	}
	if record.TerminalID != "term_a" || record.WorkspaceID != "w1" || record.TabID != "w1:t1" || record.Cwd != "/tmp/fixture" {
		t.Fatalf("record pane identity = %#v", record)
	}
	if record.ProcessPID != sleep.pid || record.ProcessName != "claude" || record.ProcessStart != sleep.start {
		t.Fatalf("record process identity = pid %d name %q start %d; want %d claude %d", record.ProcessPID, record.ProcessName, record.ProcessStart, sleep.pid, sleep.start)
	}
	if result.Claim == nil || result.Claim.Key != "coordinator" || result.Claim.Agent != "t10-coord-g1" {
		t.Fatalf("claim = %#v", result.Claim)
	}
	stored := readAgent(t, home, "t10-coord-g1")
	if stored.ProcessPID != sleep.pid || stored.Status != records.AgentLive {
		t.Fatalf("stored = %#v", stored)
	}
	startRequests := fake.recorded("agent.start")
	if len(startRequests) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(startRequests))
	}
	wantStart := `{"name":"t10-coord-g1","kind":"claude","pane_id":"w1:p1","args":["--model","m"],"timeout_ms":120000}`
	if startRequests[0].Params != wantStart {
		t.Fatalf("start params = %s; want %s", startRequests[0].Params, wantStart)
	}
	waitRequests := fake.recorded("agent.wait")
	if len(waitRequests) != 1 {
		t.Fatalf("agent.wait requests = %d; want 1", len(waitRequests))
	}
	wantWait := `{"target":"w1:p1","until":["idle","blocked","done"],"timeout_ms":120000}`
	if waitRequests[0].Params != wantWait {
		t.Fatalf("wait params = %s; want %s", waitRequests[0].Params, wantWait)
	}
	getRequests := fake.recorded("agent.get")
	if len(getRequests) != 1 {
		t.Fatalf("agent.get requests = %d; want 1", len(getRequests))
	}
	if getRequests[0].Params != `{"target":"w1:p1"}` {
		t.Fatalf("agent.get params = %s", getRequests[0].Params)
	}
}

func TestSpawnRecordExistsBeforeStart(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "idle")
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "agent.start" {
			return
		}
		var record records.Agent
		registry, err := store.Open(home)
		if err != nil {
			t.Errorf("store open: %v", err)
			return
		}
		if err := registry.ReadIn([]string{"agents"}, "t10-coord-g1", &record); err != nil {
			t.Errorf("record absent at agent.start: %v", err)
			return
		}
		if !record.Spawned || record.PaneID != "w1:p1" || record.Harness != "claude" || record.Model != "m" {
			t.Errorf("record at start = %#v", record)
		}
	})
	if envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord")); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestSpawnDone(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "done")
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
}

func TestSpawnBlockedKeepsLivePartial(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "blocked")
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if !strings.Contains(toString(envelope.Error), "agent blocked: accept the dialog in w1:p1") {
		t.Fatalf("error = %v", envelope.Error)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentLive {
		t.Fatalf("record status = %q; want live", record.Status)
	}
	if record.ProcessPID != sleep.pid {
		t.Fatalf("record process pid = %d; want %d", record.ProcessPID, sleep.pid)
	}
	if !hasEffect(envelope.Effects, "agent", "started") {
		t.Fatalf("effects = %#v; want agent started", envelope.Effects)
	}
}

func TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.failure("agent.wait", "timeout", "timed out waiting for agent status")
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "working", "t10-coord-g1")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial (%#v)", envelope.Outcome, envelope)
	}
	if !strings.Contains(toString(envelope.Error), "agent not ready after 120 s") {
		t.Fatalf("error = %v; want readiness timeout", envelope.Error)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentLive {
		t.Fatalf("record status = %q; want live", record.Status)
	}
	if record.ProcessPID != sleep.pid {
		t.Fatalf("record process pid = %d; want %d", record.ProcessPID, sleep.pid)
	}
}

func TestSpawnWaitTimeoutWithoutOwnershipIsUnresolved(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.failure("agent.wait", "timeout", "timed out waiting for agent status")
	fake.failure("agent.get", "timeout", "timed out")
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", record.Status)
	}
}

func TestSpawnStartServerErrorExitsRecord(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	fake.failure("agent.start", "invalid_agent_name", "name is too long")
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentExited {
		t.Fatalf("record status = %q; want exited", record.Status)
	}
	if !strings.Contains(record.LastError, "invalid_agent_name") || !strings.Contains(record.LastError, "name is too long") {
		t.Fatalf("last_error = %q", record.LastError)
	}
	if !hasEffect(envelope.Effects, "agent", "exited") && !hasEffect(envelope.Effects, "agent", "created") {
		t.Fatalf("effects = %#v; want known register effects", envelope.Effects)
	}
	if len(fake.recorded("pane.get")) != 0 {
		t.Fatalf("pane.get ran after a start error")
	}
}

func TestSpawnStartTransportErrorIsUnresolved(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	fake.drop("agent.start")
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want partial start", envelope)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", record.Status)
	}
}

func TestSpawnWaitNotRunningInspectsExited(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.failure("agent.wait", "agent_not_running", "agent is no longer running in the target pane")
	fake.failure("agent.get", "agent_not_found", "agent target w1:p1 not found")
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoShellJSON("w1:p1", sleep.pid, processEntry(sleep.pid, "bash", "bash"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if !strings.Contains(toString(envelope.Error), "agent exited during start") {
		t.Fatalf("error = %v", envelope.Error)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.Status != records.AgentExited {
		t.Fatalf("record status = %q; want exited", record.Status)
	}
}

func TestSpawnAgentGetMissingNameIsOk(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "idle")
	fake.mu.Lock()
	fake.replies["agent.get"] = []fakeReply{{error: `{"code":"agent_not_found","message":"agent target w1:p1 not found"}`}}
	fake.mu.Unlock()
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome = %q; want ok (%#v)", envelope.Outcome, envelope)
	}
	logText := readLog(t, home)
	if !strings.Contains(logText, "missing name") && !strings.Contains(logText, "no name") {
		t.Fatalf("log = %s; want a missing-name event", logText)
	}
}

func TestSpawnProcessIdentityCases(t *testing.T) {
	sleep := startSleep(t)
	cases := []struct {
		name    string
		entries []string
	}{
		{name: "absent"},
		{name: "multiple", entries: []string{processEntry(sleep.pid, "claude", "claude"), processEntry(sleep.pid, "claude", "claude")}},
		{name: "shell", entries: []string{processEntry(sleep.pid, "bash", "bash")}},
		{name: "unknown name", entries: []string{processEntry(sleep.pid, "node", "node", "cli.js")}},
		{name: "wrapper only", entries: []string{processEntry(sleep.pid, "node", "node", "claude")}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			fake := newFakeHerdr(t)
			fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
			fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", test.entries...)+`}`)
			envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
			if envelope.Outcome != "partial" {
				t.Fatalf("outcome = %q; want partial", envelope.Outcome)
			}
			record := readAgent(t, home, "t10-coord-g1")
			if record.Status != records.AgentUnresolved {
				t.Fatalf("record status = %q; want unresolved", record.Status)
			}
			if !hasEffect(envelope.Effects, "agent", "started") {
				t.Fatalf("effects = %#v; want started", envelope.Effects)
			}
		})
	}
}

func TestSpawnUnreadableProcessIsUnresolved(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(1<<40, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if record := readAgent(t, home, "t10-coord-g1"); record.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", record.Status)
	}
}

func TestSpawnSessionRules(t *testing.T) {
	sleep := startSleep(t)
	cases := []struct {
		name       string
		waitJSON   string
		getJSON    string
		paneJSON   string
		wantStatus string
		wantRef    bool
	}{
		{name: "absent passes", waitJSON: agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1"), getJSON: agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1"), paneJSON: paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture"), wantStatus: records.AgentLive},
		{name: "present records", waitJSON: agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "s1"), getJSON: agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "s1"), paneJSON: paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture"), wantStatus: records.AgentLive, wantRef: true},
		{name: "different sources fail", waitJSON: agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "s1"), getJSON: agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1", "herdr:claude", "id", "s2"), paneJSON: paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture"), wantStatus: records.AgentUnresolved},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			fake := newFakeHerdr(t)
			fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
			fake.result("pane.get", `{"type":"pane_info","pane":`+test.paneJSON+`}`)
			fake.result("agent.wait", `{"type":"agent_info","agent":`+test.waitJSON+`}`)
			fake.result("agent.get", `{"type":"agent_info","agent":`+test.getJSON+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
			envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
			record := readAgent(t, home, "t10-coord-g1")
			if record.Status != test.wantStatus {
				t.Fatalf("record status = %q; want %q (envelope %#v)", record.Status, test.wantStatus, envelope)
			}
			if test.wantRef && record.AgentSessionRef == "" {
				t.Fatalf("agent_session_ref is empty; want a reference")
			}
		})
	}
}

func TestSpawnClaimReplacementDuringWaitPreventsUpdate(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "t10-coord-g1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	replaced := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "agent.wait" || replaced {
			return
		}
		replaced = true
		writeRecord(t, home, records.Agent{Key: "other-g2", Name: "other", Role: "coordinator", Status: records.AgentLive, PaneID: "w2:p1", Spawned: true, Generation: 2})
		writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "other-g2", Generation: 2})
	})
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "t10-coord-g1")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
	if envelope.Outcome == "ok" {
		t.Fatalf("envelope = %#v; want a refusal after claim replacement", envelope)
	}
	record := readAgent(t, home, "t10-coord-g1")
	if record.ProcessPID != 0 {
		t.Fatalf("record process pid = %d; must not update after claim replacement", record.ProcessPID)
	}
}

func TestSpawnDoesNotHoldLockAcrossWait(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptSpawnSuccess(t, fake, home, sleep, "idle")
	lockResult := make(chan error, 1)
	once := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "agent.wait" || once {
			return
		}
		once = true
		go func() {
			registry, err := store.Open(home)
			if err != nil {
				lockResult <- err
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			lockResult <- registry.WithLock(ctx, func(*store.LockedStore) error { return nil })
		}()
	})
	if envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord")); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	select {
	case err := <-lockResult:
		if err != nil {
			t.Fatalf("store lock during wait: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("store lock during wait timed out")
	}
}

func TestSpawnWorkerWithRealTask(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	writeProject(t, home, "hive")
	writeTask(t, home, "hive", "12-spawn-worker")
	fake := newFakeHerdr(t)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "worker-hive-1")+`,"argv":["claude"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	options := SpawnOptions{Role: "worker", Project: "hive", Task: "12-spawn-worker", Harness: "claude", Model: "m", PaneID: "w1:p1", PaneSet: true}
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(SpawnResult)
	if !ok {
		t.Fatalf("result = %T", envelope.Result)
	}
	if result.Agent.Key != "worker-hive-1" || result.Agent.Name != "worker-hive-1" {
		t.Fatalf("worker key/name = %q/%q", result.Agent.Key, result.Agent.Name)
	}
	if result.Claim != nil {
		t.Fatalf("worker claim = %#v; want nil", result.Claim)
	}
}

func TestSpawnWithoutPaneFails(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	options := baseSpawnOptions("t10-coord")
	options.PaneSet = false
	options.PaneID = ""
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if !strings.Contains(toString(envelope.Error), "pane required: pass --pane") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("agent.start ran without a pane")
	}
}

func TestSpawnUnknownHarness(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	options := baseSpawnOptions("t10-coord")
	options.Harness = "nope"
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "unknown harness: nope") {
		t.Fatalf("envelope = %#v", envelope)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("agent.start ran for an unknown harness")
	}
}

func TestSpawnStoresNoRecordBeforeValidation(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	options := baseSpawnOptions("t10-coord")
	options.Harness = "nope"
	_ = Spawn(context.Background(), home, options)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ReadIn([]string{"agents"}, "t10-coord-g1", &records.Agent{}); err == nil {
		t.Fatalf("agent record exists after a validation failure")
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("agent.start ran")
	}
}

func TestManualRegisterLeavesSpawnedFalse(t *testing.T) {
	home := testHome(t)
	envelope := Register(home, RegisterOptions{
		Role:    "coordinator",
		Harness: "claude",
		PaneID:  "w1:p1",
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("register = %#v", envelope)
	}
	record := readAgent(t, home, "coordinator-g1")
	if record.Spawned {
		t.Fatalf("manual register spawned = true; want false")
	}
}

func hasEffect(effects []output.Effect, resource, action string) bool {
	for _, effect := range effects {
		if effect.Resource == resource && effect.Action == action {
			return true
		}
	}
	return false
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	data, _ := json.Marshal(value)
	return string(data)
}

func readLog(t *testing.T, home string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text += string(data)
	}
	return text
}
