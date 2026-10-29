package agenttest

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// CommandEvents returns each log event in the home that has the request ID of
// the "command start" event of command.
func CommandEvents(t testing.TB, home, command string) []map[string]any {
	t.Helper()
	file, err := os.Open(filepath.Join(home, ".hive", "logs", time.Now().Local().Format("2006-01-02")+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	var events []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	var requestID any
	for _, event := range events {
		if event["msg"] == "command start" && event["command"] == command {
			requestID = event["request_id"]
		}
	}
	if requestID == nil {
		t.Fatalf("no command start event for %q", command)
	}
	var matched []map[string]any
	for _, event := range events {
		if event["request_id"] == requestID {
			matched = append(matched, event)
		}
	}
	return matched
}

// CheckAgentFields fails the test unless each event has agent_id caller and
// target target, or no target when target is empty, and the events include
// each message of messages.
func CheckAgentFields(t testing.TB, events []map[string]any, caller, target string, messages ...string) {
	t.Helper()
	var wantTarget any
	if target != "" {
		wantTarget = target
	}
	seen := map[any]bool{}
	for _, event := range events {
		seen[event["msg"]] = true
		if event["agent_id"] != caller || event["target"] != wantTarget {
			t.Errorf("event %v: agent_id %v, target %v; want %q, %q", event["msg"], event["agent_id"], event["target"], caller, target)
		}
	}
	for _, message := range messages {
		if !seen[message] {
			t.Errorf("no %q event in %v", message, events)
		}
	}
}
