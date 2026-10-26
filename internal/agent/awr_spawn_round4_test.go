package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func removeCoordinatorClaim(locked *store.LockedStore) error {
	return locked.RemoveIn([]string{"claims"}, "coordinator")
}

func loseCoordinatorIdentity(locked *store.LockedStore) error {
	return locked.UpdateIn([]string{"agents"}, "coordinator-g1", func(raw json.RawMessage) (json.RawMessage, error) {
		var record records.Agent
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, err
		}
		record.Kind = ""
		record.Profile = records.ProfileBuilder
		record.Owns = ""
		record.Project = "demo"
		record.Task = ""
		record.Worktree = ""
		return json.Marshal(record)
	})
}

func removeCoordinatorCaller(locked *store.LockedStore) error {
	if err := removeCoordinatorClaim(locked); err != nil {
		return err
	}
	return locked.RemoveIn([]string{"agents"}, "coordinator-g1")
}

// TestSpawnOwnerRechecksExplicitCallerBeforeRegistration proves ruling 904 for
// an explicit --agent caller. The final registration lock must reread identity
// and claim facts even when HERDR_PANE_ID is empty. A caller that loses its
// claim or identity refuses before any owner record, claim, start, prompt, or
// registration effect. An invalid explicit caller must not fall back to the
// user path.
func TestSpawnOwnerRechecksExplicitCallerBeforeRegistration(t *testing.T) {
	cases := []struct {
		name   string
		pane   string
		mutate func(locked *store.LockedStore) error
		want   string
	}{
		{name: "no pane claim removed", mutate: removeCoordinatorClaim, want: "coordinator-g1"},
		{name: "no pane identity lost", mutate: loseCoordinatorIdentity, want: "coordinator"},
		{name: "no pane caller gone", mutate: removeCoordinatorCaller, want: "coordinator-g1"},
		{name: "pane claim removed", pane: "wHome:pC", mutate: removeCoordinatorClaim, want: "coordinator-g1"},
		{name: "pane identity lost", pane: "wHome:pC", mutate: loseCoordinatorIdentity, want: "coordinator"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			dir := projectDirectory(t, home, "demo")
			writeManagerProject(t, home, nil)
			writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
			writeRecord(t, home, records.Agent{
				Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator,
				PaneID: "wHome:pC", WorkspaceID: "wHome", Status: records.AgentLive, Generation: 1,
			})
			t.Setenv("HERDR_PANE_ID", test.pane)
			t.Chdir(t.TempDir())
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
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
				Profile: records.ProfilePlanner, Owner: true, Project: "demo", Harness: "claude", Model: "m",
				Name: "owner-check", NameSet: true, Agent: "coordinator-g1",
			})
			if mutationErr != nil {
				t.Fatalf("caller change during workspace.create = %v; want the store write to complete", mutationErr)
			}
			want := "--owner needs the coordinator or the user: " + test.want
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != want {
				t.Fatalf("envelope = %#v; want the fresh owner refusal %q", envelope, want)
			}
			if len(envelope.Effects) != 0 {
				t.Fatalf("effects = %#v; want none", envelope.Effects)
			}
			if starts := fake.recorded("agent.start"); len(starts) != 0 {
				t.Fatalf("agent.start requests = %d; want none", len(starts))
			}
			if prompts := fake.recorded("agent.prompt"); len(prompts) != 0 {
				t.Fatalf("agent.prompt requests = %d; want none", len(prompts))
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", "owner-check-g1.json")); !os.IsNotExist(err) {
				t.Fatalf("owner record exists after the refusal: %v", err)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "owner-demo.json")); !os.IsNotExist(err) {
				t.Fatalf("owner claim exists after the refusal: %v", err)
			}
			closes := fake.recorded("workspace.close")
			if len(closes) != 1 || !strings.Contains(closes[0].Params, `"workspace_id":"wBD"`) {
				t.Fatalf("workspace.close = %#v; want one cleanup close of wBD", closes)
			}
			project := readProjectWorkspaceRecord(t, home, "demo")
			if project.Workspace != nil {
				t.Fatalf("project workspace = %#v; want no persisted workspace after cleanup", project.Workspace)
			}
		})
	}
}

// TestSpawnOwnerAcceptsFreshExplicitCallerWithoutPane proves the fresh check
// does not refuse a valid explicit coordinator that still holds its claim.
func TestSpawnOwnerAcceptsFreshExplicitCallerWithoutPane(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	writeClaimRecord(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	writeRecord(t, home, records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator,
		PaneID: "wHome:pC", WorkspaceID: "wHome", Status: records.AgentLive, Generation: 1,
	})
	t.Setenv("HERDR_PANE_ID", "")
	t.Chdir(t.TempDir())
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := SpawnFromCLI(context.Background(), home, SpawnOptions{
		Profile: records.ProfilePlanner, Owner: true, Project: "demo", Harness: "claude", Model: "m",
		Name: "owner-check", NameSet: true, Agent: "coordinator-g1",
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if claim := readAgent(t, home, "owner-check-g1"); claim.Status != records.AgentLive {
		t.Fatalf("owner record status = %q; want live", claim.Status)
	}
	ownerClaim := readClaim(t, home, "owner-demo")
	if ownerClaim.Agent != "owner-check-g1" {
		t.Fatalf("owner claim = %#v; want owner-check-g1", ownerClaim)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start requests = %d; want one", len(starts))
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want one", len(prompts))
	}
}

// TestSpawnOwnerKeepsGenuineUserWithoutPane proves the user path still works
// after the explicit-caller check reads fresh facts.
func TestSpawnOwnerKeepsGenuineUserWithoutPane(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	t.Setenv("HERDR_PANE_ID", "")
	t.Chdir(t.TempDir())
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := SpawnFromCLI(context.Background(), home, SpawnOptions{
		Profile: records.ProfilePlanner, Project: "demo", Harness: "claude", Model: "m",
		Name: "owner-user", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start requests = %d; want one", len(starts))
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want one", len(prompts))
	}
}
