package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func tabInfoReply(tabID, workspaceID, label string) string {
	return `{"type":"tab_info","tab":{"tab_id":"` + tabID + `","workspace_id":"` + workspaceID + `","number":1,"label":"` + label + `","focused":false,"pane_count":1,"agent_status":"unknown"}}`
}

func assertTabRename(t *testing.T, fake *fakeHerdr, tabID, label string, wantCount int) {
	t.Helper()
	requests := fake.recorded("tab.rename")
	if len(requests) != wantCount {
		t.Fatalf("tab.rename requests = %d; want %d", len(requests), wantCount)
	}
	if wantCount == 0 {
		return
	}
	want := `{"tab_id":"` + tabID + `","label":"` + label + `"}`
	if requests[0].Params != want {
		t.Fatalf("tab.rename params = %s; want %s", requests[0].Params, want)
	}
}

func TestSpawnRenamesWorkerAndVerifierTabs(t *testing.T) {
	for _, role := range []string{"worker", "verifier"} {
		t.Run(role, func(t *testing.T) {
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
			paneID, tabID := "w5:p1", "w5:t1"
			if role == "worker" {
				scriptWorkerSuccess(fake, sleep, paneID, "w5", tabID, worktree, "worker-1")
			} else {
				paneID, tabID = "w5:p2", "w5:t2"
				fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
				fake.result("tab.create", tabCreateJSON("w5", tabID, paneID, worktree))
				fake.result("pane.get", workerPaneGetJSON(paneID, "w5", tabID, worktree, ""))
				fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
				scriptReadyShell(fake, paneID, "term_"+paneID, "w5", tabID, worktree)
				fake.result("agent.start", workerAgentStartJSON(paneID, "w5", tabID, "verifier-1"))
				fake.result("pane.get", workerPaneGetJSON(paneID, "w5", tabID, worktree, ""))
				fake.result("agent.wait", workerAgentInfoJSON(paneID, "w5", tabID, "idle", "verifier-1"))
				fake.result("agent.get", workerAgentInfoJSON(paneID, "w5", tabID, "idle", "verifier-1"))
				fake.result("pane.process_info", workerProcessJSON(paneID, sleep))
				fake.result("agent.get", workerAgentInfoJSON(paneID, "w5", tabID, "idle", "verifier-1"))
				fake.result("agent.prompt", workerAgentPromptedJSON(paneID, "w5", tabID, "verifier-1"))
				fake.result("agent.get", workerAgentInfoJSON(paneID, "w5", tabID, "working", "verifier-1"))
			}
			label := "worker-1"
			if role == "verifier" {
				label = "verifier-1"
			}
			fake.result("tab.rename", tabInfoReply(tabID, "w5", label))

			envelope := Spawn(context.Background(), home, SpawnOptions{
				Role: role, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: label, NameSet: true,
			})
			if envelope.Outcome != "ok" || envelope.Error != nil {
				t.Fatalf("Spawn() = %#v; want success", envelope)
			}
			assertTabRename(t, fake, tabID, label, 1)
		})
	}
}

func TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane(t *testing.T) {
	t.Run("manager root workspace", func(t *testing.T) {
		home := testHome(t)
		dir := projectDirectory(t, home, "demo")
		writeManagerProject(t, home, nil)
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		fake.result("worktree.list", managerWorktreeListJSON(""))
		fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
		fake.result("tab.rename", tabInfoReply("wBD:t1", "wBD", "manager-demo"))
		scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

		envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Spawn() = %#v; want success", envelope)
		}
		assertTabRename(t, fake, "wBD:t1", "manager-demo", 1)
	})

	t.Run("manager hive tab", func(t *testing.T) {
		home := testHome(t)
		dir := projectDirectory(t, home, "demo")
		writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", RootPane: "wBC:p1", TabID: "wBC:t1", TabCreatedByHive: true})
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "hive"))
		fake.result("worktree.list", managerWorktreeListJSON("wBC"))
		fake.result("tab.list", managerTabListJSON(managerTabJSON("wBC", "wBC:t1", 1, 1)))
		fake.result("pane.list", managerPaneListJSON(managerPaneInfoJSON("wBC:p1", "term_wBC", "wBC", "wBC:t1", dir, "")))
		fake.result("pane.process_info", managerProcessInfoJSON("wBC:p1", sleep))
		scriptManagerSpawnTail(t, home, fake, sleep, "wBC:p1", "wBC", "wBC:t1")

		envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Spawn() = %#v; want success", envelope)
		}
		assertTabRename(t, fake, "", "", 0)
	})

	t.Run("explicit pane", func(t *testing.T) {
		home := testHome(t)
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptSpawnSuccess(t, fake, home, sleep, "idle")
		envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Spawn() = %#v; want success", envelope)
		}
		assertTabRename(t, fake, "", "", 0)
	})
}

func TestSpawnRenameFailureIsNonFatalAndLogged(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.failure("tab.rename", "tab_not_found", "tab disappeared")
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Spawn() = %#v; want success after rename failure", envelope)
	}
	if !hasTabEffect(envelope.Effects, "rename failed", "wBD:t1") {
		t.Fatalf("effects = %#v; want failed tab rename", envelope.Effects)
	}
	record := readAgent(t, home, "manager-demo-g1")
	if record.Status != records.AgentLive {
		t.Fatalf("agent status = %q; want live", record.Status)
	}
	files, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("log files = %v, %v; want a log file", files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "agent tab rename failed") || !strings.Contains(string(data), "tab_not_found") {
		t.Fatalf("log does not report the tab rename error: %s", data)
	}
}

func hasTabEffect(effects []output.Effect, action, tabID string) bool {
	for _, effect := range effects {
		if effect.Resource == "tab" && effect.Path == tabID && effect.Action == action {
			return true
		}
	}
	return false
}
