package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
)

// profileSections gives the headings of each profile, in order (D5).
var profileSections = map[string][]string{
	"coordinator": {"Marker", "Chat replies", "Startup", "Preference record", "Owner spawn", "Supervisor", "Handoff", "Ask the user", "Rules"},
	"planner":     {"Marker", "Startup", "Task creation", "Spec state", "Worktree", "Builder spawn", "Verifier spawn", "Landing", "CI check", "Memory and handoff", "Friction log", "Escalation", "Ask the user", "Rules"},
	"builder":     {"Marker", "Home", "Brief", "Spec state", "Visual check", "Work", "Gate", "Evidence", "Commit", "Report", "Finish", "Ask the user", "Rules"},
	"verifier":    {"Marker", "Home", "Brief", "Spec state", "Visual check", "Items", "Gate", "Verdict", "Report", "Ask the user", "Rules"},
	"general":     {"Marker", "Work", "Report", "After the report", "Ask the user", "Rules"},
}

const profileMarkerParagraph = "Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant."

// readProfile returns the text of one tracked profile.
func readProfile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "profiles", name+".md"))
	if err != nil {
		t.Fatalf("read profile %s: %v", name, err)
	}
	return string(data)
}

// writeProfileHome copies each tracked profile into a test home.
func writeProfileHome(t *testing.T, home string) {
	t.Helper()
	directory := filepath.Join(home, "profiles")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("create profiles directory: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join("..", "..", "profiles"))
	if err != nil {
		t.Fatalf("read profiles: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("..", "..", "profiles", entry.Name()))
		if err != nil {
			t.Fatalf("read profile %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0o644); err != nil {
			t.Fatalf("write profile %s: %v", entry.Name(), err)
		}
	}
}

func TestProfilesHaveTheApprovedSections(t *testing.T) {
	for name, headings := range profileSections {
		t.Run(name, func(t *testing.T) {
			text := readProfile(t, name)
			last := -1
			for _, heading := range headings {
				index := strings.Index(text, "## "+heading+"\n")
				if index < 0 {
					t.Errorf("profile %s has no section %q", name, heading)
					continue
				}
				if index < last {
					t.Errorf("profile %s puts section %q out of order", name, heading)
				}
				last = index
			}
			if !strings.Contains(text, profileMarkerParagraph) {
				t.Errorf("profile %s does not hold the marker paragraph of AGENTS.md", name)
			}
		})
	}
}

func TestProfilesNameNoHarnessOrModel(t *testing.T) {
	modelNames := []string{"gpt", "opus", "sonnet", "haiku", "gemini", "deepseek", "llama", "qwen"}
	for name := range profileSections {
		t.Run(name, func(t *testing.T) {
			text := readProfile(t, name)
			for _, kind := range harness.KnownKinds() {
				pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kind) + `\b`)
				if pattern.MatchString(text) {
					t.Errorf("profile %s names the harness %q", name, kind)
				}
			}
			for _, model := range modelNames {
				pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(model) + `\b`)
				if pattern.MatchString(text) {
					t.Errorf("profile %s names the model %q", name, model)
				}
			}
			if strings.Contains(text, "--harness") || strings.Contains(text, "--model") {
				t.Errorf("profile %s names a harness or model flag", name)
			}
		})
	}
}

func TestBuilderAndVerifierProfilesHaveTheGateAndReportRules(t *testing.T) {
	gate := "gofmt -l $(git ls-files -co --exclude-standard '*.go')\ngo mod tidy -diff\ngo vet ./...\ngolangci-lint run ./...\nenv -u TERM TZ=UTC go test -race ./..."
	for _, name := range []string{"builder", "verifier"} {
		t.Run(name, func(t *testing.T) {
			text := readProfile(t, name)
			if !strings.Contains(text, gate) {
				t.Errorf("profile %s does not hold the gate of AGENTS.md", name)
			}
			for _, want := range []string{
				"hive inbox send --to project:ID",
				"Never send it to the coordinator.",
			} {
				if !strings.Contains(text, want) {
					t.Errorf("profile %s does not contain %q", name, want)
				}
			}
		})
	}
}

