package task_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	hivecmd "github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestFeatureTaskCLIFlow(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "")
	projectPath := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(projectPath, ".hive", "spec", "f"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitTask(t, projectPath, "init", "-b", "main")
	gitTask(t, projectPath, "config", "user.name", "Hive Test")
	gitTask(t, projectPath, "config", "user.email", "hive-test@example.invalid")
	if err := os.WriteFile(filepath.Join(projectPath, ".hive", "spec", "f", "README.md"), []byte("feature spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitTask(t, projectPath, "add", ".hive/spec/f/README.md")
	gitTask(t, projectPath, "commit", "-m", "add feature spec")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("projects", "one", records.Project{ID: "one", Path: projectPath, Base: "main"}); err != nil {
		t.Fatal(err)
	}

	withoutFeature := runHiveFailure(t, "--json", "task", "create", "--project", "one", "--title", "No feature")
	var invalid struct {
		Outcome string `json:"outcome"`
		Phase   string `json:"phase"`
	}
	decodeJSON(t, withoutFeature, &invalid)
	if invalid.Outcome != "failed" || invalid.Phase != "validate" {
		t.Fatalf("create without feature = %#v; want validate failure", invalid)
	}

	missing := runHiveFailure(t, "--json", "task", "create", "--project", "one", "--title", "Missing spec", "--feature", "not-ready")
	var missingResult struct {
		Outcome string `json:"outcome"`
		Phase   string `json:"phase"`
		Error   string `json:"error"`
	}
	decodeJSON(t, missing, &missingResult)
	if missingResult.Outcome != "failed" || missingResult.Phase != "validate" || !strings.Contains(missingResult.Error, ".hive/spec/not-ready/README.md") {
		t.Fatalf("create without base README = %#v; want validate error naming the path", missingResult)
	}
	for _, feature := range []string{"Bad", "a--b", strings.Repeat("a", 41)} {
		invalidFeature := runHiveFailure(t, "--json", "task", "create", "--project", "one", "--title", "Bad feature", "--feature", feature)
		var invalidFeatureResult struct {
			Phase string `json:"phase"`
			Error string `json:"error"`
		}
		decodeJSON(t, invalidFeature, &invalidFeatureResult)
		if invalidFeatureResult.Phase != "validate" || !strings.Contains(invalidFeatureResult.Error, "invalid feature") {
			t.Fatalf("invalid feature create = %#v; want validation error", invalidFeatureResult)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "counters", "task.json")); !os.IsNotExist(err) {
		t.Fatalf("failed creates changed the counter: %v", err)
	}
	for _, path := range []string{
		filepath.Join(home, ".hive", "state", "tasks", "1-missing-spec.json"),
		filepath.Join(home, ".hive", "state", "briefs", "1-missing-spec.md"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("failed create left %s: %v", path, err)
		}
	}

	_ = createFeatureTask(t, "Maximum feature slug", strings.Repeat("a", 40), true)
	spec := createFeatureTask(t, "Spec task", "not-ready", true)
	specShow := runHive(t, "--json", "task", "show", spec)
	var specWarnings struct {
		Result struct {
			Warnings []string `json:"warnings"`
		} `json:"result"`
	}
	decodeJSON(t, specShow, &specWarnings)
	if len(specWarnings.Result.Warnings) != 2 || specWarnings.Result.Warnings[0] != "linked file missing: .hive/spec/not-ready/README.md" || specWarnings.Result.Warnings[1] != "linked file missing: .hive/spec/not-ready/tasks/"+spec+"/brief.md" {
		t.Fatalf("show warnings for spec task = %#v; want missing README and brief", specWarnings.Result.Warnings)
	}
	first := createFeatureTask(t, "First feature task", "f", false)
	second := createFeatureTask(t, "Second feature task", "f", false)
	other := createFeatureTask(t, "Other feature task", "g", true)
	listed := runHive(t, "--json", "task", "list", "--feature", "f")
	var listResult struct {
		Result []records.Task `json:"result"`
	}
	decodeJSON(t, listed, &listResult)
	if len(listResult.Result) != 2 || listResult.Result[0].ID != first || listResult.Result[1].ID != second {
		t.Fatalf("feature list = %#v; want %s and %s (not %s)", listResult.Result, first, second, other)
	}

	shown := runHive(t, "task", "show", second)
	for _, want := range []string{
		"feature: f",
		"brief: .hive/spec/f/tasks/" + second + "/brief.md",
		"evidence: .hive/spec/f/tasks/" + second + "/evidence/",
		"warning: linked file missing: .hive/spec/f/tasks/" + second + "/brief.md",
	} {
		if !strings.Contains(shown, want) {
			t.Errorf("task show output %q; want %q", shown, want)
		}
	}
	jsonShown := runHive(t, "--json", "task", "show", second)
	var showResult struct {
		Result struct {
			Warnings []string `json:"warnings"`
		} `json:"result"`
	}
	decodeJSON(t, jsonShown, &showResult)
	if len(showResult.Result.Warnings) != 1 || showResult.Result.Warnings[0] != "linked file missing: .hive/spec/f/tasks/"+second+"/brief.md" {
		t.Fatalf("show warnings = %#v; want the missing task brief", showResult.Result.Warnings)
	}

	activeID := "92-active-feature"
	activePath := t.TempDir()
	activeTask := featureTaskRecord{
		Task: records.Task{
			ID: activeID, Project: "one", Title: "Active worktree", ControllingRole: "manager-one",
			Status: "created", DependsOn: []string{}, Worktree: &records.Worktree{Path: activePath},
		},
		Feature: "f",
	}
	if err := registry.Write("tasks", activeID, activeTask); err != nil {
		t.Fatal(err)
	}
	briefDir := filepath.Join(home, ".hive", "state", "briefs")
	if err := os.MkdirAll(briefDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(briefDir, activeID+".md"), []byte("active task"), 0o644); err != nil {
		t.Fatal(err)
	}
	activeShow := runHive(t, "--json", "task", "show", activeID)
	var activeWarnings struct {
		Result struct {
			Warnings []string `json:"warnings"`
		} `json:"result"`
	}
	decodeJSON(t, activeShow, &activeWarnings)
	if len(activeWarnings.Result.Warnings) != 2 || activeWarnings.Result.Warnings[0] != "linked file missing: .hive/spec/f/README.md" || activeWarnings.Result.Warnings[1] != "linked file missing: .hive/spec/f/tasks/"+activeID+"/brief.md" {
		t.Fatalf("active worktree warnings = %#v; want both missing worktree files", activeWarnings.Result.Warnings)
	}
	if spec == "" {
		t.Fatal("spec task ID is empty")
	}

	oldID := "90-old-task"
	oldTask := records.Task{
		ID: oldID, Project: "one", Title: "Old record", ControllingRole: "manager-one",
		Status: "verified", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.Write("tasks", oldID, oldTask); err != nil {
		t.Fatal(err)
	}
	agent := records.Agent{Key: "manager-agent", Name: "manager-agent", Role: "manager", Project: "one", Status: records.AgentLive, Generation: 1}
	if err := registry.Write("agents", agent.Key, agent); err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("claims", "manager-one", records.Claim{Key: "manager-one", Agent: agent.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
	workerTask := oldTask
	workerTask.ID = "91-worker-task"
	if err := registry.Write("tasks", workerTask.ID, workerTask); err != nil {
		t.Fatal(err)
	}
	worker := records.Agent{Key: "worker-agent", Name: "worker-agent", Role: "worker", Project: "one", Status: records.AgentLive}
	if err := registry.Write("agents", worker.Key, worker); err != nil {
		t.Fatal(err)
	}
	unauthorized := runHiveFailure(t, "--json", "task", "edit", "--task", workerTask.ID, "--feature", "g", "--agent", worker.Name)
	var unauthorizedResult struct {
		Error string `json:"error"`
	}
	decodeJSON(t, unauthorized, &unauthorizedResult)
	if !strings.Contains(unauthorizedResult.Error, "not the controlling role") {
		t.Fatalf("non-controlling edit error = %q", unauthorizedResult.Error)
	}
	edited := runHive(t, "--json", "task", "edit", "--task", oldID, "--feature", "g", "--agent", agent.Name)
	var editResult struct {
		Outcome string `json:"outcome"`
		Result  struct {
			Feature string `json:"feature"`
		} `json:"result"`
	}
	decodeJSON(t, edited, &editResult)
	if editResult.Outcome != "ok" || editResult.Result.Feature != "g" {
		t.Fatalf("edit feature result = %#v; want feature g", editResult)
	}
	briefFile := filepath.Join(t.TempDir(), "brief.md")
	if err := os.WriteFile(briefFile, []byte("new brief"), 0o644); err != nil {
		t.Fatal(err)
	}
	both := runHiveFailure(t, "--json", "task", "edit", "--task", oldID, "--feature", "f", "--brief", briefFile, "--agent", agent.Name)
	var bothResult struct {
		Phase string `json:"phase"`
		Error string `json:"error"`
	}
	decodeJSON(t, both, &bothResult)
	if bothResult.Phase != "validate" || !strings.Contains(bothResult.Error, "not both") {
		t.Fatalf("edit with both flags = %#v; want validation failure", bothResult)
	}
	repeated := runHiveFailure(t, "--json", "task", "edit", "--task", oldID, "--feature", "f", "--agent", agent.Name)
	var repeatedResult struct {
		Outcome string `json:"outcome"`
		Error   string `json:"error"`
	}
	decodeJSON(t, repeated, &repeatedResult)
	if repeatedResult.Outcome != "failed" || !strings.Contains(repeatedResult.Error, "already has a feature") {
		t.Fatalf("second feature edit = %#v; want refusal", repeatedResult)
	}
}

type featureTaskRecord struct {
	records.Task
	Feature string `json:"feature,omitempty"`
}

func createFeatureTask(t *testing.T, title, feature string, spec bool) string {
	t.Helper()
	args := []string{"--json", "task", "create", "--project", "one", "--title", title, "--feature", feature}
	if spec {
		args = append(args, "--spec")
	}
	result := runHive(t, args...)
	var envelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			ID      string `json:"id"`
			Feature string `json:"feature"`
			Spec    bool   `json:"spec"`
		} `json:"result"`
	}
	decodeJSON(t, result, &envelope)
	if envelope.Outcome != "ok" || envelope.Result.ID == "" || envelope.Result.Feature != feature || envelope.Result.Spec != spec {
		t.Fatalf("create result = %#v; want feature %q and spec %v", envelope, feature, spec)
	}
	return envelope.Result.ID
}

func runHive(t *testing.T, args ...string) string {
	t.Helper()
	output, err := executeHive(t, args...)
	if err != nil {
		t.Fatalf("hive %s: %v; output %s", strings.Join(args, " "), err, output)
	}
	return output
}

func runHiveFailure(t *testing.T, args ...string) string {
	t.Helper()
	output, err := executeHive(t, args...)
	if err == nil {
		t.Fatalf("hive %s succeeded; want failure", strings.Join(args, " "))
	}
	return output
}

func executeHive(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var stdout bytes.Buffer
	err := hivecmd.ExecuteWithArgs(args, nil, &stdout)
	return stdout.String(), err
}

func decodeJSON(t *testing.T, data string, result any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), result); err != nil {
		t.Fatalf("decode output %q: %v", data, err)
	}
}

func gitTask(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
