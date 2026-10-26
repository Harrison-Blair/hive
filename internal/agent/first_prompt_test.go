package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func firstPromptWorkerFixture(t *testing.T) (string, *fakeHerdr) {
	t.Helper()
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
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	return home, fake
}

func workerPromptWaitReply(status string) string {
	return `{"type":"agent_prompted","agent":` + agentJSON("w5:p1", "term_w5:p1", "w5", "w5:t1", status, "worker-1") + `}`
}

func requireFirstPromptWait(t *testing.T, requests []fakeRequest) {
	t.Helper()
	if len(requests) == 0 {
		t.Fatal("agent.prompt requests = 0; want at least one")
	}
	for index, request := range requests {
		var params herdr.AgentPromptParams
		if err := json.Unmarshal([]byte(request.Params), &params); err != nil {
			t.Fatalf("decode prompt %d params: %v", index+1, err)
		}
		if params.Wait == nil {
			t.Fatalf("prompt %d params = %s; want wait", index+1, request.Params)
		}
		if got, want := params.Wait.Until, []string{"working", "blocked"}; fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("prompt %d wait.until = %v; want %v", index+1, got, want)
		}
		if params.Wait.TimeoutMS != 10000 {
			t.Fatalf("prompt %d wait.timeout_ms = %d; want 10000", index+1, params.Wait.TimeoutMS)
		}
	}
}

func TestWorkerFirstPromptRetriesOneStall(t *testing.T) {
	home, fake := firstPromptWorkerFixture(t)
	fake.failure("agent.prompt", herdr.CodeAgentPromptStalled, "the prompt did not start")
	fake.result("agent.prompt", workerPromptWaitReply("working"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome=%s error=%v; want successful spawn after one retry", envelope.Outcome, envelope.Error)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 2 {
		t.Fatalf("agent.prompt requests = %d; want first prompt and one retry", len(requests))
	}
	requireFirstPromptWait(t, requests)
	if text := SpawnText(envelope); !strings.Contains(text, "agent worker-1 live\n") {
		t.Fatalf("spawn text = %q; want live after confirmed prompt", text)
	}
}

func TestWorkerFirstPromptFailsAfterTwoStalls(t *testing.T) {
	home, fake := firstPromptWorkerFixture(t)
	fake.failure("agent.prompt", herdr.CodeAgentPromptStalled, "first wait expired")
	fake.failure("agent.prompt", herdr.CodeAgentPromptStalled, "second wait expired")

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome == "ok" {
		t.Fatalf("envelope = %#v; want failed spawn", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "worker-1") || !strings.Contains(fmt.Sprint(envelope.Error), "agent_prompt_stalled") {
		t.Fatalf("error = %v; want agent name and prompt-stalled cause", envelope.Error)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 2 {
		t.Fatalf("agent.prompt requests = %d; want exactly two", len(requests))
	}
	requireFirstPromptWait(t, requests)
	if text := SpawnText(envelope); strings.Contains(text, "agent worker-1 live") {
		t.Fatalf("spawn text = %q; must not report live without a confirmed prompt", text)
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal failed spawn: %v", err)
	}
	if strings.Contains(string(data), `"status":"live"`) {
		t.Fatalf("JSON spawn result = %s; must not report live without a confirmed prompt", data)
	}
}

func TestWorkerFirstPromptBlockedDoesNotRetry(t *testing.T) {
	home, fake := firstPromptWorkerFixture(t)
	fake.failure("agent.prompt", herdr.CodeAgentBlocked, "trust dialog is open")

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome == "ok" {
		t.Fatalf("envelope = %#v; want blocked spawn", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "agent blocked") || !strings.Contains(fmt.Sprint(envelope.Error), "worker-1") {
		t.Fatalf("error = %v; want blocked status and agent name", envelope.Error)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want one blocked prompt", len(requests))
	}
	requireFirstPromptWait(t, requests)
}
