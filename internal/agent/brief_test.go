package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestPiSpawnDeliversInstructionsInSafeOrder(t *testing.T) {
	_, fake, _ := spawnProfileTask(t, "pi")
	args := startArgs(t, fake)
	if !reflect.DeepEqual(args, []string{"--model", "m"}) {
		t.Errorf("Pi start arguments must contain only the model; got %d arguments", len(args))
	}
	for index, arg := range args {
		if strings.ContainsAny(arg, "\r\n") {
			t.Errorf("Pi start argument %d contains multiline text", index)
		}
	}
	prompt := promptText(t, fake)
	profile := strings.TrimSpace(readProfile(t, "builder"))
	guideText := strings.TrimSpace(guide.Text())
	reminder := "Read `profiles/builder.md` and run `hive guide` after a context reset."
	wantPrefix := "⬢ " + profile + "\n\n" + guideText + "\n\n" + reminder + "\n\nYou are builder-1."
	if !strings.Contains(prompt, profile) {
		t.Error("Pi first prompt misses the builder profile")
	}
	if !strings.Contains(prompt, guideText) {
		t.Error("Pi first prompt misses the guide")
	}
	if !strings.HasPrefix(prompt, wantPrefix) {
		t.Error("Pi first prompt must start with the profile, guide, reset reminder, and facts in that order")
	}
}

func TestFirstPromptStartsWithHiveMarker(t *testing.T) {
	got := firstPrompt(records.Agent{Name: "builder-1"}, nil, "")
	if !strings.HasPrefix(got, "⬢ ") {
		t.Fatalf("prompt = %q; want the hive marker prefix", got)
	}
}

func TestFirstPromptGivesFactsThenSpawnerText(t *testing.T) {
	record := records.Agent{
		Name:     "builder-1",
		Parent:   "manager-hive-g1",
		Project:  "hive",
		Task:     "154-awr-spawn",
		Worktree: "/home/user/projects/hive/worktrees/154-awr-spawn",
	}
	task := &records.Task{ID: "154-awr-spawn", Project: "hive", Feature: "agents-without-roles"}
	got := firstPrompt(record, task, "Do the task.")
	want := "⬢ You are builder-1. Your parent is manager-hive-g1. Your project is hive. " +
		"Your task is 154-awr-spawn. Your worktree is /home/user/projects/hive/worktrees/154-awr-spawn. " +
		"Your brief is .hive/spec/agents-without-roles/tasks/154-awr-spawn/brief.md. " +
		"Your evidence is .hive/spec/agents-without-roles/tasks/154-awr-spawn/evidence/.\n\nDo the task."
	if got != want {
		t.Fatalf("prompt = %q; want %q", got, want)
	}
}

func TestFirstPromptSkipsEmptyFacts(t *testing.T) {
	got := firstPrompt(records.Agent{Name: "general-1"}, nil, "")
	want := "⬢ You are general-1."
	if got != want {
		t.Fatalf("prompt = %q; want %q", got, want)
	}
}

func TestFirstPromptUsesTheTaskBriefWithoutFeature(t *testing.T) {
	record := records.Agent{Name: "builder-1", Task: "1-simple"}
	task := &records.Task{ID: "1-simple"}
	got := firstPrompt(record, task, "")
	if !strings.Contains(got, "Your brief is .hive/state/briefs/1-simple.md.") {
		t.Fatalf("prompt = %q; want the task brief path", got)
	}
	if strings.Contains(got, "Your evidence is") {
		t.Fatalf("prompt = %q; want no evidence path for a task with no feature", got)
	}
}

// claudeFileFlag is the Claude Code flag that names the instruction file.
const claudeFileFlag = "--append-system-prompt-file"

// knownProfile is a profile with white space around its text. The
// instruction file holds the text without that white space.
const knownProfile = "\n  # Known profile\n\nT186-CANARY: this line comes from the profile file.\n\n"

// writeKnownProfile writes knownProfile as profiles/NAME.md in home.
func writeKnownProfile(t *testing.T, home, name string) {
	t.Helper()
	directory := filepath.Join(home, "profiles")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("create profiles directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, name+".md"), []byte(knownProfile), 0o644); err != nil {
		t.Fatalf("write profile %s: %v", name, err)
	}
}

// wantKnownInstructions gives the instruction bytes for knownProfile as
// profile NAME: the profile, the guide, and the reset reminder (D4).
func wantKnownInstructions(name string) []byte {
	text := "# Known profile\n\nT186-CANARY: this line comes from the profile file." +
		"\n\n" + strings.TrimSpace(guide.Text()) +
		"\n\nRead `profiles/" + name + ".md` and run `hive guide` after a context reset."
	return []byte(text)
}

