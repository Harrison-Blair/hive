package stop

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func stopRun(t *testing.T, home string, options Options) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Run(ctx, home, options)
}

func errorText(envelope output.Envelope) string {
	if envelope.Error == nil {
		return ""
	}
	return envelope.Error.(string)
}

func git(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

// worktreeAgent makes a checkout workspace, a worktree "feat" with its own
// workspace, and the record "wt" in the root pane of that workspace.
func worktreeAgent(t *testing.T) (*agenttest.Herdr, string, string, string, agent.Record) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	fake.AddWorkspace("", repo, false)
	path := filepath.Join(home, "projects", "p1", "worktrees", "feat")
	git(t, "-C", repo, "worktree", "add", "-q", "-b", "feat", path, "main")
	ws, pane := fake.AddWorkspace("", path, true)
	fake.AddPane(ws.ID, path)
	record := agent.Record{Name: "wt", Pane: pane.ID, Cwd: path, Worktree: path}
	if err := agent.Save(home, record); err != nil {
		t.Fatal(err)
	}
	return fake, home, repo, ws.ID, record
}

func TestStopUnknownName(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	envelope := stopRun(t, home, Options{Name: "ghost"})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "unknown agent: ghost") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestStopClosesPaneAndRemovesRecord(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	if envelope := stopRun(t, home, Options{Name: "kid"}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, ok := fake.Pane(pane.ID); ok {
		t.Fatalf("pane %s still open", pane.ID)
	}
	if got := len(fake.Workspaces()); got != 1 {
		t.Fatalf("workspaces = %d, want the caller workspace kept", got)
	}
	if _, err := agent.Load(home, "kid"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
}

func TestStopLogsHarnessSessionID(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	fake.StartAgent(pane.ID, "kid", "codex", "idle")
	fake.SetSession(pane.ID, "019a-codex")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	log := agenttest.NewLog(t)
	if envelope := stopRun(t, home, Options{Name: "kid", Logger: log.Logger}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := log.Event("agent stop")["harness_session_id"]; got != "019a-codex" {
		t.Fatalf("harness_session_id = %v, want 019a-codex", got)
	}
}

func TestStopGonePaneRemovesRecord(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: "w7:p7"}); err != nil {
		t.Fatal(err)
	}
	if envelope := stopRun(t, home, Options{Name: "kid"}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("pane.close")); got != 0 {
		t.Fatalf("pane.close calls = %d", got)
	}
	if _, err := agent.Load(home, "kid"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
}

func TestStopWorktreeAgentClosesWorkspaceAndKeepsCheckout(t *testing.T) {
	fake, home, repo, wsID, record := worktreeAgent(t)
	if envelope := stopRun(t, home, Options{Name: "wt"}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	closes := fake.Requests("workspace.close")
	if len(closes) != 1 || closes[0].Params["workspace_id"] != wsID {
		t.Fatalf("workspace.close = %+v, want %s", closes, wsID)
	}
	for _, ws := range fake.Workspaces() {
		if ws.ID == wsID {
			t.Fatalf("worktree workspace still open")
		}
	}
	if len(fake.Workspaces()) != 1 {
		t.Fatalf("workspaces = %+v, want the checkout workspace kept", fake.Workspaces())
	}
	if _, err := os.Stat(record.Worktree); err != nil {
		t.Fatalf("checkout removed: %v", err)
	}
	if !strings.Contains(git(t, "-C", repo, "branch", "--list", "feat"), "feat") {
		t.Fatalf("branch feat removed")
	}
	if _, err := agent.Load(home, "wt"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
}

func TestStopRemoveWorktreeRefusesDirtyCheckout(t *testing.T) {
	fake, home, _, _, record := worktreeAgent(t)
	if err := os.WriteFile(filepath.Join(record.Worktree, "new.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	envelope := stopRun(t, home, Options{Name: "wt", RemoveWorktree: true})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "dirty") || !strings.Contains(errorText(envelope), record.Worktree) {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, ok := fake.Pane(record.Pane); !ok {
		t.Fatalf("pane closed")
	}
	if got := len(fake.Requests("workspace.close")) + len(fake.Requests("pane.close")); got != 0 {
		t.Fatalf("close calls = %d, want 0", got)
	}
	if _, err := os.Stat(filepath.Join(record.Worktree, "new.txt")); err != nil {
		t.Fatalf("checkout changed: %v", err)
	}
	if _, err := agent.Load(home, "wt"); err != nil {
		t.Fatalf("record removed: %v", err)
	}
}

func TestStopRemoveWorktreeRemovesCleanCheckoutAndKeepsBranch(t *testing.T) {
	fake, home, repo, wsID, record := worktreeAgent(t)
	if envelope := stopRun(t, home, Options{Name: "wt", RemoveWorktree: true}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := fake.Requests("workspace.close"); len(got) != 1 || got[0].Params["workspace_id"] != wsID {
		t.Fatalf("workspace.close = %+v", got)
	}
	if _, err := os.Stat(record.Worktree); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkout still at %s: %v", record.Worktree, err)
	}
	if strings.Contains(git(t, "-C", repo, "worktree", "list"), record.Worktree) {
		t.Fatalf("git still lists the worktree")
	}
	if !strings.Contains(git(t, "-C", repo, "branch", "--list", "feat"), "feat") {
		t.Fatalf("branch feat removed")
	}
	if _, err := agent.Load(home, "wt"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
}

func TestStopRemoveWorktreeWithoutWorktreeClosesPane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	envelope := stopRun(t, home, Options{Name: "kid", RemoveWorktree: true})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, ok := fake.Pane(pane.ID); ok {
		t.Fatalf("pane %s still open", pane.ID)
	}
	if _, err := agent.Load(home, "kid"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("record err = %v, want not found", err)
	}
	if text := Text(envelope); !strings.Contains(text, "no worktree for kid\n") {
		t.Fatalf("text = %q, want the line no worktree for kid", text)
	}
}

func TestStopRemoveWorktreeWithoutWorktreeHasNoWorktreeEffect(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	envelope := stopRun(t, home, Options{Name: "kid", RemoveWorktree: true})
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Effects []output.Effect `json:"effects"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := []output.Effect{
		{Resource: "pane", Path: pane.ID, Action: "closed"},
		{Resource: "record", Path: "agents/kid.json", Action: "removed"},
	}
	if !slices.Equal(got.Effects, want) {
		t.Fatalf("effects = %+v, want %+v", got.Effects, want)
	}
}
