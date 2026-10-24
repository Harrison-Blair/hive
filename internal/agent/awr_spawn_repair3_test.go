package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// queueTaskPaneCheck scripts the worktree pane check of one task spawn.
func queueTaskPaneCheck(fake *fakeHerdr, paneID, workspaceID, tabID, cwd, taskID, repo string) {
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON(workspaceID, taskID, repo, cwd))
}

// queueTaskSpawnTail scripts the start, wait, and prompt of one placed agent.
func queueTaskSpawnTail(fake *fakeHerdr, sleep sleepProcess, paneID, workspaceID, tabID, cwd, name string) {
	fake.result("agent.start", workerAgentStartJSON(paneID, workspaceID, tabID, name))
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("agent.wait", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("pane.process_info", workerProcessJSON(paneID, sleep))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.prompt", workerAgentPromptedJSON(paneID, workspaceID, tabID, name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "working", name))
}

// scriptTaskBind scripts a task spawn that binds in a new tab of the worktree
// workspace. Such a spawn changes no task lifecycle field.
func scriptTaskBind(fake *fakeHerdr, sleep sleepProcess, workspaceID, tabID, paneID, cwd, taskID, repo, name string) {
	fake.result("workspace.get", workerWorkspaceGetJSON(workspaceID, taskID, repo, cwd))
	fake.result("tab.create", tabCreateJSON(workspaceID, tabID, paneID, cwd))
	queueTaskPaneCheck(fake, paneID, workspaceID, tabID, cwd, taskID, repo)
	scriptReadyShell(fake, paneID, "term_"+paneID, workspaceID, tabID, cwd)
	queueTaskSpawnTail(fake, sleep, paneID, workspaceID, tabID, cwd, name)
}

// TestSpawnTaskBindingEveryStatus proves the task binding rule of coordinator
// ruling 902 for each status. Only created and assigned tasks without a live
// acting agent select one. Each other spawn binds and keeps the status, the
// acting agent, the delivery, and the cancel reason.
func TestSpawnTaskBindingEveryStatus(t *testing.T) {
	delivery := &records.Delivery{Mode: "merge", Target: "dev", Grant: "g13", Commit: "abc123", Pushed: true}
	cases := []struct {
		name        string
		status      string
		actor       string
		actorStatus string
		delivery    *records.Delivery
		cancel      string
		selects     bool
		wantStatus  string
		wantActor   string
	}{
		{name: "created without an actor selects", status: "created", selects: true, wantStatus: "assigned", wantActor: "new-worker"},
		{name: "assigned without an actor selects", status: "assigned", selects: true, wantStatus: "assigned", wantActor: "new-worker"},
		{name: "assigned with a stopped actor selects", status: "assigned", actor: "old-worker", actorStatus: records.AgentExited, selects: true, wantStatus: "assigned", wantActor: "new-worker"},
		{name: "created with a live actor binds", status: "created", actor: "old-worker", actorStatus: records.AgentLive, wantStatus: "created", wantActor: "old-worker"},
		{name: "assigned with a live actor binds", status: "assigned", actor: "old-worker", actorStatus: records.AgentLive, wantStatus: "assigned", wantActor: "old-worker"},
		{name: "completed binds", status: "completed", actor: "old-worker", actorStatus: records.AgentLive, wantStatus: "completed", wantActor: "old-worker"},
		{name: "verified binds and keeps delivery", status: "verified", actor: "old-worker", actorStatus: records.AgentLive, delivery: delivery, wantStatus: "verified", wantActor: "old-worker"},
		{name: "verified without an actor binds", status: "verified", delivery: delivery, wantStatus: "verified", wantActor: ""},
		{name: "cancelled binds and keeps the cancel reason", status: "cancelled", actor: "old-worker", actorStatus: records.AgentLive, cancel: "superseded", wantStatus: "cancelled", wantActor: "old-worker"},
		{name: "cancelled without an actor binds", status: "cancelled", cancel: "superseded", wantStatus: "cancelled", wantActor: ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			repo, worktree, branch := taskWorktree(t)
			writeProjectPath(t, home, "demo", repo)
			task := workerTask(home, repo, worktree, branch, test.status)
			task.ActingAgent = test.actor
			task.Delivery = test.delivery
			task.CancelReason = test.cancel
			writeTaskRecord(t, home, task)
			if test.actor != "" {
				writeRecord(t, home, records.Agent{
					Key: test.actor, Name: test.actor, Profile: records.ProfileBuilder,
					Project: "demo", Task: task.ID, PaneID: "wTask:pOld", WorkspaceID: "w5",
					Status: test.actorStatus,
				})
			}
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
			if test.selects {
				scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "new-worker")
			} else {
				scriptTaskBind(fake, sleep, "w5", "w5:t2", "w5:p2", worktree, task.ID, repo, "new-worker")
			}

			envelope := Spawn(context.Background(), home, SpawnOptions{
				Role: "worker", Task: task.ID, Harness: "claude", Model: "m",
				Name: "new-worker", NameSet: true,
			})
			if envelope.Outcome != "ok" {
				t.Fatalf("envelope = %#v; want ok", envelope)
			}
			after := readTaskRecord(t, home, task.ID)
			if after.Status != test.wantStatus || after.ActingAgent != test.wantActor {
				t.Fatalf("task status/acting = %q/%q; want %q/%q", after.Status, after.ActingAgent, test.wantStatus, test.wantActor)
			}
			if !reflect.DeepEqual(after.Delivery, test.delivery) {
				t.Fatalf("delivery = %#v; want %#v", after.Delivery, test.delivery)
			}
			if after.CancelReason != test.cancel {
				t.Fatalf("cancel reason = %q; want %q", after.CancelReason, test.cancel)
			}
			if !hasEffectAction(envelope.Effects, "prompted") {
				t.Fatalf("effects = %#v; want the first prompt", envelope.Effects)
			}
			if got := hasEffectAction(envelope.Effects, "assigned"); got != test.selects {
				t.Fatalf("assigned effect = %v; want %v", got, test.selects)
			}
		})
	}
}

