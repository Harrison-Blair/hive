package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
	}
	var params herdr.AgentPromptParams
	if err := json.Unmarshal([]byte(requests[0].Params), &params); err != nil {
		t.Fatalf("decode prompt params: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(home, "projects", "demo", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if params.Target != "wBD:p1" || params.Text != string(content) {
		t.Fatalf("prompt target/text = %q/%q; want wBD:p1 and generated instructions", params.Target, params.Text)
	}
	if !strings.HasPrefix(params.Text, "You are the manager of project demo.") {
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
			"After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`.",
			"After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.",
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
