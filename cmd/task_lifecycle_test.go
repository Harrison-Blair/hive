package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type taskCLI struct {
	home string
}

func newTaskCLI(t *testing.T) *taskCLI {
	t.Helper()
	home := testHome(t)
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cli := &taskCLI{home: home}
	status, stdout, stderr := cli.run("project", "register", repo)
	if status != 0 {
		t.Fatalf("project register status = %d; stdout %q, stderr %q", status, stdout, stderr)
	}
	return cli
}

func (c *taskCLI) run(args ...string) (int, string, string) {
	return c.runWithInput(strings.NewReader(""), args...)
}

func (c *taskCLI) runWithInput(input io.Reader, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(args, input, &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func (c *taskCLI) seedAgent(t *testing.T, name, role, taskID string) {
	t.Helper()
	registry, err := store.Open(c.home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Agent{
		Key: name, Name: name, Role: role, Project: "repo", Task: taskID,
		Status: records.AgentLive, PaneID: "w1:" + name, Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"agents"}, name, record); err != nil {
		t.Fatal(err)
	}
}

func (c *taskCLI) seedController(t *testing.T) {
	t.Helper()
	registry, err := store.Open(c.home)
	if err != nil {
		t.Fatal(err)
	}
	agent := records.Agent{
		Key: "manager-repo", Name: "manager-repo", Role: "manager", Project: "repo",
		Status: records.AgentLive, PaneID: "manager-pane", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"agents"}, agent.Key, agent); err != nil {
		t.Fatal(err)
	}
	claim := records.Claim{Key: agent.Key, Agent: agent.Key, Generation: agent.Generation}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

func (c *taskCLI) markVerified(t *testing.T, id string) {
	t.Helper()
	registry, err := store.Open(c.home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, id, &task); err != nil {
		t.Fatal(err)
	}
	task.Status = "verified"
	task.Verification = &records.Verification{Agent: "verifier-old", Result: "pass", At: "2026-01-02T00:00:00Z"}
	if err := registry.WriteIn([]string{"tasks"}, id, task); err != nil {
		t.Fatal(err)
	}
}

func (c *taskCLI) createTask(t *testing.T, title string) string {
	t.Helper()
	status, stdout, stderr := c.run("task", "create", "--project", "repo", "--title", title, "--feature", "f", "--spec")
	if status != 0 {
		t.Fatalf("task create status = %d; stdout %q, stderr %q", status, stdout, stderr)
	}
	first, _, _ := strings.Cut(stdout, "\n")
	if !strings.HasPrefix(first, "id: ") {
		t.Fatalf("task create output = %q; want an id line", stdout)
	}
	return strings.TrimPrefix(first, "id: ")
}

func (c *taskCLI) mustRun(t *testing.T, args ...string) string {
	t.Helper()
	status, stdout, stderr := c.run(args...)
	if status != 0 {
		t.Fatalf("%v status = %d; stdout %q, stderr %q", args, status, stdout, stderr)
	}
	return stdout
}

func taskLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(dailyLogPath(home))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	events := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line is not JSON: %v; line %q", err, line)
		}
		if event["component"] == "task" {
			events = append(events, event)
		}
	}
	return events
}

func eventField(event map[string]any, name string) string {
	value, _ := event[name].(string)
	return value
}

func countEvents(events []map[string]any, message string, fields map[string]string) int {
	count := 0
	for _, event := range events {
		if eventField(event, "msg") != message {
			continue
		}
		matches := true
		for name, want := range fields {
			if eventField(event, name) != want {
				matches = false
				break
			}
		}
		if matches {
			count++
		}
	}
	return count
}

