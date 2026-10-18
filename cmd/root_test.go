package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestBareHiveRunsTUIUntilQuit(t *testing.T) {
	home := newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams([]string{"--home", home}, strings.NewReader("q"), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; want 0", status)
	}
	// A Bubble Tea program hides the cursor on start.
	if !strings.Contains(out.String(), "\x1b[?25l") {
		t.Fatalf("output %q shows no TUI ran", out.String())
	}
	if !strings.Contains(out.String(), home) {
		t.Fatalf("output %q does not show the resolved home", out.String())
	}
	if strings.Contains(out.String(), `"outcome"`) {
		t.Fatalf("successful TUI wrote an envelope: %q", out.String())
	}
}

func TestBareJSONRefusesBeforeTUI(t *testing.T) {
	home := newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams([]string{"--json", "--home", home}, strings.NewReader(""), &out, &out)
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	got := out.String()
	if strings.Contains(got, "\x1b[?25l") {
		t.Fatalf("JSON refusal started the TUI: %q", got)
	}
	for _, want := range []string{`"outcome":"failed"`, `"phase":"validate"`, `"error":"TUI requires text output"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %s", got, want)
		}
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("output %q is not one envelope line", got)
	}
}
