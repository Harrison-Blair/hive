package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
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
			systemIndex := -1
			for index, arg := range start.Args {
				if arg == "--append-system-prompt" {
					systemIndex = index
					break
				}
			}
			contextText := prompt.Text
			if profile.SystemPromptArg() != "" {
				if systemIndex < 0 || systemIndex+1 >= len(start.Args) {
					t.Fatalf("agent.start args = %#v; want --append-system-prompt TEXT", start.Args)
				}
				contextText = start.Args[systemIndex+1] + "\n\n" + prompt.Text
			} else if systemIndex >= 0 {
				t.Fatalf("agent.start args = %#v; want the text in the first prompt", start.Args)
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