func TestTaskCLIEditBrief(t *testing.T) {
	cli := newTaskCLI(t)
	cli.seedController(t)
	briefFile := filepath.Join(t.TempDir(), "brief.md")

	createdID := cli.createTask(t, "Edit in created")
	const updatedBrief = "# Updated brief\n\nKeep this text out of logs.\n"
	if err := os.WriteFile(briefFile, []byte(updatedBrief), 0o644); err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr := cli.run("task", "edit", "--task", createdID, "--brief", briefFile, "--agent", "manager-repo")
	if status != 0 || stderr != "" {
		t.Fatalf("created task edit status = %d; stdout %q, stderr %q; want success", status, stdout, stderr)
	}
	// AC3: task edit --brief also writes the brief of the feature to the home.
	if want := "updated brief .hive/state/briefs/" + createdID + ".md\nupdated brief .hive/spec/f/tasks/" + createdID + "/brief.md\n"; stdout != want {
		t.Fatalf("created task edit output = %q; want %q", stdout, want)
	}
	if got := cli.mustRun(t, "task", "show", "--brief", createdID); got != updatedBrief {
		t.Fatalf("edited brief = %q; want %q", got, updatedBrief)
	}
	if err := os.WriteFile(briefFile, []byte("rejected completed update"), 0o644); err != nil {
		t.Fatal(err)
	}
	completedID := cli.createTask(t, "Edit in completed")
	cli.seedAgent(t, "worker-completed", "worker", completedID)
	cli.mustRun(t, "task", "assign", "--task", completedID, "--agent", "worker-completed")
	cli.mustRun(t, "task", "complete", "--task", completedID, "--agent", "worker-completed")
	status, stdout, stderr = cli.run("task", "edit", "--task", completedID, "--brief", briefFile, "--agent", "manager-repo")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "task is completed: edit needs created or assigned") {
		t.Fatalf("completed task edit status = %d; stdout %q, stderr %q; want status refusal", status, stdout, stderr)
	}
	if got := cli.mustRun(t, "task", "show", "--brief", completedID); got != "Edit in completed" {
		t.Fatalf("completed task brief = %q; want the original text", got)
	}

	workerID := cli.createTask(t, "Edit by worker")
	cli.seedAgent(t, "worker-edit", "worker", workerID)
	status, stdout, stderr = cli.run("task", "edit", "--task", workerID, "--brief", briefFile, "--agent", "worker-edit")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "not the controlling role: worker-edit") {
		t.Fatalf("worker edit status = %d; stdout %q, stderr %q; want role refusal", status, stdout, stderr)
	}
	if got := cli.mustRun(t, "task", "show", "--brief", workerID); got != "Edit by worker" {
		t.Fatalf("worker task brief = %q; want the original text", got)
	}

	stdinID := cli.createTask(t, "Edit from stdin")
	const stdinBrief = "Brief from standard input\n"
	status, stdout, stderr = cli.runWithInput(strings.NewReader(stdinBrief), "task", "edit", "--task", stdinID, "--brief", "-", "--agent", "manager-repo", "--json")
	if status != 0 || stderr != "" {
		t.Fatalf("stdin edit status = %d; stdout %q, stderr %q; want success", status, stdout, stderr)
	}
	var jsonEnvelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &jsonEnvelope); err != nil {
		t.Fatalf("decode JSON edit response: %v; output %q", err, stdout)
	}
	for _, field := range []string{"outcome", "phase", "effects", "error", "result"} {
		if _, ok := jsonEnvelope[field]; !ok {
			t.Errorf("JSON edit response lacks %q: %s", field, stdout)
		}
	}
	if string(jsonEnvelope["outcome"]) != `"ok"` || string(jsonEnvelope["phase"]) != `"complete"` || string(jsonEnvelope["error"]) != "null" {
		t.Fatalf("JSON edit response = %s; want successful complete envelope", stdout)
	}
	if got := cli.mustRun(t, "task", "show", "--brief", stdinID); got != stdinBrief {
		t.Fatalf("stdin edited brief = %q; want %q", got, stdinBrief)
	}

	largeID := cli.createTask(t, "Edit with oversized brief")
	if err := os.WriteFile(briefFile, bytes.Repeat([]byte("x"), (1<<20)+1), 0o644); err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr = cli.run("task", "edit", "--task", largeID, "--brief", briefFile, "--agent", "manager-repo")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "brief too large") {
		t.Fatalf("oversized edit status = %d; stdout %q, stderr %q; want size refusal", status, stdout, stderr)
	}
	if got := cli.mustRun(t, "task", "show", "--brief", largeID); got != "Edit with oversized brief" {
		t.Fatalf("oversized task brief = %q; want the original text", got)
	}

	logText, err := os.ReadFile(dailyLogPath(cli.home))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(logText), "Keep this text out of logs") {
		t.Fatal("task edit log contains brief text")
	}
	events := taskLogEvents(t, cli.home)
	editEvents := 0
	for _, event := range events {
		if eventField(event, "msg") != "task brief edit" || eventField(event, "task_id") != createdID {
			continue
		}
		editEvents++
		if eventField(event, "command") != "task edit" || eventField(event, "project_id") != "repo" ||
			eventField(event, "agent_id") != "manager-repo" || event["old_size"] != float64(15) || event["new_size"] != float64(45) {
			t.Fatalf("brief edit event = %#v; want task, caller, and byte sizes", event)
		}
	}
	if editEvents != 1 {
		t.Fatalf("task events = %#v; want one edit event", events)
	}
}

