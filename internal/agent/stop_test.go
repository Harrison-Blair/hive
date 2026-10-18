package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func spawnedRecord(sleep sleepProcess) records.Agent {
	return records.Agent{
		Key:          "t10-coord-g1",
		Name:         "t10-coord",
		Role:         "coordinator",
		Harness:      "claude",
		PaneID:       "w1:p1",
		WorkspaceID:  "w1",
		TabID:        "w1:t1",
		TerminalID:   "term_a",
		Cwd:          "/tmp/fixture",
		Status:       records.AgentLive,
		Spawned:      true,
		Generation:   1,
		ProcessPID:   sleep.pid,
		ProcessName:  "claude",
		ProcessStart: sleep.start,
	}
}

func writeStopRecord(t *testing.T, home string, record records.Agent) {
	t.Helper()
	writeRecord(t, home, record)
	if key := claimKeyFor(record.Role, record.Project); key != "" {
		writeClaimRecord(t, home, records.Claim{Key: key, Agent: record.Key, Generation: record.Generation})
	}
}

func scriptStop(t *testing.T, fake *fakeHerdr, sleep sleepProcess, closeResult string, postWorkspace string) {
	t.Helper()
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	if closeResult == "" {
		fake.result("pane.close", `{"type":"ok"}`)
	} else {
		fake.result("pane.close", closeResult)
	}
	if postWorkspace != "" {
		fake.failure("workspace.get", "workspace_not_found", "workspace w1 not found")
	}
}

func TestStopRefusesManualRegistration(t *testing.T) {
	home := testHome(t)
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", Status: records.AgentLive, Spawned: false, Generation: 1}
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if !strings.Contains(toString(envelope.Error), "agent coordinator was not spawned by hive: exit it by hand") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if result.Agent.Status == records.AgentStopped {
		t.Fatalf("record stopped after refusal")
	}
	if len(fake.recorded("pane.get")) != 0 || len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("refusal sent a request")
	}
}

func TestStopRefusesNonLiveRecord(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	record.Status = records.AgentStopped
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "agent is stopped") {
		t.Fatalf("envelope = %#v", envelope)
	}
	if len(fake.recorded("pane.get")) != 0 {
		t.Fatalf("pane.get ran for a stopped record")
	}
}

func TestStopPaneNotFound(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.failure("pane.get", "pane_not_found", "pane w1:p1 not found")
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q", envelope.Outcome)
	}
	if !strings.Contains(toString(envelope.Error), "pane w1:p1 not found: record unresolved") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if got := readAgent(t, home, record.Key).Status; got != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", got)
	}
}

func TestStopTerminalDiffers(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_other", "w1", "w1:t1", "/tmp/fixture")+`}`)
	_, envelope := Stop(context.Background(), home, record)
	if !strings.Contains(toString(envelope.Error), "pane w1:p1 holds another terminal: record unresolved") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if got := readAgent(t, home, record.Key).Status; got != records.AgentUnresolved {
		t.Fatalf("record status = %q", got)
	}
}

func TestStopProcessIdentityCases(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*records.Agent)
		entry  func(sleepProcess) string
	}{
		{name: "pid missing", entry: func(sleep sleepProcess) string { return processEntry(sleep.pid+1, "claude", "claude") }},
		{name: "name mismatch", entry: func(sleep sleepProcess) string { return processEntry(sleep.pid, "node", "node") }},
		{name: "start mismatch", mutate: func(record *records.Agent) { record.ProcessStart++ }, entry: func(sleep sleepProcess) string { return processEntry(sleep.pid, "claude", "claude") }},
		{name: "unreadable process", mutate: func(record *records.Agent) { record.ProcessPID = 1 << 40 }, entry: func(sleep sleepProcess) string { return processEntry(1<<40, "claude", "claude") }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			sleep := startSleep(t)
			record := spawnedRecord(sleep)
			if test.mutate != nil {
				test.mutate(&record)
			}
			writeStopRecord(t, home, record)
			fake := newFakeHerdr(t)
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", test.entry(sleep))+`}`)
			_, envelope := Stop(context.Background(), home, record)
			if envelope.Outcome != "failed" {
				t.Fatalf("outcome = %q; want failed", envelope.Outcome)
			}
			if got := readAgent(t, home, record.Key).Status; got != records.AgentUnresolved {
				t.Fatalf("record status = %q; want unresolved", got)
			}
			if len(fake.recorded("pane.close")) != 0 {
				t.Fatalf("pane.close ran after an identity mismatch")
			}
		})
	}
}

func TestStopSessionRules(t *testing.T) {
	cases := []struct {
		name    string
		ref     string
		session string
		want    string
	}{
		{name: "absent passes", ref: "", session: "", want: "ok"},
		{name: "missing fails", ref: `{"source":"herdr:claude","agent":"claude","kind":"id","value":"s1"}`, session: "", want: "failed"},
		{name: "different fails", ref: `{"source":"herdr:claude","agent":"claude","kind":"id","value":"s1"}`, session: `,"agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"s2"}`, want: "failed"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			sleep := startSleep(t)
			record := spawnedRecord(sleep)
			record.AgentSessionRef = test.ref
			writeStopRecord(t, home, record)
			fake := newFakeHerdr(t)
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSONSession("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture", test.session)+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
			if test.want == "ok" {
				fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
				fake.result("pane.close", `{"type":"ok"}`)
				fake.failure("workspace.get", "workspace_not_found", "workspace w1 not found")
			}
			_, envelope := Stop(context.Background(), home, record)
			if envelope.Outcome != test.want {
				t.Fatalf("outcome = %q; want %q (%#v)", envelope.Outcome, test.want, envelope)
			}
		})
	}
}

