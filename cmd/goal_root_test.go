package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestGoalCommandsManageWavesAndEnforceOwner(t *testing.T) {
	cli := newGoalCLI(t)
	cli.seedAgent(t, records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo"}, true)
	cli.seedAgent(t, records.Agent{Key: "worker-demo-g1", Name: "worker-demo", Role: "worker", Project: "demo"}, false)
	cli.seedAgent(t, records.Agent{Key: "manager-other-g1", Name: "manager-other", Role: "manager", Project: "other"}, true)
	cli.seedTask(t, records.Task{ID: "1-first", Project: "demo", Title: "First task"})
	cli.seedTask(t, records.Task{ID: "2-foreign", Project: "other", Title: "Foreign task"})

	created := cli.mustRun(t, "goal", "create", "--project", "demo", "--title", "Demo goal", "--slug", "demo-goal", "--grant", "grant line", "--agent", "manager-demo")
	if !strings.Contains(created, "id: g1-demo-goal") {
		t.Fatalf("create output = %q; want generated goal ID", created)
	}
	goalID := "g1-demo-goal"
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--add-wave", "First wave", "--done-when", "dev", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--add-wave", "Second wave", "--done-when", "release", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--wave", "A", "--add-task", "1-first", "--agent", "manager-demo")

	status, _, stderr := cli.run("goal", "edit", "--goal", goalID, "--wave", "B", "--add-task", "1-first", "--agent", "manager-demo")
	if status == 0 || !strings.Contains(stderr, "already belongs to wave A") {
		t.Fatalf("duplicate task add status = %d, stderr %q; want refusal", status, stderr)
	}
	status, _, stderr = cli.run("goal", "edit", "--goal", goalID, "--wave", "B", "--add-task", "2-foreign", "--agent", "manager-demo")
	if status == 0 || !strings.Contains(stderr, "not in goal project demo") {
		t.Fatalf("foreign task add status = %d, stderr %q; want project refusal", status, stderr)
	}

	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--wave", "A", "--needs-user", "Need approval", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--wave", "A", "--clear-needs-user", "--agent", "manager-demo")
	shown := cli.mustRun(t, "goal", "show", goalID)
	for _, want := range []string{"wave A: First wave", "tasks: 1-first", "landed: 0/1 landed", "wave B: Second wave"} {
		if !strings.Contains(shown, want) {
			t.Errorf("show output %q does not contain %q", shown, want)
		}
	}
	listed := cli.mustRun(t, "goal", "list", "--project", "demo")
	if !strings.Contains(listed, goalID) || !strings.Contains(listed, "Demo goal") {
		t.Fatalf("list output = %q; want goal", listed)
	}
	status, _, stderr = cli.run("goal", "edit", "--goal", goalID, "--status", "done", "--agent", "worker-demo")
	if status == 0 || !strings.Contains(stderr, "not authorized") {
		t.Fatalf("worker edit status = %d, stderr %q; want authorization refusal", status, stderr)
	}
	status, _, stderr = cli.run("goal", "edit", "--goal", goalID, "--status", "done", "--agent", "manager-other")
	if status == 0 || !strings.Contains(stderr, "not authorized") {
		t.Fatalf("other manager edit status = %d, stderr %q; want project authorization refusal", status, stderr)
	}
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--wave", "B", "--status", "done", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", goalID, "--status", "done", "--agent", "manager-demo")
	var record records.Goal
	registry, err := store.Open(cli.home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ReadIn([]string{"goals"}, goalID, &record); err != nil {
		t.Fatal(err)
	}
	if record.Status != "done" || record.Waves[1].Status != "done" {
		t.Fatalf("goal after status edits = %#v; want goal and wave done", record)
	}
	loggedCommand, loggedChange := false, false
	for _, event := range readRootLogEvents(t, cli.home) {
		if event["msg"] == "command start" && event["command"] == "goal create" {
			loggedCommand = true
		}
		if event["component"] == "goal" && event["msg"] == "goal created" && event["project_id"] == "demo" {
			loggedChange = true
		}
	}
	if !loggedCommand || !loggedChange {
		t.Fatalf("logged command = %v, goal change = %v; want both events", loggedCommand, loggedChange)
	}
}

