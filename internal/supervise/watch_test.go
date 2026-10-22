package supervise

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func startRun(t *testing.T, home string, ops *fakeOps, clock Clock) (context.CancelFunc, <-chan output.Envelope) {
	t.Helper()
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatalf("process start: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan output.Envelope, 1)
	go func() {
		_, envelope := Run(ctx, Options{
			Home:        home,
			Out:         io.Discard,
			PaneID:      "w1:p0",
			WorkspaceID: "w1",
			TabID:       "w1:t0",
			PID:         pid,
			PIDStart:    start,
			Operations:  ops,
			Clock:       clock,
		})
		done <- envelope
	}()
	deadline := time.Now().Add(3 * time.Second)
	for ops.lastStream() == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if ops.lastStream() == nil {
		cancel()
		t.Fatalf("supervisor did not open a subscription")
	}
	return cancel, done
}

func TestWatchInboxNewRoleReachesNudge(t *testing.T) {
	home := testHome(t)
	manager := managerRecord("manager-demo-g1", "demo", "w1:p1")
	writeAgent(t, home, manager)
	ops := newFakeOps()
	ops.agent["w1:p1"] = herdr.AgentInfo{PaneID: "w1:p1", TerminalID: "term-m", WorkspaceID: "w1", TabID: "w1:t1", AgentStatus: "idle"}
	clock := newFakeClock()
	cancel, done := startRun(t, home, ops, clock)
	defer func() {
		cancel()
		<-done
	}()

	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")
	time.Sleep(100 * time.Millisecond)
	appendInbox(t, home, "manager-demo", records.InboxCompleted, "1-demo", "worker-1", "completed")

	deadline := time.Now().Add(time.Second)
	for ops.promptCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if ops.promptCount() == 0 {
		t.Fatalf("nudge did not arrive within one second")
	}
}

func TestWatchFinishedMarkReachesStop(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	ops.agent["w1:p2"] = herdr.AgentInfo{PaneID: "w1:p2", TerminalID: "term-w", WorkspaceID: "w1", TabID: "w1:t2", AgentStatus: "idle"}
	clock := newFakeClock()
	cancel, done := startRun(t, home, ops, clock)
	defer func() {
		cancel()
		<-done
	}()

	worker := workerRecord("worker-1", "demo", "1-demo", "w1:p2")
	worker.TerminalID = "term-w"
	worker.AgentStatus = "idle"
	worker.Finished = true
	writeAgent(t, home, worker)

	deadline := time.Now().Add(time.Second)
	for ops.stopCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if ops.stopCount() == 0 {
		t.Fatalf("stop did not arrive within one second")
	}
}
