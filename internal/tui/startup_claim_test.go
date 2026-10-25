package tui

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func readStartupClaim(t *testing.T, homePath string) (records.Claim, error) {
	t.Helper()
	registry, err := store.Open(homePath)
	if err != nil {
		t.Fatal(err)
	}
	var claim records.Claim
	err = registry.ReadIn([]string{"claims"}, "coordinator", &claim)
	return claim, err
}

// TestStartupTakesStartingClaimBeforeCoordinatorTab covers decision 31 for
// bare hive: the starting claim exists before hive makes the coordinator tab.
func TestStartupTakesStartingClaimBeforeCoordinatorTab(t *testing.T) {
	homePath := newTestHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var claim records.Claim
	var claimErr error
	starter := func(_ context.Context, launch *agent.CoordinatorLaunch, _, _, _ string) error {
		claim, claimErr = readStartupClaim(t, homePath)
		return launch.Release()
	}

	if _, err := setupSessionWith(homePath, nil, starter); err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	if claimErr != nil || claim.Status != records.ClaimStarting || claim.PID != int64(os.Getpid()) || claim.PIDStart != currentProcessStart(t) {
		t.Fatalf("claim at coordinator start = %#v, %v; want the starting claim of this launcher", claim, claimErr)
	}
	requireStartupLabels(t, server.recorded(), "coordinator", "supervisor")
}

// TestStartupRefusesWhileAnotherLauncherStarts covers decision 31: a starting
// claim with a live launcher makes the second start fail with no new tab.
func TestStartupRefusesWhileAnotherLauncherStarts(t *testing.T) {
	homePath := newTestHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	launcher := exec.Command("sleep", "60")
	if err := launcher.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = launcher.Process.Kill()
		_, _ = launcher.Process.Wait()
	})
	start, err := process.Start(int64(launcher.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(homePath)
	if err != nil {
		t.Fatal(err)
	}
	held := records.Claim{Key: "coordinator", Generation: 1, PID: int64(launcher.Process.Pid), PIDStart: start, Status: records.ClaimStarting, PaneID: "w1:p7"}
	if err := registry.WriteIn([]string{"claims"}, "coordinator", held); err != nil {
		t.Fatal(err)
	}
	writeStartupSupervisor(t, homePath, int64(os.Getpid()), currentProcessStart(t))
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	starter := func(context.Context, *agent.CoordinatorLaunch, string, string, string) error {
		t.Fatal("coordinator starter ran while another launcher holds the claim")
		return nil
	}

	_, err = setupSessionWith(homePath, nil, starter)
	if err == nil || !strings.Contains(err.Error(), "coordinator is starting") {
		t.Fatalf("setupSessionWith() = %v; want the starting claim refusal", err)
	}
	if created := requestsFor(server.recorded(), "tab.create"); len(created) != 0 {
		t.Fatalf("tab.create = %#v; want no coordinator tab", created)
	}
	if claim, err := readStartupClaim(t, homePath); err != nil || claim != held {
		t.Fatalf("claim = %#v, %v; want the held claim unchanged", claim, err)
	}
}
