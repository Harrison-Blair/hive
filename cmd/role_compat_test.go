package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestRoleFlagsAreUnknown proves AC2/D13: the old role flags are gone from
// the agent and inbox commands.
func TestRoleFlagsAreUnknown(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"spawn role", []string{"agent", "spawn", "--role", "worker"}},
		{"register role", []string{"agent", "register", "--role", "worker"}},
		{"list role", []string{"agent", "list", "--role", "worker"}},
		{"inbox role", []string{"inbox", "--role", "coordinator"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			t.Setenv("HIVE_HOME", home)
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr)
			if status == 0 {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want a refusal", status, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), "unknown flag: --role") {
				t.Fatalf("stderr = %q; want unknown flag: --role", stderr.String())
			}
		})
	}
}
