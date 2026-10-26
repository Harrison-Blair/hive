package resolve

import (
	"context"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestResolveManualUnresolvedExits proves B2: a record that a person registered
// by hand (Spawned false) settles as exited. The exited path checks no live
// identity.
func TestResolveManualUnresolvedExits(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteProjectPath(t, home, "demo", t.TempDir())
	agenttest.WriteTaskRecord(t, home, records.Task{ID: "1-task", Project: "demo", Title: "Task"})
	registered := agentlib.Register(home, agentlib.RegisterOptions{
		Profile: records.ProfileBuilder, Project: "demo", Task: "1-task", Harness: "claude", PaneID: "w1:p1",
	})
	if registered.Outcome != "ok" {
		t.Fatalf("Register() = %+v, want ok", registered)
	}
	record := registered.Result.(agentlib.RegisterResult).Agent
	if record.Spawned {
		t.Fatalf("registered record spawned = true, want false")
	}
	record.Status = records.AgentUnresolved
	agenttest.WriteRecord(t, home, record)

	envelope := Run(context.Background(), home, Options{Agent: record.Key, As: records.AgentExited})
	if envelope.Outcome != "ok" || envelope.Phase != "complete" {
		t.Fatalf("Run() = %+v, want ok complete", envelope)
	}
	got := agenttest.ReadAgent(t, home, record.Key)
	if got.Status != records.AgentExited {
		t.Fatalf("status = %q, want exited", got.Status)
	}
	if got.Spawned {
		t.Fatalf("resolved record spawned = true, want false")
	}
}

// TestResolveMatchesCanonicalCwd proves B3: a directory link and its target
// match when hive compares the record with the live agent.
func TestResolveMatchesCanonicalCwd(t *testing.T) {
	home := agenttest.NewHome(t)
	realDir := filepath.Join(t.TempDir(), "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	linkDir := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(linkDir)
	if err != nil || resolved != realDir {
		t.Fatalf("EvalSymlinks(%q) = %q, %v; want %q", linkDir, resolved, err, realDir)
	}
	record := agenttest.ResolveSeedRecord()
	record.Cwd = linkDir
	agenttest.WriteRecord(t, home, record)
	fake := agenttest.NewFakeHerdr(t)
	fake.Result("agent.get", liveAgentInfo("worker-3", "w1:p1", "term_a", "w1", "w1:t1", "claude", realDir))
	envelope := Run(context.Background(), home, Options{Agent: "worker-3", As: records.AgentLive})
	if envelope.Outcome != "ok" || envelope.Phase != "complete" {
		t.Fatalf("Run() = %+v, want ok complete", envelope)
	}
	if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentLive {
		t.Fatalf("status = %q, want live", got.Status)
	}
}
