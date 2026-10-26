package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func startCoordinatorWith(ctx context.Context, home, paneID, harnessKind, model string, logger *slog.Logger) error {
	launch, err := ClaimCoordinatorStart(context.Background(), home, logger)
	if err != nil {
		return err
	}
	return launch.Start(ctx, paneID, harnessKind, model)
}

func TestStartCoordinatorSendsStartupPromptAfterHiveRegisters(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", home)
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "coordinator")+`,"argv":["claude","--dangerously-skip-permissions","--model","m"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "coordinator")+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(harness.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "working", "coordinator")+`}`)

	if err := startCoordinatorWith(context.Background(), home, "w1:p1", "claude", "m", nil); err != nil {
		t.Fatalf("start coordinator: %v", err)
	}
	startRequests := fake.recorded("agent.start")
	if len(startRequests) != 1 {
		t.Fatalf("agent.start requests = %#v; want one claude start", startRequests)
	}
	if got, want := startRequests[0].Params, `{"name":"coordinator","kind":"claude","pane_id":"w1:p1","args":["--dangerously-skip-permissions","--model","m"],"timeout_ms":120000}`; got != want {
		t.Fatalf("agent.start params = %s; want %s", got, want)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
	}
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(requests[0].Params), &params); err != nil {
		t.Fatalf("decode prompt params: %v", err)
	}
	for _, want := range []string{"coordinator", "profiles/coordinator.md", "startup"} {
		if !strings.Contains(params.Text, want) {
			t.Errorf("first prompt %q does not contain %q", params.Text, want)
		}
	}
	agents := List(home, ListOptions{Profile: records.ProfileCoordinator, All: true})
	if agents.Outcome != "ok" || len(agents.Result.([]records.Agent)) != 1 || agents.Result.([]records.Agent)[0].Status != records.AgentLive {
		t.Fatalf("coordinator records after start = %#v; want the live record that hive wrote", agents)
	}
}

func TestStartCoordinatorWaitsForHerdrReadiness(t *testing.T) {
	home := testHome(t)
	harness := startSleep(t)
	fake := newFakeHerdr(t)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", home)
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "coordinator")+`,"argv":["claude","--model","m"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", true, false)+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", true, true)+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", false, true)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(harness.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "working", "coordinator")+`}`)

	if err := startCoordinatorWith(context.Background(), home, "w1:p1", "claude", "m", nil); err != nil {
		t.Fatalf("start coordinator: %v", err)
	}
	methods := fake.requestedMethods()
	if got := len(fake.recorded("agent.get")); got != 2 {
		t.Fatalf("agent.get requests = %d; want pending and ready checks", got)
	}
	promptIndex := indexOf(methods, "agent.prompt")
	readyChecksBeforePrompt := 0
	for index, method := range methods {
		if index >= promptIndex {
			break
		}
		if method == "agent.get" {
			readyChecksBeforePrompt++
		}
	}
	if readyChecksBeforePrompt != 2 {
		t.Fatalf("methods = %v; want both readiness checks before the first prompt", methods)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt requests = %d; want one first prompt", got)
	}
	if indexOf(methods, "pane.wait_for_output") < 0 || indexOf(methods, "pane.wait_for_output") > indexOf(methods, "agent.start") {
		t.Fatalf("methods = %v; want shell-prompt wait before agent.start", methods)
	}
}

