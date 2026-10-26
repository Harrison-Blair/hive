package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// spawnPromptText returns the text of the one prompt that a spawn sent.
func spawnPromptText(t *testing.T, fake *fakeHerdr) string {
	t.Helper()
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(prompts[0].Params), &params); err != nil {
		t.Fatalf("decode prompt: %v", err)
	}
	return params.Text
}

// TestSpawnPromptFactsThenSpawnerText proves AC1: --prompt works with each
// placement and follows the facts in the first prompt.
func TestSpawnPromptFactsThenSpawnerText(t *testing.T) {
	t.Run("task", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "builder-1")
		envelope := Spawn(context.Background(), home, SpawnOptions{
			Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "builder-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
			Prompt: "Do the work.", PromptSet: true,
		})
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		want := "⬢ You are builder-1. Your project is demo. Your task is 1-worker-demo. " +
			"Your worktree is " + worktree + ". Your brief is .hive/state/briefs/1-worker-demo.md.\n\nDo the work."
		if got := spawnPromptText(t, fake); got != want {
			t.Fatalf("prompt = %q; want %q", got, want)
		}
	})

	t.Run("project", func(t *testing.T) {
		home := testHome(t)
		dir := projectDirectory(t, home, "demo")
		writeManagerProject(t, home, nil)
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
		scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")
		options := managerSpawnOptions("demo")
		options.Prompt, options.PromptSet = "Plan the work.", true
		envelope := Spawn(context.Background(), home, options)
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		want := "⬢ You are manager-demo. Your project is demo.\n\nPlan the work."
		if got := spawnPromptText(t, fake); got != want {
			t.Fatalf("prompt = %q; want %q", got, want)
		}
	})

	t.Run("cwd", func(t *testing.T) {
		home := testHome(t)
		writeGeneralParent(t, home)
		t.Setenv("HERDR_PANE_ID", "w1:p9")
		dir := t.TempDir()
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptGeneralNewTab(fake, sleep, dir, "x", "pi", "idle")
		scriptGeneralPrompt(fake, "w1:p5", "w1:t3", "x")
		envelope := Spawn(context.Background(), home, generalOptions("x", "pi", dir))
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		want := "⬢ You are x. Your parent is boss.\n\nReview the plan."
		if got := spawnPromptText(t, fake); got != want {
			t.Fatalf("prompt = %q; want %q", got, want)
		}
	})

	t.Run("pane", func(t *testing.T) {
		home := testHome(t)
		writeGeneralParent(t, home)
		t.Setenv("HERDR_PANE_ID", "w1:p9")
		dir := t.TempDir()
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p4", "term_x", "w1", "w1:t1", dir)+`}`)
		scriptGeneralStart(fake, sleep, "w1:p4", "w1:t1", dir, "x", "pi", "idle")
		scriptGeneralPrompt(fake, "w1:p4", "w1:t1", "x")
		options := generalOptions("x", "pi", dir)
		options.PaneID, options.PaneSet, options.CwdSet = "w1:p4", true, false
		envelope := Spawn(context.Background(), home, options)
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		want := "⬢ You are x. Your parent is boss.\n\nReview the plan."
		if got := spawnPromptText(t, fake); got != want {
			t.Fatalf("prompt = %q; want %q", got, want)
		}
	})
}

