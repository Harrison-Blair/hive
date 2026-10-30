package log

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOutcomeSetsTheLevel(t *testing.T) {
	home := t.TempDir()
	clock := time.Now()
	writer, err := NewWriter(home, "cli", func() time.Time { return clock })
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	logger.Debug("ok", "outcome", "ok")
	logger.Debug("partial", "outcome", "partial")
	logger.Debug("failed", "outcome", "failed")
	logger.With("outcome", "failed").Debug("failed from With")
	logger.Info("info without outcome")
	logger.Debug("debug without outcome")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(home, "logs", clock.Local().Format("2006-01-02")+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"ok":                    "DEBUG",
		"partial":               "WARN",
		"failed":                "ERROR",
		"failed from With":      "ERROR",
		"info without outcome":  "INFO",
		"debug without outcome": "DEBUG",
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(want) {
		t.Fatalf("%d lines; want %d:\n%s", len(lines), len(want), data)
	}
	for _, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event["level"] != want[event["msg"].(string)] {
			t.Errorf("%v: level %v; want %s", event["msg"], event["level"], want[event["msg"].(string)])
		}
	}
}
