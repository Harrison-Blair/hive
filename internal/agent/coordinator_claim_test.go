package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// scriptCoordinatorStart scripts one coordinator harness start in paneID. The
// harness process is the sleep process with the name claude.
func scriptCoordinatorStart(fake *fakeHerdr, paneID, tabID string, harness sleepProcess) {
	scriptReadyShell(fake, paneID, "term_"+paneID, "w1", tabID, "/tmp")
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON(paneID, "term_"+paneID, "w1", tabID, "unknown", "coordinator")+`,"argv":["claude"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON(paneID, "term_"+paneID, "w1", tabID, "idle", "coordinator")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON(paneID, "term_"+paneID, "w1", tabID, "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(harness.pid, "claude", "claude"))+`}`)
}

func startTestCoordinator(home, paneID string) error {
	launch, err := ClaimCoordinatorStart(context.Background(), home, nil)
	if err != nil {
		return err
	}
	return launch.Start(context.Background(), paneID, "claude", "m")
}

func readCoordinatorClaim(t *testing.T, home string) (records.Claim, bool) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var claim records.Claim
	err = registry.ReadIn([]string{"claims"}, "coordinator", &claim)
	if errors.Is(err, os.ErrNotExist) {
		return records.Claim{}, false
	}
	if err != nil {
		t.Fatal(err)
	}
	return claim, true
}

func coordinatorRecords(t *testing.T, home string) []records.Agent {
	t.Helper()
	listed := List(home, ListOptions{Role: "coordinator", All: true})
	if listed.Outcome != "ok" {
		t.Fatalf("list coordinators = %#v", listed)
	}
	return listed.Result.([]records.Agent)
}

// deadLauncher returns the process ID and start time of a process that ended.
func deadLauncher(t *testing.T) (int64, uint64) {
	t.Helper()
	command := exec.Command("sleep", "0.2")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	pid := int64(command.Process.Pid)
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err != nil {
		t.Fatal(err)
	}
	return pid, start
}

func stateFiles(t *testing.T, home string) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, collection := range []string{"agents", "claims"} {
		dir := filepath.Join(home, ".hive", "state", collection)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			files[collection+"/"+entry.Name()] = string(data)
		}
	}
	return files
}

// TestCoordinatorStartClaimsThenRecordsThenPrompts covers decision 31.
func TestCoordinatorStartClaimsThenRecordsThenPrompts(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptCoordinatorStart(fake, "w1:p1", "w1:t1", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "working", "coordinator")+`}`)
	var mu sync.Mutex
	var atStart, atPrompt records.Claim
	var recordsAtStart, recordsAtPrompt []records.Agent
	fake.setBefore(func(method string, _ json.RawMessage) {
		mu.Lock()
		defer mu.Unlock()
		switch method {
		case "agent.start":
			atStart, _ = readCoordinatorClaim(t, home)
			recordsAtStart = coordinatorRecords(t, home)
		case "agent.prompt":
			atPrompt, _ = readCoordinatorClaim(t, home)
			recordsAtPrompt = coordinatorRecords(t, home)
		}
	})

	if err := startTestCoordinator(home, "w1:p1"); err != nil {
		t.Fatalf("start coordinator: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	start, _ := process.Start(int64(os.Getpid()))
	if atStart.Status != records.ClaimStarting || atStart.PID != int64(os.Getpid()) || atStart.PIDStart != start || atStart.PaneID != "w1:p1" || atStart.Agent != "" {
		t.Fatalf("claim at harness start = %#v; want a starting claim of this launcher for pane w1:p1", atStart)
	}
	if len(recordsAtStart) != 0 {
		t.Fatalf("records at harness start = %#v; want none before the start", recordsAtStart)
	}
	if len(recordsAtPrompt) != 1 {
		t.Fatalf("records at prompt = %#v; want the record before the prompt", recordsAtPrompt)
	}
	record := recordsAtPrompt[0]
	if record.Key != "coordinator-g1" || record.Status != records.AgentLive || record.PaneID != "w1:p1" || !record.StartedByHive || record.ProcessPID != harness.pid || record.ProcessStart != harness.start {
		t.Fatalf("record at prompt = %#v; want live coordinator-g1 with the harness process", record)
	}
	if atPrompt.Status == records.ClaimStarting || atPrompt.Agent != record.Key || atPrompt.Generation != record.Generation {
		t.Fatalf("claim at prompt = %#v; want a live claim of %s", atPrompt, record.Key)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 || strings.Contains(strings.ToLower(prompts[0].Params), "register") {
		t.Fatalf("prompts = %#v; want one startup prompt that does not tell the coordinator to register", prompts)
	}
}

// TestCoordinatorStartTwoStartsAtTheSameTime covers decision 35, case 1.
func TestCoordinatorStartTwoStartsAtTheSameTime(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptCoordinatorStart(fake, "w1:p1", "w1:t1", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "working", "coordinator")+`}`)

	var wg sync.WaitGroup
	results := make([]error, 2)
	gate := make(chan struct{})
	for index := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-gate
			results[index] = startTestCoordinator(home, "w1:p1")
		}(index)
	}
	close(gate)
	wg.Wait()
	failures := 0
	for _, err := range results {
		if err != nil {
			failures++
			if !strings.Contains(err.Error(), "coordinator") {
				t.Errorf("refused start error = %v; want a coordinator claim refusal", err)
			}
		}
	}
	if failures != 1 {
		t.Fatalf("start results = %v; want one start and one refusal", results)
	}
	if got := len(fake.recorded("agent.start")); got != 1 {
		t.Fatalf("agent.start requests = %d; want 1", got)
	}
	if got := coordinatorRecords(t, home); len(got) != 1 {
		t.Fatalf("coordinator records = %#v; want one", got)
	}
}

func TestCoordinatorStartRefusesStartingClaimWithLiveLauncher(t *testing.T) {
	home := testHome(t)
	launcher := startSleep(t)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Generation: 1, PID: launcher.pid, PIDStart: launcher.start, Status: records.ClaimStarting, PaneID: "w1:p1"})
	newFakeHerdr(t)
	before := stateFiles(t, home)
	_, err := ClaimCoordinatorStart(context.Background(), home, nil)
	if err == nil || !strings.Contains(err.Error(), "starting") {
		t.Fatalf("claim = %v; want a refusal for the live launcher", err)
	}
	if after := stateFiles(t, home); len(after) != len(before) || after["claims/coordinator.json"] != before["claims/coordinator.json"] {
		t.Fatalf("state after refusal = %#v; want no change", after)
	}
}

func TestCoordinatorStartRefusesLiveClaim(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1})
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	newFakeHerdr(t)
	before := stateFiles(t, home)
	if _, err := ClaimCoordinatorStart(context.Background(), home, nil); err == nil || !strings.Contains(err.Error(), "coordinator-g1") {
		t.Fatalf("claim = %v; want a refusal that names the live record", err)
	}
	if after := stateFiles(t, home); after["claims/coordinator.json"] != before["claims/coordinator.json"] {
		t.Fatal("claim changed after the refusal")
	}
}

// TestCoordinatorStartHarnessFailureRemovesClaim covers decision 32.
func TestCoordinatorStartHarnessFailureRemovesClaim(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", "/tmp")
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.failure("agent.start", "invalid_agent_kind", "no harness")

	err := startTestCoordinator(home, "w1:p1")
	if err == nil || !strings.Contains(err.Error(), "no harness") {
		t.Fatalf("start = %v; want the harness start failure", err)
	}
	if claim, ok := readCoordinatorClaim(t, home); ok {
		t.Fatalf("claim after the failed start = %#v; want none", claim)
	}
	if got := coordinatorRecords(t, home); len(got) != 0 {
		t.Fatalf("records = %#v; want none", got)
	}
}

// TestCoordinatorStartPromptFailureAfterRecordWrite covers decision 35, case 2.
func TestCoordinatorStartPromptFailureAfterRecordWrite(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptCoordinatorStart(fake, "w1:p1", "w1:t1", harness)
	fake.drop("agent.prompt")

	if err := startTestCoordinator(home, "w1:p1"); err == nil {
		t.Fatal("start succeeded; want the prompt failure")
	}
	record := readAgent(t, home, "coordinator-g1")
	if record.Status != records.AgentUnresolved || record.LastError == "" {
		t.Fatalf("record = %#v; want unresolved with the prompt error", record)
	}
	claim, ok := readCoordinatorClaim(t, home)
	if !ok || claim.Status == records.ClaimStarting || claim.Agent != record.Key {
		t.Fatalf("claim = %#v; want a live claim of the unresolved record", claim)
	}
	if _, err := ClaimCoordinatorStart(context.Background(), home, nil); err == nil || !strings.Contains(err.Error(), "Health › Agents") {
		t.Fatalf("second start = %v; want a refusal that names the restart", err)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt requests = %d; want one and no second prompt", got)
	}
}

// TestCoordinatorStartLauncherCrashBeforeHarnessStart covers decision 35,
// case 3, and the first rule of decision 33.
func TestCoordinatorStartLauncherCrashBeforeHarnessStart(t *testing.T) {
	home := testHome(t)
	pid, start := deadLauncher(t)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Generation: 1, PID: pid, PIDStart: start, Status: records.ClaimStarting, PaneID: "w1:old"})
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:old", processEntry(1, "bash", "bash"))+`}`)
	scriptCoordinatorStart(fake, "w1:p1", "w1:t1", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "working", "coordinator")+`}`)

	if err := startTestCoordinator(home, "w1:p1"); err != nil {
		t.Fatalf("start after the crash: %v", err)
	}
	got := coordinatorRecords(t, home)
	if len(got) != 1 || got[0].PaneID != "w1:p1" || got[0].Status != records.AgentLive {
		t.Fatalf("records = %#v; want one live record in the new pane", got)
	}
	claim, _ := readCoordinatorClaim(t, home)
	if claim.Agent != got[0].Key || claim.Status == records.ClaimStarting {
		t.Fatalf("claim = %#v; want a live claim of %s", claim, got[0].Key)
	}
	if got := len(fake.recorded("agent.start")); got != 1 {
		t.Fatalf("agent.start requests = %d; want 1", got)
	}
}

// TestCoordinatorLauncherCrashAfterHarnessStartThenRestart covers decision
// 35, case 4, and the second rule of decision 33.
func TestCoordinatorLauncherCrashAfterHarnessStartThenRestart(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	pid, start := deadLauncher(t)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Generation: 1, PID: pid, PIDStart: start, Status: records.ClaimStarting, PaneID: "w1:old"})
	orphan := startSleep(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:old", processEntry(orphan.pid, "claude", "claude"))+`}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:old", "term_old", "w1", "w1:t0", "/tmp")+`}`)

	_, err := ClaimCoordinatorStart(context.Background(), home, nil)
	if err == nil || !strings.Contains(err.Error(), "Health › Agents") {
		t.Fatalf("start after the crash = %v; want a stop that names the restart", err)
	}
	if got := len(fake.recorded("agent.start")); got != 0 {
		t.Fatalf("agent.start requests = %d; want none", got)
	}
	settled := coordinatorRecords(t, home)
	if len(settled) != 1 || settled[0].Status != records.AgentUnresolved || settled[0].PaneID != "w1:old" || settled[0].WorkspaceID != "w1" || !settled[0].StartedByHive {
		t.Fatalf("records = %#v; want one unresolved record for pane w1:old", settled)
	}
	claim, _ := readCoordinatorClaim(t, home)
	if claim.Agent != settled[0].Key || claim.Status == records.ClaimStarting {
		t.Fatalf("claim = %#v; want the claim to name %s", claim, settled[0].Key)
	}

	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:old", "term_old", "w1", "w1:t0", "/tmp")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:old", processEntry(orphan.pid, "claude", "claude"))+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t2", "w1:p2"))
	scriptCoordinatorStart(fake, "w1:p2", "w1:t2", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p2", "term_w1:p2", "w1", "w1:t2", "working", "coordinator")+`}`)
	if _, err := Restart(context.Background(), home, settled[0], nil); err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	closes := fake.recorded("pane.close")
	if len(closes) != 1 || !strings.Contains(closes[0].Params, `"w1:old"`) {
		t.Fatalf("pane.close = %#v; want the orphan harness pane closed", closes)
	}
	running := 0
	for _, record := range coordinatorRecords(t, home) {
		if record.Status == records.AgentLive || record.Status == records.AgentUnresolved {
			running++
			if record.PaneID != "w1:p2" {
				t.Fatalf("running record = %#v; want only the new pane", record)
			}
		}
	}
	if running != 1 || len(fake.recorded("agent.start")) != 1 {
		t.Fatalf("running records = %d, agent.start = %d; want one coordinator harness", running, len(fake.recorded("agent.start")))
	}
}