// TestSpawnProjectUsesHiveWorkspaceAtProjectDirectory proves AC3.
func TestSpawnProjectUsesHiveWorkspaceAtProjectDirectory(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || creates[0].Params != `{"cwd":"`+dir+`","label":"hive:demo","focus":false}` {
		t.Fatalf("workspace.create = %#v; want hive:demo at the project directory", creates)
	}
	starts := fake.recorded("agent.start")
	if len(starts) != 1 || !strings.Contains(starts[0].Params, `"pane_id":"wBD:p1"`) {
		t.Fatalf("agent.start = %#v; want the root pane of the first tab", starts)
	}
	record := readAgent(t, home, "manager-demo-g1")
	if record.Profile != records.ProfilePlanner || record.Owns != "demo" || record.Project != "demo" || record.Worktree != "" {
		t.Fatalf("record facts = %#v; want planner owning demo", record)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var claim records.Claim
	if err := registry.ReadIn([]string{"claims"}, "owner-demo", &claim); err != nil {
		t.Fatalf("owner claim: %v", err)
	}
	if claim.Agent != "manager-demo-g1" || claim.Generation != 1 {
		t.Fatalf("owner claim = %#v", claim)
	}
}

// TestSpawnTaskSetsActingAgentByTaskFacts proves AC4.
func TestSpawnTaskSetsActingAgentByTaskFacts(t *testing.T) {
	t.Run("created task sets the acting agent", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "builder-1")
		envelope := Spawn(context.Background(), home, SpawnOptions{
			Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m",
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

	t.Run("completed task binds without an acting agent", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
		fake.result("tab.create", tabCreateJSON("w5", "w5:t2", "w5:p2", worktree))
		scriptWorkerSuccess(fake, sleep, "w5:p2", "w5", "w5:t2", worktree, "builder-1")
		envelope := Spawn(context.Background(), home, SpawnOptions{
			Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m",
			Name: "builder-1", NameSet: true,
		})
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok", envelope)
		}
		task := readTaskRecord(t, home, "1-worker-demo")
		if task.Status != "completed" || task.ActingAgent != "" {
			t.Fatalf("task = %#v; want completed with no acting agent", task)
		}
		if len(fake.recorded("tab.create")) != 1 {
			t.Fatalf("a bound agent did not take a new tab")
		}
	})
}

// TestSpawnOwnerExclusivity proves AC5: one live owner per project.
func TestSpawnOwnerExclusivity(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")
	if envelope := Spawn(context.Background(), home, managerSpawnOptions("demo")); envelope.Outcome != "ok" {
		t.Fatalf("first owner = %#v; want ok", envelope)
	}

	before := len(fake.requests)
	second := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if second.Outcome != "failed" || !strings.Contains(toString(second.Error), "project demo has an owner: manager-demo") {
		t.Fatalf("second owner = %#v; want the live owner refusal", second)
	}
	if len(fake.requests) != before {
		t.Fatalf("refused owner sent herdr requests: %#v", fake.requests[before:])
	}

	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdateIn([]string{"agents"}, "manager-demo-g1", func(raw json.RawMessage) (json.RawMessage, error) {
		var record records.Agent
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, err
		}
		record.Status = records.AgentStopped
		return json.Marshal(record)
	}); err != nil {
		t.Fatal(err)
	}

	next := newFakeHerdr(t)
	next.result("workspace.get", managerWorkspaceGetJSON("wBD", "hive:demo"))
	next.result("tab.list", managerTabListJSON())
	next.result("tab.create", managerTabCreatedJSON("wBD", "wBD:t2", "wBD:p2", dir))
	scriptManagerSpawnTail(t, home, next, sleep, "wBD:p2", "wBD", "wBD:t2")
	third := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if third.Outcome != "ok" {
		t.Fatalf("owner after stop = %#v; want ok", third)
	}
	var claim records.Claim
	if err := registry.ReadIn([]string{"claims"}, "owner-demo", &claim); err != nil {
		t.Fatal(err)
	}
	if claim.Generation != 2 || claim.Agent != "manager-demo-g2" {
		t.Fatalf("claim = %#v; want generation 2", claim)
	}
}

// TestSpawnRecordsProfileFactsAndNoRoleShim proves AC1: a spawn writes the
// profile facts and no old role shim event.
func TestSpawnRecordsProfileFactsAndNoRoleShim(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "t1-builder-1")
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "t1-builder-1", NameSet: true,
		PaneID: "w5:p1", PaneSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	record := readAgent(t, home, "t1-builder-1")
	if record.Profile != records.ProfileBuilder || record.Kind != "" || record.Owns != "" {
		t.Fatalf("record facts = %#v; want profile builder and no built-in facts", record)
	}
	logText := readLog(t, home)
	for _, line := range strings.Split(strings.TrimSpace(logText), "\n") {
		var event struct {
			Message string `json:"msg"`
		}
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}
		if event.Message == "agent spawn old role flag" {
			t.Fatalf("log has the removed old role shim event: %s", line)
		}
	}
}