func TestStartCoordinatorReadinessCancellationDoesNotPromptAgain(t *testing.T) {
	home := testHome(t)
	fake := newFakeHerdr(t)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", home)
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "coordinator")+`,"argv":["claude","--model","m"]}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", true, false)+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", true, false)+`}`)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method == "agent.get" {
			cancel()
		}
	})
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := startCoordinatorWith(ctx, home, "w1:p1", "claude", "m", logger)
	if err == nil || !strings.Contains(err.Error(), "wait for coordinator harness readiness") {
		t.Fatalf("start coordinator error = %v; want readiness step failure", err)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("agent.prompt requests = %d; want no prompt before readiness", got)
	}
	if !strings.Contains(logs.String(), `"msg":"coordinator harness readiness wait"`) || !strings.Contains(logs.String(), `"duration_ms"`) {
		t.Fatalf("logs = %s; want a readiness wait event with duration", logs.String())
	}
}

func TestWaitForCoordinatorReadyTimesOut(t *testing.T) {
	fake := newFakeHerdr(t)
	fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", true, false)+`}`)

	started := time.Now()
	_, err := waitForCoordinatorReady(context.Background(), "w1:p1", 10*time.Millisecond, nil)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("waitForCoordinatorReady() error = %v; want readiness timeout", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("readiness wait took %s; want a bounded timeout", elapsed)
	}
	if got := len(fake.recorded("agent.get")); got != 1 {
		t.Fatalf("agent.get requests = %d; want one pending readiness check", got)
	}
}

func coordinatorReadinessJSON(paneID, terminalID, workspaceID, tabID, status string, launchPending, interactiveReady bool) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"coordinator","launch_pending":` + jsonBool(launchPending) + `,"interactive_ready":` + jsonBool(interactiveReady) + `}`
}

func indexOf(values []string, want string) int {
	for index, value := range values {
		if value == want {
			return index
		}
	}
	return -1
}

func (f *fakeHerdr) requestedMethods() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	methods := make([]string, len(f.requests))
	for index, request := range f.requests {
		methods[index] = request.Method
	}
	return methods
}

func jsonBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// TestStartCoordinatorDeliversProfileAndGuide proves AC2 and AC3 for bare
// hive: each harness gets the coordinator profile and the guide, with both
// approved coordinator instructions and the fixed manager inbox key.
func TestStartCoordinatorDeliversProfileAndGuide(t *testing.T) {
	for _, kind := range []string{"claude", "pi", "codex", "opencode", "cursor"} {
		t.Run(kind, func(t *testing.T) {
			home := testHome(t)
			writeProfileHome(t, home)
			profile, err := harness.Lookup(kind)
			if err != nil {
				t.Fatalf("Lookup(%s): %v", kind, err)
			}
			processName := profile.ProcessNames()[0]
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
			reader := readStartFile(t, fake)
			scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", home)
			fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
			fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "coordinator")+`,"argv":["`+kind+`"]}`)
			fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "coordinator")+`}`)
			fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", false, true)+`}`)
			fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, processName, processName))+`}`)
			fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "working", "coordinator")+`}`)

			if err := startCoordinatorWith(context.Background(), home, "w1:p1", kind, "m", nil); err != nil {
				t.Fatalf("start coordinator: %v", err)
			}
			starts := fake.recorded("agent.start")
			if len(starts) != 1 {
				t.Fatalf("agent.start requests = %d; want 1", len(starts))
			}
			var start struct {
				Args []string `json:"args"`
			}
			if err := json.Unmarshal([]byte(starts[0].Params), &start); err != nil {
				t.Fatalf("decode agent.start params: %v", err)
			}
			prompts := fake.recorded("agent.prompt")
			if len(prompts) != 1 {
				t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
			}
			var prompt struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal([]byte(prompts[0].Params), &prompt); err != nil {
				t.Fatalf("decode agent.prompt params: %v", err)
			}
			if kind == "pi" {
				for index, arg := range start.Args {
					if arg == "--append-system-prompt" || strings.ContainsAny(arg, "\r\n") {
						t.Errorf("Pi coordinator start argument %d carries unsafe instructions", index)
					}
				}
				want := strings.TrimSpace(readProfile(t, "coordinator")) + "\n\n" + strings.TrimSpace(guide.Text()) +
					"\n\nRead `profiles/coordinator.md` and run `hive guide` after a context reset.\n\n" + CoordinatorPrompt
				if prompt.Text != want {
					t.Error("Pi coordinator first prompt must contain the profile, guide, reset reminder, and startup prompt in order")
				}
			}
			contextText := prompt.Text
			if profile.SystemPromptArg() != "" {
				read := reader.only(t)
				want := strings.TrimSpace(readProfile(t, "coordinator")) + "\n\n" + strings.TrimSpace(guide.Text()) +
					"\n\nRead `profiles/coordinator.md` and run `hive guide` after a context reset."
				checkStartFile(t, read, home, []byte(want))
				if prompt.Text != CoordinatorPrompt {
					t.Errorf("Claude coordinator first prompt = %q; want only the startup prompt", prompt.Text)
				}
				if _, err := os.Lstat(read.path); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("coordinator instruction file after a ready start: %v; want it removed", err)
				}
				contextText = string(read.data) + "\n\n" + prompt.Text
			} else {
				for _, arg := range start.Args {
					if arg == "--append-system-prompt" || arg == claudeFileFlag {
						t.Fatalf("agent.start args = %#v; want the text in the first prompt", start.Args)
					}
				}
			}
			profileIndex := strings.Index(contextText, readProfile(t, "coordinator"))
			guideIndex := strings.Index(contextText, guide.Text())
			startupIndex := strings.Index(contextText, "profiles/coordinator.md")
			if profileIndex < 0 || guideIndex < 0 || startupIndex < 0 {
				t.Fatalf("starting context misses the profile, the guide, or the startup prompt")
			}
			if profileIndex >= guideIndex || guideIndex >= startupIndex {
				t.Fatalf("starting context order is profile %d, guide %d, startup %d", profileIndex, guideIndex, startupIndex)
			}
			for _, want := range []string{
				"Hive-started exchanges get replies through Hive. Use the native harness only when the user starts the exchange there.",
				"Address project messages to the inbox address `project:ID`. For project hive, this is `project:hive`.",
				"The address goes to the live owner of the project, or to the coordinator when the project has no live owner.",
			} {
				if !strings.Contains(contextText, want) {
					t.Errorf("starting context does not contain %q", want)
				}
			}
			if strings.Contains(contextText, "manager-hive-g1") {
				t.Error("starting context names the generation-specific record key")
			}
		})
	}
}

