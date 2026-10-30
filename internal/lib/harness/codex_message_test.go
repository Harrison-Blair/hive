package harness

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

// fakeCodex puts a codex script on PATH. The script writes each argument,
// ended by a NUL byte, to the returned file, prints out, and exits with
// status.
func fakeCodex(t *testing.T, out string, status int) string {
	t.Helper()
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	script := "#!/bin/sh\nfor arg in \"$@\"; do printf '%s\\0' \"$arg\" >> '" + calls + "'; done\necho '" + out + "' >&2\nexit " + string(rune('0'+status)) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return calls
}

// codexArgs returns the arguments of each fake codex call.
func codexArgs(t *testing.T, calls string) []string {
	t.Helper()
	data, err := os.ReadFile(calls)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
}

func codexPane(t *testing.T, session string) (*agenttest.Herdr, string) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.StartAgent(pane.ID, "kid", "codex", "idle")
	fake.SetSession(pane.ID, session)
	return fake, pane.ID
}

func TestCodexDeliverQueuesTheMessageOnTheThread(t *testing.T) {
	calls := fakeCodex(t, "Queued message", 0)
	fake, paneID := codexPane(t, "01a18ed7-6603-7fc1-9d78-2a5e693e23fb")
	profile, _ := Lookup("codex")
	if err := profile.Deliver(deliverCtx(t), paneID, "⬢ boss: say \"hi\"\nnow", nil); err != nil {
		t.Fatalf("Deliver error = %v", err)
	}
	want := []string{"queue", "--thread", "01a18ed7-6603-7fc1-9d78-2a5e693e23fb", "--message", "⬢ boss: say \"hi\"\nnow"}
	if got := codexArgs(t, calls); !slices.Equal(got, want) {
		t.Fatalf("codex args = %q; want %q", got, want)
	}
	for _, method := range []string{"agent.prompt", "pane.send_input"} {
		if got := fake.Requests(method); len(got) != 0 {
			t.Fatalf("%s calls = %d; want none", method, len(got))
		}
	}
}

func TestCodexDeliverFailsWithoutSession(t *testing.T) {
	calls := fakeCodex(t, "Queued message", 0)
	_, paneID := codexPane(t, "")
	profile, _ := Lookup("codex")
	err := profile.Deliver(deliverCtx(t), paneID, "hi", nil)
	if err == nil || !strings.Contains(err.Error(), "no session ID") || !strings.Contains(err.Error(), "first turn") {
		t.Fatalf("Deliver error = %v; want no session ID and the fix", err)
	}
	if got := codexArgs(t, calls); got != nil {
		t.Fatalf("codex args = %q; want no call", got)
	}
}

func TestCodexDeliverReportsAFailedQueue(t *testing.T) {
	fakeCodex(t, "no rollout found for thread id", 1)
	_, paneID := codexPane(t, "01a18ed7-6603-7fc1-9d78-2a5e693e23fb")
	profile, _ := Lookup("codex")
	err := profile.Deliver(deliverCtx(t), paneID, "hi", nil)
	if err == nil || !strings.Contains(err.Error(), "no rollout found") {
		t.Fatalf("Deliver error = %v; want the codex error", err)
	}
}
