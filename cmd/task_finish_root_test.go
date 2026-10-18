package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func seedFinishCLI(t *testing.T, status string) (string, string) {
	t.Helper()
	home := testHome(t)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	taskID := "1-step"
	task := records.Task{
		ID: taskID, Project: "one", Title: "Step", ControllingRole: "manager-one",
		ActingAgent: "worker-1", Status: status, DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"tasks"}, taskID, task); err != nil {
		t.Fatal(err)
	}
	for _, record := range []records.Agent{
		{Key: "worker-1", Name: "worker-1", Role: "worker", Project: "one", Task: taskID, Status: records.AgentLive, PaneID: "w1:p1", Generation: 1},
		{Key: "verifier-1", Name: "verifier-1", Role: "verifier", Project: "one", Task: taskID, Status: records.AgentLive, PaneID: "w1:p2", Generation: 1},
	} {
		if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
			t.Fatal(err)
		}
	}
	return home, taskID
}

func assertFinishReport(t *testing.T, stderr string) {
	t.Helper()
	for _, want := range []string{
		"outcome: ok",
		"phase: complete",
		"effect: updated task .hive/state/tasks/1-step.json",
		"inspect: hive task show 1-step",
		"inspect: hive agent show worker-1",
		"inspect: hive inbox --role manager-one",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr = %q; want %q", stderr, want)
		}
	}
}

func TestTaskCompleteShortStdoutReportsKnownEffectsAndDoesNotReplay(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "complete", "--task", taskID, "--agent", "worker-1"}, strings.NewReader(""), &shortWriter{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	assertFinishReport(t, stderr.String())
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil || task.Status != "completed" {
		t.Fatalf("task = %#v, %v; want one completed write", task, err)
	}
	var worker records.Agent
	if err := registry.ReadIn([]string{"agents"}, "worker-1", &worker); err != nil || !worker.Finished {
		t.Fatalf("worker = %#v, %v; want finished", worker, err)
	}
	inboxRecords, err := registry.ListIn([]string{"inbox", "manager-one"}, nil)
	if err != nil || len(inboxRecords) != 1 {
		t.Fatalf("inbox = %d, %v; want one record and no replay", len(inboxRecords), err)
	}
}

func TestTaskVerifyFailedStdoutReportsKnownEffects(t *testing.T) {
	home, taskID := seedFinishCLI(t, "completed")
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "verify", "--task", taskID, "--agent", "verifier-1"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	for _, want := range []string{"outcome: ok", "phase: complete", "inspect: hive task show 1-step", "inspect: hive agent show verifier-1", "inspect: hive inbox --role manager-one"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q; want %q", stderr.String(), want)
		}
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil || task.Status != "verified" {
		t.Fatalf("task = %#v, %v; want verified", task, err)
	}
	if !recordsAgentFinished(t, registry, "verifier-1") {
		t.Fatal("verifier is not finished")
	}
}

type truncatedWriter struct {
	buffer bytes.Buffer
}

func (w *truncatedWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	written := len(data) - 1
	_, _ = w.buffer.Write(data[:written])
	return written, nil
}

func TestTaskCompleteJSONShortStdoutWritesOneEnvelope(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	var stderr bytes.Buffer
	stdout := &truncatedWriter{}
	status := executeWithStreams([]string{"--home", home, "--json", "task", "complete", "--task", taskID, "--agent", "worker-1"}, strings.NewReader(""), stdout, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if got := strings.Count(stdout.buffer.String(), `"outcome"`); got != 1 {
		t.Fatalf("stdout = %q; want one JSON envelope attempt", stdout.buffer.String())
	}
	if !strings.Contains(stderr.String(), "inspect: hive task show 1-step") {
		t.Fatalf("stderr = %q; want the failure report", stderr.String())
	}
}

func recordsAgentFinished(t *testing.T, registry *store.Store, key string) bool {
	t.Helper()
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	return record.Finished
}

func TestTaskCompleteLogsResolvedActorAndEffects(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "complete", "--task", taskID, "--pane", "w1:p1"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; stdout %q, stderr %q; want 0", status, stdout.String(), stderr.String())
	}
	events := taskLogEvents(t, home)
	if countEvents(events, "task transition", map[string]string{
		"command": "task complete", "task_id": taskID, "project_id": "one",
		"agent_id": "worker-1", "from_status": "assigned", "to_status": "completed",
	}) != 1 {
		t.Fatalf("log events = %#v; want one pane-resolved transition", events)
	}
	for _, resource := range []string{"task", "counter", "inbox", "agent"} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task complete", "task_id": taskID, "project_id": "one",
			"agent_id": "worker-1", "resource": resource,
		}) != 1 {
			t.Fatalf("log events = %#v; want one %s effect", events, resource)
		}
	}
	for _, resource := range []string{"counter", "inbox"} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task complete", "task_id": taskID, "role": "manager-one", "resource": resource,
		}) != 1 {
			t.Fatalf("log events = %#v; want the controlling role on the %s effect", events, resource)
		}
	}
}

