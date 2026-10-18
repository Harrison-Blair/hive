package agent

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// processArgument returns a pointer to one argv0 or argv value.
func processArgument(value string) *string { return &value }

// TestCaptureIdentityArgvMatch proves the identity selection of T19.
func TestCaptureIdentityArgvMatch(t *testing.T) {
	profile, err := harness.Lookup("cursor")
	if err != nil {
		t.Fatalf("lookup cursor: %v", err)
	}
	t.Run("argv0 selects a wrapper launcher", func(t *testing.T) {
		sleep := startSleep(t)
		entry := herdr.PaneProcessInfoProcess{
			PID:   uint32(sleep.pid),
			Name:  "MainThread",
			Argv0: processArgument("cursor-agent"),
			Argv:  []string{"/opt/cursor-agent/index.js"},
		}
		identity, reason := captureIdentity(profile, records.Agent{}, herdr.PaneProcessInfo{ForegroundProcesses: []herdr.PaneProcessInfoProcess{entry}}, nil)
		if reason != "" {
			t.Fatalf("reason = %q", reason)
		}
		if identity.pid != sleep.pid || identity.name != "MainThread" {
			t.Fatalf("identity = %+v", identity)
		}
	})
	t.Run("argv0 with a path selects a wrapper launcher", func(t *testing.T) {
		sleep := startSleep(t)
		entry := herdr.PaneProcessInfoProcess{
			PID:   uint32(sleep.pid),
			Name:  "MainThread",
			Argv0: processArgument("/home/penguin/.local/bin/cursor-agent"),
		}
		identity, reason := captureIdentity(profile, records.Agent{}, herdr.PaneProcessInfo{ForegroundProcesses: []herdr.PaneProcessInfoProcess{entry}}, nil)
		if reason != "" {
			t.Fatalf("reason = %q", reason)
		}
		if identity.pid != sleep.pid {
			t.Fatalf("identity = %+v", identity)
		}
	})
	t.Run("argv with a path selects a wrapper launcher", func(t *testing.T) {
		sleep := startSleep(t)
		entry := herdr.PaneProcessInfoProcess{
			PID:  uint32(sleep.pid),
			Name: "MainThread",
			Argv: []string{"/home/penguin/.local/bin/cursor-agent", "--use-system-ca", "/home/penguin/.local/share/cursor-agent/versions/index.js"},
		}
		identity, reason := captureIdentity(profile, records.Agent{}, herdr.PaneProcessInfo{ForegroundProcesses: []herdr.PaneProcessInfoProcess{entry}}, nil)
		if reason != "" {
			t.Fatalf("reason = %q", reason)
		}
		if identity.pid != sleep.pid {
			t.Fatalf("identity = %+v", identity)
		}
	})
	t.Run("no name or argv match is rejected", func(t *testing.T) {
		entry := herdr.PaneProcessInfoProcess{
			PID:   4242,
			Name:  "node",
			Argv0: processArgument("node"),
			Argv:  []string{"node", "/opt/other/index.js"},
		}
		_, reason := captureIdentity(profile, records.Agent{}, herdr.PaneProcessInfo{ForegroundProcesses: []herdr.PaneProcessInfoProcess{entry}}, nil)
		if reason != "no foreground process has a profile name" {
			t.Fatalf("reason = %q", reason)
		}
	})
	t.Run("two matches are multiple", func(t *testing.T) {
		first := herdr.PaneProcessInfoProcess{PID: 4242, Name: "cursor-agent"}
		second := herdr.PaneProcessInfoProcess{PID: 4243, Name: "MainThread", Argv0: processArgument("cursor-agent")}
		_, reason := captureIdentity(profile, records.Agent{}, herdr.PaneProcessInfo{ForegroundProcesses: []herdr.PaneProcessInfoProcess{first, second}}, nil)
		if reason != "multiple foreground processes have a profile name" {
			t.Fatalf("reason = %q", reason)
		}
	})
}
