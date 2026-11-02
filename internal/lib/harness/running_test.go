package harness

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestRunningFindsTheHarnessProcessWhateverTheHerdrAgent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.StartAgent(pane.ID, "kid", "opencode", "unknown")
	fake.SetProcess(pane.ID, "claude")
	profile, _ := Lookup("claude")
	running, err := Running(deliverCtx(t), profile, pane.ID, nil)
	if err != nil || !running {
		t.Fatalf("Running = %v, %v; want true", running, err)
	}
}

func TestRunningIsFalseForAShell(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.SetProcess(pane.ID, "bash")
	for _, kind := range []string{"claude", "codex", "pi"} {
		profile, _ := Lookup(kind)
		running, err := Running(deliverCtx(t), profile, pane.ID, nil)
		if err != nil || running {
			t.Fatalf("%s: Running = %v, %v; want false", kind, running, err)
		}
	}
}

func TestRunningReturnsTheHerdrError(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", "", false)
	fake.StartAgent(pane.ID, "kid", "claude", "idle")
	fake.Fail("pane.process_info", "server_error", "boom")
	profile, _ := Lookup("claude")
	if _, err := Running(deliverCtx(t), profile, pane.ID, nil); err == nil {
		t.Fatal("Running error = nil; want boom")
	}
}
