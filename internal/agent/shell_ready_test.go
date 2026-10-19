package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func shellProcessInfoJSON(paneID string, shellPID, foregroundPID int64) string {
	return shellProcessInfoJSONWithEntries(paneID, shellPID, foregroundPID, processEntry(shellPID, "bash", "/usr/bin/bash"))
}

func shellProcessInfoJSONWithEntries(paneID string, shellPID, foregroundPID int64, entries ...string) string {
	return fmt.Sprintf(`{"type":"pane_process_info","process_info":{"pane_id":%q,"shell_pid":%d,"foreground_process_group_id":%d,"foreground_processes":[%s]}}`, paneID, shellPID, foregroundPID, strings.Join(entries, ","))
}

func requestsBefore(fake *fakeHerdr, stop, method string) int {
	count := 0
	for _, requestMethod := range requestMethods(fake) {
		if requestMethod == stop {
			return count
		}
		if requestMethod == method {
			count++
		}
	}
	return count
}

func TestManagerSpawnWaitsForShellSecondRead(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	for range 2 {
		fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("wBD:p1", "term_wBD", "wBD", "wBD:t1", managerRepoPath, "")+`}`)
	}
	fake.failure("pane.process_info", "pane_not_found", "shell process data is not ready")
	fake.result("pane.process_info", shellProcessInfoJSON("wBD:p1", 41, 41))
	scriptManagerSpawnTail(t, fake, sleep, "wBD:p1", "wBD", "wBD:t1", false)
	var shellReadStatuses []string
	startStatus := ""
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "pane.get" && method != "agent.start" {
			return
		}
		record := readAgent(t, home, "manager-demo-g1")
		if method == "pane.get" && len(shellReadStatuses) < 2 {
			shellReadStatuses = append(shellReadStatuses, record.Status)
		}
		if method == "agent.start" {
			startStatus = record.Status
		}
	})

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if starts := fake.recorded("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(starts))
	}
	if got := requestsBefore(fake, "agent.start", "pane.get"); got != 2 {
		t.Fatalf("pane.get requests before agent.start = %d; want 2", got)
	}
	if got := requestsBefore(fake, "agent.start", "pane.process_info"); got != 2 {
		t.Fatalf("pane.process_info requests before agent.start = %d; want 2", got)
	}
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(shellReadStatuses) != 2 || shellReadStatuses[0] != "starting" || shellReadStatuses[1] != "starting" {
		t.Fatalf("statuses during shell reads = %v; want [starting starting]", shellReadStatuses)
	}
	if startStatus != "starting" {
		t.Fatalf("status at agent.start = %q; want starting", startStatus)
	}
	if record := readAgent(t, home, "manager-demo-g1"); record.Status != "live" {
		t.Fatalf("final status = %q; want live", record.Status)
	}
}

func TestManagerSpawnWaitsForShellStartupProcessesToClear(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	for range 2 {
		fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("wBD:p1", "term_wBD", "wBD", "wBD:t1", managerRepoPath, "")+`}`)
	}
	fake.result("pane.process_info", shellProcessInfoJSONWithEntries(
		"wBD:p1", 41, 41,
		processEntry(41, "bash", "/usr/bin/bash"),
		processEntry(42, "bash", "/usr/bin/bash"),
		processEntry(43, "bash", "/usr/bin/bash"),
		processEntry(44, "bash", "/usr/bin/bash"),
		processEntry(45, "bash", "/usr/bin/bash"),
		processEntry(46, "tail", "tail", "-1"),
	))
	fake.result("pane.process_info", shellProcessInfoJSONWithEntries(
		"wBD:p1", 41, 41, processEntry(41, "bash", "/usr/bin/bash"),
	))
	scriptManagerSpawnTail(t, fake, sleep, "wBD:p1", "wBD", "wBD:t1", false)

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if got := requestsBefore(fake, "agent.start", "pane.process_info"); got != 2 {
		t.Fatalf("pane.process_info requests before agent.start = %d; want 2", got)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(starts))
	}
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
}

func TestWorkerSpawnWaitsForShellSecondRead(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("pane.process_info", shellProcessInfoJSON("w5:p1", 51, 52))
	fake.result("pane.process_info", shellProcessInfoJSON("w5:p1", 51, 51))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	startStatus := ""
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method == "agent.start" {
			startStatus = readAgent(t, home, "worker-1").Status
		}
	})
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "working", "worker-1"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if starts := fake.recorded("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(starts))
	}
	if got := requestsBefore(fake, "agent.start", "pane.get"); got != 3 {
		t.Fatalf("pane.get requests before agent.start = %d; want 3", got)
	}
	if got := requestsBefore(fake, "agent.start", "pane.process_info"); got != 2 {
		t.Fatalf("pane.process_info requests before agent.start = %d; want 2", got)
	}
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if startStatus != "starting" || spawnResultAgent(envelope).Status != "live" {
		t.Fatalf("status at agent.start/final = %q/%q; want starting/live", startStatus, spawnResultAgent(envelope).Status)
	}
}

func TestManagerSpawnStopsWhenShellDeadlineExpires(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("wBD:p1", "term_wBD", "wBD", "wBD:t1", managerRepoPath, "")+`}`)
	fake.result("pane.process_info", shellProcessInfoJSON("wBD:p1", 41, 42))
	fake.failure("agent.start", "invalid_agent_name", "must not start")
	fake.result("workspace.close", `{"type":"workspace_closed","workspace_id":"wBD"}`)
	options := managerSpawnOptions("demo")
	options.shellReadyTimeout = 25 * time.Millisecond

	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	if message := fmt.Sprint(envelope.Error); !strings.Contains(message, "wBD:p1") || !strings.Contains(message, "25ms") {
		t.Fatalf("error = %q; want pane and deadline", message)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 0 {
		t.Fatalf("agent.start requests = %d; want none", len(starts))
	}
}
