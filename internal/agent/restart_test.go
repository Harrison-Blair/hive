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

func TestRestartCoordinatorClosesOwnedPaneStartsConfiguredCoordinatorAndKeepsRecord(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_1", "w1", "w1:t1", "/tmp")+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t2", "w1:p2"))
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p2","terminal_id":"term_2","workspace_id":"w1","tab_id":"w1:t2","focused":false,"agent_status":"unknown","revision":1}}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p2","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["bash"],"cmdline":"bash","cwd":"/tmp"}]}}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p2", "term_2", "w1", "w1:t2", "unknown", "coordinator")+`,"argv":["claude"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p2", "term_2", "w1", "w1:t2", "idle", "coordinator")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p2", "term_2", "w1", "w1:t2", "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p2", processEntry(harness.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p2", "term_2", "w1", "w1:t2", "working", "coordinator")+`}`)

	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, PaneID: "w1:p1", WorkspaceID: "w1", Status: records.AgentLive, StartedByHive: true}
	writeRestartRecord(t, home, record)
	result, err := Restart(context.Background(), home, record, nil)
	if err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	if result.Tab.Tab.TabID != "w1:t2" || result.Tab.RootPane.PaneID != "w1:p2" {
		t.Fatalf("restart tab = %#v, want coordinator tab w1:t2 and pane w1:p2", result.Tab)
	}
	if len(fake.recorded("pane.close")) != 1 {
		t.Fatalf("pane.close calls = %d, want one for the Hive-started pane", len(fake.recorded("pane.close")))
	}
	if len(fake.recorded("pane.wait_for_output")) != 1 {
		t.Fatalf("pane.wait_for_output calls = %d, want a wait for the new coordinator shell prompt", len(fake.recorded("pane.wait_for_output")))
	}
	if len(fake.recorded("agent.prompt")) != 1 {
		t.Fatalf("agent.prompt calls = %d, want the coordinator startup prompt", len(fake.recorded("agent.prompt")))
	}
	var create struct {
		WorkspaceID string            `json:"workspace_id"`
		Cwd         string            `json:"cwd"`
		Label       string            `json:"label"`
		Env         map[string]string `json:"env"`
	}
	if err := json.Unmarshal([]byte(fake.recorded("tab.create")[0].Params), &create); err != nil {
		t.Fatal(err)
	}
	if create.WorkspaceID != "w1" || create.Cwd != home || create.Label != "coordinator" || create.Env["HIVE_STARTUP_KIND"] != "coordinator" {
		t.Fatalf("tab.create = %#v, want the coordinator startup environment", create)
	}
	old := readRestartRecord(t, home, record.Key)
	if old.Status != records.AgentStale {
		t.Fatalf("old coordinator status = %q, want stale (decision 34)", old.Status)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", record.Key+".json")); err != nil {
		t.Fatalf("old coordinator record was removed: %v", err)
	}
}

func TestRestartSupervisorKeepsUnownedPaneAndReportsIt(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	fake.result("tab.create", restartTabCreatedJSON("supervisor", "w1:t3", "w1:p3"))
	fake.result("pane.send_input", `{"type":"ok"}`)

	record := records.Agent{Key: "supervisor-g2", Name: "supervisor", Kind: records.AgentKindSupervisor, PaneID: "w1:p2", WorkspaceID: "w1", Status: records.AgentLive}
	writeRestartRecord(t, home, record)
	result, err := Restart(context.Background(), home, record, nil)
	if err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	if len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("pane.close calls = %d, want none for a pane Hive did not start", len(fake.recorded("pane.close")))
	}
	if len(fake.recorded("pane.send_input")) != 1 || !strings.Contains(fake.recorded("pane.send_input")[0].Params, "hive supervise") {
		t.Fatalf("pane.send_input = %#v, want hive supervise", fake.recorded("pane.send_input"))
	}
	if result.Warning == "" {
		t.Fatal("restart result does not tell the user that the old pane was kept")
	}
	old := readRestartRecord(t, home, record.Key)
	if old.Status != records.AgentStale {
		t.Fatalf("old supervisor status = %q, want stale", old.Status)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", record.Key+".json")); err != nil {
		t.Fatalf("old supervisor record was removed: %v", err)
	}
}

func TestRestartSupervisorClosesOwnedPaneAndRunsSupervise(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("tab.create", restartTabCreatedJSON("supervisor", "w1:t5", "w1:p5"))
	fake.result("pane.send_input", `{"type":"ok"}`)
	record := records.Agent{Key: "supervisor-g3", Name: "supervisor", Kind: records.AgentKindSupervisor, PaneID: "w1:p4", WorkspaceID: "w1", Status: records.AgentLive, StartedByHive: true}
	writeRestartRecord(t, home, record)

	result, err := Restart(context.Background(), home, record, nil)
	if err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	if len(fake.recorded("pane.close")) != 1 {
		t.Fatalf("pane.close calls = %d, want one for the Hive-started supervisor", len(fake.recorded("pane.close")))
	}
	inputs := fake.recorded("pane.send_input")
	if len(inputs) != 1 || !strings.Contains(inputs[0].Params, "hive supervise") {
		t.Fatalf("pane.send_input = %#v, want hive supervise", inputs)
	}
	if result.Tab.Tab.Label != "supervisor" || result.Warning != "" {
		t.Fatalf("restart result = %#v, want supervisor tab without warning", result)
	}
	if old := readRestartRecord(t, home, record.Key); old.Status != records.AgentStopped {
		t.Fatalf("old supervisor status = %q, want stopped", old.Status)
	}
}

func TestRestartCoordinatorDoesNotCloseUnownedPaneAndWarns(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t6", "w1:p6"))
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p6","terminal_id":"term_6","workspace_id":"w1","tab_id":"w1:t6","focused":false,"agent_status":"unknown","revision":1}}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p6","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["bash"],"cmdline":"bash","cwd":"/tmp"}]}}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p6", "term_6", "w1", "w1:t6", "unknown", "coordinator")+`,"argv":["claude"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p6", "term_6", "w1", "w1:t6", "idle", "coordinator")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p6", "term_6", "w1", "w1:t6", "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p6", processEntry(harness.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p6", "term_6", "w1", "w1:t6", "working", "coordinator")+`}`)
	record := records.Agent{Key: "coordinator-g4", Name: "coordinator", Kind: records.AgentKindCoordinator, PaneID: "w1:p5", WorkspaceID: "w1", Status: records.AgentLive}
	writeRestartRecord(t, home, record)

	result, err := Restart(context.Background(), home, record, nil)
	if err != nil {
		t.Fatalf("Restart(): %v", err)
	}
	if len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("pane.close calls = %d, want none for an unowned coordinator pane", len(fake.recorded("pane.close")))
	}
	if !strings.Contains(result.Warning, "old pane kept") {
		t.Fatalf("restart warning = %q, want the old pane warning", result.Warning)
	}
	if old := readRestartRecord(t, home, record.Key); old.Status != records.AgentStale {
		t.Fatalf("old coordinator status = %q, want stale", old.Status)
	}
}

