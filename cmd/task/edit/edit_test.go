package edit_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hivecmd "github.com/Harrison-Blair/hive/cmd"
	editcmd "github.com/Harrison-Blair/hive/cmd/task/edit"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestNeedsUserFlagsAreAvailable(t *testing.T) {
	command := editcmd.New()
	for _, name := range []string{"needs-user", "clear-needs-user"} {
		if command.Flags().Lookup(name) == nil {
			t.Errorf("task edit has no --%s flag", name)
		}
	}
}

func TestNeedsUserEditFlow(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	task := records.Task{
		ID: "90-decision", Project: "one", Title: "Decision task", ControllingRole: "manager-one",
		Status: "verified", DependsOn: []string{},
		Verification: &records.Verification{Agent: "verifier", Result: "pass", At: "2026-01-01T00:00:00Z"},
		CreatedAt:    "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
	for _, agent := range []records.Agent{
		{Key: "manager", Name: "manager", Role: "manager", Project: "one", Status: records.AgentLive, Generation: 1},
		{Key: "worker", Name: "worker", Role: "worker", Project: "one", Status: records.AgentLive, Generation: 1},
	} {
		if err := registry.WriteIn([]string{"agents"}, agent.Key, agent); err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.WriteIn([]string{"claims"}, "manager-one", records.Claim{Key: "manager-one", Agent: "manager", Generation: 1}); err != nil {
		t.Fatal(err)
	}
	briefPath := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	if err := os.MkdirAll(filepath.Dir(briefPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(briefPath, []byte("Decision task"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, invalid := range []struct {
		text string
		want string
	}{
		{text: strings.Repeat("é", 201), want: "at most 200 characters"},
		{text: "first\nsecond", want: "must be one line"},
	} {
		response, err := runFailure("--json", "task", "edit", "--task", task.ID, "--needs-user", invalid.text, "--agent", "manager")
		if err == nil || !strings.Contains(response, invalid.want) {
			t.Fatalf("needs_user %q = %q, %v; want validation error %q", invalid.text, response, err, invalid.want)
		}
	}

	set := run(t, "--json", "task", "edit", "--task", task.ID, "--needs-user", "choose an option", "--agent", "manager")
	var setEnvelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			NeedsUser string `json:"needs_user"`
		} `json:"result"`
	}
	decode(t, set, &setEnvelope)
	if setEnvelope.Outcome != "ok" || setEnvelope.Result.NeedsUser != "choose an option" {
		t.Fatalf("set needs_user = %#v; want success and the text", setEnvelope)
	}

	shown := run(t, "task", "show", task.ID)
	if !strings.Contains(shown, "needs user: choose an option") {
		t.Fatalf("task show = %q; want needs_user text", shown)
	}
	jsonShown := run(t, "--json", "task", "show", task.ID)
	var showEnvelope struct {
		Result struct {
			NeedsUser string `json:"needs_user"`
		} `json:"result"`
	}
	decode(t, jsonShown, &showEnvelope)
	if showEnvelope.Result.NeedsUser != "choose an option" {
		t.Fatalf("JSON task show = %#v; want needs_user text", showEnvelope.Result)
	}
	listed := run(t, "task", "list")
	if !strings.Contains(listed, "[needs user]") {
		t.Fatalf("task list = %q; want needs_user marker", listed)
	}
	jsonListed := run(t, "--json", "task", "list")
	var listEnvelope struct {
		Result []struct {
			NeedsUser string `json:"needs_user"`
		} `json:"result"`
	}
	decode(t, jsonListed, &listEnvelope)
	if len(listEnvelope.Result) != 1 || listEnvelope.Result[0].NeedsUser != "choose an option" {
		t.Fatalf("JSON task list = %#v; want needs_user text", listEnvelope.Result)
	}

	cleared := run(t, "--json", "task", "edit", "--task", task.ID, "--clear-needs-user", "--agent", "manager")
	var clearEnvelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			NeedsUser string `json:"needs_user"`
		} `json:"result"`
	}
	decode(t, cleared, &clearEnvelope)
	if clearEnvelope.Outcome != "ok" || clearEnvelope.Result.NeedsUser != "" {
		t.Fatalf("clear needs_user = %#v; want success and empty text", clearEnvelope)
	}
	shown = run(t, "task", "show", task.ID)
	if strings.Contains(shown, "needs user:") {
		t.Fatalf("task show after clear = %q; must omit needs_user", shown)
	}
	listed = run(t, "task", "list")
	if strings.Contains(listed, "[needs user]") {
		t.Fatalf("task list after clear = %q; must omit needs_user marker", listed)
	}

	unauthorized, err := runFailure("--json", "task", "edit", "--task", task.ID, "--needs-user", "worker text", "--agent", "worker")
	if err == nil || !strings.Contains(unauthorized, "not the controlling role") {
		t.Fatalf("worker edit = %q, %v; want controlling-role refusal", unauthorized, err)
	}

	completed := task
	completed.ID = "91-completed"
	completed.Status = "completed"
	completed.Verification = nil
	if err := registry.WriteIn([]string{"tasks"}, completed.ID, completed); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--json", "task", "edit", "--task", completed.ID, "--needs-user", "decision for completed task", "--agent", "manager"},
		{"--json", "task", "edit", "--task", completed.ID, "--clear-needs-user", "--agent", "manager"},
	} {
		var result struct {
			Outcome string `json:"outcome"`
		}
		decode(t, run(t, args...), &result)
		if result.Outcome != "ok" {
			t.Fatalf("edit in %s status = %#v; want success", completed.Status, result)
		}
	}
}

func TestNeedsUserEditRejectsMultipleOperations(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	response, err := runFailure("--json", "task", "edit", "--task", "90-decision", "--needs-user", "choose", "--clear-needs-user", "--agent", "manager")
	if err == nil || !strings.Contains(response, "exactly one") {
		t.Fatalf("task edit with set and clear = %q, %v; want validation failure", response, err)
	}
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	text, err := execute(args...)
	if err != nil {
		t.Fatalf("hive %s: %v; output %s", strings.Join(args, " "), err, text)
	}
	return text
}

func runFailure(args ...string) (string, error) {
	return execute(args...)
}

func execute(args ...string) (string, error) {
	var output bytes.Buffer
	err := hivecmd.ExecuteWithArgs(args, strings.NewReader(""), &output)
	return output.String(), err
}

func decode(t *testing.T, text string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(text), target); err != nil {
		t.Fatalf("decode JSON %q: %v", text, err)
	}
}
