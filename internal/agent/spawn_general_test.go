package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// writeGeneralParent writes the live registered caller boss in pane w1:p9.
func writeGeneralParent(t *testing.T, home string) {
	t.Helper()
	writeRecord(t, home, records.Agent{Key: "boss", Name: "boss", Role: "agent", Status: records.AgentLive, PaneID: "w1:p9", Generation: 1, Spawned: true})
}

func generalOptions(name, harnessKind, cwd string) SpawnOptions {
	return SpawnOptions{
		Role:      "agent",
		Harness:   harnessKind,
		Model:     "m",
		Name:      name,
		NameSet:   true,
		Prompt:    "Review the plan.",
		PromptSet: true,
		Cwd:       cwd,
		CwdSet:    true,
	}
}

func tabCreatedJSON(tabID, paneID, cwd string) string {
	return `{"type":"tab_created","tab":{"tab_id":"` + tabID + `","workspace_id":"w1","number":3,"label":"x","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":` + paneInfoJSON(paneID, "term_x", "w1", tabID, cwd) + `}`
}

func interactiveAgentJSON(paneID, tabID, status, name string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"term_x","workspace_id":"w1","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":3,"name":"` + name + `","interactive_ready":true}`
}

// scriptGeneralStart scripts the spawn core in pane paneID up to a settled
// wait with waitStatus.
func scriptGeneralStart(fake *fakeHerdr, sleep sleepProcess, paneID, tabID, cwd, name, kind, waitStatus string) {
	scriptReadyShell(fake, paneID, "term_x", "w1", tabID, cwd)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON(paneID, "term_x", "w1", tabID, "unknown", name)+`,"argv":["`+kind+`"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON(paneID, "term_x", "w1", tabID, cwd)+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON(paneID, "term_x", "w1", tabID, waitStatus, name)+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON(paneID, "term_x", "w1", tabID, waitStatus, name)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, kind, kind))+`}`)
}

// scriptGeneralNewTab scripts a general spawn in a new tab w1:t3 with root
// pane w1:p5.
func scriptGeneralNewTab(fake *fakeHerdr, sleep sleepProcess, cwd, name, kind, waitStatus string) {
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p9", "term_p", "w1", "w1:t1", "/tmp")+`}`)
	fake.result("tab.create", tabCreatedJSON("w1:t3", "w1:p5", cwd))
	scriptGeneralStart(fake, sleep, "w1:p5", "w1:t3", cwd, name, kind, waitStatus)
}

func scriptGeneralPrompt(fake *fakeHerdr, paneID, tabID, name string) {
	fake.result("agent.get", `{"type":"agent_info","agent":`+interactiveAgentJSON(paneID, tabID, "idle", name)+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+interactiveAgentJSON(paneID, tabID, "working", name)+`}`)
}

func decodeParams(t *testing.T, raw string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		t.Fatalf("decode params %s: %v", raw, err)
	}
}

func TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce(t *testing.T) {
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
	stored := readAgent(t, home, "x")
	if stored.Role != "agent" || stored.Parent != "boss" || stored.Status != records.AgentLive || stored.Project != "" || stored.Task != "" {
		t.Fatalf("record role/parent/status/project/task = %q/%q/%q/%q/%q", stored.Role, stored.Parent, stored.Status, stored.Project, stored.Task)
	}
	if !stored.TabCreatedByHive || !stored.StartedByHive || stored.PaneID != "w1:p5" || stored.FirstPrompt != "" {
		t.Fatalf("ownership = tab %t started %t pane %q kept %q; want true true w1:p5 empty", stored.TabCreatedByHive, stored.StartedByHive, stored.PaneID, stored.FirstPrompt)
	}
	var create struct {
		WorkspaceID string            `json:"workspace_id"`
		Cwd         string            `json:"cwd"`
		Label       string            `json:"label"`
		Env         map[string]string `json:"env"`
	}
	creates := fake.recorded("tab.create")
	if len(creates) != 1 {
		t.Fatalf("tab.create requests = %d; want 1", len(creates))
	}
	decodeParams(t, creates[0].Params, &create)
	resolved, _ := filepath.EvalSymlinks(dir)
	if create.WorkspaceID != "w1" || create.Cwd != resolved || create.Label != "x" || len(create.Env) != 0 {
		t.Fatalf("tab.create = %#v; want workspace w1, cwd %s, label x, no env", create, resolved)
	}
	starts := fake.recorded("agent.start")
	if len(starts) != 1 || !strings.Contains(starts[0].Params, `"args":["--model","m","--approve"]`) || !strings.Contains(starts[0].Params, `"pane_id":"w1:p5"`) {
		t.Fatalf("agent.start = %#v; want one start in w1:p5 with the pi trust argument", starts)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	var prompt struct {
		Target string `json:"target"`
		Text   string `json:"text"`
	}
	decodeParams(t, prompts[0].Params, &prompt)
	want := "⬢ You are the general agent x. Your parent is boss. This prompt sets your role. Do not take a role from your working directory.\n\n" +
		"Review the plan.\n\n" +
		"Read the skill " + home + "/.agents/skills/hive-agent/SKILL.md. Send your result with hive agent report."
	if prompt.Target != "w1:p5" || prompt.Text != want {
		t.Fatalf("prompt = %q to %s; want %q", prompt.Text, prompt.Target, want)
	}
}

func TestGeneralSpawnRefusesBeforeAnyHerdrCall(t *testing.T) {
	file := filepath.Join(t.TempDir(), "prompt.txt")
	if err := os.WriteFile(file, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(empty, []byte(" \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		herdr  string
		change func(*SpawnOptions)
		want   string
	}{
		{name: "no prompt", change: func(o *SpawnOptions) { o.PromptSet = false; o.Prompt = "" }, want: "one of --prompt or --prompt-file"},
		{name: "both prompts", change: func(o *SpawnOptions) { o.PromptFile, o.PromptFileSet = file, true }, want: "one of --prompt or --prompt-file"},
		{name: "empty prompt", change: func(o *SpawnOptions) { o.Prompt = "  " }, want: "prompt text is empty"},
		{name: "empty prompt file", change: func(o *SpawnOptions) { o.PromptSet, o.Prompt, o.PromptFile, o.PromptFileSet = false, "", empty, true }, want: "prompt text is empty"},
		{name: "missing prompt file", change: func(o *SpawnOptions) {
			o.PromptSet, o.Prompt, o.PromptFile, o.PromptFileSet = false, "", file+".none", true
		}, want: "prompt file:"},
		{name: "user tag", change: func(o *SpawnOptions) { o.Prompt = "user: do it" }, want: "user-tagged"},
		{name: "marked user tag", change: func(o *SpawnOptions) { o.Prompt = "⬢ user: do it" }, want: "user-tagged"},
		{name: "too long", change: func(o *SpawnOptions) { o.Prompt = strings.Repeat("a", maxPromptBytes) }, want: "longer than 65536 bytes"},
		{name: "no name", change: func(o *SpawnOptions) { o.Name, o.NameSet = "", false }, want: "agent requires --name"},
		{name: "bad name", change: func(o *SpawnOptions) { o.Name = "Bad" }, want: "invalid agent name"},
		{name: "coordinator name", change: func(o *SpawnOptions) { o.Name = "coordinator" }, want: "reserved"},
		{name: "supervisor name", change: func(o *SpawnOptions) { o.Name = "supervisor" }, want: "reserved"},
		{name: "manager name", change: func(o *SpawnOptions) { o.Name = "manager-hive" }, want: "reserved"},
		{name: "old name", change: func(o *SpawnOptions) { o.Name = "gone" }, want: "agent exists: gone"},
		{name: "task", change: func(o *SpawnOptions) { o.Task = "1-demo" }, want: "agent takes no --task"},
		{name: "no caller", herdr: "w1:p77", want: "no registered caller"},
		{name: "agent override in herdr", change: func(o *SpawnOptions) { o.Agent = "other" }, want: "--agent other is not the caller boss"},
		{name: "no caller outside herdr", herdr: "-", want: "no registered caller"},
		{name: "cwd not a directory", change: func(o *SpawnOptions) { o.Cwd = file }, want: "is not a directory"},
		{name: "cwd missing", change: func(o *SpawnOptions) { o.Cwd = file + ".none" }, want: "is not a directory"},
		{name: "no trust method", change: func(o *SpawnOptions) { o.Harness = "codex" }, want: "harness codex can show a trust dialog for "},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			writeGeneralParent(t, home)
			writeRecord(t, home, records.Agent{Key: "other", Name: "other", Role: "agent", Status: records.AgentLive, PaneID: "w1:p8", Generation: 1})
			writeRecord(t, home, records.Agent{Key: "gone", Name: "gone", Role: "agent", Status: records.AgentStopped, PaneID: "w1:p7", Generation: 1})
			pane := "w1:p9"
			if test.herdr == "-" {
				pane = ""
			} else if test.herdr != "" {
				pane = test.herdr
			}
			t.Setenv("HERDR_PANE_ID", pane)
			fake := newFakeHerdr(t)
			options := generalOptions("x", "pi", t.TempDir())
			if test.change != nil {
				test.change(&options)
			}
			envelope := Spawn(context.Background(), home, options)
			if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), test.want) {
				t.Fatalf("envelope = %#v; want failed with %q", envelope, test.want)
			}
			if len(fake.requests) != 0 {
				t.Fatalf("herdr requests = %#v; want none", fake.requests)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", "x.json")); err == nil {
				t.Fatalf("record x exists after a refusal")
			}
		})
	}
}

