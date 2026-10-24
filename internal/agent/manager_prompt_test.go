package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

func scriptManagerPromptSpawn(t *testing.T, home string, fake *fakeHerdr, sleep sleepProcess, replaceInstructions bool) {
	t.Helper()
	writeManagerProject(t, home, nil)
	if replaceInstructions {
		original := projectAgentsTemplate
		projectAgentsTemplate = "You are the manager of project {{PROJECT_ID}}.\n\n" + strings.Repeat("instruction ", 500)
		t.Cleanup(func() { projectAgentsTemplate = original })
	}
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	scriptManagerSpawnTail(t, fake, sleep, "wBD:p1", "wBD", "wBD:t1")
}

func TestManagerPromptSendsGeneratedInstructionsAfterReady(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptManagerPromptSpawn(t, home, fake, sleep, false)
	fake.mu.Lock()
	fake.replies["agent.prompt"] = nil
	fake.mu.Unlock()
	fake.failure("agent.prompt", herdr.CodeAgentPromptStalled, "first manager prompt stalled")
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+managerAgentInfoJSON("wBD:p1", "wBD", "wBD:t1", "working", "manager-demo-g1")+`}`)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("outcome=%s error=%v; want ok", envelope.Outcome, envelope.Error)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 2 {
		t.Fatalf("agent.prompt requests = %d; want first prompt and one retry", len(requests))
	}
	requireFirstPromptWait(t, requests)
	var params herdr.AgentPromptParams
	if err := json.Unmarshal([]byte(requests[0].Params), &params); err != nil {
		t.Fatalf("decode prompt params: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(home, "projects", "demo", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(params.Text, "⬢ ") {
		t.Fatalf("manager prompt = %q; want the hive marker prefix", params.Text)
	}
	if params.Target != "wBD:p1" || params.Text != "⬢ "+string(content) {
		t.Fatalf("prompt target/text = %q/%q; want wBD:p1 and marked generated instructions", params.Target, params.Text)
	}
	if params.Wait == nil {
		t.Fatal("manager prompt has no wait")
	}
	if got, want := params.Wait.Until, []string{"working", "blocked"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("manager prompt wait.until = %v; want %v", got, want)
	}
	if params.Wait.TimeoutMS != 10000 {
		t.Fatalf("manager prompt wait.timeout_ms = %d; want 10000", params.Wait.TimeoutMS)
	}
	if !strings.HasPrefix(params.Text, "⬢ You are the manager of project demo.") {
		t.Fatalf("manager instructions do not start with the role: %q", params.Text[:min(len(params.Text), 80)])
	}
	frictionPath := filepath.Join(home, "reference", "dogfood", "friction") + string(filepath.Separator)
	resolvedPath := filepath.Join(frictionPath, "entries", "resolved") + string(filepath.Separator)
	normalizedInstructions := strings.Join(strings.Fields(params.Text), " ")
	if !strings.Contains(normalizedInstructions, "You never change a source file except `"+frictionPath+"` as described below.") {
		t.Fatal("manager instructions do not name the home friction folder exception")
	}
	if !strings.Contains(normalizedInstructions, "You can edit `"+frictionPath+"` directly.") {
		t.Fatal("manager instructions do not allow direct friction folder edits")
	}
	if !strings.Contains(normalizedInstructions, "To resolve an entry, set its status, move its file to `"+resolvedPath+"`, and update its row.") {
		t.Fatal("manager instructions do not give the friction entry resolve steps")
	}
	fake.mu.Lock()
	var waitIndex, promptIndex = -1, -1
	for index, request := range fake.requests {
		if request.Method == "agent.wait" {
			waitIndex = index
		}
		if request.Method == "agent.prompt" {
			promptIndex = index
		}
	}
	fake.mu.Unlock()
	if waitIndex < 0 || promptIndex <= waitIndex {
		t.Fatalf("request order wait=%d prompt=%d; prompt must follow readiness", waitIndex, promptIndex)
	}
}

func TestManagerTemplateAndSkillHaveMarkerAndEscalationRules(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-manager", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	markerRule := "Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant."
	escalationRule := "Before you stop for a user decision, or when a harness blocks an action, send the coordinator one record. The record names the decision or the block. Then end your turn."
	command := "hive inbox send --to coordinator TEXT --agent NAME"
	for name, text := range map[string]string{"template": projectAgentsTemplate, "manager skill": string(skill)} {
		normalized := strings.Join(strings.Fields(text), " ")
		for _, want := range []string{markerRule, escalationRule, command} {
			if !strings.Contains(normalized, want) {
				t.Errorf("%s does not contain %q", name, want)
			}
		}
	}
}