func TestRestartStaleCoordinatorStartsWithoutClosingAnUnownedPane(t *testing.T) {
	home := testHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("tab.create", restartTabCreatedJSON("coordinator", "w1:t4", "w1:p4"))
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("pane.get", `{"type":"pane_info","pane":{"pane_id":"w1:p4","terminal_id":"term_4","workspace_id":"w1","tab_id":"w1:t4","focused":false,"agent_status":"unknown","revision":1}}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p4","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["bash"],"cmdline":"bash","cwd":"/tmp"}]}}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p4", "term_4", "w1", "w1:t4", "unknown", "coordinator")+`,"argv":["claude"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p4", "term_4", "w1", "w1:t4", "idle", "coordinator")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p4", "term_4", "w1", "w1:t4", "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p4", processEntry(harness.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p4", "term_4", "w1", "w1:t4", "working", "coordinator")+`}`)

	record := records.Agent{Key: "coordinator-g3", Name: "coordinator", Kind: records.AgentKindCoordinator, PaneID: "w1:p3", WorkspaceID: "w1", Status: records.AgentStale}
	writeRestartRecord(t, home, record)
	if _, err := Restart(context.Background(), home, record, nil); err != nil {
		t.Fatalf("Restart() for stale coordinator: %v", err)
	}
	if len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("pane.close calls = %d, want none for the unowned stale pane", len(fake.recorded("pane.close")))
	}
	if len(fake.recorded("agent.prompt")) != 1 {
		t.Fatal("stale coordinator did not start in a new tab")
	}
}

func restartTabCreatedJSON(label, tabID, paneID string) string {
	return `{"type":"tab_created","tab":{"tab_id":"` + tabID + `","workspace_id":"w1","number":2,"label":"` + label + `","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"w1","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp"}}`
}

func writeRestartRecord(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func readRestartRecord(t *testing.T, home, key string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