// TestCoordinatorRestartNormal covers decision 35, case 5, and decision 34.
func TestCoordinatorRestartNormal(t *testing.T) {
	testCoordinatorRestartFrom(t, records.AgentLive)
}

// TestCoordinatorRestartUnresolved covers decision 35, case 6.
func TestCoordinatorRestartUnresolved(t *testing.T) {
	testCoordinatorRestartFrom(t, records.AgentUnresolved)
}

func testCoordinatorRestartFrom(t *testing.T, status string) {
	t.Helper()
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	old := startSleep(t)
	harness := startSleep(t)
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1", TerminalID: "term_w1:p1", Status: status, Generation: 1, StartedByHive: true, ProcessPID: old.pid, ProcessName: "claude", ProcessStart: old.start}
	writeRecord(t, home, record)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1, Status: records.ClaimLive})
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "/tmp")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(old.pid, "claude", "claude"))+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t2", "w1:p2"))
	scriptCoordinatorStart(fake, "w1:p2", "w1:t2", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p2", "term_w1:p2", "w1", "w1:t2", "working", "coordinator")+`}`)
	var mu sync.Mutex
	var claimAtClose records.Claim
	var oldAtClose records.Agent
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "pane.close" {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		claimAtClose, _ = readCoordinatorClaim(t, home)
		oldAtClose = readAgent(t, home, "coordinator-g1")
	})

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if _, err := Restart(context.Background(), home, record, logger); err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	for _, msg := range []string{"coordinator harness start", "coordinator record written"} {
		found := false
		for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
			var event map[string]any
			if json.Unmarshal([]byte(line), &event) != nil || event["msg"] != msg {
				continue
			}
			found = true
			if event["command"] != "tui agent restart" {
				t.Fatalf("%q command = %v; want tui agent restart", msg, event["command"])
			}
		}
		if !found {
			t.Fatalf("no %q log event in:\n%s", msg, logs.String())
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if claimAtClose.Status != records.ClaimStarting || claimAtClose.PID != int64(os.Getpid()) || oldAtClose.Status != records.AgentStale {
		t.Fatalf("at close: claim %#v, old record %#v; want a starting claim and a stale old record", claimAtClose, oldAtClose)
	}
	if closes := fake.recorded("pane.close"); len(closes) != 1 || !strings.Contains(closes[0].Params, `"w1:p1"`) {
		t.Fatalf("pane.close = %#v; want the old pane", closes)
	}
	if got := readAgent(t, home, "coordinator-g1"); got.Status != records.AgentStale {
		t.Fatalf("old record = %q; want stale", got.Status)
	}
	next := readAgent(t, home, "coordinator-g2")
	if next.Status != records.AgentLive || next.PaneID != "w1:p2" || next.Replaces != "coordinator-g1" {
		t.Fatalf("new record = %#v; want live coordinator-g2 in w1:p2 that replaces coordinator-g1", next)
	}
	if claim, _ := readCoordinatorClaim(t, home); claim.Agent != "coordinator-g2" || claim.Generation != 2 || claim.Status == records.ClaimStarting {
		t.Fatalf("claim = %#v; want a live claim of coordinator-g2", claim)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt = %d; want 1", got)
	}
}