// startKnownCoordinator starts a coordinator with knownProfile. script gives
// the herdr replies after the shell checks.
func startKnownCoordinator(t *testing.T, setup func(home string), script func(fake *fakeHerdr, sleep sleepProcess)) (string, *fakeHerdr, *startReader, error) {
	t.Helper()
	home := testHome(t)
	writeKnownProfile(t, home, "coordinator")
	if setup != nil {
		setup(home)
	}
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	reader := readStartFile(t, fake)
	scriptReadyShell(fake, "w1:p1", "term_a", "w1", "w1:t1", home)
	fake.result("pane.wait_for_output", `{"type":"output_matched"}`)
	script(fake, sleep)
	err := startCoordinatorWith(context.Background(), home, "w1:p1", "claude", "m", nil)
	return home, fake, reader, err
}

func coordinatorStarted(fake *fakeHerdr) {
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "unknown", "coordinator")+`,"argv":["claude"]}`)
}

func TestClaudeInstructionFileCoordinatorReadyRemovesOwnFile(t *testing.T) {
	var other string
	home, fake, reader, err := startKnownCoordinator(t, func(home string) { other = writeOtherAttemptFile(t, home) }, func(fake *fakeHerdr, sleep sleepProcess) {
		coordinatorStarted(fake)
		fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "coordinator")+`}`)
		fake.result("agent.get", `{"type":"agent_info","agent":`+coordinatorReadinessJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", false, true)+`}`)
		fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
		fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "working", "coordinator")+`}`)
	})
	if err != nil {
		t.Fatalf("start coordinator: %v", err)
	}
	read := reader.only(t)
	checkStartFile(t, read, home, wantKnownInstructions("coordinator"))
	if want := []string{"--dangerously-skip-permissions", "--model", "m", claudeFileFlag, read.path}; !reflect.DeepEqual(read.args, want) {
		t.Fatalf("agent.start args = %q; want %q", read.args, want)
	}
	if prompts := fake.recorded("agent.prompt"); len(prompts) != 1 || strings.Contains(prompts[0].Params, "T186-CANARY") {
		t.Fatalf("agent.prompt requests = %#v; want one startup prompt with no instructions", prompts)
	}
	if _, err := os.Lstat(read.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("own instruction file: %v; want it removed", err)
	}
	if data, err := os.ReadFile(other); err != nil || string(data) != "other attempt\n" {
		t.Fatalf("other attempt file = %q, %v; want it unchanged", data, err)
	}
}