// TestSpawnTaskConcurrentSelectionKeepsLiveActingAgent proves the atomic
// selection rule of coordinator ruling 902. Two spawns pass the preflight
// before either assignment. The first selection wins, and the second spawn
// still binds and prompts without a lifecycle change.
func TestSpawnTaskConcurrentSelectionKeepsLiveActingAgent(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)

	firstReached := make(chan struct{})
	secondReached := make(chan struct{})
	firstRelease := make(chan struct{})
	secondRelease := make(chan struct{})
	var mu sync.Mutex
	seen := map[string]bool{}
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "agent.start" {
			return
		}
		var request struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(params, &request); err != nil {
			return
		}
		mu.Lock()
		if seen[request.Name] {
			mu.Unlock()
			return
		}
		seen[request.Name] = true
		mu.Unlock()
		switch request.Name {
		case "task-first":
			close(firstReached)
			<-firstRelease
		case "task-second":
			close(secondReached)
			<-secondRelease
		}
	})

	// The first spawn reaches agent.start after its worktree pane check.
	queueTaskPaneCheck(fake, "w5:p1", "w5", "w5:t1", worktree, "1-worker-demo", repo)
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "task-first"))
	firstCh := make(chan output.Envelope, 1)
	go func() {
		firstCh <- Spawn(context.Background(), home, SpawnOptions{
			Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "task-first", NameSet: true, PaneID: "w5:p1", PaneSet: true,
		})
	}()
	<-firstReached

	// The second spawn passes the same preflight while the task has no acting
	// agent. It reaches agent.start before the first assignment.
	queueTaskPaneCheck(fake, "w5:p2", "w5", "w5:t2", worktree, "1-worker-demo", repo)
	scriptReadyShell(fake, "w5:p2", "term_w5:p2", "w5", "w5:t2", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p2", "w5", "w5:t2", "task-second"))
	secondCh := make(chan output.Envelope, 1)
	go func() {
		secondCh <- Spawn(context.Background(), home, SpawnOptions{
			Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "task-second", NameSet: true, PaneID: "w5:p2", PaneSet: true,
		})
	}()
	<-secondReached

	// Let the first spawn finish its start, select the acting agent, and prompt.
	queueTaskSpawnTail(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "task-first")
	close(firstRelease)
	first := <-firstCh
	if first.Outcome != "ok" {
		t.Fatalf("first spawn = %#v; want ok", first)
	}
	winner := readTaskRecord(t, home, "1-worker-demo")
	if winner.Status != "assigned" || winner.ActingAgent != "task-first" {
		t.Fatalf("task after first spawn = %q/%q; want assigned/task-first", winner.Status, winner.ActingAgent)
	}

	// Let the second spawn finish: it binds and prompts but keeps the winner.
	queueTaskSpawnTail(fake, sleep, "w5:p2", "w5", "w5:t2", worktree, "task-second")
	close(secondRelease)
	second := <-secondCh
	if second.Outcome != "ok" {
		t.Fatalf("second spawn = %#v; want ok", second)
	}
	after := readTaskRecord(t, home, "1-worker-demo")
	if after.Status != "assigned" || after.ActingAgent != "task-first" {
		t.Fatalf("task after both spawns = %q/%q; want assigned/task-first", after.Status, after.ActingAgent)
	}
	if hasEffectAction(second.Effects, "assigned") {
		t.Fatalf("second effects = %#v; want no assigned effect", second.Effects)
	}
	if !hasEffectAction(second.Effects, "prompted") {
		t.Fatalf("second effects = %#v; want the first prompt", second.Effects)
	}
	if firstRecord := readAgent(t, home, "task-first"); firstRecord.Status != records.AgentLive {
		t.Fatalf("first record = %q; want live", firstRecord.Status)
	}
	if secondRecord := readAgent(t, home, "task-second"); secondRecord.Status != records.AgentLive {
		t.Fatalf("second record = %q; want live", secondRecord.Status)
	}
}