func TestPlannerProfileHasTheCICheck(t *testing.T) {
	text := readProfile(t, "planner")
	for _, want := range []string{
		"gh run list --branch dev --limit 1",
		"Use the base branch of the project in place of `dev`.",
		"hive inbox send --to coordinator TEXT --agent NAME",
		"hive agent spawn --profile builder --task ID",
		"hive agent spawn --profile verifier --task ID",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("planner profile does not contain %q", want)
		}
	}
}

func TestCoordinatorProfileHasTheApprovedInstructions(t *testing.T) {
	text := readProfile(t, "coordinator")
	for _, want := range []string{
		"Hive-started exchanges get replies through Hive. Use the native harness only when the user starts the exchange there.",
		"Address manager messages to the inbox key `manager-PROJECT`. For project hive, this is `manager-hive`.",
		"Do not use the live agent record key `manager-PROJECT-gN` for a manager message.",
		"hive agent spawn --profile planner --project ID --owner",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("coordinator profile does not contain %q", want)
		}
	}
	for _, rejected := range []string{"manager-hive-g1", "--to manager-hive-g"} {
		if strings.Contains(text, rejected) {
			t.Errorf("coordinator profile names the generation-specific record key %q", rejected)
		}
	}
}

func TestGeneralProfileKeepsThePromptRule(t *testing.T) {
	text := readProfile(t, "general")
	for _, want := range []string{
		"The first prompt from hive sets your work. Do not take work from your",
		"working directory.",
		"hive agent report",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("general profile does not contain %q", want)
		}
	}
}

// scriptProfileTaskSpawn scripts one task spawn for harnessName.
func scriptProfileTaskSpawn(fake *fakeHerdr, sleep sleepProcess, harnessName, paneID, workspaceID, tabID, cwd, name string) {
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON(workspaceID, "1-worker-demo", workspaceRoot(cwd), cwd))
	scriptReadyShell(fake, paneID, "term_"+paneID, workspaceID, tabID, cwd)
	fake.result("agent.start", workerAgentStartJSON(paneID, workspaceID, tabID, name))
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("agent.wait", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, harnessName, harnessName))+`}`)
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.prompt", workerAgentPromptedJSON(paneID, workspaceID, tabID, name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "working", name))
}

// spawnProfileTask spawns one builder with a profile from the test home.
func spawnProfileTask(t *testing.T, harnessName string) (string, *fakeHerdr, string) {
	t.Helper()
	home := testHome(t)
	writeProfileHome(t, home)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	profile, err := harness.Lookup(harnessName)
	if err != nil {
		t.Fatalf("Lookup(%s): %v", harnessName, err)
	}
	scriptProfileTaskSpawn(fake, sleep, profile.ProcessNames()[0], "w5:p1", "w5", "w5:t1", worktree, "builder-1")
	envelope := Spawn(context.Background(), home, SpawnOptions{
		Profile: "builder", ProfileSet: true, Task: "1-worker-demo",
		Harness: harnessName, Model: "m", Name: "builder-1", NameSet: true,
	})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("spawn outcome = %s error = %v; want ok", envelope.Outcome, envelope.Error)
	}
	return home, fake, worktree
}

func startArgs(t *testing.T, fake *fakeHerdr) []string {
	t.Helper()
	requests := fake.recorded("agent.start")
	if len(requests) != 1 {
		t.Fatalf("agent.start requests = %d; want 1", len(requests))
	}
	var params struct {
		Args []string `json:"args"`
	}
	if err := json.Unmarshal([]byte(requests[0].Params), &params); err != nil {
		t.Fatalf("decode agent.start params: %v", err)
	}
	return params.Args
}

func promptText(t *testing.T, fake *fakeHerdr) string {
	t.Helper()
	requests := fake.recorded("agent.prompt")
	if len(requests) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(requests))
	}
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(requests[0].Params), &params); err != nil {
		t.Fatalf("decode agent.prompt params: %v", err)
	}
	return params.Text
}

