package supervise

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"log/slog"
)

func TestLoopPassLogsOneRequestLifecycle(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	s := handlerSupervisor(t, home, ops, newFakeClock())
	var output bytes.Buffer
	s.logger = slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if err := s.openWatch(); err != nil {
		t.Fatal(err)
	}
	defer s.watch.Close()
	stream := newFakeStream()
	s.stream = stream
	defer func() { _ = stream.Close() }()

	stream.events <- eventOf("unknown", `{}`)
	stream.events <- eventOf("unknown", `{}`)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := s.loop(ctx); err != nil {
		t.Fatalf("loop: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 6 {
		t.Fatalf("loop wrote %d events; want at least two passes:\n%s", len(lines), output.String())
	}
	var requestIDs []string
	for index := 0; index < len(lines); {
		var start, end map[string]any
		if err := json.Unmarshal([]byte(lines[index]), &start); err != nil {
			t.Fatal(err)
		}
		if start["msg"] != "command start" || start["command"] != "supervise" {
			t.Fatalf("loop event = %v; want supervise command start", start)
		}
		requestID, _ := start["request_id"].(string)
		endIndex := index + 1
		if endIndex < len(lines) {
			var middle map[string]any
			if err := json.Unmarshal([]byte(lines[endIndex]), &middle); err != nil {
				t.Fatal(err)
			}
			if middle["msg"] == "supervisor event" {
				if middle["request_id"] != requestID {
					t.Fatalf("loop event request_id = %v; want %q", middle["request_id"], requestID)
				}
				endIndex++
			}
		}
		if endIndex >= len(lines) {
			t.Fatalf("missing command end after %v", start)
		}
		if err := json.Unmarshal([]byte(lines[endIndex]), &end); err != nil {
			t.Fatal(err)
		}
		if end["msg"] != "command end" || end["command"] != "supervise" || end["request_id"] != requestID || !regexp.MustCompile(`^r-[0-9a-f]{8}$`).MatchString(requestID) {
			t.Fatalf("loop start/end request mismatch: start=%v end=%v", start, end)
		}
		if _, ok := end["duration_ms"]; !ok {
			t.Fatalf("command end has no duration_ms: %v", end)
		}
		if end["outcome"] != "ok" {
			t.Fatalf("command end outcome = %v; want ok", end["outcome"])
		}
		requestIDs = append(requestIDs, requestID)
		index = endIndex + 1
	}
	if len(requestIDs) < 2 || requestIDs[0] == requestIDs[1] {
		t.Fatalf("loop pass request IDs = %v; want a new ID per pass", requestIDs)
	}
}