func TestTaskCLILogsTransitionsAndErrors(t *testing.T) {
	cli := newTaskCLI(t)
	id := cli.createTask(t, "Logged")
	cli.seedAgent(t, "worker-1", "worker", id)
	cli.seedAgent(t, "verifier-1", "verifier", id)

	steps := []struct {
		args    []string
		command string
		from    string
		to      string
		agent   string
	}{
		{[]string{"task", "assign", "--task", id, "--agent", "worker-1"}, "task assign", "created", "assigned", "worker-1"},
		{[]string{"task", "complete", "--task", id, "--agent", "worker-1"}, "task complete", "assigned", "completed", "worker-1"},
		{[]string{"task", "verify", "--task", id, "--agent", "verifier-1"}, "task verify", "completed", "verified", "verifier-1"},
	}
	for _, step := range steps {
		cli.mustRun(t, step.args...)
	}

	events := taskLogEvents(t, cli.home)
	if countEvents(events, "task transition", map[string]string{
		"command": "task create", "task_id": id, "project_id": "repo", "to_status": "created",
	}) != 1 {
		t.Fatalf("log events = %#v; want one create transition", events)
	}
	for _, step := range steps {
		if countEvents(events, "task transition", map[string]string{
			"command": step.command, "task_id": id, "project_id": "repo",
			"agent_id": step.agent, "from_status": step.from, "to_status": step.to,
		}) != 1 {
			t.Fatalf("log events = %#v; want a %s transition %s -> %s", events, step.command, step.from, step.to)
		}
	}

	actionID := cli.createTask(t, "Action")
	cli.seedAgent(t, "worker-2", "worker", actionID)
	cli.mustRun(t, "task", "assign", "--task", actionID, "--agent", "worker-2")
	cli.mustRun(t, "task", "complete", "--task", actionID, "--agent", "worker-2")
	before := countEvents(taskLogEvents(t, cli.home), "task transition", map[string]string{"task_id": actionID})

	status, _, stderr := cli.run("task", "verify", "--task", actionID, "--agent", "worker-2")
	if status != 1 {
		t.Fatalf("self verification status = %d; stderr %q; want 1", status, stderr)
	}
	afterEvents := taskLogEvents(t, cli.home)
	if got := countEvents(afterEvents, "task transition", map[string]string{"task_id": actionID}); got != before {
		t.Fatalf("failed verification logged %d transitions; want the prior %d", got, before)
	}
	if countEvents(afterEvents, "task error", map[string]string{
		"command": "task verify", "task_id": actionID, "project_id": "repo", "agent_id": "worker-2",
	}) != 1 {
		t.Fatalf("log events = %#v; want one self-verification error", afterEvents)
	}
	for _, event := range afterEvents {
		if eventField(event, "msg") == "task error" && eventField(event, "task_id") == actionID {
			if eventField(event, "error") == "" {
				t.Fatalf("error event = %#v; want an error field", event)
			}
		}
	}
}