func TestCoordinatorRestartRefusesRecordThatTheClaimDoesNotName(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	old := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", WorkspaceID: "w1", Status: records.AgentUnresolved, Generation: 1, StartedByHive: true}
	current := records.Agent{Key: "coordinator-g2", Name: "coordinator", Role: "coordinator", PaneID: "w1:p2", WorkspaceID: "w1", Status: records.AgentLive, Generation: 2, StartedByHive: true}
	writeRecord(t, home, old)
	writeRecord(t, home, current)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g2", Generation: 2, Status: records.ClaimLive})
	fake := newFakeHerdr(t)
	before := stateFiles(t, home)

	if _, err := Restart(context.Background(), home, old, nil); err == nil || !strings.Contains(err.Error(), "coordinator-g2") {
		t.Fatalf("Restart() = %v; want a refusal that names the claim holder", err)
	}
	after := stateFiles(t, home)
	if len(after) != len(before) {
		t.Fatalf("state files = %d; want %d", len(after), len(before))
	}
	for path, data := range before {
		if after[path] != data {
			t.Fatalf("%s changed after the refused restart", path)
		}
	}
	if methods := fake.requestedMethods(); len(methods) != 0 {
		t.Fatalf("herdr requests = %#v; want none", methods)
	}
}

// TestCoordinatorRestartWithNoClaim covers decision 34: with no claim, the
// restart takes a new starting claim under the store lock.
func TestCoordinatorRestartWithNoClaim(t *testing.T) {
	testCoordinatorRestartOf(t, records.AgentLive, false)
}

