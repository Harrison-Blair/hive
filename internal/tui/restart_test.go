package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestAgentsRestartConfirmsLiveCoordinatorAndCancelsOnOtherKey(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	server := newStartupHerdr(t)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups: []agentGroup{{id: "home", title: "home", keys: []string{"coordinator-g1"}}},
		nodes: []agentNode{{
			record: records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p2", Status: records.AgentLive},
			status: "idle", current: true,
		}},
		supervisorLive: true,
	})

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil {
		t.Fatal("restart request did not start a coordinator pane check")
	}
	m = runCoordinatorRestartCheck(t, m, cmd)
	if got := len(requestsFor(server.requests, "pane.get")); got != 1 {
		t.Fatalf("pane.get requests = %d; want 1", got)
	}
	requireView(t, m, "restart coordinator? y/n")

	m, cmd = raise(t, m, key('n', "n"))
	if cmd == nil {
		t.Fatal("restart cancellation did not log a user action")
	}
	m = deliverAgentCommands(t, m, cmd)
	requireView(t, m, "restart cancelled")
	if strings.Contains(m.View().Content, "restarting") {
		t.Fatal("cancelled restart is still running")
	}
	logData, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent restart request"`, `"action":"agent restart confirmation"`, `"result":"cancelled"`} {
		if !strings.Contains(string(logData), want) {
			t.Fatalf("restart action log omits %q:\n%s", want, logData)
		}
	}
}

func TestAgentsRestartMissingCoordinatorPaneTakesStalePath(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	server := newStartupHerdr(t)
	server.paneFailures = map[string]startupPaneFailure{
		"w1:p9": {Code: "pane_not_found", Message: "pane w1:p9 not found"},
	}
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p9", WorkspaceID: "w1", Status: records.AgentLive}
	writeRecord(t, home, []string{"agents"}, record.Key, record)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups:         []agentGroup{{id: "home", title: "home", keys: []string{record.Key}}},
		nodes:          []agentNode{{record: record, status: "idle", current: true}},
		supervisorLive: true,
	})

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil {
		t.Fatal("restart request did not start the coordinator pane check")
	}
	m = runCoordinatorRestartCheck(t, m, cmd)
	if got := len(requestsFor(server.requests, "pane.get")); got != 1 {
		t.Fatalf("pane.get requests = %d; want 1 before the restart decision", got)
	}
	if m.agents.restartConfirm || !m.agents.restartRunning || m.agents.restartTarget.Status != records.AgentStale {
		t.Fatalf("restart state = confirm %v, running %v, target status %q; want immediate stale restart", m.agents.restartConfirm, m.agents.restartRunning, m.agents.restartTarget.Status)
	}
	requireView(t, m, "restarting")
	if got := readStartupAgent(t, home, record.Key); got.Status != records.AgentStale {
		t.Fatalf("coordinator status = %q; want stale", got.Status)
	}
}

func TestAgentsRestartCoordinatorLookupErrorKeepsConfirmation(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	server := newStartupHerdr(t)
	server.paneFailures = map[string]startupPaneFailure{
		"w1:p9": {Code: "server_error", Message: "pane lookup unavailable"},
	}
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p9", WorkspaceID: "w1", Status: records.AgentLive}
	writeRecord(t, home, []string{"agents"}, record.Key, record)
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups:         []agentGroup{{id: "home", title: "home", keys: []string{record.Key}}},
		nodes:          []agentNode{{record: record, status: "idle", current: true}},
		supervisorLive: true,
	})

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil {
		t.Fatal("restart request did not start the coordinator pane check")
	}
	m = runCoordinatorRestartCheck(t, m, cmd)
	if got := len(requestsFor(server.requests, "pane.get")); got != 1 {
		t.Fatalf("pane.get requests = %d; want 1 before the restart decision", got)
	}
	if !m.agents.restartConfirm || m.agents.restartRunning {
		t.Fatalf("restart state = confirm %v, running %v; want confirmation after lookup error", m.agents.restartConfirm, m.agents.restartRunning)
	}
	requireView(t, m, "restart coordinator? y/n")
	if got := readStartupAgent(t, home, record.Key); got.Status != records.AgentLive {
		t.Fatalf("coordinator status = %q; want live after lookup error", got.Status)
	}
}

func runCoordinatorRestartCheck(t *testing.T, model *Model, command tea.Cmd) *Model {
	t.Helper()
	msg := command()
	if msg == nil {
		return model
	}
	commands, ok := msg.(tea.BatchMsg)
	if !ok {
		updated, _ := model.Update(msg)
		return updated.(*Model)
	}
	for _, child := range commands {
		childMsg := child()
		if childMsg == nil {
			continue
		}
		updated, _ := model.Update(childMsg)
		model = updated.(*Model)
	}
	return model
}

