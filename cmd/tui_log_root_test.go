package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoStartupTUIWritesCLIStartAndEnd covers AC3 with the actual CLI. The
// outer CLI writer keeps the TUI command end after the TUI writer closes, and
// the accepted TUI quit event stays exactly once.
func TestNoStartupTUIWritesCLIStartAndEnd(t *testing.T) {
	home := newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams([]string{"--no-startup"}, strings.NewReader("\x03\x03"), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; want 0; output %q", status, out.String())
	}

	logDir := filepath.Join(home, ".hive", "logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatal(err)
	}
	var logFiles []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".jsonl") {
			logFiles = append(logFiles, filepath.Join(logDir, entry.Name()))
		}
	}
	if len(logFiles) != 1 {
		t.Fatalf("log files = %v; want one daily file", logFiles)
	}

	var starts, ends, quits int
	file, err := os.Open(logFiles[0])
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("decode log event: %v\n%s", err, scanner.Text())
		}
		if event["component"] == "tui" && event["action"] == "quit request" {
			quits++
		}
		if event["component"] != "cli" || event["command"] != "tui" {
			continue
		}
		switch event["msg"] {
		case "command start":
			starts++
		case "command end":
			ends++
			if event["outcome"] != "ok" {
				t.Fatalf("TUI command end outcome = %v; want ok", event["outcome"])
			}
			if _, ok := event["duration_ms"]; !ok {
				t.Fatalf("TUI command end has no duration_ms: %v", event)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if starts != 1 || ends != 1 {
		t.Fatalf("TUI CLI events: %d start, %d end; want one each", starts, ends)
	}
	if quits != 1 {
		t.Fatalf("TUI quit events = %d; want exactly 1", quits)
	}
}