func TestTaskCompletePartialLogsExactPrefix(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	if err := os.MkdirAll(filepath.Join(home, ".hive", "state", "inbox", "manager-one", "1.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "complete", "--task", taskID, "--pane", "w1:p1"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; stdout %q, stderr %q; want 1", status, stdout.String(), stderr.String())
	}
	events := taskLogEvents(t, home)
	for _, resource := range []string{"task", "counter"} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task complete", "task_id": taskID, "resource": resource, "agent_id": "worker-1",
		}) != 1 {
			t.Fatalf("log events = %#v; want one retained %s effect", events, resource)
		}
	}
	for _, resource := range []string{"inbox", "agent"} {
		if countEvents(events, "task effect", map[string]string{
			"command": "task complete", "task_id": taskID, "resource": resource,
		}) != 0 {
			t.Fatalf("log events = %#v; want no %s effect", events, resource)
		}
	}
	if countEvents(events, "task error", map[string]string{
		"command": "task complete", "task_id": taskID, "agent_id": "worker-1", "phase": "record",
	}) != 1 {
		t.Fatalf("log events = %#v; want one record error", events)
	}
}

func TestTaskCompleteShortStdoutLogsStreamError(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "complete", "--task", taskID, "--pane", "w1:p1"}, strings.NewReader(""), &shortWriter{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	events := taskLogEvents(t, home)
	if countEvents(events, "task output error", map[string]string{
		"command": "task complete", "task_id": taskID, "agent_id": "worker-1", "phase": "output",
	}) != 1 {
		t.Fatalf("log events = %#v; want one stream error", events)
	}
}

func TestVerifyFailLogsNoReasonText(t *testing.T) {
	home, taskID := seedFinishCLI(t, "completed")
	const reason = "reason-sentinel-8a91"
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "verify", "--task", taskID, "--fail", "--reason", reason, "--pane", "w1:p2"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; stdout %q, stderr %q; want 0", status, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(dailyLogPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), reason) {
		t.Fatalf("log holds the reason text: %s", data)
	}
}

func TestTaskCompleteShortStdoutReportsOperationActor(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
		t.Fatal(err)
	}
	task.ActingAgent = "worker-3"
	task.Verification = &records.Verification{Agent: "verifier-old", Result: "fail", Reason: "old", At: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"tasks"}, taskID, task); err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, "worker-3", records.Agent{
		Key: "worker-3", Name: "worker-3", Role: "worker", Project: "one", Task: taskID,
		Status: records.AgentLive, PaneID: "w1:p3", Generation: 1,
	}); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "task", "complete", "--task", taskID, "--pane", "w1:p3"}, strings.NewReader(""), &shortWriter{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	if !strings.Contains(stderr.String(), "inspect: hive agent show worker-3") {
		t.Fatalf("stderr = %q; want the current operation actor", stderr.String())
	}
	if strings.Contains(stderr.String(), "verifier-old") {
		t.Fatalf("stderr = %q; want no prior verifier", stderr.String())
	}
	if !recordsAgentFinished(t, registry, "worker-3") {
		t.Fatal("the current worker is not finished")
	}
}