func TestAgentsHistoryCanShowStoppedRestartGeneration(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups: []agentGroup{{id: "home", title: "home", keys: []string{"coordinator-g1"}}},
		nodes: []agentNode{{
			record: records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p2", Status: records.AgentStopped},
			status: "stopped", current: true, stopped: true,
		}},
	})
	if strings.Contains(m.currentScreen().View(120, 30), "coordinator · stopped") {
		t.Fatal("stopped coordinator is visible before history is enabled")
	}
	m, _ = raise(t, m, key('h', "h"))
	requireAgentsView(t, m, "coordinator")
}

func TestAgentsRestartLiveSupervisorAsksForConfirmation(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups: []agentGroup{{id: "home", title: "home", keys: []string{"supervisor-g1"}}},
		nodes: []agentNode{{
			record: records.Agent{Key: "supervisor-g1", Name: "supervisor", Role: "supervisor", PaneID: "w1:p3", Status: records.AgentLive},
			status: "idle", current: true,
		}},
		supervisorLive: true,
	})

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil || !m.agents.restartConfirm || m.agents.restartRunning {
		t.Fatalf("supervisor restart state = confirm %v, running %v, cmd %v; want confirmation", m.agents.restartConfirm, m.agents.restartRunning, cmd != nil)
	}
	requireView(t, m, "restart supervisor? y/n")
}

func TestAgentsRestartResultIsRenderedAndLogged(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m, cmd := raise(t, m, agentRestartMsg{
		agentKey: "coordinator-g2",
		result:   agent.RestartResult{Tab: herdr.TabCreated{Tab: herdr.TabInfo{TabID: "w1:t2", Label: "coordinator"}}},
	})
	requireView(t, m, "restarted coordinator")
	m = deliverAgentCommands(t, m, cmd)
	logData, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logData), `"action":"agent restart result"`) || !strings.Contains(string(logData), `"result":"started"`) {
		t.Fatalf("restart result was not logged:\n%s", logData)
	}
}

func TestAgentsRestartStaleCoordinatorStartsWithoutConfirmation(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups: []agentGroup{{id: "home", title: "home", keys: []string{"coordinator-g1"}}},
		nodes: []agentNode{{
			record: records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p2", WorkspaceID: "w1", Status: records.AgentStale},
			status: "unknown", current: true,
		}},
	})

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil || !m.agents.restartRunning || m.agents.restartConfirm {
		t.Fatalf("stale restart state = running %v, confirm %v, cmd %v; want immediate restart", m.agents.restartRunning, m.agents.restartConfirm, cmd != nil)
	}
	requireView(t, m, "restarting")
}

func TestHeaderMarksOldSupervisorBinaryAndShowsRestartHint(t *testing.T) {
	home := newTestHome(t)
	source, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	oldDir := t.TempDir()
	oldExecutable := filepath.Join(oldDir, "supervisor")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	if err := os.WriteFile(oldExecutable, data, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(oldExecutable, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	processCmd := exec.Command(oldExecutable, "-test.run=^TestSupervisorHeaderOldProcess$")
	if err := processCmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = processCmd.Process.Kill()
		_ = processCmd.Wait()
	})
	pidStart, err := process.Start(int64(processCmd.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor-g1", Generation: 1, PID: int64(processCmd.Process.Pid), PIDStart: pidStart,
	})
	installedDir := t.TempDir()
	installed := filepath.Join(installedDir, "hive")
	if err := os.WriteFile(installed, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	installedTime := oldTime.Add(time.Hour)
	if err := os.Chtimes(installed, installedTime, installedTime); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", installedDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	state, err := readHeader(home)
	if err != nil {
		t.Fatalf("readHeader(): %v", err)
	}
	if state.supervisor != "fresh" || !state.supervisorOld {
		t.Fatalf("supervisor header state = %#v, want fresh old supervisor", state)
	}
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.width = 120
	m.header = state
	requireView(t, m, "supervisor old", "r to restart")
}

func TestSupervisorHeaderOldProcess(t *testing.T) {
	if filepath.Base(os.Args[0]) != "supervisor" {
		return
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestAgentsRestartKeyExplainsUnsupportedNode(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 1
	m.agents.accept(agentsData{
		groups: []agentGroup{{id: "demo", title: "demo", keys: []string{"manager-demo"}}},
		nodes: []agentNode{{
			record: records.Agent{Key: "manager-demo", Name: "manager-demo", Role: "manager", Status: records.AgentLive},
			status: "working", current: true,
		}},
		supervisorLive: true,
	})

	m, _ = raise(t, m, key('r', "r"))
	requireView(t, m, "restart applies only to the coordinator and the supervisor")
}
