package agent

import (
	"context"
	"fmt"
	"testing"
)

func setupTaskPaneSpawn(t *testing.T, role string) string {
	t.Helper()
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	status := "created"
	if role == "verifier" {
		status = "completed"
	}
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, status))

	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	paneID, workspaceID, tabID, name := "w5:p1", "w5", "w5:t1", role+"-demo-1"
	scriptReadyShell(fake, paneID, "term_"+paneID, workspaceID, tabID, worktree)
	fake.result("agent.start", workerAgentStartJSON(paneID, workspaceID, tabID, name))
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("pane.process_info", workerProcessJSON(paneID, sleep))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	return home
}

func TestVerifierSpawnPaneUsesTaskProject(t *testing.T) {
	home := setupTaskPaneSpawn(t, "verifier")
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m",
		Name: "verifier-demo-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	record, found, err := loadAgentByName(home, "verifier-demo-1")
	if err != nil || !found {
		t.Fatalf("agent record found = %t, error = %v; want a record", found, err)
	}
	if record.Project != "demo" || record.Task != "1-worker-demo" || record.PaneID != "w5:p1" {
		t.Fatalf("agent record = %#v; want project demo, the task, and pane w5:p1", record)
	}
}

func TestWorkerSpawnPaneUsesTaskProject(t *testing.T) {
	home := setupTaskPaneSpawn(t, "worker")
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m",
		Name: "worker-demo-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	record, found, err := loadAgentByName(home, "worker-demo-1")
	if err != nil || !found {
		t.Fatalf("agent record found = %t, error = %v; want a record", found, err)
	}
	if record.Project != "demo" || record.Task != "1-worker-demo" || record.PaneID != "w5:p1" {
		t.Fatalf("agent record = %#v; want project demo, the task, and pane w5:p1", record)
	}
}

func TestSpawnPaneRejectsDifferentTaskProject(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	fake := newFakeHerdr(t)

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Project: "other", Task: "1-worker-demo", Harness: "claude", Model: "m",
		Name: "worker-demo-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
	})
	want := "task 1-worker-demo is in project demo"
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != want {
		t.Fatalf("envelope = %#v; want failure %q", envelope, want)
	}
	if _, found, err := loadAgentByName(home, "worker-demo-1"); err != nil || found {
		t.Fatalf("agent record found = %t, error = %v; want no record", found, err)
	}
	if len(fake.requests) != 0 {
		t.Fatalf("herdr requests = %#v; want none", fake.requests)
	}
}
