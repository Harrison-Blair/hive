package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// TestSpawnTaskPaneMustBeInWorktree proves that a supplied pane obeys the
// worktree placement of a task agent. A pane outside the worktree is refused
// before any record or start. A pane in the worktree still places and assigns.
func TestSpawnTaskPaneMustBeInWorktree(t *testing.T) {
	t.Run("outside pane is refused before effects", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptWorkerSuccess(fake, sleep, "wWrong:p1", "wWrong", "wWrong:t1", home, "wrong-place")

		envelope := Spawn(context.Background(), home, SpawnOptions{
			Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "wrong-place", NameSet: true, PaneID: "wWrong:p1", PaneSet: true,
		})
		if envelope.Outcome != "failed" {
			t.Fatalf("envelope = %#v; want a failed placement", envelope)
		}
		if got := toString(envelope.Error); !strings.Contains(got, "worktree pane mismatch") {
			t.Fatalf("error = %q; want a worktree pane mismatch", got)
		}
		if starts := len(fake.recorded("agent.start")); starts != 0 {
			t.Fatalf("agent.start requests = %d; want 0", starts)
		}
		if creates := len(fake.recorded("workspace.create")); creates != 0 {
			t.Fatalf("workspace.create requests = %d; want 0", creates)
		}
		task := readTaskRecord(t, home, "1-worker-demo")
		if task.Status != "created" || task.ActingAgent != "" {
			t.Fatalf("task = %#v; want an unchanged created task", task)
		}
		if _, err := os.Stat(filepath.Join(home, ".hive/state/agents/wrong-place.json")); !os.IsNotExist(err) {
			t.Fatalf("a refused placement wrote an agent record: %v", err)
		}
	})

	t.Run("worktree pane still assigns", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "builder-1")

		envelope := Spawn(context.Background(), home, SpawnOptions{
			Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "builder-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
		})
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		task := readTaskRecord(t, home, "1-worker-demo")
		if task.Status != "assigned" || task.ActingAgent != "builder-1" {
			t.Fatalf("task = %#v; want assigned to builder-1", task)
		}
	})
}

// TestRegisterOwnerRefusesLiveHolder proves the owner rule under the store
// lock, which is the cross-process boundary of a spawn.
func TestRegisterOwnerRefusesLiveHolder(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "demo")
	first := Register(home, RegisterOptions{Role: "manager", Project: "demo", PaneID: "w1:p1"})
	if first.Outcome != "ok" {
		t.Fatalf("first owner = %#v; want ok", first)
	}
	before := readClaim(t, home, "manager-demo")

	second := Register(home, RegisterOptions{Role: "manager", Project: "demo", PaneID: "w1:p2", Name: "owner-two", NameSet: true})
	if second.Outcome != "failed" || toString(second.Error) != "project demo has an owner: manager-demo" {
		t.Fatalf("second owner = %#v; want the live owner refusal", second)
	}
	if len(second.Effects) != 0 {
		t.Fatalf("second owner effects = %#v; want none", second.Effects)
	}
	if after := readClaim(t, home, "manager-demo"); after != before {
		t.Fatalf("claim changed: before %#v after %#v", before, after)
	}
	if record := readAgent(t, home, firstResultKey(t, first)); record.Status != "live" {
		t.Fatalf("first record = %#v; want live", record)
	}
}

func firstResultKey(t *testing.T, envelope output.Envelope) string {
	t.Helper()
	result, ok := envelope.Result.(RegisterResult)
	if !ok {
		t.Fatalf("result = %T; want RegisterResult", envelope.Result)
	}
	return result.Agent.Key
}

// TestSpawnOwnerConcurrentExclusivity proves that one project keeps one live
// owner when two spawns pass the preflight together. The second spawn must
// refuse under the store lock, send no start request, and leave the first
// claim, record, and workspace in place.
func TestSpawnOwnerConcurrentExclusivity(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBE", "hive:demo", dir, "wBE:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")
	scriptManagerSpawnTail(t, home, fake, sleep, "wBE:p1", "wBE", "wBE:t1")

	firstReached := make(chan struct{})
	secondReached := make(chan struct{})
	firstDone := make(chan struct{})
	var mu sync.Mutex
	count := 0
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method != "workspace.create" {
			return
		}
		mu.Lock()
		count++
		number := count
		mu.Unlock()
		switch number {
		case 1:
			close(firstReached)
			<-secondReached
		case 2:
			close(secondReached)
			<-firstDone
		}
	})

	firstCh := make(chan output.Envelope, 1)
	go func() {
		firstCh <- Spawn(context.Background(), home, managerSpawnOptions("demo"))
	}()
	<-firstReached
	secondCh := make(chan output.Envelope, 1)
	go func() {
		secondCh <- Spawn(context.Background(), home, managerSpawnOptions("demo"))
	}()
	<-secondReached
	first := <-firstCh
	close(firstDone)
	second := <-secondCh

	if first.Outcome != "ok" {
		t.Fatalf("first owner = %#v; want ok", first)
	}
	if second.Outcome != "failed" {
		t.Fatalf("second owner = %#v; want the live owner refusal", second)
	}
	if got := toString(second.Error); !strings.Contains(got, "project demo has an owner: manager-demo") {
		t.Fatalf("second error = %q; want the live owner refusal", got)
	}
	if effects := len(second.Effects); effects != 0 {
		t.Fatalf("second owner effects = %#v; want none", second.Effects)
	}
	if starts := len(fake.recorded("agent.start")); starts != 1 {
		t.Fatalf("agent.start requests = %d; want 1", starts)
	}
	claim := readClaim(t, home, "owner-demo")
	if claim.Generation != 1 || claim.Agent != "manager-demo-g1" {
		t.Fatalf("claim = %#v; want the first owner at generation 1", claim)
	}
	record := readAgent(t, home, "manager-demo-g1")
	if record.Status != "live" || record.PaneID != "wBD:p1" {
		t.Fatalf("first record = %#v; want live in wBD:p1", record)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBD" {
		t.Fatalf("project workspace = %#v; want the first workspace wBD", project.Workspace)
	}
}
