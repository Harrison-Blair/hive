package harness

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestPiInboxArgsWritesTheExtension(t *testing.T) {
	want, err := os.ReadFile("hive-inbox.ts")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	path := filepath.Join(home, "extensions", "hive-inbox.ts")
	profile, _ := Lookup("pi")
	args, err := profile.InboxArgs(home)
	if err != nil {
		t.Fatalf("InboxArgs error = %v", err)
	}
	if !slices.Equal(args, []string{"-e", path}) {
		t.Fatalf("InboxArgs = %q; want -e %s", args, path)
	}
	if got, _ := os.ReadFile(path); string(got) != string(want) {
		t.Fatalf("extension = %q; want hive-inbox.ts", got)
	}
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := profile.InboxArgs(home); err != nil {
		t.Fatalf("InboxArgs error = %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != string(want) {
		t.Fatalf("extension after a change = %q; want hive-inbox.ts again", got)
	}
}

func TestClaudeAndCodexNeedNoInboxArgs(t *testing.T) {
	for _, kind := range []string{"claude", "codex"} {
		profile, _ := Lookup(kind)
		home := t.TempDir()
		if args, err := profile.InboxArgs(home); args != nil || err != nil {
			t.Fatalf("%s InboxArgs = %q, %v; want nil", kind, args, err)
		}
		if entries, _ := os.ReadDir(home); len(entries) != 0 {
			t.Fatalf("%s InboxArgs wrote %v", kind, entries)
		}
	}
}

func piPane(t *testing.T) agenttest.Pane {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.StartAgent(pane.ID, "kid", "pi", "idle")
	got, _ := fake.Pane(pane.ID)
	return got
}

func TestPiDeliverWritesOneLineToTheExtensionSocket(t *testing.T) {
	pane := piPane(t)
	lines := agenttest.ListenPi(t, pane.PID)
	profile, _ := Lookup("pi")
	if err := profile.Deliver(deliverCtx(t), pane.ID, "⬢ boss: say \"hi\"\nnow", nil); err != nil {
		t.Fatalf("Deliver error = %v", err)
	}
	select {
	case line := <-lines:
		if want := `{"text":"⬢ boss: say \"hi\"\nnow"}` + "\n"; line != want {
			t.Fatalf("line = %q; want %q", line, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no line reached the socket")
	}
}

func TestPiDeliverFailsWithoutSocket(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	pane := piPane(t)
	profile, _ := Lookup("pi")
	err := profile.Deliver(deliverCtx(t), pane.ID, "hi", nil)
	if err == nil || !strings.Contains(err.Error(), "no message socket") || !strings.Contains(err.Error(), "restart") {
		t.Fatalf("Deliver error = %v; want no message socket and restart", err)
	}
}
