package spawn

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func gitOut(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// caller adds the caller workspace with the record boss in its root pane.
func caller(t *testing.T, fake *agenttest.Herdr, home string, options *Options) agenttest.Workspace {
	t.Helper()
	ws, pane := fake.AddWorkspace("mine", "", false)
	if err := agent.Save(home, agent.Record{Name: "boss", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	options.CallerPane = pane.ID
	return *ws
}

func paneOf(t *testing.T, fake *agenttest.Herdr, home, name string) (agent.Record, agenttest.Pane) {
	t.Helper()
	record, err := agent.Load(home, name)
	if err != nil {
		t.Fatal(err)
	}
	pane, ok := fake.Pane(record.Pane)
	if !ok {
		t.Fatalf("pane %s of %s not found", record.Pane, name)
	}
	return record, pane
}

func TestSpawnBranchMakesTheCheckoutAndANewWorkspaceOnIt(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	options := base("kid")
	callerWS := caller(t, fake, home, &options)
	options.Cwd, options.Branch = repo, "feat/x"

	envelope := spawnRun(t, home, options)
	if envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, pane := paneOf(t, fake, home, "kid")
	path := record.Worktree
	if !strings.HasPrefix(path, filepath.Join(home, "worktrees")+string(filepath.Separator)) || record.Cwd != path {
		t.Fatalf("record = %+v; want cwd and worktree below %s/worktrees", record, home)
	}
	if got := gitOut(t, "-C", path, "branch", "--show-current"); got != "feat/x" {
		t.Fatalf("branch of %s = %s; want feat/x", path, got)
	}
	if got, want := gitOut(t, "-C", path, "rev-parse", "HEAD"), gitOut(t, "-C", repo, "rev-parse", "HEAD"); got != want {
		t.Fatalf("HEAD = %s; want the current commit %s", got, want)
	}
	creates := fake.Requests("workspace.create")
	if len(creates) != 1 || creates[0].Params["cwd"] != path {
		t.Fatalf("workspace.create = %+v; want one with cwd %s", creates, path)
	}
	if pane.WorkspaceID == callerWS.ID || pane.TabID != pane.WorkspaceID+":t1" {
		t.Fatalf("pane = %+v; want the root pane of a new workspace, not the caller workspace %s", pane, callerWS.ID)
	}
	if !hasEffect(envelope, output.Effect{Resource: "worktree", Path: path, Action: "created"}) {
		t.Fatalf("effects = %+v; want worktree created %s", envelope.Effects, path)
	}
}

// TestSpawnBranchSharesTheCheckoutAndItsWorkspace proves that a second agent
// on the branch gets a tab in the workspace of the first one, and that both
// agents stay.
func TestSpawnBranchSharesTheCheckoutAndItsWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	for _, name := range []string{"one", "two"} {
		options := base(name)
		options.Cwd, options.Branch = repo, "feat"
		if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
			t.Fatalf("spawn %s = %+v", name, envelope)
		}
	}
	oneRecord, one := paneOf(t, fake, home, "one")
	twoRecord, two := paneOf(t, fake, home, "two")
	if got := len(fake.Requests("workspace.create")); got != 1 {
		t.Fatalf("workspace.create calls = %d; want 1", got)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != one.WorkspaceID || tabs[0].Params["cwd"] != oneRecord.Worktree {
		t.Fatalf("tab.create = %+v; want one tab in %s with cwd %s", tabs, one.WorkspaceID, oneRecord.Worktree)
	}
	if two.WorkspaceID != one.WorkspaceID || one.ID == two.ID || twoRecord.Worktree != oneRecord.Worktree {
		t.Fatalf("panes %+v and %+v; want two panes of one workspace on one checkout", one, two)
	}
	if one.Kind == "" || two.Kind == "" {
		t.Fatalf("agents = %q, %q; want both running", one.Kind, two.Kind)
	}
}

func TestSpawnBranchUsesTheOpenWorkspaceOfAnExistingCheckout(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	fake.AddWorkspace("other", agenttest.GitRepo(t), false)
	userWS, _ := fake.AddWorkspace("my label", repo, false)
	options := base("kid")
	caller(t, fake, home, &options)
	options.Cwd, options.Branch = repo, "main"

	envelope := spawnRun(t, home, options)
	if envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.create")); got != 0 {
		t.Fatalf("workspace.create calls = %d; want 0", got)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != userWS.ID || tabs[0].Params["cwd"] != repo {
		t.Fatalf("tab.create = %+v; want a tab in %s with cwd %s", tabs, userWS.ID, repo)
	}
	if record, _ := paneOf(t, fake, home, "kid"); record.Worktree != repo || record.Cwd != repo {
		t.Fatalf("record = %+v; want worktree and cwd %s", record, repo)
	}
	for _, ws := range fake.Workspaces() {
		if ws.ID == userWS.ID && ws.Label != "my label" {
			t.Fatalf("label = %q; want it unchanged", ws.Label)
		}
	}
	for _, effect := range envelope.Effects {
		if effect.Resource == "worktree" {
			t.Fatalf("effects = %+v; want no worktree effect", envelope.Effects)
		}
	}
}

func TestSpawnBranchFailureStartsNothing(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	options := base("kid")
	options.Cwd, options.Branch, options.Base = repo, "main", "main"
	envelope := spawnRun(t, home, options)
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "worktree" || !strings.Contains(errorText(envelope), "--base") {
		t.Fatalf("envelope = %+v; want a failed worktree phase about --base", envelope)
	}
	for _, method := range []string{"workspace.create", "tab.create", "agent.start"} {
		if got := len(fake.Requests(method)); got != 0 {
			t.Fatalf("%s calls = %d; want 0", method, got)
		}
	}
}

func TestSpawnBranchStartsAtTheBase(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	first := gitOut(t, "-C", repo, "rev-parse", "HEAD")
	gitOut(t, "-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "two")
	options := base("kid")
	options.Cwd, options.Branch, options.Base = repo, "feat", first
	if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, err := agent.Load(home, "kid")
	if err != nil {
		t.Fatal(err)
	}
	if got := gitOut(t, "-C", record.Worktree, "rev-parse", "HEAD"); got != first {
		t.Fatalf("HEAD = %s; want the base %s", got, first)
	}
}

// TestSpawnDirectoryJoinsTheWorkspaceOfALiveAgent proves that a directory
// outside Git gets one workspace, which a later agent of that directory
// shares, and that a gone agent does not hold it.
func TestSpawnDirectoryJoinsTheWorkspaceOfALiveAgent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	dir := t.TempDir()
	for _, name := range []string{"one", "two"} {
		options := base(name)
		caller(t, fake, home, &options)
		options.Directory = dir
		if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
			t.Fatalf("spawn %s = %+v", name, envelope)
		}
	}
	oneRecord, one := paneOf(t, fake, home, "one")
	_, two := paneOf(t, fake, home, "two")
	if one.WorkspaceID != two.WorkspaceID || oneRecord.Cwd != dir || oneRecord.Worktree != "" {
		t.Fatalf("panes %+v, %+v, record %+v; want one workspace on %s and no worktree", one, two, oneRecord, dir)
	}
	creates := fake.Requests("workspace.create")
	if len(creates) != 1 || creates[0].Params["cwd"] != dir {
		t.Fatalf("workspace.create = %+v; want one with cwd %s", creates, dir)
	}
	fake.ClosePane(one.ID)
	fake.ClosePane(two.ID)
	options := base("three")
	options.Directory = dir
	if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("spawn three = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.create")); got != 2 {
		t.Fatalf("workspace.create calls = %d; want a new workspace once the agents are gone", got)
	}
}

