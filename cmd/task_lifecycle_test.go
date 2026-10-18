package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	full := append([]string{"--home", c.home}, args...)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(full, strings.NewReader(""), &stdout, &stderr)
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

func (c *taskCLI) createTask(t *testing.T, title string) string {
	t.Helper()
	status, stdout, stderr := c.run("task", "create", "--project", "repo", "--title", title)
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

func TestTaskCLILogsPartialCreatePrefixes(t *testing.T) {
	cli := newTaskCLI(t)
	taskPath := filepath.Join(cli.home, ".hive", "state", "tasks", "1-topic.json")
	if err := os.MkdirAll(taskPath, 0o755); err != nil {
		t.Fatal(err)
	}

	status, stdout, stderr := cli.run("task", "create", "--project", "repo", "--title", "Topic")
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
	for _, resource := range []string{"counter", "brief"} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task create", "task_id": "1-topic", "project_id": "repo", "resource": resource,
		}) != 1 {
			t.Fatalf("log events = %#v; want one completed %s effect", events, resource)
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

	status, _, stderr := cli.run("task", "create", "--project", "repo", "--title", "Seed")
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
