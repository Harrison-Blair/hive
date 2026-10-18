package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

func logEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	events := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(readLog(t, home)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line = %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func findLogEvent(t *testing.T, events []map[string]any, message string) map[string]any {
	t.Helper()
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	t.Fatalf("log has no %q event", message)
	return nil
}

func checkBody(t *testing.T, event map[string]any, prefix string) {
	t.Helper()
	body, ok := event["body"].(string)
	if !ok {
		t.Fatalf("event body = %#v; want a string", event["body"])
	}
	if len(body) > 4096 {
		t.Fatalf("body bytes = %d; want at most 4096", len(body))
	}
	if !utf8.ValidString(body) {
		t.Fatalf("body is not valid UTF-8")
	}
	if !strings.HasPrefix(prefix, body) {
		t.Fatalf("body is not a prefix of the source text")
	}
	if event["truncated"] != true {
		t.Fatalf("truncated = %#v; want true", event["truncated"])
	}
}

func TestArgvLogBodyStaysBounded(t *testing.T) {
	home := testHome(t)
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		t.Fatalf("log writer: %v", err)
	}
	logger := slog.New(writer)
	long := "a" + strings.Repeat("é", 3000) + "END"
	logArgv(logger, "agent-key", []string{long})
	if err := writer.Close(); err != nil {
		t.Fatalf("log close: %v", err)
	}
	event := findLogEvent(t, logEvents(t, home), "agent start argv")
	checkBody(t, event, `["`+long+`"]`)
	if event["command"] != "agent spawn" {
		t.Fatalf("command = %#v", event["command"])
	}
}

func TestPromptLogBodyStaysBounded(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, liveRecord("worker-hive-1", "worker-hive-1"))
	fake := newFakeHerdr(t)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+agentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	long := "a" + strings.Repeat("é", 3000) + "END"
	if envelope := Send(context.Background(), home, SendOptions{Agent: "worker-hive-1", Text: long}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	event := findLogEvent(t, logEvents(t, home), "agent prompt")
	checkBody(t, event, long)
	if event["command"] != "agent send" {
		t.Fatalf("command = %#v", event["command"])
	}
}