func TestGeneralSpawnNoTrustMethodNamesHarnessAndDirectory(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	newFakeHerdr(t)
	envelope := Spawn(context.Background(), home, func() SpawnOptions {
		options := generalOptions("x", "codex", dir)
		return options
	}())
	resolved, _ := filepath.EvalSymlinks(dir)
	message, _ := envelope.Error.(string)
	if !strings.Contains(message, "harness codex") || !strings.Contains(message, resolved) || !strings.Contains(message, "--allow-trust-dialog") {
		t.Fatalf("error = %q; want the harness, the directory %s, and --allow-trust-dialog", message, resolved)
	}
}

func TestGeneralSpawnAllowTrustDialogKeepsPromptAndDoesNotSend(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "codex", "blocked")
	options := generalOptions("x", "codex", dir)
	options.AllowTrustDialog = true

	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "partial" || !strings.Contains(envelope.Error.(string), "agent blocked") {
		t.Fatalf("envelope = %#v; want partial blocked", envelope)
	}
	stored := readAgent(t, home, "x")
	if stored.Status != records.AgentLive || !strings.HasPrefix(stored.FirstPrompt, "⬢ You are the general agent x.") {
		t.Fatalf("record status %q kept prompt %q; want live and the kept first prompt", stored.Status, stored.FirstPrompt)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 0 {
		t.Fatalf("agent.prompt requests = %d; want 0", len(prompts))
	}
}

// TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent covers decision
// 15: a trust dialog gives a live record with the agent status blocked and a
// blocked record in the inbox of the parent. Hive does not answer the dialog.
func TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "codex", "blocked")
	options := generalOptions("x", "codex", dir)
	options.AllowTrustDialog = true

	Spawn(context.Background(), home, options)
	stored := readAgent(t, home, "x")
	if stored.Status != records.AgentLive || stored.AgentStatus != "blocked" {
		t.Fatalf("record status %q agent status %q; want live and blocked", stored.Status, stored.AgentStatus)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var held []records.Inbox
	if err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		held, err = inbox.AllHeld(locked, "boss")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(held) != 1 || held[0].Kind != records.InboxBlocked || held[0].Agent != "x" {
		t.Fatalf("parent inbox = %+v; want one blocked record for x", held)
	}
	for _, method := range []string{"agent.prompt", "pane.send_text", "pane.send_keys"} {
		if requests := fake.recorded(method); len(requests) != 0 {
			t.Fatalf("%s requests = %d; want 0", method, len(requests))
		}
	}
}

func TestGeneralSpawnClaudeTrustEnvironment(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "claude", "idle")
	scriptGeneralPrompt(fake, "w1:p5", "w1:t3", "x")
	if envelope := Spawn(context.Background(), home, generalOptions("x", "claude", dir)); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	var create struct {
		Env map[string]string `json:"env"`
	}
	decodeParams(t, fake.recorded("tab.create")[0].Params, &create)
	if create.Env["CLAUDE_CODE_SANDBOXED"] != "1" {
		t.Fatalf("tab.create env = %#v; want CLAUDE_CODE_SANDBOXED=1", create.Env)
	}
}