// TestCoordinatorRestartStale covers decision 34 for a stale record.
func TestCoordinatorRestartStale(t *testing.T) {
	testCoordinatorRestartOf(t, records.AgentStale, true)
}

// TestCoordinatorRestartStopped covers decision 34 for a stopped record.
func TestCoordinatorRestartStopped(t *testing.T) {
	testCoordinatorRestartOf(t, records.AgentStopped, true)
}

// TestCoordinatorRestartExited covers decision 34 for an exited record.
func TestCoordinatorRestartExited(t *testing.T) {
	testCoordinatorRestartOf(t, records.AgentExited, true)
}

// testCoordinatorRestartOf restarts a coordinator record with status. With
// claimed, the claim names that record. Without it, no claim exists. Only a
// live or unresolved old record has a pane to close.
func testCoordinatorRestartOf(t *testing.T, status string, claimed bool) {
	t.Helper()
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	old := startSleep(t)
	harness := startSleep(t)
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1", TerminalID: "term_w1:p1", Status: status, Generation: 1, StartedByHive: true, ProcessPID: old.pid, ProcessName: "claude", ProcessStart: old.start}
	writeRecord(t, home, record)
	if claimed {
		writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1, Status: records.ClaimLive})
	}
	closes := status == records.AgentLive || status == records.AgentUnresolved
	fake := newFakeHerdr(t)
	if closes {
		fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_w1:p1", "w1", "w1:t1", "/tmp")+`}`)
		fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(old.pid, "claude", "claude"))+`}`)
		fake.result("pane.close", `{"type":"ok"}`)
	}
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t2", "w1:p2"))
	scriptCoordinatorStart(fake, "w1:p2", "w1:t2", harness)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p2", "term_w1:p2", "w1", "w1:t2", "working", "coordinator")+`}`)

	if _, err := Restart(context.Background(), home, record, nil); err != nil {
		t.Fatalf("Restart() of a %s record, claimed=%v: %v", status, claimed, err)
	}
	wantCloses := 0
	if closes {
		wantCloses = 1
	}
	if got := len(fake.recorded("pane.close")); got != wantCloses {
		t.Fatalf("pane.close = %d; want %d", got, wantCloses)
	}
	if got := readAgent(t, home, "coordinator-g1"); got.Status != records.AgentStale {
		t.Fatalf("old record = %q; want stale", got.Status)
	}
	next := readAgent(t, home, "coordinator-g2")
	if next.Status != records.AgentLive || next.PaneID != "w1:p2" || next.Replaces != "coordinator-g1" {
		t.Fatalf("new record = %#v; want live coordinator-g2 in w1:p2 that replaces coordinator-g1", next)
	}
	if claim, ok := readCoordinatorClaim(t, home); !ok || claim.Agent != "coordinator-g2" || claim.Generation != 2 || claim.Status != records.ClaimLive {
		t.Fatalf("claim = %#v, %v; want a live claim of coordinator-g2", claim, ok)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt = %d; want 1", got)
	}
}