// startRead is what the fake herdr saw at one agent.start request: the
// arguments and the instruction file as it was at that time.
type startRead struct {
	args    []string
	path    string
	data    []byte
	mode    os.FileMode
	readErr error
}

type startReader struct {
	mu    sync.Mutex
	reads []startRead
}

// readStartFile reads the instruction file of each agent.start request at
// the time herdr gets the request, before the reply.
func readStartFile(t *testing.T, fake *fakeHerdr) *startReader {
	t.Helper()
	reader := &startReader{}
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "agent.start" {
			return
		}
		var start struct {
			Args []string `json:"args"`
		}
		if err := json.Unmarshal(params, &start); err != nil {
			t.Errorf("decode agent.start params: %v", err)
			return
		}
		read := startRead{args: start.Args}
		for index, arg := range start.Args {
			if arg == claudeFileFlag && index+1 < len(start.Args) {
				read.path = start.Args[index+1]
			}
		}
		if read.path != "" {
			info, err := os.Lstat(read.path)
			if err == nil {
				read.mode = info.Mode()
				read.data, err = os.ReadFile(read.path)
			}
			read.readErr = err
		}
		reader.mu.Lock()
		reader.reads = append(reader.reads, read)
		reader.mu.Unlock()
	})
	return reader
}

func (r *startReader) only(t *testing.T) startRead {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.reads) != 1 {
		t.Fatalf("agent.start reads = %d; want 1", len(r.reads))
	}
	return r.reads[0]
}

// checkStartFile checks one Claude start at the agent.start boundary: no
// argument holds a line break, one flag names an absolute private file in the
// home's .hive/tmp/ folder, and the file holds want.
func checkStartFile(t *testing.T, read startRead, home string, want []byte) {
	t.Helper()
	for index, arg := range read.args {
		if strings.ContainsAny(arg, "\r\n") {
			t.Fatalf("agent.start argument %d holds a line break", index)
		}
		if arg == "--append-system-prompt" {
			t.Fatalf("agent.start args = %q; want no text flag", read.args)
		}
	}
	flags := 0
	for _, arg := range read.args {
		if arg == claudeFileFlag {
			flags++
		}
	}
	if flags != 1 || read.path == "" {
		t.Fatalf("agent.start args = %q; want one %s PATH", read.args, claudeFileFlag)
	}
	if !filepath.IsAbs(read.path) {
		t.Fatalf("instruction file path %q is not absolute", read.path)
	}
	if got, want := filepath.Dir(read.path), filepath.Join(home, ".hive", "tmp"); got != want {
		t.Fatalf("instruction file folder = %q; want %q", got, want)
	}
	if read.readErr != nil {
		t.Fatalf("instruction file at agent.start: %v", read.readErr)
	}
	if !read.mode.IsRegular() || read.mode.Perm() != 0o600 {
		t.Fatalf("instruction file mode = %v; want a regular file with mode 0600", read.mode)
	}
	if !bytes.Equal(read.data, want) {
		t.Fatalf("instruction file holds %d bytes; want the exact %d instruction bytes", len(read.data), len(want))
	}
}

// tmpEntries returns the names in the home's .hive/tmp/ folder.
func tmpEntries(t *testing.T, home string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "tmp"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read .hive/tmp: %v", err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// writeOtherAttemptFile writes a file of another attempt in .hive/tmp/.
func writeOtherAttemptFile(t *testing.T, home string) string {
	t.Helper()
	directory := filepath.Join(home, ".hive", "tmp")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create .hive/tmp: %v", err)
	}
	path := filepath.Join(directory, "instructions-other-attempt.md")
	if err := os.WriteFile(path, []byte("other attempt\n"), 0o600); err != nil {
		t.Fatalf("write other attempt file: %v", err)
	}
	return path
}

func keptEffect(envelope output.Envelope, home, path string) bool {
	relative, err := filepath.Rel(home, path)
	if err != nil {
		return false
	}
	for _, effect := range envelope.Effects {
		if effect.Resource == "instructions" && effect.Action == "kept" && effect.Path == filepath.ToSlash(relative) {
			return true
		}
	}
	return false
}

// taskScript gives the herdr replies of one task spawn.
type taskScript func(fake *fakeHerdr, sleep sleepProcess, processName, worktree string)

// spawnKnownTask spawns one builder task agent with knownProfile. setup runs
// on the home before the spawn.
func spawnKnownTask(t *testing.T, harnessName string, setup func(home string), script taskScript) (string, *fakeHerdr, *startReader, output.Envelope) {
	t.Helper()
	home := testHome(t)
	writeKnownProfile(t, home, "builder")
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	if setup != nil {
		setup(home)
	}
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	reader := readStartFile(t, fake)
	profile, err := harness.Lookup(harnessName)
	if err != nil {
		t.Fatalf("Lookup(%s): %v", harnessName, err)
	}
	script(fake, sleep, profile.ProcessNames()[0], worktree)
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: "builder", ProfileSet: true, Task: "1-worker-demo",
		Harness: harnessName, Model: "m", Name: "builder-1", NameSet: true,
	})
	return home, fake, reader, envelope
}