func TestStopSuccessClosed(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	scriptStop(t, fake, sleep, "", `{"error":{"code":"workspace_not_found","message":"workspace w1 not found"}}`)
	closedWhileStopped := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "pane.close" {
			return
		}
		if got := readAgent(t, home, record.Key).Status; got != records.AgentStopped {
			t.Errorf("record status at close = %q; want stopped", got)
		}
		closedWhileStopped = true
	})
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if !closedWhileStopped {
		t.Fatalf("pane.close did not run")
	}
	if result.Topology != "closed" {
		t.Fatalf("topology = %q; want closed", result.Topology)
	}
	if result.Agent.Status != records.AgentStopped {
		t.Fatalf("record status = %q; want stopped", result.Agent.Status)
	}
	if !hasEffect(envelope.Effects, "pane", "closed") {
		t.Fatalf("effects = %#v; want pane closed", envelope.Effects)
	}
	if got := readAgent(t, home, record.Key).Status; got != records.AgentStopped {
		t.Fatalf("stored status = %q", got)
	}
	if len(fake.recorded("workspace.get")) != 2 {
		t.Fatalf("workspace.get requests = %d; want pre-close and post-close", len(fake.recorded("workspace.get")))
	}
}

func TestStopSurvivingTopology(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || result.Topology != "surviving" || result.Workspace == nil || result.Workspace.PaneCount != 1 {
		t.Fatalf("result = %#v envelope = %#v", result, envelope)
	}
}

func TestStopUnknownTopology(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.failure("workspace.get", "timeout", "timed out")
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || result.Topology != "unknown" {
		t.Fatalf("result = %#v envelope = %#v", result, envelope)
	}
	if result.Workspace != nil {
		t.Fatalf("workspace = %#v; want nil for unknown", result.Workspace)
	}
}

func TestStopCloseErrorRestoresLive(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.failure("pane.close", "invalid_request", "close failed")
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	stored := readAgent(t, home, record.Key)
	if stored.Status != records.AgentLive {
		t.Fatalf("record status = %q; want live", stored.Status)
	}
	if !strings.Contains(stored.LastError, "close failed") {
		t.Fatalf("last_error = %q", stored.LastError)
	}
}

func TestStopClosePaneNotFound(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.failure("pane.close", "pane_not_found", "pane w1:p1 not found")
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "pane w1:p1 not found: record unresolved") {
		t.Fatalf("envelope = %#v", envelope)
	}
	if got := readAgent(t, home, record.Key).Status; got != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", got)
	}
}

func TestStopRecordChangedBeforeStop(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeRecord(t, home, record)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: record.Key, Generation: 1})
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	replaced := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "workspace.get" || replaced {
			return
		}
		replaced = true
		writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "other-g2", Generation: 2})
	})
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("pane.close ran after a claim replacement")
	}
	if got := readAgent(t, home, record.Key).Status; got != records.AgentLive {
		t.Fatalf("record status = %q; want live", got)
	}
}

func TestStopDoesNotHoldLockAcrossPostClose(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.failure("workspace.get", "workspace_not_found", "workspace w1 not found")
	lockResult := make(chan error, 1)
	postClose := false
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "workspace.get" || !postClose {
			if method == "pane.close" {
				postClose = true
			}
			return
		}
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
	if _, envelope := Stop(context.Background(), home, record); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	select {
	case err := <-lockResult:
		if err != nil {
			t.Fatalf("store lock during post-close read: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("store lock during post-close read timed out")
	}
}

func TestStopDoesNotMutateProject(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	writeProject(t, home, "hive")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdateIn([]string{"projects"}, "hive", func(raw json.RawMessage) (json.RawMessage, error) {
		var project records.Project
		if err := json.Unmarshal(raw, &project); err != nil {
			return nil, err
		}
		project.RepoWorkspaceID = "w1"
		return json.Marshal(project)
	}); err != nil {
		t.Fatal(err)
	}
	var before records.Project
	if err := registry.ReadIn([]string{"projects"}, "hive", &before); err != nil {
		t.Fatal(err)
	}
	fake := newFakeHerdr(t)
	scriptStop(t, fake, sleep, "", `{"error":{"code":"workspace_not_found","message":"workspace w1 not found"}}`)
	if _, envelope := Stop(context.Background(), home, record); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	var after records.Project
	if err := registry.ReadIn([]string{"projects"}, "hive", &after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("project changed: before %#v after %#v", before, after)
	}
}

func TestStopLogsTransition(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	scriptStop(t, fake, sleep, "", `{"error":{"code":"workspace_not_found","message":"workspace w1 not found"}}`)
	if _, envelope := Stop(context.Background(), home, record); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	logText := readLog(t, home)
	if !strings.Contains(logText, "agent stop") || !strings.Contains(logText, "stopped") {
		t.Fatalf("log does not contain the stop transition")
	}
}

func TestStopEffects(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	scriptStop(t, fake, sleep, "", `{"error":{"code":"workspace_not_found","message":"workspace w1 not found"}}`)
	_, envelope := Stop(context.Background(), home, record)
	if !hasEffect(envelope.Effects, "pane", "closed") {
		t.Fatalf("effects = %#v; want pane closed", envelope.Effects)
	}
	if !hasEffect(envelope.Effects, "agent", "stopped") {
		t.Fatalf("effects = %#v; want agent stopped", envelope.Effects)
	}
}

func paneInfoJSONSession(paneID, terminalID, workspaceID, tabID, cwd, session string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + cwd + `"` + session + `}`
}

var _ = output.Effect{}
