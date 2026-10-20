package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCommandRequestIDAndLogsRequestFilter(t *testing.T) {
	home := requestTestHome(t)
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	if err := ExecuteWithArgs([]string{"task", "list", "--project", "demo"}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("task list: %v\n%s", err, output.String())
	}

	events := requestTestEvents(t, home)
	requestID := commandRequestID(t, events, "task list", []string{"command start", "command end"})
	output.Reset()
	if err := ExecuteWithArgs([]string{"logs", "--json", "--request", requestID}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("logs --request: %v\n%s", err, output.String())
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("request filter returned %d events; want only start and end:\n%s", len(lines), output.String())
	}
	for _, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event["request_id"] != requestID {
			t.Fatalf("request filter returned event with request_id %v; want %q", event["request_id"], requestID)
		}
	}
}

func TestCommandRequestIDReachesOperationLogs(t *testing.T) {
	home := requestTestHome(t)
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	err := ExecuteWithArgs([]string{"task", "create", "--project", "demo", "--title", "sample", "--slug", "sample"}, strings.NewReader(""), &output)
	if err == nil {
		t.Fatal("task create for an unknown project succeeded")
	}
	events := requestTestEvents(t, home)
	requestID := commandRequestID(t, events, "task create", []string{"command start", "task error", "command end"})
	for _, event := range events {
		if event["command"] == "task create" && event["request_id"] != requestID {
			t.Fatalf("task create event has request_id %v; want %q", event["request_id"], requestID)
		}
	}
}

func TestCommandRequestIDReachesCommandOwnedLoggers(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		home := requestTestHome(t)
		t.Setenv("HIVE_HOME", home)
		if err := os.MkdirAll(filepath.Join(home, "memory"), 0o755); err != nil {
			t.Fatal(err)
		}
		input := strings.NewReader("---\nname: fact\ndescription: A fact.\n---\nBody.\n")
		var output strings.Builder
		if err := ExecuteWithArgs([]string{"memory", "write", "--scope", "home", "fact"}, input, &output); err != nil {
			t.Fatalf("memory write: %v\n%s", err, output.String())
		}
		commandRequestID(t, requestTestEvents(t, home), "memory write", []string{"command start", "effect", "effect", "command end"})
	})

	t.Run("handoff", func(t *testing.T) {
		home := requestTestHome(t)
		t.Setenv("HIVE_HOME", home)
		var output strings.Builder
		if err := ExecuteWithArgs([]string{"handoff", "write", "--scope", "home"}, strings.NewReader("synthetic handoff content\n"), &output); err != nil {
			t.Fatalf("handoff write: %v\n%s", err, output.String())
		}
		commandRequestID(t, requestTestEvents(t, home), "handoff write", []string{"command start", "effect", "command end"})
	})
}

func requestTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func requestTestEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	path := filepath.Join(home, ".hive", "logs", time.Now().Format("2006-01-02")+".jsonl")
	file, err := os.Open(path)
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
	return events
}

func commandRequestID(t *testing.T, events []map[string]any, command string, wantMessages []string) string {
	t.Helper()
	var requestID string
	var messages []string
	for _, event := range events {
		if event["command"] != command {
			continue
		}
		id, _ := event["request_id"].(string)
		if requestID == "" {
			requestID = id
		} else if id != requestID {
			t.Fatalf("%s events have different request IDs: %q and %q", command, requestID, id)
		}
		messages = append(messages, fmt.Sprint(event["msg"]))
	}
	if !regexp.MustCompile(`^r-[0-9a-f]{8}$`).MatchString(requestID) {
		t.Fatalf("request ID %q does not match r- plus 8 lowercase hex characters", requestID)
	}
	if strings.Join(messages, ",") != strings.Join(wantMessages, ",") {
		t.Fatalf("%s events = %v; want %v", command, messages, wantMessages)
	}
	return requestID
}
