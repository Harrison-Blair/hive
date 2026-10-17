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
	if got := out.String(); got != "dev\n" {
		t.Fatalf("output = %q, want %q", got, "dev\n")
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