func TestTaskCLIReopenTextJSONShowHistoryAndLog(t *testing.T) {
	cli := newTaskCLI(t)
	cli.seedController(t)

	missingID := cli.createTask(t, "Reopen without reason")
	cli.markVerified(t, missingID)
	status, stdout, stderr := cli.run("task", "reopen", "--task", missingID, "--agent", "manager-repo")
	if status == 0 || stdout != "" || !strings.Contains(stderr, "reason") {
		t.Fatalf("reopen without reason: status %d, stdout %q, stderr %q; want a reason error", status, stdout, stderr)
	}
	status, stdout, stderr = cli.run("task", "reopen", "--task", missingID, "--reason", "", "--agent", "manager-repo")
	if status == 0 || stdout != "" || !strings.Contains(stderr, "--reason is required") {
		t.Fatalf("reopen with empty reason: status %d, stdout %q, stderr %q; want a required reason error", status, stdout, stderr)
	}
	registry, err := store.Open(cli.home)
	if err != nil {
		t.Fatal(err)
	}
	var unchanged records.Task
	if err := registry.ReadIn([]string{"tasks"}, missingID, &unchanged); err != nil {
		t.Fatal(err)
	}
	if unchanged.Status != "verified" || unchanged.Verification == nil || len(unchanged.VerificationHistory) != 0 {
		t.Fatalf("task after rejected reopen = %#v; want the original verified record", unchanged)
	}

	textID := cli.createTask(t, "Reopen in text mode")
	cli.markVerified(t, textID)
	text := cli.mustRun(t, "task", "reopen", "--task", textID, "--reason", "merge conflict", "--agent", "manager-repo")
	if text != "id: "+textID+"\nstatus: assigned\n" {
		t.Fatalf("reopen text = %q; want task ID and assigned status", text)
	}
	shown := cli.mustRun(t, "task", "show", textID)
	for _, want := range []string{"verification: -", `verification_history: [{"verification":`, `"reason":"merge conflict"`, `"caller":"manager-repo"`} {
		if !strings.Contains(shown, want) {
			t.Fatalf("task show = %q; want %q", shown, want)
		}
	}

	jsonID := cli.createTask(t, "Reopen in JSON mode")
	cli.markVerified(t, jsonID)
	status, stdout, stderr = cli.run("task", "reopen", "--task", jsonID, "--reason", "json repair", "--pane", "manager-pane", "--json")
	if status != 0 || stderr != "" {
		t.Fatalf("JSON reopen status = %d; stdout %q, stderr %q; want success", status, stdout, stderr)
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			Status              string                        `json:"status"`
			Verification        json.RawMessage               `json:"verification"`
			VerificationHistory []records.VerificationHistory `json:"verification_history"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode JSON reopen: %v; output %q", err, stdout)
	}
	if envelope.Outcome != "ok" || envelope.Result.Status != "assigned" || string(envelope.Result.Verification) != "null" ||
		len(envelope.Result.VerificationHistory) != 1 || envelope.Result.VerificationHistory[0].Reason != "json repair" {
		t.Fatalf("JSON reopen = %#v; want assigned with cleared verification and history", envelope)
	}

	events := taskLogEvents(t, cli.home)
	if countEvents(events, "task transition", map[string]string{
		"command": "task reopen", "task_id": jsonID, "project_id": "repo", "agent_id": "manager-repo",
		"from_status": "verified", "to_status": "assigned", "reason": "json repair",
	}) != 1 {
		t.Fatalf("log events = %#v; want the reopen transition with its reason", events)
	}
}

func TestTaskCLIReopenRegisteredManagerUsesClaimKey(t *testing.T) {
	cli := newTaskCLI(t)
	// The CLI refuses the manager role, so hive code registers the manager.
	if envelope := agentlib.Register(cli.home, agentlib.RegisterOptions{Role: "manager", Project: "repo", Name: "manager-repo", NameSet: true, PaneID: "w1:manager", Harness: "pi", Model: "test"}); envelope.Outcome != "ok" {
		t.Fatalf("register manager = %#v", envelope)
	}

	registry, err := store.Open(cli.home)
	if err != nil {
		t.Fatal(err)
	}
	var manager records.Agent
	if err := registry.ReadIn([]string{"agents"}, "manager-repo-g1", &manager); err != nil {
		t.Fatal(err)
	}
	var claim records.Claim
	if err := registry.ReadIn([]string{"claims"}, "manager-repo", &claim); err != nil {
		t.Fatal(err)
	}
	if manager.Key != "manager-repo-g1" || manager.Name != "manager-repo" || claim.Key != "manager-repo" || claim.Agent != manager.Key {
		t.Fatalf("manager = %#v, claim = %#v; want generation-suffixed agent and inbox-key claim", manager, claim)
	}

	for _, caller := range []struct {
		name string
		args []string
	}{
		{name: "agent", args: []string{"--agent", "manager-repo"}},
		{name: "pane", args: []string{"--pane", "w1:manager"}},
	} {
		t.Run(caller.name, func(t *testing.T) {
			id := cli.createTask(t, "Reopen by registered manager")
			cli.markVerified(t, id)
			args := []string{"task", "reopen", "--task", id, "--reason", "merge conflict"}
			status, stdout, stderr := cli.run(append(args, caller.args...)...)
			if status != 0 || stderr != "" {
				t.Fatalf("reopen status = %d; stdout %q, stderr %q; want success", status, stdout, stderr)
			}
			var reopened records.Task
			if err := registry.ReadIn([]string{"tasks"}, id, &reopened); err != nil {
				t.Fatal(err)
			}
			if reopened.Status != "assigned" || reopened.Verification != nil || len(reopened.VerificationHistory) != 1 ||
				reopened.VerificationHistory[0].Caller != "manager-repo" {
				t.Fatalf("reopened task = %#v; want assigned with manager-repo in history", reopened)
			}
		})
	}
}

func TestTaskCLILogsPartialCreatePrefixes(t *testing.T) {
	cli := newTaskCLI(t)
	taskPath := filepath.Join(cli.home, ".hive", "state", "tasks", "1-topic.json")
	if err := os.MkdirAll(taskPath, 0o755); err != nil {
		t.Fatal(err)
	}

	status, stdout, stderr := cli.run("task", "create", "--project", "repo", "--title", "Topic", "--feature", "f", "--spec")
	if status != 1 {
		t.Fatalf("partial create status = %d; stdout %q, stderr %q; want 1", status, stdout, stderr)
	}
	for _, want := range []string{
		"updated counter .hive/state/counters/task.json",
		"created brief .hive/state/briefs/1-topic.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("partial create stdout = %q; want %q", stdout, want)
		}
	}

	events := taskLogEvents(t, cli.home)
	// AC2: task create writes two briefs, the state brief and the home brief
	// of the feature.
	for resource, want := range map[string]int{"counter": 1, "brief": 2} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task create", "task_id": "1-topic", "project_id": "repo", "resource": resource,
		}) != want {
			t.Fatalf("log events = %#v; want %d completed %s effects", events, want, resource)
		}
	}
	if countEvents(events, "task effect", map[string]string{
		"command": "task create", "task_id": "1-topic", "resource": "task",
	}) != 0 {
		t.Fatalf("log events = %#v; want no completed task effect", events)
	}
	if countEvents(events, "task transition", map[string]string{"task_id": "1-topic"}) != 0 {
		t.Fatalf("log events = %#v; want no transition for the partial create", events)
	}
	if countEvents(events, "task error", map[string]string{
		"command": "task create", "task_id": "1-topic", "project_id": "repo", "phase": "task",
	}) != 1 {
		t.Fatalf("log events = %#v; want one partial create error", events)
	}
}

func TestTaskCLIRejectsLinkedBriefAncestor(t *testing.T) {
	cli := newTaskCLI(t)
	external := t.TempDir()
	briefs := filepath.Join(cli.home, ".hive", "state", "briefs")
	if err := os.Symlink(external, briefs); err != nil {
		t.Fatal(err)
	}

	status, _, stderr := cli.run("task", "create", "--project", "repo", "--title", "Seed", "--feature", "f", "--spec")
	if status != 1 {
		t.Fatalf("linked brief create status = %d; stderr %q; want 1", status, stderr)
	}
	if strings.Count(stderr, "\n") != 1 || !strings.Contains(stderr, "brief path invalid") {
		t.Fatalf("stderr = %q; want one brief path diagnostic", stderr)
	}
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("external directory holds %d entries; want none", len(entries))
	}
	for _, path := range []string{
		filepath.Join(cli.home, ".hive", "state", "tasks", "1-seed.json"),
		filepath.Join(cli.home, ".hive", "state", "counters", "task.json"),
	} {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("linked brief create made %s: %v", path, err)
		}
	}
}

func TestTaskCLIRejectsLinkedBriefOnShow(t *testing.T) {
	cli := newTaskCLI(t)
	id := cli.createTask(t, "Second step")
	briefs := filepath.Join(cli.home, ".hive", "state", "briefs")
	archived := filepath.Join(cli.home, ".hive", "state", "briefs-original")
	if err := os.Rename(briefs, archived); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	const externalBrief = "external brief"
	if err := os.WriteFile(filepath.Join(external, id+".md"), []byte(externalBrief), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, briefs); err != nil {
		t.Fatal(err)
	}

	status, stdout, stderr := cli.run("task", "show", id)
	if status != 1 || stdout != "" {
		t.Fatalf("linked brief show status = %d, stdout = %q; want 1 and empty stdout", status, stdout)
	}
	if strings.Count(stderr, "\n") != 1 || !strings.Contains(stderr, "brief path invalid") {
		t.Fatalf("stderr = %q; want one brief path diagnostic", stderr)
	}
	data, err := os.ReadFile(filepath.Join(external, id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != externalBrief {
		t.Fatalf("external brief = %q; want %q", data, externalBrief)
	}
}

func TestTaskCLIShowRendersCancelReason(t *testing.T) {
	cli := newTaskCLI(t)
	id := cli.createTask(t, "Stop")
	cli.mustRun(t, "task", "cancel", "--task", id, "--reason", "stop reason")

	status, stdout, stderr := cli.run("task", "show", id)
	if status != 0 || stderr != "" {
		t.Fatalf("show status = %d, stderr = %q; want 0 and empty stderr", status, stderr)
	}
	if !strings.Contains(stdout, "cancel_reason: stop reason\n") {
		t.Fatalf("show stdout = %q; want the stored cancel reason", stdout)
	}
}

func TestTaskCLIPlainDiagnosticsStayOneLine(t *testing.T) {
	cli := newTaskCLI(t)

	status, stdout, stderr := cli.run("task", "show", "bad\nsecond-line")
	want := "hive: task not found: \"bad\\nsecond-line\"\n"
	if status != 1 || stdout != "" || stderr != want {
		t.Fatalf("invalid ID: status = %d, stdout = %q, stderr = %q; want one line %q", status, stdout, stderr, want)
	}

	id := cli.createTask(t, "Actor")
	cli.seedAgent(t, "worker-1", "worker", id)
	cli.mustRun(t, "task", "assign", "--task", id, "--agent", "worker-1")
	status, stdout, stderr = cli.run("task", "complete", "--task", id, "--agent", "bad\nactor")
	if status != 1 || stdout != "" {
		t.Fatalf("wrong actor: status = %d, stdout = %q; want 1 and empty stdout", status, stdout)
	}
	if strings.Count(stderr, "\n") != 1 || strings.ContainsAny(strings.TrimSuffix(stderr, "\n"), "\r\t") {
		t.Fatalf("wrong actor stderr = %q; want one plain line", stderr)
	}
	if !strings.Contains(stderr, `agent not found: "bad\nactor"`) {
		t.Fatalf("wrong actor stderr = %q; want the quoted unknown actor", stderr)
	}
	status, stdout, stderr = cli.run("task", "show", id)
	if status != 0 || stderr != "" || !strings.Contains(stdout, "status: assigned") {
		t.Fatalf("record after wrong actor: status = %d, stderr = %q, stdout = %q; want assigned", status, stderr, stdout)
	}
}
