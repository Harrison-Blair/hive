package spawn

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func spawnRun(t *testing.T, home string, options Options) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return Run(ctx, home, options)
}

func base(name string) Options {
	return Options{Name: name, Harness: "pi", Model: "prov/m1", Prompt: "do the work"}
}

func errorText(envelope output.Envelope) string {
	if envelope.Error == nil {
		return ""
	}
	return envelope.Error.(string)
}

func TestSpawnRequiresNameHarnessAndModel(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	for _, tc := range []struct {
		name    string
		options Options
		want    string
	}{
		{"no name", Options{Harness: "pi", Model: "m"}, "--name"},
		{"no harness", Options{Name: "a", Model: "m"}, "--harness"},
		{"no model", Options{Name: "a", Harness: "pi"}, "--model"},
		{"bad harness", Options{Name: "a", Harness: "vim", Model: "m"}, "unknown harness"},
		{"both prompts", Options{Name: "a", Harness: "pi", Model: "m", Prompt: "x", PromptFile: "f"}, "--prompt"},
		{"worktree without project", Options{Name: "a", Harness: "pi", Model: "m", Worktree: "x"}, "--worktree needs --project"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			envelope := spawnRun(t, home, tc.options)
			if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), tc.want) {
				t.Fatalf("envelope = %+v, want failed with %q", envelope, tc.want)
			}
		})
	}
	if got := fake.Requests("agent.start"); len(got) != 0 {
		t.Fatalf("agent.start calls = %d, want 0", len(got))
	}
}

func TestSpawnWithoutProjectOpensTabInCallerWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	callerWS, callerPane := fake.AddWorkspace("mine", "", false)
	if err := agent.Save(home, agent.Record{Name: "boss", Pane: callerPane.ID}); err != nil {
		t.Fatal(err)
	}
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Cwd = "/work/dir"

	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != callerWS.ID || tabs[0].Params["cwd"] != "/work/dir" {
		t.Fatalf("tab.create = %+v, want one tab in %s with cwd /work/dir", tabs, callerWS.ID)
	}
	record, err := agent.Load(home, "kid")
	if err != nil {
		t.Fatal(err)
	}
	pane, ok := fake.Pane(record.Pane)
	if !ok || pane.WorkspaceID != callerWS.ID || pane.ID == callerPane.ID {
		t.Fatalf("record pane %q is not a new pane in %s", record.Pane, callerWS.ID)
	}
	if record.Parent != "boss" || record.Harness != "pi" || record.Model != "prov/m1" || record.Cwd != "/work/dir" || record.Worktree != "" || record.StartedAt.IsZero() {
		t.Fatalf("record = %+v", record)
	}
	if pane.Kind != "pi" || pane.Name != "kid" || !slices.Contains(pane.Args, "prov/m1") {
		t.Fatalf("started agent = %+v", pane)
	}
	if !slices.Equal(pane.Prompts, []string{"do the work"}) {
		t.Fatalf("prompts = %q, want the first prompt once", pane.Prompts)
	}
}

