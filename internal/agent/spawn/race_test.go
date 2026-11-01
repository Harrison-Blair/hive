package spawn

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
	"github.com/Harrison-Blair/hive/internal/worktree/remove"
)

// spawnAll runs each options at the same time and returns the envelopes in
// the same order.
func spawnAll(t *testing.T, home string, all ...Options) []output.Envelope {
	t.Helper()
	envelopes := make([]output.Envelope, len(all))
	var group sync.WaitGroup
	for i, options := range all {
		group.Add(1)
		go func() {
			defer group.Done()
			envelopes[i] = spawnRun(t, home, options)
		}()
	}
	group.Wait()
	return envelopes
}

// TestSpawnDirectoryFindsAWorkspaceByTheDirectoryOfItsPanes proves that
// --directory joins a workspace that a person opened on a directory outside
// Git, or on a subdirectory of a checkout, with no hive agent in it.
func TestSpawnDirectoryFindsAWorkspaceByTheDirectoryOfItsPanes(t *testing.T) {
	repo := agenttest.GitRepo(t)
	sub := filepath.Join(repo, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{"outside git": t.TempDir(), "git subdirectory": sub} {
		t.Run(name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			userWS, _ := fake.AddWorkspace("user", dir, false)
			options := base("kid")
			options.Directory = dir
			if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
				t.Fatalf("envelope = %+v", envelope)
			}
			if creates := fake.Requests("workspace.create"); len(creates) != 0 {
				t.Fatalf("workspace.create = %+v; want none", creates)
			}
			if tabs := fake.Requests("tab.create"); len(tabs) != 1 || tabs[0].Params["workspace_id"] != userWS.ID {
				t.Fatalf("tab.create = %+v; want a tab in %s", tabs, userWS.ID)
			}
		})
	}
}

// TestConcurrentSpawnsInOneDirectoryShareOneWorkspace proves that spawns in
// one directory at the same time make one workspace.
func TestConcurrentSpawnsInOneDirectoryShareOneWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	dir := t.TempDir()
	fake.Before("workspace.create", func() { time.Sleep(200 * time.Millisecond) })
	one, two := base("one"), base("two")
	one.Directory, two.Directory = dir, dir
	for _, envelope := range spawnAll(t, home, one, two) {
		if envelope.Outcome != output.OutcomeOK {
			t.Fatalf("spawn = %+v", envelope)
		}
	}
	if creates := fake.Requests("workspace.create"); len(creates) != 1 {
		t.Fatalf("workspace.create calls = %d; want 1", len(creates))
	}
	_, onePane := paneOf(t, fake, home, "one")
	_, twoPane := paneOf(t, fake, home, "two")
	if onePane.WorkspaceID != twoPane.WorkspaceID {
		t.Fatalf("workspaces %s and %s; want one", onePane.WorkspaceID, twoPane.WorkspaceID)
	}
}

// TestConcurrentSpawnsOfOneNameStartOneAgent proves that two spawns with one
// name at the same time start one agent, and the other fails on the name.
func TestConcurrentSpawnsOfOneNameStartOneAgent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.Before("agent.start", func() { time.Sleep(200 * time.Millisecond) })
	one, two := base("kid"), base("kid")
	one.Directory, two.Directory = t.TempDir(), t.TempDir()
	ok, nameFailures := 0, 0
	for _, envelope := range spawnAll(t, home, one, two) {
		switch {
		case envelope.Outcome == output.OutcomeOK:
			ok++
		case envelope.Phase == "name":
			nameFailures++
		default:
			t.Fatalf("spawn = %+v; want ok or a name failure", envelope)
		}
	}
	if ok != 1 || nameFailures != 1 {
		t.Fatalf("ok = %d, name failures = %d; want 1 and 1", ok, nameFailures)
	}
	if starts := fake.Requests("agent.start"); len(starts) != 1 {
		t.Fatalf("agent.start calls = %d; want 1", len(starts))
	}
}

// TestRemoveDuringASpawnOnTheCheckoutKeepsTheCheckout proves that hive
// worktree remove that starts while a spawn on the checkout starts its agent
// waits for the spawn, then refuses, and keeps the checkout and its workspace.
func TestRemoveDuringASpawnOnTheCheckoutKeepsTheCheckout(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	checkout, err := worktree.Ensure(context.Background(), home, repo, "feat", "", logger)
	if err != nil {
		t.Fatal(err)
	}
	removed := make(chan output.Envelope, 1)
	var once sync.Once
	fake.Before("agent.start", func() {
		once.Do(func() {
			go func() {
				removed <- remove.Run(context.Background(), home, remove.Options{Dir: repo, Branch: "feat", Logger: logger})
			}()
			time.Sleep(300 * time.Millisecond)
		})
	})
	options := base("kid")
	options.Cwd, options.Branch = repo, "feat"
	if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("spawn = %+v", envelope)
	}
	envelope := <-removed
	if envelope.Outcome != output.OutcomeFailed || !strings.Contains(envelope.Error.(string), "kid") {
		t.Fatalf("remove = %+v; want a failure that names agent kid", envelope)
	}
	if _, err := os.Stat(checkout.Path); err != nil {
		t.Fatalf("checkout stat = %v; want it kept", err)
	}
	if closes := fake.Requests("workspace.close"); len(closes) != 0 {
		t.Fatalf("workspace.close = %+v; want none", closes)
	}
}
