package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	for _, key := range []string{"HERDR_ENV", "HERDR_PANE_ID", "HERDR_SOCKET_PATH", "HIVE_HOME", "HIVE_NO_STARTUP"} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	// A test never reaches the real ~/.hive: the fallback home is below a
	// temporary user home.
	userHome, err := os.MkdirTemp("", "hive-cmd-test-home-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("HOME", userHome); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(userHome)
	os.Exit(code)
}

func TestHomeFlagIsUnknown(t *testing.T) {
	t.Setenv("HIVE_HOME", newRootTestHome(t))
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--" + "home", "/tmp/ignored", "version"}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 {
		t.Fatalf("status = %d; want 2; stdout %q, stderr %q", status, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "unknown flag: --"+"home") {
		t.Fatalf("stderr = %q; want unknown flag error", stderr.String())
	}
}

func TestVersionSubcommandPrintsVersion(t *testing.T) {
	var out bytes.Buffer
	if err := ExecuteWithArgs([]string{"version"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("version: %v", err)
	}
	if got := out.String(); !strings.HasPrefix(got, "hive dev (commit unknown, go") || !strings.HasSuffix(got, ")\n") || strings.Count(got, "\n") != 1 {
		t.Fatalf("output = %q, want one version line", got)
	}
}

func newRootTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HIVE_HOME", home)
	return home
}

func TestBareHiveInHomeCtrlCQuitsWithNoRecord(t *testing.T) {
	home := newRootTestHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader("\x03"), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; want 0; stderr %q", status, stderr.String())
	}
	logs, err := filepath.Glob(filepath.Join(home, "logs", "*.jsonl"))
	if err != nil || len(logs) != 1 {
		t.Fatalf("log files = %v, %v; want one", logs, err)
	}
	events, err := os.ReadFile(logs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(events), `"msg":"picker quit"`) {
		t.Fatalf("log = %s; want a picker quit event", events)
	}
	if strings.Contains(stdout.String(), "Available Commands:") {
		t.Fatalf("stdout = %q; want no help text", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(home, "agents", "coordinator.json")); !os.IsNotExist(err) {
		t.Fatalf("coordinator record stat = %v; want none", err)
	}
}

func TestRootListsOnlyKeptCommands(t *testing.T) {
	want := []string{"agent", "completion", "guide", "help", "logs", "project", "version"}
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"--help"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d; stderr %q", status, stderr.String())
	}
	if got := helpCommands(stdout.String()); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("commands = %v; want %v", got, want)
	}
	stdout.Reset()
	if status := executeWithStreams([]string{"project", "--help"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d; stderr %q", status, stderr.String())
	}
	if got := helpCommands(stdout.String()); strings.Join(got, " ") != "help list register" && strings.Join(got, " ") != "list register" {
		t.Fatalf("project commands = %v; want list register", got)
	}
}

// helpCommands returns the command names of the Available Commands section.
func helpCommands(help string) []string {
	var names []string
	in := false
	for _, line := range strings.Split(help, "\n") {
		switch {
		case strings.HasPrefix(line, "Available Commands:"):
			in = true
		case in && strings.HasPrefix(line, "  "):
			names = append(names, strings.Fields(line)[0])
		case in:
			return names
		}
	}
	return names
}

func TestBareHiveWithoutHiveHomeUsesDotHiveInUserHome(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)
	t.Setenv("HIVE_HOME", "")
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader("\x03"), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; want 0; stderr %q", status, stderr.String())
	}
	logs, err := filepath.Glob(filepath.Join(userHome, ".hive", "logs", "*.jsonl"))
	if err != nil || len(logs) != 1 {
		t.Fatalf("log files in ~/.hive = %v, %v; want one", logs, err)
	}
}

func TestBareHiveWithoutUserHomeFailsWithNoPicker(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("HIVE_HOME", "")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader(""), &stdout, &stderr)
	if status == 0 {
		t.Fatalf("status = 0; want non-zero; stdout %q", stdout.String())
	}
	if !strings.HasPrefix(stderr.String(), "hive: hive home: ") {
		t.Fatalf("stderr = %q; want a hive home error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q; want no picker output", stdout.String())
	}
}