func TestSpawnReadsPromptFile(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	file := filepath.Join(t.TempDir(), "p.txt")
	if err := os.WriteFile(file, []byte("from the file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	options := Options{Name: "kid", Harness: "pi", Model: "m", PromptFile: file, CallerPane: callerPane.ID}
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, _ := agent.Load(home, "kid")
	pane, _ := fake.Pane(record.Pane)
	if !slices.Equal(pane.Prompts, []string{"from the file\n"}) {
		t.Fatalf("prompts = %q", pane.Prompts)
	}
}

func TestSpawnWithoutCallerRecordHasNoParent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	options := base("kid")
	options.CallerPane = callerPane.ID
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, _ := agent.Load(home, "kid")
	if record.Parent != "" {
		t.Fatalf("parent = %q, want empty", record.Parent)
	}
}

func TestSpawnWithoutProjectNeedsCallerPane(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	envelope := spawnRun(t, home, base("kid"))
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "HERDR_PANE_ID") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestSpawnProjectCreatesOneCheckoutWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")

	first := base("one")
	first.Project = "p1"
	if envelope := spawnRun(t, home, first); envelope.Outcome != "ok" {
		t.Fatalf("first = %+v", envelope)
	}
	creates := fake.Requests("workspace.create")
	if len(creates) != 1 || creates[0].Params["cwd"] != repo || creates[0].Params["label"] != nil {
		t.Fatalf("workspace.create = %+v, want one with cwd %s and no label", creates, repo)
	}
	workspaces := fake.Workspaces()
	if len(workspaces) != 1 {
		t.Fatalf("workspaces = %+v", workspaces)
	}
	one, _ := agent.Load(home, "one")
	if pane, _ := fake.Pane(one.Pane); pane.WorkspaceID != workspaces[0].ID || pane.TabID != workspaces[0].ID+":t1" {
		t.Fatalf("first agent pane = %+v, want the root pane of %s", pane, workspaces[0].ID)
	}
	if one.Cwd != repo {
		t.Fatalf("cwd = %q, want %q", one.Cwd, repo)
	}

	second := base("two")
	second.Project = "p1"
	if envelope := spawnRun(t, home, second); envelope.Outcome != "ok" {
		t.Fatalf("second = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.create")); got != 1 {
		t.Fatalf("workspace.create calls = %d, want 1", got)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != workspaces[0].ID || tabs[0].Params["cwd"] != repo {
		t.Fatalf("tab.create = %+v", tabs)
	}
	if got := len(fake.Workspaces()); got != 1 {
		t.Fatalf("workspaces = %d, want 1", got)
	}
}

func TestSpawnProjectUsesExistingCheckoutWorkspaceWithAnyLabel(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")
	fake.AddWorkspace("other", agenttest.GitRepo(t), false)
	fake.AddWorkspace("linked", filepath.Join(home, "elsewhere"), true)
	userWS, _ := fake.AddWorkspace("my label", repo, false)

	options := base("one")
	options.Project = "p1"
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.create")); got != 0 {
		t.Fatalf("workspace.create calls = %d, want 0", got)
	}
	tabs := fake.Requests("tab.create")
	if len(tabs) != 1 || tabs[0].Params["workspace_id"] != userWS.ID {
		t.Fatalf("tab.create = %+v, want a tab in %s", tabs, userWS.ID)
	}
	for _, ws := range fake.Workspaces() {
		if ws.ID == userWS.ID && ws.Label != "my label" {
			t.Fatalf("label = %q, want it unchanged", ws.Label)
		}
	}
}

func TestSpawnLinkedWorkspaceOnCheckoutIsNotTheCheckoutWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")
	fake.AddWorkspace("odd", repo, true)

	options := base("one")
	options.Project = "p1"
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.create")); got != 1 {
		t.Fatalf("workspace.create calls = %d, want 1", got)
	}
}

func TestSpawnUnknownProject(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	options := base("one")
	options.Project = "nope"
	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "nope") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestSpawnInvalidProjectIDUsesTheProjectRule(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	options := base("one")
	options.Project = "Repo_One"
	envelope := spawnRun(t, home, options)
	const want = "invalid project id: Repo_One: an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters"
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "project" || errorText(envelope) != want {
		t.Fatalf("envelope = %+v; want failed project phase with %q", envelope, want)
	}
}

func TestSpawnWorktreeNestsUnderCheckoutWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")
	checkoutWS, _ := fake.AddWorkspace("any", repo, false)

	options := base("wt-agent")
	options.Project = "p1"
	options.Worktree = "feat-x"
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	path := filepath.Join(home, "projects", "p1", "worktrees", "feat-x")
	creates := fake.Requests("worktree.create")
	if len(creates) != 1 {
		t.Fatalf("worktree.create = %+v", creates)
	}
	params := creates[0].Params
	if params["workspace_id"] != checkoutWS.ID || params["path"] != path || params["branch"] != "feat-x" || params["base"] != "main" {
		t.Fatalf("worktree.create params = %+v", params)
	}
	if _, has := params["cwd"]; has {
		t.Fatalf("worktree.create params = %+v, want no cwd", params)
	}
	if info, err := os.Stat(filepath.Join(path, ".git")); err != nil || info.IsDir() {
		t.Fatalf("worktree checkout missing at %s: %v", path, err)
	}
	record, _ := agent.Load(home, "wt-agent")
	pane, _ := fake.Pane(record.Pane)
	var linked agenttest.Workspace
	for _, ws := range fake.Workspaces() {
		if ws.Linked {
			linked = ws
		}
	}
	if pane.WorkspaceID != linked.ID || pane.TabID != linked.ID+":t1" {
		t.Fatalf("agent pane = %+v, want the root pane of %s", pane, linked.ID)
	}
	if record.Worktree != path || record.Cwd != path {
		t.Fatalf("record = %+v, want worktree and cwd %s", record, path)
	}
	if got := len(fake.Requests("workspace.create")); got != 0 {
		t.Fatalf("workspace.create calls = %d, want 0", got)
	}
}