// TestSpawnRefusesReservedNamesWithoutEffects proves the carried reserved-name
// finding: both names are refused in each placement, with no start request, no
// record, and no acting-agent change.
func TestSpawnRefusesReservedNamesWithoutEffects(t *testing.T) {
	for _, name := range []string{"coordinator", "supervisor"} {
		t.Run(name, func(t *testing.T) {
			home := testHome(t)
			repo, worktree, branch := taskWorktree(t)
			writeProjectPath(t, home, "demo", repo)
			writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
			writeManagerProject(t, home, nil)
			fake := newFakeHerdr(t)
			options := SpawnOptions{
				Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m",
				Name: name, NameSet: true, PaneID: "w5:p1", PaneSet: true,
			}
			envelope := Spawn(context.Background(), home, options)
			if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "name is reserved: "+name) {
				t.Fatalf("task placement = %#v; want the reserved-name refusal", envelope)
			}
			projectOptions := managerSpawnOptions("demo")
			projectOptions.Name, projectOptions.NameSet = name, true
			if envelope := Spawn(context.Background(), home, projectOptions); envelope.Outcome != "failed" {
				t.Fatalf("project placement = %#v; want the reserved-name refusal", envelope)
			}
			t.Setenv("HERDR_PANE_ID", "w1:p9")
			writeGeneralParent(t, home)
			general := generalOptions(name, "pi", t.TempDir())
			if envelope := Spawn(context.Background(), home, general); envelope.Outcome != "failed" {
				t.Fatalf("cwd placement = %#v; want the reserved-name refusal", envelope)
			}
			paneOptions := general
			paneOptions.CwdSet = false
			paneOptions.PaneID, paneOptions.PaneSet = "w1:p9", true
			if envelope := Spawn(context.Background(), home, paneOptions); envelope.Outcome != "failed" {
				t.Fatalf("pane placement = %#v; want the reserved-name refusal", envelope)
			}
			if len(fake.requests) != 0 {
				t.Fatalf("refusals sent herdr requests: %#v", fake.requests)
			}
			entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "agents"))
			if err == nil {
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), name) {
						t.Fatalf("refusals wrote agent record %s", entry.Name())
					}
				}
			}
			task := readTaskRecord(t, home, "1-worker-demo")
			if task.ActingAgent != "" || task.Status != "created" {
				t.Fatalf("task = %#v; want created with no acting agent", task)
			}
		})
	}
}

// TestSpawnOwnerRefusesAnotherAgent proves the internal owner rule for an
// agent caller that is not the coordinator.
func TestSpawnOwnerRefusesAnotherAgent(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	writeRecord(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Profile: records.ProfileBuilder, Project: "demo", Task: "1-demo", PaneID: "w1:p9", Status: records.AgentLive, Spawned: true})
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	newFakeHerdr(t)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "--owner needs the coordinator or the user: worker-1") {
		t.Fatalf("envelope = %#v; want the caller refusal", envelope)
	}
}

// TestSpawnPromptFileFollowsFacts proves that --prompt-file reaches the first
// prompt with the same order as --prompt.
func TestSpawnPromptFileFollowsFacts(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	file := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(file, []byte("From the file."), 0o644); err != nil {
		t.Fatal(err)
	}
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "builder-1")
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: records.ProfileBuilder, Task: "1-worker-demo", Harness: "claude", Model: "m",
		Name: "builder-1", NameSet: true, PaneID: "w5:p1", PaneSet: true,
		PromptFile: file, PromptFileSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	want := "⬢ You are builder-1. Your project is demo. Your task is 1-worker-demo. " +
		"Your worktree is " + worktree + ". Your brief is .hive/state/briefs/1-worker-demo.md.\n\nFrom the file."
	if got := spawnPromptText(t, fake); got != want {
		t.Fatalf("prompt = %q; want %q", got, want)
	}
}

// TestSpawnFirstPromptBoundary proves the 64 KiB boundary: exactly 65536 bytes
// is accepted and one more byte is refused.
func TestSpawnFirstPromptBoundary(t *testing.T) {
	base := "⬢ You are x. Your parent is boss."
	t.Run("exactly 65536 bytes", func(t *testing.T) {
		home := testHome(t)
		writeGeneralParent(t, home)
		t.Setenv("HERDR_PANE_ID", "w1:p9")
		dir := t.TempDir()
		sleep := startSleep(t)
		fake := newFakeHerdr(t)
		scriptGeneralNewTab(fake, sleep, dir, "x", "pi", "idle")
		scriptGeneralPrompt(fake, "w1:p5", "w1:t3", "x")
		options := generalOptions("x", "pi", dir)
		options.Prompt = strings.Repeat("a", 65536-len(base)-2)
		envelope := Spawn(context.Background(), home, options)
		if envelope.Outcome != "ok" {
			t.Fatalf("envelope = %#v; want ok at the boundary", envelope)
		}
		if got := len(spawnPromptText(t, fake)); got != 65536 {
			t.Fatalf("prompt bytes = %d; want 65536", got)
		}
	})
	t.Run("one byte above", func(t *testing.T) {
		home := testHome(t)
		writeGeneralParent(t, home)
		t.Setenv("HERDR_PANE_ID", "w1:p9")
		fake := newFakeHerdr(t)
		options := generalOptions("x", "pi", t.TempDir())
		options.Prompt = strings.Repeat("a", 65536-len(base)-1)
		envelope := Spawn(context.Background(), home, options)
		if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "first prompt is longer than 65536 bytes") {
			t.Fatalf("envelope = %#v; want the length refusal", envelope)
		}
		if len(fake.requests) != 0 {
			t.Fatalf("refusal sent herdr requests: %#v", fake.requests)
		}
	})
}