// TestSpawnOwnerRechecksCallerBeforeRegistration proves ruling 904: the owner
// registration lock reads the caller identity and claim facts again after the
// remote calls. A caller that becomes an agent, or a coordinator claim that
// goes away, refuses the registration with no record, claim, start, or effect.
func TestSpawnOwnerRechecksCallerBeforeRegistration(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(locked *store.LockedStore) error
		want   string
	}{
		{
			name: "identity",
			mutate: func(locked *store.LockedStore) error {
				record := records.Agent{
					Key: "manager-late-g1", Name: "manager-late", Role: "agent",
					Project: "demo", PaneID: "wHome:pUser", WorkspaceID: "wHome",
					Status: records.AgentLive, Generation: 1,
				}
				record.ApplyRoleFacts("")
				return locked.WriteIn([]string{"agents"}, record.Key, record)
			},
			want: "--owner needs the coordinator or the user: manager-late",
		},
		{
			name: "claim",
			mutate: func(locked *store.LockedStore) error {
				if err := locked.RemoveIn([]string{"claims"}, "coordinator"); err != nil {
					return err
				}
				record := records.Agent{
					Key: "worker-late", Name: "worker-late", Role: "worker",
					Project: "demo", PaneID: "wHome:p9", WorkspaceID: "wHome",
					Status: records.AgentLive, Generation: 1,
				}
				record.ApplyRoleFacts("")
				return locked.WriteIn([]string{"agents"}, record.Key, record)
			},
			want: "--owner needs the coordinator or the user: the caller",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			dir := projectDirectory(t, home, "demo")
			writeManagerProject(t, home, nil)
			writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
			writeRecord(t, home, records.Agent{
				Key: "coordinator-g1", Name: "coordinator", Role: "coordinator",
				Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator,
				PaneID: "wHome:pC", WorkspaceID: "wHome", Status: records.AgentLive, Generation: 1,
			})
			t.Setenv("HERDR_PANE_ID", "wHome:pUser")
			t.Chdir(t.TempDir())
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("wHome:pUser", "term_user", "wHome", "wHome:t1", home)+`}`)
			fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
			fake.result("workspace.close", `{"type":"ok"}`)
			scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")
			registry, err := store.Open(home)
			if err != nil {
				t.Fatal(err)
			}
			var mutationErr error
			fake.setBefore(func(method string, params json.RawMessage) {
				if method != "workspace.create" {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()
				mutationErr = registry.WithLock(ctx, test.mutate)
			})

			envelope := SpawnFromCLI(context.Background(), home, SpawnOptions{
				Role: "manager", Project: "demo", Harness: "claude", Model: "m",
				Name: "owner-check", NameSet: true,
			})
			if mutationErr != nil {
				t.Fatalf("caller change during workspace.create = %v; want the store write to complete", mutationErr)
			}
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != test.want {
				t.Fatalf("envelope = %#v; want the fresh owner refusal %q", envelope, test.want)
			}
			if len(envelope.Effects) != 0 {
				t.Fatalf("effects = %#v; want none", envelope.Effects)
			}
			if starts := fake.recorded("agent.start"); len(starts) != 0 {
				t.Fatalf("agent.start requests = %d; want none", len(starts))
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", "owner-check-g1.json")); !os.IsNotExist(err) {
				t.Fatalf("owner record exists after the refusal: %v", err)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "owner-demo.json")); !os.IsNotExist(err) {
				t.Fatalf("owner claim exists after the refusal: %v", err)
			}
		})
	}
}