func TestSpawnWorktreeCreatesCheckoutWorkspaceWhenNoneIsOpen(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")

	options := base("wt-agent")
	options.Project = "p1"
	options.Worktree = "feat-x"
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	creates := fake.Requests("workspace.create")
	if len(creates) != 1 || creates[0].Params["cwd"] != repo {
		t.Fatalf("workspace.create = %+v", creates)
	}
	var checkout string
	for _, ws := range fake.Workspaces() {
		if !ws.Linked {
			checkout = ws.ID
		}
	}
	if got := fake.Requests("worktree.create"); len(got) != 1 || got[0].Params["workspace_id"] != checkout {
		t.Fatalf("worktree.create = %+v, want source %s", got, checkout)
	}
}

func TestSpawnWorktreeFailureReportsTheCreatedCheckoutWorkspace(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	agenttest.RegisterProject(t, home, "p1", repo, "main")
	fake.Fail("worktree.create", "worktree_create_failed", "create boom")

	options := base("wt-agent")
	options.Project = "p1"
	options.Worktree = "feat-x"
	envelope := spawnRun(t, home, options)
	workspaces := fake.Workspaces()
	if len(workspaces) != 1 {
		t.Fatalf("workspaces = %+v; want the one checkout workspace", workspaces)
	}
	want := []output.Effect{{Resource: "workspace", Path: workspaces[0].ID, Action: "created"}}
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "worktree" || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("envelope = %+v; want failed worktree phase with effects %+v", envelope, want)
	}
}

func TestSpawnBlockedAgentReportsPaneAndSendsNoPrompt(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	fake.StartStatus = "blocked"
	options := base("kid")
	options.CallerPane = callerPane.ID

	envelope := spawnRun(t, home, options)
	if envelope.Outcome == "ok" || output.Status(envelope.Outcome) == 0 {
		t.Fatalf("envelope = %+v, want a non-zero outcome", envelope)
	}
	record, err := agent.Load(home, "kid")
	if err != nil {
		t.Fatalf("record of a blocked agent: %v", err)
	}
	if !strings.Contains(errorText(envelope), record.Pane) || !strings.Contains(errorText(envelope), "blocked") {
		t.Fatalf("error = %q, want the pane %s and the word blocked", errorText(envelope), record.Pane)
	}
	if got := fake.Requests("agent.prompt"); len(got) != 0 {
		t.Fatalf("agent.prompt calls = %d, want 0", len(got))
	}
}

func TestSpawnNameMustBeUniqueAmongLiveRecords(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	live := fake.AddPane("w1", "")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: live.ID}); err != nil {
		t.Fatal(err)
	}
	options := base("kid")
	options.CallerPane = callerPane.ID
	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "kid") {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := fake.Requests("tab.create"); len(got) != 0 {
		t.Fatalf("tab.create calls = %d, want 0", len(got))
	}

	fake.ClosePane(live.ID)
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("spawn over a gone record = %+v", envelope)
	}
	record, _ := agent.Load(home, "kid")
	if record.Pane == live.ID {
		t.Fatalf("record not replaced: %+v", record)
	}
}

func TestSpawnStartFailureClosesNewPaneAndWritesNoRecord(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	fake.Fail("agent.start", "unsupported_agent_kind", "nope")
	options := base("kid")
	options.CallerPane = callerPane.ID
	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "unsupported_agent_kind") {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, err := agent.Load(home, "kid"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
	if got := fake.Requests("pane.close"); len(got) != 1 {
		t.Fatalf("pane.close calls = %d, want 1", len(got))
	}
}

func TestSpawnTriesAgainWhilePaneIsBusy(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	fake.Fail("agent.start", "agent_pane_busy", "agent target pane is not an available shell")
	options := base("kid")
	options.CallerPane = callerPane.ID
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("agent.start")); got != 2 {
		t.Fatalf("agent.start calls = %d, want 2", got)
	}
}