func TestSpawnDirectoryUsesTheOpenWorkspaceOfACheckout(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	userWS, _ := fake.AddWorkspace("mine", repo, false)
	options := base("kid")
	options.Directory = repo
	if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != userWS.ID || tabs[0].Params["cwd"] != repo {
		t.Fatalf("tab.create = %+v; want a tab in %s", tabs, userWS.ID)
	}
}

func TestSpawnDirectoryResolvesARelativePathFromTheCaller(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	parent := t.TempDir()
	dir := filepath.Join(parent, "sub")
	gitOut(t, "init", "-q", dir)
	options := base("kid")
	options.Cwd, options.Directory = parent, "sub"
	if envelope := spawnRun(t, home, options); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	if creates := fake.Requests("workspace.create"); len(creates) != 1 || creates[0].Params["cwd"] != dir {
		t.Fatalf("workspace.create = %+v; want cwd %s", creates, dir)
	}
	if !slices.Contains(fake.Methods(), "agent.start") {
		t.Fatal("no agent.start")
	}
}

// TestSpawnBranchReportsTheCheckoutThatAFailedHookLeft proves that a spawn
// whose git worktree add fails after it made the checkout fails in phase
// worktree, starts no agent, and reports the branch and the checkout.
func TestSpawnBranchReportsTheCheckoutThatAFailedHookLeft(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".git", "hooks", "post-checkout"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	options := base("kid")
	caller(t, fake, home, &options)
	options.Cwd, options.Branch = repo, "feat"

	envelope := spawnRun(t, home, options)
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "worktree" {
		t.Fatalf("envelope = %+v; want failed in phase worktree", envelope)
	}
	path := gitOut(t, "-C", repo, "worktree", "list", "--porcelain")
	if !strings.Contains(path, "branch refs/heads/feat") {
		t.Fatalf("worktree list = %s; want the checkout of feat", path)
	}
	if !hasEffect(envelope, output.Effect{Resource: "branch", Path: "feat", Action: "created"}) || len(envelope.Effects) != 2 || envelope.Effects[1].Resource != "worktree" || envelope.Effects[1].Action != "created" {
		t.Fatalf("effects = %+v; want branch created feat and worktree created", envelope.Effects)
	}
	if got := len(fake.Requests("agent.start")); got != 0 {
		t.Fatalf("agent.start calls = %d; want 0", got)
	}
}