// scriptTaskUntilStart scripts a task spawn up to the start request.
func scriptTaskUntilStart(fake *fakeHerdr, worktree string) {
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", workspaceRoot(worktree), worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
}

func readyTaskScript(fake *fakeHerdr, sleep sleepProcess, processName, worktree string) {
	scriptProfileTaskSpawn(fake, sleep, processName, "w5:p1", "w5", "w5:t1", worktree, "builder-1")
}

func refusedTaskScript(fake *fakeHerdr, _ sleepProcess, _ string, worktree string) {
	scriptTaskUntilStart(fake, worktree)
	fake.failure("agent.start", "invalid_agent_argument", "refused")
}

func uncertainTaskScript(fake *fakeHerdr, _ sleepProcess, _ string, worktree string) {
	scriptTaskUntilStart(fake, worktree)
	fake.drop("agent.start")
}

// blockedTaskScript scripts a start that waits on a dialog. The harness can
// show the dialog before it reads the instruction file.
func blockedTaskScript(fake *fakeHerdr, sleep sleepProcess, processName, worktree string) {
	scriptTaskUntilStart(fake, worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "builder-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "blocked", "builder-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "blocked", "builder-1"))
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w5:p1", processEntry(sleep.pid, processName, processName))+`}`)
}

func TestClaudeInstructionFileTaskSpawnDeliversExactPrivateFile(t *testing.T) {
	home, fake, reader, envelope := spawnKnownTask(t, "claude", nil, readyTaskScript)
	if envelope.Outcome != "ok" {
		t.Fatalf("spawn = %#v; want ok", envelope)
	}
	read := reader.only(t)
	checkStartFile(t, read, home, wantKnownInstructions("builder"))
	if want := []string{"--dangerously-skip-permissions", "--model", "m", claudeFileFlag, read.path}; !reflect.DeepEqual(read.args, want) {
		t.Fatalf("agent.start args = %q; want %q", read.args, want)
	}
	prompt := promptText(t, fake)
	if strings.Contains(prompt, strings.TrimSpace(guide.Text())) || strings.Contains(prompt, "T186-CANARY") {
		t.Fatal("Claude first prompt holds the instructions; want only the file to carry them")
	}
	if !strings.HasPrefix(prompt, "⬢ You are builder-1.") {
		t.Fatalf("first prompt = %q; want the facts after the marker", prompt)
	}
	if _, err := os.Lstat(read.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("instruction file after a ready start: %v; want it removed", err)
	}
	if names := tmpEntries(t, home); len(names) != 0 {
		t.Fatalf(".hive/tmp holds %q after a ready start; want nothing", names)
	}
}

// TestClaudeInstructionFileRemovedAfterKnownOutcome covers a ready start and
// a start that herdr refused. Hive removes only its own file. The file of
// another attempt stays.
func TestClaudeInstructionFileRemovedAfterKnownOutcome(t *testing.T) {
	for _, test := range []struct {
		name    string
		outcome string
		status  string
		script  taskScript
	}{
		{name: "ready", outcome: "ok", status: records.AgentLive, script: readyTaskScript},
		{name: "refused", outcome: "failed", status: records.AgentExited, script: refusedTaskScript},
	} {
		t.Run(test.name, func(t *testing.T) {
			var other string
			home, _, reader, envelope := spawnKnownTask(t, "claude", func(home string) { other = writeOtherAttemptFile(t, home) }, test.script)
			if envelope.Outcome != test.outcome {
				t.Fatalf("spawn = %#v; want %s", envelope, test.outcome)
			}
			read := reader.only(t)
			checkStartFile(t, read, home, wantKnownInstructions("builder"))
			if _, err := os.Lstat(read.path); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("own instruction file: %v; want it removed", err)
			}
			if data, err := os.ReadFile(other); err != nil || string(data) != "other attempt\n" {
				t.Fatalf("other attempt file = %q, %v; want it unchanged", data, err)
			}
			if keptEffect(envelope, home, read.path) {
				t.Fatalf("effects = %#v; want no kept instruction file", envelope.Effects)
			}
			if stored := readAgent(t, home, "builder-1"); stored.Status != test.status {
				t.Fatalf("record status = %q; want %q", stored.Status, test.status)
			}
		})
	}
}