func TestSpawnLogsHarnessSessionID(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	fake.StartSession = "sess-42"
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := log.Event("agent spawn")["harness_session_id"]; got != "sess-42" {
		t.Fatalf("harness_session_id = %v, want sess-42", got)
	}
}

func TestSpawnLogsHarnessSessionIDFromStartWhenSpawnEndsAfterStart(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	fake.LaunchSession = "start-6"
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	agents := filepath.Join(home, "agents")
	if err := os.MkdirAll(agents, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(agents, 0o755) })
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "partial" || envelope.Phase != "record" {
		t.Fatalf("envelope = %+v, want partial in phase record", envelope)
	}
	if got := fake.Requests("agent.get"); len(got) != 0 {
		t.Fatalf("agent.get calls = %d, want 0", len(got))
	}
	if got := log.Event("agent spawn")["harness_session_id"]; got != "start-6" {
		t.Fatalf("harness_session_id = %v, want start-6", got)
	}
}

func TestSpawnLogsHarnessSessionIDFromAgentGetBeforeReady(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	fake.PendingSession = "early-7"
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := log.Event("agent spawn")["harness_session_id"]; got != "early-7" {
		t.Fatalf("harness_session_id = %v, want early-7", got)
	}
}

func TestSpawnLogsHarnessSessionIDFromPromptReply(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	fake.PromptSession = "prompt-9"
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := log.Event("agent spawn")["harness_session_id"]; got != "prompt-9" {
		t.Fatalf("harness_session_id = %v, want prompt-9", got)
	}
}

func TestSpawnLogsHarnessSessionIDWhenThePromptFails(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	fake.LaunchSession = "early-8"
	fake.Fail("agent.prompt", "internal_error", "boom")
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "partial" || envelope.Phase != "prompt" {
		t.Fatalf("envelope = %+v, want partial in phase prompt", envelope)
	}
	if got := log.Event("agent spawn")["harness_session_id"]; got != "early-8" {
		t.Fatalf("harness_session_id = %v, want early-8", got)
	}
}

func TestSpawnLogsNoHarnessSessionIDWhenHerdrReportsNone(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.CallerPane = callerPane.ID
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got, ok := log.Event("agent spawn")["harness_session_id"]; ok {
		t.Fatalf("harness_session_id = %v, want absent", got)
	}
}

// tabLabel returns the tab label of the pane of agent name.
func tabLabel(t *testing.T, fake *agenttest.Herdr, home, name string) string {
	t.Helper()
	record, err := agent.Load(home, name)
	if err != nil {
		t.Fatal(err)
	}
	pane, ok := fake.Pane(record.Pane)
	if !ok {
		t.Fatalf("pane %s of agent %s not found", record.Pane, name)
	}
	return pane.TabLabel
}

func hasEffect(envelope output.Envelope, want output.Effect) bool {
	return slices.Contains(envelope.Effects, want)
}

