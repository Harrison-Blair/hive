package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	for _, key := range []string{"HERDR_ENV", "HERDR_PANE_ID", "HERDR_SOCKET_PATH", "HIVE_NO_STARTUP"} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	configDir, err := os.MkdirTemp("", "hive-cmd-test-config-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(configDir)
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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	return home
}

func TestBareHivePrintsHelp(t *testing.T) {
	newRootTestHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; want 0; stderr %q", status, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "Usage:") || !strings.Contains(got, "Available Commands:") {
		t.Fatalf("output %q is not the help text", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("output %q has terminal escapes; want no TUI", got)
	}
}

func TestRootListsOnlyKeptCommands(t *testing.T) {
	want := []string{"agent", "completion", "help", "init", "logs", "project", "version"}
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