func TestClaudeInstructionFileCoordinatorRefusalRemovesFile(t *testing.T) {
	home, _, reader, err := startKnownCoordinator(t, nil, func(fake *fakeHerdr, _ sleepProcess) {
		fake.failure("agent.start", "invalid_agent_argument", "refused")
	})
	if err == nil {
		t.Fatal("start coordinator succeeded; want the refusal")
	}
	read := reader.only(t)
	checkStartFile(t, read, home, wantKnownInstructions("coordinator"))
	if _, statErr := os.Lstat(read.path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("instruction file after a refused start: %v; want it removed", statErr)
	}
	if strings.Contains(err.Error(), read.path) {
		t.Fatalf("error = %v; want no kept file", err)
	}
}

// TestClaudeInstructionFileCoordinatorUncertainKeepsFile covers a lost start
// reply and a harness that does not become ready. Hive keeps the file and
// names it in the error.
func TestClaudeInstructionFileCoordinatorUncertainKeepsFile(t *testing.T) {
	for _, test := range []struct {
		name   string
		script func(fake *fakeHerdr, sleep sleepProcess)
	}{
		{name: "start reply lost", script: func(fake *fakeHerdr, _ sleepProcess) { fake.drop("agent.start") }},
		{name: "not ready", script: func(fake *fakeHerdr, _ sleepProcess) {
			coordinatorStarted(fake)
			fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "blocked", "coordinator")+`}`)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			home, _, reader, err := startKnownCoordinator(t, nil, test.script)
			if err == nil {
				t.Fatal("start coordinator succeeded; want an uncertain start")
			}
			read := reader.only(t)
			checkStartFile(t, read, home, wantKnownInstructions("coordinator"))
			data, readErr := os.ReadFile(read.path)
			if readErr != nil || !bytes.Equal(data, wantKnownInstructions("coordinator")) {
				t.Fatalf("kept instruction file = %d bytes, %v; want the exact instructions", len(data), readErr)
			}
			if !strings.Contains(err.Error(), "instruction file kept") || !strings.Contains(err.Error(), read.path) {
				t.Fatalf("error = %v; want it to name the kept file %s", err, read.path)
			}
		})
	}
}

func TestClaudeInstructionFileCoordinatorWriteFailureStartsNothing(t *testing.T) {
	block := func(home string) {
		if err := os.WriteFile(filepath.Join(home, ".hive", "tmp"), []byte("not a folder\n"), 0o644); err != nil {
			t.Fatalf("write .hive/tmp file: %v", err)
		}
	}
	home := testHome(t)
	writeKnownProfile(t, home, "coordinator")
	block(home)
	fake := newFakeHerdr(t)
	if err := startCoordinatorWith(context.Background(), home, "w1:p1", "claude", "m", nil); err == nil {
		t.Fatal("start coordinator succeeded; want the file write failure")
	}
	for _, method := range []string{"agent.start", "pane.wait_for_shell_prompt", "agent.prompt"} {
		if requests := fake.recorded(method); len(requests) != 0 {
			t.Fatalf("%s requests = %d; want 0", method, len(requests))
		}
	}
	if data, err := os.ReadFile(filepath.Join(home, ".hive", "tmp")); err != nil || string(data) != "not a folder\n" {
		t.Fatalf(".hive/tmp = %q, %v; want it unchanged", data, err)
	}
}
