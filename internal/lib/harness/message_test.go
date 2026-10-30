package harness

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

// claudePane starts a fake claude agent with args in a pane of a fake herdr.
func claudePane(t *testing.T, args []string) (*agenttest.Herdr, agenttest.Pane) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.StartAgent(pane.ID, "kid", "claude", "idle")
	fake.SetArgs(pane.ID, args)
	got, _ := fake.Pane(pane.ID)
	return fake, got
}

func deliverCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestClaudeArgsAcceptCrossSessionMessages(t *testing.T) {
	profile, _ := Lookup("claude")
	args := profile.Args("m", nil)
	for index, arg := range args {
		if arg == "--settings" && index+1 < len(args) && args[index+1] == `{"crossSessionInbound":"accept"}` {
			return
		}
	}
	t.Fatalf("Args() = %q; want --settings with crossSessionInbound accept", args)
}

func TestClaudeDeliverWritesOneUserLineToTheSocket(t *testing.T) {
	profile, _ := Lookup("claude")
	_, pane := claudePane(t, profile.Args("m", nil))
	lines := agenttest.ListenClaude(t, pane.PID)
	if err := profile.Deliver(deliverCtx(t), pane.ID, "⬢ boss: say \"hi\"\nnow"); err != nil {
		t.Fatalf("Deliver error = %v", err)
	}
	select {
	case line := <-lines:
		want := `{"type":"user","message":{"role":"user","content":"⬢ boss: say \"hi\"\nnow"}}` + "\n"
		if line != want {
			t.Fatalf("line = %q; want %q", line, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no line reached the socket")
	}
}

func TestClaudeDeliverRefusesAClaudeWithoutTheSetting(t *testing.T) {
	profile, _ := Lookup("claude")
	_, pane := claudePane(t, []string{"--dangerously-skip-permissions", "--model", "m"})
	lines := agenttest.ListenClaude(t, pane.PID)
	err := profile.Deliver(deliverCtx(t), pane.ID, "hi")
	if err == nil || !strings.Contains(err.Error(), "crossSessionInbound") || !strings.Contains(err.Error(), "restart") {
		t.Fatalf("Deliver error = %v; want the missing setting and restart", err)
	}
	select {
	case line := <-lines:
		t.Fatalf("socket got %q; want nothing", line)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestClaudeDeliverFailsWithoutSocket(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	profile, _ := Lookup("claude")
	_, pane := claudePane(t, profile.Args("m", nil))
	err := profile.Deliver(deliverCtx(t), pane.ID, "hi")
	if err == nil || !strings.Contains(err.Error(), "no message socket") || !strings.Contains(err.Error(), "restart") {
		t.Fatalf("Deliver error = %v; want no message socket and restart", err)
	}
}

func TestClaudeDeliverFailsWithoutClaudeProcess(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	profile, _ := Lookup("claude")
	err := profile.Deliver(deliverCtx(t), pane.ID, "hi")
	if err == nil || !strings.Contains(err.Error(), "no claude process") {
		t.Fatalf("Deliver error = %v; want no claude process", err)
	}
}