func TestCoordinatorInstructionsRequireChatReplies(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-coordinator", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(strings.Fields(string(skill)), " ")
	for _, want := range []string{
		"Read each `⬢ user:` message and its ID.",
		"Answer each message with `hive chat reply --to ID`.",
		"Give the reply text on standard input.",
		"Send each answer to the user through `hive chat reply`.",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("coordinator skill does not contain %q", want)
		}
	}
}

func TestManagerInstructionsRequireFeatureTaskRules(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-manager", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{"template": projectAgentsTemplate, "skill": string(skill)} {
		normalized := strings.Join(strings.Fields(text), " ")
		command := "hive task create --project ID --feature F --title TITLE --slug SLUG --depends-on IDS --brief FILE"
		if name == "template" {
			command = "hive task create --project {{PROJECT_ID}} --feature F --title TITLE --slug SLUG --depends-on IDS --brief FILE"
		}
		for _, want := range []string{
			command,
			"Use `--spec` to mark the spec task. Land its `.hive/spec/F/README.md` before you create any other task for the feature.",
		} {
			if !strings.Contains(normalized, want) {
				t.Errorf("%s does not contain %q", name, want)
			}
		}
	}
}

func TestManagerInstructionsRequireRespawnRules(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-manager", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	instructions := map[string]string{
		"template": projectAgentsTemplate,
		"skill":    string(skill),
	}
	for name, text := range instructions {
		normalized := strings.Join(strings.Fields(text), " ")
		for _, want := range []string{
			"Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name.",
			"hive agent spawn --role worker --task ID --harness KIND --model MODEL",
			"hive agent spawn --role verifier --task ID --harness KIND --model MODEL",
			"Use `hive task edit --task ID --brief FILE` to change a brief before the work is complete.",
			"Before you stop for a user decision on a task, set `needs_user` with `hive task edit --task ID --needs-user TEXT`. Clear it when the user gives the decision with `hive task edit --task ID --clear-needs-user`.",
			"After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`.",
			"After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.",
			"If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`.",
		} {
			if !strings.Contains(normalized, want) {
				t.Errorf("%s does not contain %q", name, want)
			}
		}
		if strings.Contains(normalized, "hive task assign --task") {
			t.Errorf("%s has a manual task assign command", name)
		}
	}
}

// AC9: the verifier keeps VERIFIED.md out of each commit. Hive task verify copies it
// from the ignored evidence folder to the home.
func TestVerifierLeavesVerdictUncommitted(t *testing.T) {
	managerSkill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-manager", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	verifierSkill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-verifier", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := "The verifier writes `VERIFIED.md` in the evidence folder and does not commit it. `hive task verify` copies it to the home."
	for name, instructions := range map[string]string{
		"template":       projectAgentsTemplate,
		"manager skill":  string(managerSkill),
		"verifier skill": string(verifierSkill),
	} {
		if !strings.Contains(strings.Join(strings.Fields(instructions), " "), want) {
			t.Errorf("%s does not contain %q", name, want)
		}
	}
}

func TestManagerInstructionsRequireFreshVerifierAfterFail(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join("..", "..", ".agents", "skills", "hive-manager", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := "If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`."
	for name, instructions := range map[string]string{"template": projectAgentsTemplate, "skill": string(skill)} {
		if !strings.Contains(strings.Join(strings.Fields(instructions), " "), want) {
			t.Errorf("%s does not contain %q", name, want)
		}
	}
}

func TestManagerPromptLogCapsLargeBody(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptManagerPromptSpawn(t, home, fake, sleep, true)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
	}
	var prompt herdr.AgentPromptParams
	if err := json.Unmarshal([]byte(requests[0].Params), &prompt); err != nil {
		t.Fatalf("decode prompt params: %v", err)
	}
	if len(prompt.Text) <= 4096 {
		t.Fatalf("prompt bytes = %d; want more than 4096", len(prompt.Text))
	}
	logText := readLog(t, home)
	found := false
	for _, line := range strings.Split(strings.TrimSpace(logText), "\n") {
		var event struct {
			Message   string `json:"msg"`
			Method    string `json:"method"`
			Body      string `json:"body"`
			Truncated bool   `json:"truncated"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("decode log event: %v", err)
		}
		if event.Method != "agent.prompt" || event.Message != "herdr request" {
			continue
		}
		found = true
		if len(event.Body) > 4096 || !event.Truncated {
			t.Fatalf("prompt log body bytes=%d truncated=%t; want <=4096 and true: %s", len(event.Body), event.Truncated, line)
		}
	}
	if !found {
		t.Fatalf("log has no agent.prompt event")
	}
}
