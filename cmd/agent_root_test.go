package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func readRootLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(dailyLogPath(home))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line is not JSON: %v; line %q", err, line)
		}
		events = append(events, event)
	}
	return events
}

func findRootLogEvent(events []map[string]any, message string) map[string]any {
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	return nil
}

func TestAgentShowUnknownLogsReadFailure(t *testing.T) {
	home := testHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "agent", "show", "unknown"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 {
		t.Fatalf("status = %d, stderr = %q; want 1", status, stderr.String())
	}
	events := readRootLogEvents(t, home)
	event := findRootLogEvent(events, "agent show failed")
	if event == nil {
		t.Fatalf("events = %#v; want an agent show failure event", events)
	}
	message, _ := event["error"].(string)
	if event["name"] != "unknown" || event["phase"] != "read" || !strings.Contains(message, "agent not found: unknown") {
		t.Fatalf("failure event = %#v; want name unknown and the read error", event)
	}
	if strings.ContainsAny(message, "\n\r\x1b") {
		t.Fatalf("error = %q; want one physical line", message)
	}
	for _, candidate := range events {
		if candidate["msg"] == "agent record written" || candidate["msg"] == "agent claim written" {
			t.Fatalf("event = %#v; want no fabricated success event", candidate)
		}
	}
}

func TestAgentRegisterUsesHerdrEnvironment(t *testing.T) {
	home := testHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("HERDR_WORKSPACE_ID", "w1")
	t.Setenv("HERDR_TAB_ID", "t1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "coordinator-g1", &record); err != nil {
		t.Fatal(err)
	}
	if record.PaneID != "w1:p1" || record.WorkspaceID != "w1" || record.TabID != "t1" {
		t.Fatalf("record = %#v; want the environment pane values", record)
	}
}

func TestAgentRegisterExplicitPaneLeavesWorkspaceAndTabEmpty(t *testing.T) {
	home := testHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("HERDR_WORKSPACE_ID", "w1")
	t.Setenv("HERDR_TAB_ID", "t1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator", "--pane", "w2:p2"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "coordinator-g1", &record); err != nil {
		t.Fatal(err)
	}
	if record.PaneID != "w2:p2" || record.WorkspaceID != "" || record.TabID != "" {
		t.Fatalf("record = %#v; want explicit pane only", record)
	}
}

func TestAgentRegisterRejectsUnsafeHerdrPaneValues(t *testing.T) {
	for _, pane := range []string{"", "bad\npane", "bad\x01pane", strings.Repeat("p", 600)} {
		t.Run(strings.ReplaceAll(pane, "\n", "\\n"), func(t *testing.T) {
			home := testHome(t)
			t.Setenv("HERDR_PANE_ID", pane)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d, stderr = %q; want 1", status, stderr.String())
			}
			if strings.Count(stderr.String(), "\n") != 1 || !strings.HasPrefix(stderr.String(), "hive: ") {
				t.Fatalf("stderr = %q; want one hive error line", stderr.String())
			}
			registry, err := store.Open(home)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := registry.ListIn([]string{"agents"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(raw) != 0 {
				t.Fatalf("unsafe pane wrote %d agent records", len(raw))
			}
		})
	}
}