func TestSpawnLabelsTheAgentTabWithTheAgentName(t *testing.T) {
	for _, tc := range []struct {
		name        string
		setup       func(t *testing.T, fake *agenttest.Herdr, home string, options *Options)
		wantRenamed bool
	}{
		{"caller tab", func(t *testing.T, fake *agenttest.Herdr, home string, options *Options) {
			_, callerPane := fake.AddWorkspace("mine", "", false)
			options.CallerPane = callerPane.ID
		}, false},
		{"project tab", func(t *testing.T, fake *agenttest.Herdr, home string, options *Options) {
			repo := agenttest.GitRepo(t)
			agenttest.RegisterProject(t, home, "p1", repo, "main")
			fake.AddWorkspace("any", repo, false)
			options.Project = "p1"
		}, false},
		{"project root pane", func(t *testing.T, fake *agenttest.Herdr, home string, options *Options) {
			agenttest.RegisterProject(t, home, "p1", agenttest.GitRepo(t), "main")
			options.Project = "p1"
		}, true},
		{"worktree root pane", func(t *testing.T, fake *agenttest.Herdr, home string, options *Options) {
			repo := agenttest.GitRepo(t)
			agenttest.RegisterProject(t, home, "p1", repo, "main")
			fake.AddWorkspace("any", repo, false)
			options.Project = "p1"
			options.Worktree = "feat-x"
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			options := base("kid")
			tc.setup(t, fake, home, &options)
			envelope := spawnRun(t, home, options)
			if envelope.Outcome != "ok" {
				t.Fatalf("envelope = %+v", envelope)
			}
			if got := tabLabel(t, fake, home, "kid"); got != "kid" {
				t.Fatalf("tab label = %q, want kid", got)
			}
			record, _ := agent.Load(home, "kid")
			pane, _ := fake.Pane(record.Pane)
			renamed := output.Effect{Resource: "tab", Path: pane.TabID, Action: "renamed"}
			if hasEffect(envelope, renamed) != tc.wantRenamed {
				t.Fatalf("effects = %+v, want renamed effect %v", envelope.Effects, tc.wantRenamed)
			}
			if tc.wantRenamed {
				methods := fake.Methods()
				rename, start := slices.Index(methods, "tab.rename"), slices.Index(methods, "agent.start")
				if rename < 0 || start < 0 || rename > start {
					t.Fatalf("requests = %q, want tab.rename before agent.start", methods)
				}
			}
			for _, ws := range fake.Workspaces() {
				if ws.Label == "kid" {
					t.Fatalf("workspace %s has label kid, want workspace labels unchanged", ws.ID)
				}
			}
		})
	}
}

func TestSpawnTabRenameFailureIsPartialAndStillPrompts(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	agenttest.RegisterProject(t, home, "p1", agenttest.GitRepo(t), "main")
	fake.Fail("tab.rename", "internal_error", "rename boom")
	options := base("kid")
	options.Project = "p1"

	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "partial" || envelope.Phase != "label" || !strings.Contains(errorText(envelope), "rename boom") {
		t.Fatalf("envelope = %+v, want partial in phase label with the rename error", envelope)
	}
	record, err := agent.Load(home, "kid")
	if err != nil {
		t.Fatal(err)
	}
	pane, _ := fake.Pane(record.Pane)
	if !slices.Equal(pane.Prompts, []string{"do the work"}) {
		t.Fatalf("prompts = %q, want the prompt sent", pane.Prompts)
	}
	for _, effect := range envelope.Effects {
		if effect.Action == "renamed" {
			t.Fatalf("effects = %+v, want no renamed effect", envelope.Effects)
		}
	}
	if envelope.Result != record {
		t.Fatalf("result = %+v, want the record", envelope.Result)
	}
}

func TestSpawnTabRenameFailureJoinsALaterFailure(t *testing.T) {
	for _, tc := range []struct {
		name      string
		setup     func(fake *agenttest.Herdr)
		wantPhase string
		wantLater string
	}{
		{"start", func(fake *agenttest.Herdr) { fake.Fail("agent.start", "unsupported_agent_kind", "start boom") }, "start", "start boom"},
		{"blocked", func(fake *agenttest.Herdr) { fake.StartStatus = "blocked" }, "wait", "blocked"},
		{"prompt", func(fake *agenttest.Herdr) { fake.Fail("agent.prompt", "internal_error", "prompt boom") }, "prompt", "prompt boom"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			agenttest.RegisterProject(t, home, "p1", agenttest.GitRepo(t), "main")
			fake.Fail("tab.rename", "internal_error", "rename boom")
			tc.setup(fake)
			options := base("kid")
			options.Project = "p1"

			envelope := spawnRun(t, home, options)
			message := errorText(envelope)
			if envelope.Outcome == "ok" || envelope.Phase != tc.wantPhase {
				t.Fatalf("envelope = %+v, want phase %s", envelope, tc.wantPhase)
			}
			later, rename := strings.Index(message, tc.wantLater), strings.Index(message, "(also: tab label: ")
			if later < 0 || rename < later || !strings.Contains(message[rename:], "rename boom") {
				t.Fatalf("error = %q, want %q and then the rename error", message, tc.wantLater)
			}
		})
	}
}

