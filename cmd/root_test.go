package cmd

import (
	"bytes"
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

func TestBareHiveRunsTUIUntilQuit(t *testing.T) {
	var out bytes.Buffer
	if err := ExecuteWithArgs(nil, strings.NewReader("q"), &out); err != nil {
		t.Fatalf("hive: %v", err)
	}
	// A Bubble Tea program hides the cursor on start.
	if !strings.Contains(out.String(), "\x1b[?25l") {
		t.Fatalf("output %q shows no TUI ran", out.String())
	}
}