func TestGeneralSpawnClaudeInGivenPaneNeedsAllowTrustDialog(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p4", "term_x", "w1", "w1:t1", dir)+`}`)
	options := generalOptions("x", "claude", "")
	options.CwdSet = false
	options.PaneID, options.PaneSet = "w1:p4", true
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), "CLAUDE_CODE_SANDBOXED=1") {
		t.Fatalf("envelope = %#v; want a trust refusal", envelope)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 0 {
		t.Fatalf("agent.start requests = %d; want 0", len(starts))
	}
}

// The destination pane w1:p4 differs from the caller pane w1:p9, and --agent
// names the caller (decision 25).
func TestGeneralSpawnGivenPaneUsesItsDirectoryAndMakesNoTab(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p4", "term_x", "w1", "w1:t1", dir)+`}`)
	scriptGeneralStart(fake, sleep, "w1:p4", "w1:t1", dir, "x", "pi", "idle")
	scriptGeneralPrompt(fake, "w1:p4", "w1:t1", "x")
	options := generalOptions("x", "pi", link)
	options.PaneID, options.PaneSet = "w1:p4", true
	options.Agent = "boss"

	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if creates := fake.recorded("tab.create"); len(creates) != 0 {
		t.Fatalf("tab.create requests = %d; want 0", len(creates))
	}
	stored := readAgent(t, home, "x")
	if stored.PaneID != "w1:p4" || stored.Parent != "boss" || stored.TabCreatedByHive || !stored.StartedByHive {
		t.Fatalf("record pane %q parent %q tab %t started %t; want w1:p4 boss false true", stored.PaneID, stored.Parent, stored.TabCreatedByHive, stored.StartedByHive)
	}
}

func TestGeneralSpawnGivenPaneRefusesOtherCwd(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p4", "term_x", "w1", "w1:t1", t.TempDir())+`}`)
	options := generalOptions("x", "pi", t.TempDir())
	options.PaneID, options.PaneSet = "w1:p4", true
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), "differs from the directory") {
		t.Fatalf("envelope = %#v; want a directory refusal", envelope)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 0 {
		t.Fatalf("agent.start requests = %d; want 0", len(starts))
	}
}

func TestGeneralSpawnRelativeCwdUsesProcessDirectory(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, base, "x", "pi", "idle")
	scriptGeneralPrompt(fake, "w1:p5", "w1:t3", "x")
	if envelope := Spawn(context.Background(), home, generalOptions("x", "pi", "sub")); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	var create struct {
		Cwd string `json:"cwd"`
	}
	decodeParams(t, fake.recorded("tab.create")[0].Params, &create)
	want, _ := filepath.EvalSymlinks(filepath.Join(base, "sub"))
	if create.Cwd != want {
		t.Fatalf("tab.create cwd = %q; want %q", create.Cwd, want)
	}
}

func TestGeneralSpawnOutsideHerdrUsesAgentFlag(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "pi", "idle")
	scriptGeneralPrompt(fake, "w1:p5", "w1:t3", "x")
	options := generalOptions("x", "pi", dir)
	options.Agent = "boss"
	if envelope := Spawn(context.Background(), home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if stored := readAgent(t, home, "x"); stored.Parent != "boss" {
		t.Fatalf("parent = %q; want boss", stored.Parent)
	}
	if gets := fake.recorded("pane.get"); len(gets) == 0 || gets[0].Params != `{"pane_id":"w1:p9"}` {
		t.Fatalf("first pane.get = %#v; want the parent pane w1:p9", gets)
	}
}

func TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane(t *testing.T) {
	for _, test := range []struct {
		name      string
		panes     string
		wantClose bool
		want      string
	}{
		{name: "root pane only", panes: paneInfoJSON("w1:p5", "term_x", "w1", "w1:t3", "/tmp"), wantClose: true, want: "closed"},
		{name: "other pane", panes: paneInfoJSON("w1:p5", "term_x", "w1", "w1:t3", "/tmp") + "," + paneInfoJSON("w1:p6", "term_y", "w1", "w1:t3", "/tmp"), want: "kept: the tab holds pane w1:p6"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			writeGeneralParent(t, home)
			t.Setenv("HERDR_PANE_ID", "w1:p9")
			dir := t.TempDir()
			fake := newFakeHerdr(t)
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p9", "term_p", "w1", "w1:t1", "/tmp")+`}`)
			fake.result("tab.create", tabCreatedJSON("w1:t3", "w1:p5", dir))
			scriptReadyShell(fake, "w1:p5", "term_x", "w1", "w1:t3", dir)
			fake.failure("agent.start", "unsupported_agent_kind", "bad kind")
			fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p5", "term_x", "w1", "w1:t3", dir)+`}`)
			fake.result("pane.list", `{"type":"pane_list","panes":[`+test.panes+`]}`)
			fake.result("pane.close", `{"type":"ok"}`)
			envelope := Spawn(context.Background(), home, generalOptions("x", "pi", dir))
			if envelope.Outcome != "failed" {
				t.Fatalf("envelope = %#v; want failed", envelope)
			}
			closes := fake.recorded("pane.close")
			if test.wantClose != (len(closes) == 1) {
				t.Fatalf("pane.close requests = %#v; want close %t", closes, test.wantClose)
			}
			if test.wantClose && closes[0].Params != `{"pane_id":"w1:p5"}` {
				t.Fatalf("pane.close = %s; want w1:p5", closes[0].Params)
			}
			found := false
			for _, effect := range envelope.Effects {
				if effect.Resource == "tab" && effect.Path == "w1:t3" && effect.Action == test.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("effects = %#v; want tab w1:t3 %q", envelope.Effects, test.want)
			}
			if stored := readAgent(t, home, "x"); stored.Status != records.AgentExited || stored.StartedByHive {
				t.Fatalf("record status %q started %t; want exited false", stored.Status, stored.StartedByHive)
			}
		})
	}
}

func TestGeneralSpawnUncertainStartKeepsPane(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	fake := newFakeHerdr(t)
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p9", "term_p", "w1", "w1:t1", "/tmp")+`}`)
	fake.result("tab.create", tabCreatedJSON("w1:t3", "w1:p5", dir))
	scriptReadyShell(fake, "w1:p5", "term_x", "w1", "w1:t3", dir)
	fake.drop("agent.start")
	envelope := Spawn(context.Background(), home, generalOptions("x", "pi", dir))
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if closes := fake.recorded("pane.close"); len(closes) != 0 {
		t.Fatalf("pane.close requests = %d; want 0", len(closes))
	}
	if stored := readAgent(t, home, "x"); stored.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", stored.Status)
	}
}

func TestGeneralSpawnUncertainPromptMakesRecordUnresolved(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "pi", "idle")
	fake.result("agent.get", `{"type":"agent_info","agent":`+interactiveAgentJSON("w1:p5", "w1:t3", "idle", "x")+`}`)
	fake.drop("agent.prompt")
	envelope := Spawn(context.Background(), home, generalOptions("x", "pi", dir))
	if envelope.Outcome != "partial" || envelope.Phase != "prompt" {
		t.Fatalf("envelope = %#v; want partial prompt", envelope)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	if stored := readAgent(t, home, "x"); stored.Status != records.AgentUnresolved || stored.FirstPrompt != "" {
		t.Fatalf("record status %q kept %q; want unresolved and no kept prompt", stored.Status, stored.FirstPrompt)
	}
}

func TestGeneralSpawnStalledPromptIsNotSentAgain(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "pi", "idle")
	fake.result("agent.get", `{"type":"agent_info","agent":`+interactiveAgentJSON("w1:p5", "w1:t3", "idle", "x")+`}`)
	fake.failure("agent.prompt", "agent_prompt_stalled", "stalled")
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+interactiveAgentJSON("w1:p5", "w1:t3", "working", "x")+`}`)
	envelope := Spawn(context.Background(), home, generalOptions("x", "pi", dir))
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	if stored := readAgent(t, home, "x"); stored.Status != records.AgentUnresolved {
		t.Fatalf("record status = %q; want unresolved", stored.Status)
	}
}

func TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt(t *testing.T) {
	home := testHome(t)
	writeGeneralParent(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	dir := t.TempDir()
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptGeneralNewTab(fake, sleep, dir, "x", "codex", "idle")
	options := generalOptions("x", "codex", dir)
	options.AllowTrustDialog = true

	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if stored := readAgent(t, home, "x"); !strings.HasPrefix(stored.FirstPrompt, "⬢ You are the general agent x.") {
		t.Fatalf("kept prompt = %q; want the first prompt", stored.FirstPrompt)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 0 {
		t.Fatalf("agent.prompt requests = %d; want 0", len(prompts))
	}
}
