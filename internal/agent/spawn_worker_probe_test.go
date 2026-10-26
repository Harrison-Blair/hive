package agent

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// agentSeqInfoJSON is one agent_info reply with an explicit state change
// sequence. It lets a test model a turn that completed before the probe read.
func agentSeqInfoJSON(paneID, terminalID, workspaceID, tabID, status, name string, seq uint64) string {
	return `{"type":"agent_info","agent":{` +
		`"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `",` +
		`"agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `","state_change_seq":` + jsonNumber(int64(seq)) + `}}`
}

// workerWorkspaceCountJSON is one workspace_info reply with explicit tab and
// pane counts.
func workerWorkspaceCountJSON(workspaceID, label, repoRoot, checkoutPath string, tabs, panes int) string {
	return `{"type":"workspace_info","workspace":{"workspace_id":"` + workspaceID + `","number":1,"label":"` + label + `","focused":false,` +
		`"pane_count":` + jsonNumber(int64(panes)) + `,"tab_count":` + jsonNumber(int64(tabs)) + `,"active_tab_id":"` + workspaceID + `:t1","agent_status":"unknown",` +
		`"worktree":{"repo_key":"` + repoRoot + `/.git","repo_name":"repo","repo_root":"` + repoRoot + `","checkout_path":"` + checkoutPath + `","is_linked_worktree":true}}}`
}

// TestWorkerSpawnCompletedTurnSubmitsOnce proves F01. A turn that completed
// before the probe read must produce exactly one prompt submission. The
// production probe observes the fake server; no test replaces it.
func TestWorkerSpawnCompletedTurnSubmitsOnce(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", agentSeqInfoJSON("w5:p1", "term_w5:p1", "w5", "w5:t1", "done", "worker-1", 42))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	for attempt := 0; attempt < 200; attempt++ {
		fake.result("agent.get", agentSeqInfoJSON("w5:p1", "term_w5:p1", "w5", "w5:t1", "done", "worker-1", 42))
	}
	for attempt := 0; attempt < promptAttempts; attempt++ {
		fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	}

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok with one submission", envelope)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	if !hasEffectAction(envelope.Effects, "prompted") {
		t.Fatalf("effects = %#v; want the prompted effect", envelope.Effects)
	}
}

// TestVerifierSpawnCleanupPreservesAddedTab proves F03. Cleanup must not close
// a workspace whose topology gained a user tab after worktree.open.
func TestVerifierSpawnCleanupPreservesAddedTab(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, false))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w6:p1", "term_w6:p1", "w6", "w6:t1", worktree)
	fake.failure("agent.start", "agent_pane_busy", "pane is busy")
	fake.result("workspace.get", workerWorkspaceCountJSON("w6", "1-worker-demo", repo, worktree, 2, 2))

	envelope := Spawn(context.Background(), home, SpawnOptions{Profile: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true})
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if closes := fake.recorded("workspace.close"); len(closes) != 0 {
		t.Fatalf("cleanup closed a workspace with an added user tab: %#v", closes)
	}
}