// TestClaudeInstructionFileKeptForUncertainStart covers a start with no known
// outcome. The harness can still read the file, so hive keeps it, names it in
// the effects, and leaves its removal to the owner.
func TestClaudeInstructionFileKeptForUncertainStart(t *testing.T) {
	for _, test := range []struct {
		name   string
		script taskScript
	}{
		{name: "start reply lost", script: uncertainTaskScript},
		{name: "blocked", script: blockedTaskScript},
	} {
		t.Run(test.name, func(t *testing.T) {
			home, _, reader, envelope := spawnKnownTask(t, "claude", nil, test.script)
			if envelope.Outcome != "partial" {
				t.Fatalf("spawn = %#v; want partial", envelope)
			}
			read := reader.only(t)
			checkStartFile(t, read, home, wantKnownInstructions("builder"))
			data, err := os.ReadFile(read.path)
			if err != nil || !bytes.Equal(data, wantKnownInstructions("builder")) {
				t.Fatalf("kept instruction file = %d bytes, %v; want the exact instructions", len(data), err)
			}
			if !keptEffect(envelope, home, read.path) {
				t.Fatalf("effects = %#v; want instructions %s kept", envelope.Effects, read.path)
			}
		})
	}
}

// TestClaudeInstructionFileWriteFailureStartsNothing makes .hive/tmp a
// regular file, so hive cannot write the instruction file.
func TestClaudeInstructionFileWriteFailureStartsNothing(t *testing.T) {
	block := func(home string) {
		if err := os.WriteFile(filepath.Join(home, ".hive", "tmp"), []byte("not a folder\n"), 0o644); err != nil {
			t.Fatalf("write .hive/tmp file: %v", err)
		}
	}
	home, fake, _, envelope := spawnKnownTask(t, "claude", block, func(fake *fakeHerdr, _ sleepProcess, _ string, worktree string) {
		fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
		fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", workspaceRoot(worktree), worktree))
	})
	if envelope.Outcome != "failed" || envelope.Phase != "instructions" {
		t.Fatalf("spawn = %#v; want failed instructions", envelope)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 0 {
		t.Fatalf("agent.start requests = %d; want 0", len(starts))
	}
	if data, err := os.ReadFile(filepath.Join(home, ".hive", "tmp")); err != nil || string(data) != "not a folder\n" {
		t.Fatalf(".hive/tmp = %q, %v; want it unchanged", data, err)
	}
}

// TestClaudeInstructionFileCancelledSpawnLeavesNoFile cancels the spawn
// before the start request.
func TestClaudeInstructionFileCancelledSpawnLeavesNoFile(t *testing.T) {
	home := testHome(t)
	writeKnownProfile(t, home, "builder")
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	fake := newFakeHerdr(t)
	ctx, cancel := context.WithCancel(context.Background())
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method == "pane.get" {
			cancel()
		}
	})
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", workspaceRoot(worktree), worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	envelope := Spawn(ctx, home, SpawnOptions{
		Profile: "builder", ProfileSet: true, Task: "1-worker-demo",
		Harness: "claude", Model: "m", Name: "builder-1", NameSet: true,
	})
	if envelope.Outcome == "ok" {
		t.Fatalf("spawn = %#v; want no ok outcome after cancellation", envelope)
	}
	if starts := fake.recorded("agent.start"); len(starts) != 0 {
		t.Fatalf("agent.start requests = %d; want 0", len(starts))
	}
	if names := tmpEntries(t, home); len(names) != 0 {
		t.Fatalf(".hive/tmp holds %q after a cancelled spawn; want nothing", names)
	}
}

// TestClaudeInstructionFileNotUsedByOtherHarness keeps the first-prompt
// delivery of each other harness: no file and no file flag.
func TestClaudeInstructionFileNotUsedByOtherHarness(t *testing.T) {
	for _, harnessName := range []string{"pi", "codex", "opencode", "cursor"} {
		t.Run(harnessName, func(t *testing.T) {
			home, fake, reader, envelope := spawnKnownTask(t, harnessName, nil, readyTaskScript)
			if envelope.Outcome != "ok" {
				t.Fatalf("spawn = %#v; want ok", envelope)
			}
			read := reader.only(t)
			if read.path != "" {
				t.Fatalf("agent.start args = %q; want no instruction file", read.args)
			}
			for _, arg := range read.args {
				if arg == claudeFileFlag || strings.ContainsAny(arg, "\r\n") {
					t.Fatalf("agent.start args = %q; want no instruction file and no line break", read.args)
				}
			}
			want := "⬢ " + string(wantKnownInstructions("builder")) + "\n\nYou are builder-1."
			if prompt := promptText(t, fake); !strings.HasPrefix(prompt, want) {
				t.Fatal("first prompt must start with the marker, the instructions, and the facts")
			}
			if names := tmpEntries(t, home); len(names) != 0 {
				t.Fatalf(".hive/tmp holds %q; want nothing", names)
			}
		})
	}
}
