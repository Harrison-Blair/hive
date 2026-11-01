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
	path := filepath.Join(home, "worktrees", "repo", "feat")
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

func TestStopWorktreeAgentClosesOnlyItsPane(t *testing.T) {
	fake, home, repo, wsID, record := worktreeAgent(t)
	other := fake.AddPane(wsID, record.Worktree)
	if err := agent.Save(home, agent.Record{Name: "peer", Pane: other.ID, Cwd: record.Worktree, Worktree: record.Worktree}); err != nil {
		t.Fatal(err)
	}
	if envelope := stopRun(t, home, Options{Name: "wt"}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d, want 0", got)
	}
	if _, ok := fake.Pane(record.Pane); ok {
		t.Fatalf("pane %s still open", record.Pane)
	}
	if _, ok := fake.Pane(other.ID); !ok {
		t.Fatalf("pane %s of the other agent closed", other.ID)
	}
	if _, err := agent.Load(home, "peer"); err != nil {
		t.Fatalf("record of the other agent: %v", err)
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

func TestStopEffectsAreThePaneAndTheRecord(t *testing.T) {
	_, home, _, _, record := worktreeAgent(t)
	envelope := stopRun(t, home, Options{Name: "wt"})
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
		{Resource: "pane", Path: record.Pane, Action: "closed"},
		{Resource: "record", Path: "agents/wt.json", Action: "removed"},
	}
	if !slices.Equal(got.Effects, want) {
		t.Fatalf("effects = %+v, want %+v", got.Effects, want)
	}
}