func TestGoalShowCountsLandedTasksFromDelivery(t *testing.T) {
	cli := newGoalCLI(t)
	cli.seedAgent(t, records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo"}, true)
	cli.seedTask(t, records.Task{ID: "1-landed", Project: "demo", Title: "Landed", Delivery: &records.Delivery{Commit: "abc123", Pushed: true}})
	cli.seedTask(t, records.Task{ID: "2-unpushed", Project: "demo", Title: "Unpushed", Delivery: &records.Delivery{Commit: "def456", Pushed: false}})
	cli.mustRun(t, "goal", "create", "--project", "demo", "--title", "Progress", "--slug", "progress", "--grant", "grant line", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", "g1-progress", "--add-wave", "Delivery", "--done-when", "dev", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", "g1-progress", "--wave", "A", "--add-task", "1-landed", "--agent", "manager-demo")
	cli.mustRun(t, "goal", "edit", "--goal", "g1-progress", "--wave", "A", "--add-task", "2-unpushed", "--agent", "manager-demo")
	shown := cli.mustRun(t, "goal", "show", "g1-progress")
	if !strings.Contains(shown, "landed: 1/2 landed") {
		t.Fatalf("show output = %q; want 1/2 landed", shown)
	}
	var record records.Goal
	registry, err := store.Open(cli.home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ReadIn([]string{"goals"}, "g1-progress", &record); err != nil {
		t.Fatal(err)
	}
	if record.Waves[0].Status != "open" {
		t.Fatalf("show changed wave status to %q; want open", record.Waves[0].Status)
	}
}

func TestGoalCreateReportsCounterEffectWhenRecordWriteFails(t *testing.T) {
	cli := newGoalCLI(t)
	cli.seedAgent(t, records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo"}, true)
	if err := os.Symlink(t.TempDir(), filepath.Join(cli.home, ".hive", "state", "goals")); err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr := cli.run("goal", "create", "--project", "demo", "--title", "Partial", "--slug", "partial", "--grant", "grant line", "--agent", "manager-demo", "--json")
	if status != 1 || stderr != "" {
		t.Fatalf("create status = %d, stderr %q; want a JSON failure", status, stderr)
	}
	var envelope output.Envelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("JSON output %q: %v", stdout, err)
	}
	if envelope.Outcome != "partial" || len(envelope.Effects) != 1 || envelope.Effects[0] != (output.Effect{Resource: "counter", Path: ".hive/state/counters/goals.json", Action: "updated"}) {
		t.Fatalf("envelope = %#v; want the counter effect", envelope)
	}
}

func TestGoalCreateAcceptsKebabSlugWithoutUnspecifiedLengthLimit(t *testing.T) {
	cli := newGoalCLI(t)
	cli.seedAgent(t, records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo"}, true)
	slug := strings.Repeat("a", 41)
	if got := cli.mustRun(t, "goal", "create", "--project", "demo", "--title", "Long slug", "--slug", slug, "--grant", "grant line", "--agent", "manager-demo"); !strings.Contains(got, "g1-"+slug) {
		t.Fatalf("create output = %q; want generated ID with slug", got)
	}
}

func TestGoalCommandsResolveManagerFromPane(t *testing.T) {
	cli := newGoalCLI(t)
	agent := records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo", PaneID: "w1:p1"}
	cli.seedAgent(t, agent, true)
	t.Setenv("HERDR_PANE_ID", agent.PaneID)
	if got := cli.mustRun(t, "goal", "create", "--project", "demo", "--title", "Pane goal", "--slug", "pane-goal", "--grant", "grant line"); !strings.Contains(got, "g1-pane-goal") {
		t.Fatalf("create output = %q; want goal ID", got)
	}
}

func TestGoalCreateJSONUsesEnvelope(t *testing.T) {
	cli := newGoalCLI(t)
	cli.seedAgent(t, records.Agent{Key: "manager-demo-g1", Name: "manager-demo", Role: "manager", Project: "demo"}, true)
	status, stdout, stderr := cli.run("goal", "create", "--project", "demo", "--title", "JSON goal", "--slug", "json-goal", "--grant", "grant", "--agent", "manager-demo", "--json")
	if status != 0 || stderr != "" {
		t.Fatalf("create status = %d, stderr %q; want success", status, stderr)
	}
	var envelope struct {
		Outcome string       `json:"outcome"`
		Phase   string       `json:"phase"`
		Effects []any        `json:"effects"`
		Error   any          `json:"error"`
		Result  records.Goal `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("JSON output %q: %v", stdout, err)
	}
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Result.ID != "g1-json-goal" || envelope.Result.Status != "active" || len(envelope.Result.Waves) != 0 {
		t.Fatalf("envelope = %#v; want active goal in command envelope", envelope)
	}
}

type goalCLI struct {
	home string
}

func newGoalCLI(t *testing.T) *goalCLI {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	home := testHome(t)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	project := records.Project{ID: "demo", Path: t.TempDir(), Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"projects"}, project.ID, project); err != nil {
		t.Fatal(err)
	}
	return &goalCLI{home: home}
}

func (c *goalCLI) seedAgent(t *testing.T, agent records.Agent, controller bool) {
	t.Helper()
	registry, err := store.Open(c.home)
	if err != nil {
		t.Fatal(err)
	}
	agent.Status = records.AgentLive
	agent.Generation = 1
	if agent.Role == "manager" {
		agent.Profile = records.ProfilePlanner
		agent.Owns = agent.Project
	}
	if err := registry.WriteIn([]string{"agents"}, agent.Key, agent); err != nil {
		t.Fatal(err)
	}
	if controller {
		claimKey := "owner-" + agent.Project
		claim := records.Claim{Key: claimKey, Agent: agent.Key, Generation: agent.Generation}
		if err := registry.WriteIn([]string{"claims"}, claimKey, claim); err != nil {
			t.Fatal(err)
		}
	}
}

func (c *goalCLI) seedTask(t *testing.T, task records.Task) {
	t.Helper()
	registry, err := store.Open(c.home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
}

func (c *goalCLI) run(args ...string) (int, string, string) {
	return c.runWithInput(strings.NewReader(""), args...)
}

func (c *goalCLI) runWithInput(input io.Reader, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(args, input, &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func (c *goalCLI) mustRun(t *testing.T, args ...string) string {
	t.Helper()
	status, stdout, stderr := c.run(args...)
	if status != 0 {
		t.Fatalf("%v status = %d; stdout %q, stderr %q", args, status, stdout, stderr)
	}
	return stdout
}