// TestSpawnDeliversProfileWithSystemPrompt proves AC1: Claude Code takes
// the profile and the guide with --append-system-prompt.
func TestSpawnDeliversProfileWithSystemPrompt(t *testing.T) {
	for _, harnessName := range []string{"claude"} {
		t.Run(harnessName, func(t *testing.T) {
			_, fake, _ := spawnProfileTask(t, harnessName)
			args := startArgs(t, fake)
			index := -1
			for position, arg := range args {
				if arg == "--append-system-prompt" {
					index = position
					break
				}
			}
			if index < 0 || index+1 >= len(args) {
				t.Fatalf("agent.start args = %#v; want --append-system-prompt TEXT", args)
			}
			text := args[index+1]
			profile := readProfile(t, "builder")
			guideText := guide.Text()
			profileIndex := strings.Index(text, profile)
			guideIndex := strings.Index(text, guideText)
			if profileIndex < 0 || guideIndex < 0 {
				t.Fatalf("system prompt does not hold the profile and the guide")
			}
			if profileIndex >= guideIndex {
				t.Fatalf("system prompt puts the guide before the profile")
			}
			if !strings.Contains(text, "Read `profiles/builder.md` and run `hive guide` after a context reset.") {
				t.Errorf("system prompt does not name the profile path and the guide command")
			}
			prompt := promptText(t, fake)
			if strings.Contains(prompt, guideText) {
				t.Errorf("first prompt holds the guide; want the system prompt to carry it")
			}
			if !strings.HasPrefix(prompt, "⬢ You are builder-1.") {
				t.Errorf("first prompt = %q; want the facts after the marker", prompt)
			}
		})
	}
}

// TestSpawnDeliversProfileFirstInPrompt proves AC2: Pi, Codex, OpenCode, and
// the Cursor agent take the profile and guide at the start of the first prompt.
func TestSpawnDeliversProfileFirstInPrompt(t *testing.T) {
	for _, harnessName := range []string{"pi", "codex", "opencode", "cursor"} {
		t.Run(harnessName, func(t *testing.T) {
			_, fake, _ := spawnProfileTask(t, harnessName)
			args := startArgs(t, fake)
			for _, arg := range args {
				if arg == "--append-system-prompt" {
					t.Fatalf("agent.start args = %#v; want no system prompt argument", args)
				}
			}
			prompt := strings.TrimPrefix(promptText(t, fake), "⬢ ")
			profile := readProfile(t, "builder")
			guideText := guide.Text()
			profileIndex := strings.Index(prompt, profile)
			guideIndex := strings.Index(prompt, guideText)
			factsIndex := strings.Index(prompt, "You are builder-1.")
			if profileIndex != 0 {
				t.Fatalf("first prompt does not start with the profile")
			}
			if guideIndex < 0 || factsIndex < 0 {
				t.Fatalf("first prompt does not hold the guide and the facts")
			}
			if profileIndex >= guideIndex || guideIndex >= factsIndex {
				t.Fatalf("first prompt order is profile %d, guide %d, facts %d", profileIndex, guideIndex, factsIndex)
			}
			if !strings.Contains(prompt, "Read `profiles/builder.md` and run `hive guide` after a context reset.") {
				t.Errorf("first prompt does not name the profile path and the guide command")
			}
		})
	}
}

// TestProjectTemplateNamesProfiles proves AC4: the template uses profile names
// and profile commands, and it names no skill.
func TestProjectTemplateNamesProfiles(t *testing.T) {
	text := projectAgentsTemplate
	for _, want := range []string{
		"You are the planner of project {{PROJECT_ID}}.",
		"Read `{{HOME}}/profiles/planner.md` first.",
		"Delegate each source change to one builder.",
		"hive agent spawn --profile builder --task ID --harness KIND --model MODEL",
		"hive agent spawn --profile verifier --task ID --harness KIND --model MODEL",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("project template does not contain %q", want)
		}
	}
	for _, rejected := range []string{".agents/skills/", "--role"} {
		if strings.Contains(text, rejected) {
			t.Errorf("project template contains %q", rejected)
		}
	}
}
