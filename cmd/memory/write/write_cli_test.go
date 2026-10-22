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

func TestActualCLIMemoryLogs(t *testing.T) {
	binary := buildHiveBinary(t)

	t.Run("stdin frontmatter", func(t *testing.T) {
		home := markedCLIHome(t)
		sentinel := "T07FIX01CLISYNTHETICNAME"
		bodySentinel := "T07FIX01CLISYNTHETICBODY"
		input := []byte("---\nname: " + sentinel + "\ndescription: D.\n---\n" + bodySentinel + "\n")
		status, stdout, stderr := runHiveCLI(t, binary, home, input, "--json", "memory", "write", "--scope", "home", "fact")
		if status != 1 {
			t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout, stderr)
		}
		if len(stderr) != 0 {
			t.Fatalf("stderr = %q; want empty", stderr)
		}
		var envelope map[string]any
		if err := json.Unmarshal(stdout, &envelope); err != nil {
			t.Fatalf("stdout is not JSON: %v; output %q", err, stdout)
		}
		want := "frontmatter: name " + sentinel + " differs from fact"
		if envelope["outcome"] != "failed" || envelope["phase"] != "frontmatter" || envelope["error"] != want {
			t.Fatalf("envelope = %#v; want user error %q", envelope, want)
		}
		assertNoCLIContent(t, home, sentinel, bodySentinel)
		if !strings.Contains(cliLogText(t, home), `"error":"frontmatter: name differs"`) {
			t.Fatalf("log = %q; want safe frontmatter error category", cliLogText(t, home))
		}
		for _, event := range cliLogEvents(t, home) {
			if event["scope"] == "home" && event["project_id"] != nil {
				t.Fatalf("home event = %#v; want no invented project_id", event)
			}
		}
	})

	t.Run("existing file", func(t *testing.T) {
		home := markedCLIHome(t)
		sentinel := "T07FIX01CLIEXISTINGSYNTHETICNAME"
		bodySentinel := "T07FIX01CLIEXISTINGSYNTHETICBODY"
		memoryDir := filepath.Join(home, "memory")
		if err := os.MkdirAll(memoryDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(memoryDir, "bad.md"), []byte("---\nname: "+sentinel+"\ndescription: D.\n---\n"+bodySentinel+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(memoryDir, "MEMORY.md"), []byte("old index\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		input := []byte("---\nname: new\ndescription: New fact.\n---\nNew body.\n")
		status, stdout, stderr := runHiveCLI(t, binary, home, input, "--json", "memory", "write", "--scope", "home", "new")
		if status != 1 || len(stderr) != 0 {
			t.Fatalf("status = %d, stderr = %q; want 1 and empty", status, stderr)
		}
		if !strings.Contains(string(stdout), "memory file invalid: ") {
			t.Fatalf("stdout = %q; want memory file invalid", stdout)
		}
		assertNoCLIContent(t, home, sentinel, bodySentinel)
		if !strings.Contains(cliLogText(t, home), "memory file invalid: ") {
			t.Fatalf("log = %q; want safe memory file category", cliLogText(t, home))
		}
	})

	t.Run("project identity", func(t *testing.T) {
		home := markedCLIHome(t)
		repository := t.TempDir()
		gitInit(t, repository)
		if status, stdout, stderr := runHiveCLI(t, binary, home, nil, "project", "register", "--id", "demo", "--base", "dev", repository); status != 0 {
			t.Fatalf("register status = %d (stdout %q, stderr %q)", status, stdout, stderr)
		}
		if status, stdout, stderr := runHiveCLI(t, binary, home, memoryCLIContent("fact", "A fact.", "Body."), "memory", "write", "--scope", "project:demo", "fact"); status != 0 {
			t.Fatalf("write status = %d (stdout %q, stderr %q)", status, stdout, stderr)
		}
		if status, _, _ := runHiveCLI(t, binary, home, []byte("---\nname: other\n---\n"), "memory", "write", "--scope", "project:demo", "other"); status != 1 {
			t.Fatalf("invalid write status = %d; want 1", status)
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
		if projectEvents < 3 {
			t.Fatalf("project events = %d; want memory, index, and error events", projectEvents)
		}
	})
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

func assertNoCLIContent(t *testing.T, home string, sentinels ...string) {
	t.Helper()
	text := cliLogText(t, home)
	for _, sentinel := range sentinels {
		if strings.Contains(text, sentinel) {
			t.Fatalf("log = %q; must not contain %q", text, sentinel)
		}
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	command := exec.Command("git", "init", "-b", "dev", dir)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, output)
	}
}

func memoryCLIContent(name, description, body string) []byte {
	return []byte("---\nname: " + name + "\ndescription: " + description + "\n---\n" + body + "\n")
}
