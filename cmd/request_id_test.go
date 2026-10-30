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
	if err := ExecuteWithArgs([]string{"project", "list"}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("project list: %v\n%s", err, output.String())
	}

	events := requestTestEvents(t, home)
	requestID := commandRequestID(t, events, "project list", []string{"command start", "command end"})
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

func TestCommandRequestIDReachesCommandOwnedLoggers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	if err := ExecuteWithArgs([]string{"logs", "--limit", "0"}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("logs: %v\n%s", err, output.String())
	}
	commandRequestID(t, requestTestEvents(t, home), "logs", []string{"command start", "command end"})
}

func requestTestHome(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func requestTestEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	path := filepath.Join(home, "logs", time.Now().Format("2006-01-02")+".jsonl")
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