// TestCoordinatorRestartRefusesStaleRecordThatTheClaimDoesNotName covers
// decision 34: a stale record is accepted only when no claim names another
// record.
func TestCoordinatorRestartRefusesStaleRecordThatTheClaimDoesNotName(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	old := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", WorkspaceID: "w1", Status: records.AgentStale, Generation: 1, StartedByHive: true}
	current := records.Agent{Key: "coordinator-g2", Name: "coordinator", Role: "coordinator", PaneID: "w1:p2", WorkspaceID: "w1", Status: records.AgentLive, Generation: 2, StartedByHive: true}
	writeRecord(t, home, old)
	writeRecord(t, home, current)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g2", Generation: 2, Status: records.ClaimLive})
	fake := newFakeHerdr(t)
	before := stateFiles(t, home)

	if _, err := Restart(context.Background(), home, old, nil); err == nil || !strings.Contains(err.Error(), "coordinator-g2") {
		t.Fatalf("Restart() = %v; want a refusal that names the claim holder", err)
	}
	after := stateFiles(t, home)
	if len(after) != len(before) {
		t.Fatalf("state files = %d; want %d", len(after), len(before))
	}
	for path, data := range before {
		if after[path] != data {
			t.Fatalf("%s changed after the refused restart", path)
		}
	}
	if methods := fake.requestedMethods(); len(methods) != 0 {
		t.Fatalf("herdr requests = %#v; want none", methods)
	}
}
