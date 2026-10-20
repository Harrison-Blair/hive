package write_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestActualCLIHandoffLogsProjectID(t *testing.T) {
	binary := buildHiveBinary(t)
	home := markedCLIHome(t)
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", "-b", "dev", repository).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, output)
	}
	if status, stdout, stderr := runHiveCLI(t, binary, home, nil, "project", "register", "--id", "demo", "--base", "dev", repository); status != 0 {
		t.Fatalf("register status = %d (stdout %q, stderr %q)", status, stdout, stderr)
	}
	bodySentinel := "T07FIX01HANDOFFSYNTHETICBODY"
	if status, stdout, stderr := runHiveCLI(t, binary, home, []byte(bodySentinel+"\n"), "handoff", "write", "--scope", "project:demo"); status != 0 {
		t.Fatalf("handoff status = %d (stdout %q, stderr %q)", status, stdout, stderr)
	}
	if status, _, _ := runHiveCLI(t, binary, home, nil, "handoff", "write", "--scope", "project:demo"); status != 1 {
		t.Fatalf("empty handoff status = %d; want 1", status)
	}
	text := cliLogText(t, home)
	if strings.Contains(text, bodySentinel) {
		t.Fatalf("log = %q; must not contain handoff content", text)
	}
	projectEvents := 0
	for _, event := range cliLogEvents(t, home) {
		if event["scope"] != "project:demo" {
			continue
		}
		projectEvents++
		if event["project_id"] != "demo" {
			t.Fatalf("event = %#v; want project_id demo", event)
		}
	}
	if projectEvents != 2 {
		t.Fatalf("project events = %d; want handoff effect and input error", projectEvents)
	}
}

func buildHiveBinary(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "hive")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = root
	command.Env = append(os.Environ(), "HIVE_HOME=")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}
	return binary
}

func markedCLIHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	marker := filepath.Join(home, ".hive", "README.md")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("cli fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func runHiveCLI(t *testing.T, binary, home string, input []byte, args ...string) (int, []byte, []byte) {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Stdin = bytes.NewReader(input)
	command.Dir = t.TempDir()
	command.Env = append(os.Environ(), "HIVE_HOME="+home, "XDG_CONFIG_HOME="+t.TempDir())
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return 0, stdout.Bytes(), stderr.Bytes()
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("run hive: %v", err)
	}
	return exit.ExitCode(), stdout.Bytes(), stderr.Bytes()
}

func cliLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	events := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(cliLogText(t, home)), "\n") {
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

func cliLogText(t *testing.T, home string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var text strings.Builder
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text.Write(data)
	}
	return text.String()
}