func TestSpawnRecordFailureKeepsTheTabRenameError(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	agenttest.RegisterProject(t, home, "p1", agenttest.GitRepo(t), "main")
	fake.Fail("tab.rename", "internal_error", "rename boom")
	// A read-only agents directory lets the name check pass and the record
	// write fail.
	agents := filepath.Join(home, "agents")
	if err := os.Mkdir(agents, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(agents, 0o755) })
	options := base("kid")
	options.Project = "p1"

	envelope := spawnRun(t, home, options)
	message := errorText(envelope)
	if envelope.Outcome != "partial" || envelope.Phase != "record" {
		t.Fatalf("envelope = %+v, want partial in phase record", envelope)
	}
	if !strings.HasPrefix(message, "write record: ") || !strings.Contains(message, "(also: tab label: ") || !strings.Contains(message, "rename boom") {
		t.Fatalf("error = %q, want the record error and then the rename error", message)
	}
}

func TestSpawnStartsTheHarnessWithTheGuide(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, callerPane := fake.AddWorkspace("mine", "", false)
	log := agenttest.NewLog(t)
	options := base("kid")
	options.Harness = "claude"
	options.Model = "m"
	options.CallerPane = callerPane.ID
	options.Guide = "line one\nline two"
	options.Logger = log.Logger

	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, _ := agent.Load(home, "kid")
	pane, _ := fake.Pane(record.Pane)
	want := []string{"--dangerously-skip-permissions", "--model", "m", "--append-system-prompt", `line one\nline two`}
	if !slices.Equal(pane.Args, want) {
		t.Fatalf("agent.start args = %q; want %q", pane.Args, want)
	}
	if !slices.Equal(pane.Prompts, []string{"do the work"}) {
		t.Fatalf("prompts = %q; want the brief unchanged", pane.Prompts)
	}
	argv, _ := log.Event("agent started")["argv"].([]any)
	if len(argv) == 0 || argv[len(argv)-1] != "<guide 18 bytes>" {
		t.Fatalf("agent started argv = %q; want the guide placeholder last", argv)
	}
	for _, arg := range argv {
		if strings.Contains(arg.(string), "line one") {
			t.Fatalf("agent started argv = %q; want no guide text", argv)
		}
	}
	if text := log.Text(); strings.Contains(text, "line one") || strings.Count(text, "<guide 18 bytes>") < 3 {
		t.Fatalf("log = %s; want the placeholder in the herdr request, the herdr reply, and agent started, and no guide text", text)
	}
}

// TestSpawnLimitsTheStartLine uses pi with no guide. Its start line is
// pi '--model' 'MODEL' '--approve', which is 27 bytes plus the model.
func TestSpawnLimitsTheStartLine(t *testing.T) {
	t.Run("over the limit", func(t *testing.T) {
		fake := agenttest.NewHerdr(t)
		home := agenttest.Home(t)
		_, callerPane := fake.AddWorkspace("mine", "", false)
		options := base("kid")
		options.Model = strings.Repeat("m", 3974)
		options.CallerPane = callerPane.ID
		envelope := spawnRun(t, home, options)
		want := "the start command is 4001 bytes, more than the limit of 4000 bytes: use a shorter --model"
		if envelope.Outcome != "failed" || envelope.Phase != "validate" || errorText(envelope) != want {
			t.Fatalf("envelope = %+v; want failed in validate with %q", envelope, want)
		}
		if got := fake.RequestCount(); got != 0 {
			t.Fatalf("herdr requests = %d; want 0", got)
		}
		if _, err := agent.Load(home, "kid"); err == nil {
			t.Fatalf("record kid exists; want none")
		}
	})
	t.Run("at the limit", func(t *testing.T) {
		fake := agenttest.NewHerdr(t)
		home := agenttest.Home(t)
		_, callerPane := fake.AddWorkspace("mine", "", false)
		options := base("kid")
		options.Model = strings.Repeat("m", 3973)
		options.CallerPane = callerPane.ID
		if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
			t.Fatalf("envelope = %+v; want ok", envelope)
		}
	})
}
