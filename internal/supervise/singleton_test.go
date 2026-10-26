package supervise

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestSingletonLiveProcessBlocks(t *testing.T) {
	home := testHome(t)
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	writeClaim(t, home, records.Claim{Key: "supervisor", Agent: "supervisor-g1", Generation: 1, PID: pid, PIDStart: start})
	envelope := agent.StartSupervisor(home, agent.RegisterOptions{Kind: records.AgentKindSupervisor, PaneID: "w1:p0", PID: pid, PIDStart: start})
	if envelope.Outcome != "failed" {
		t.Fatalf("outcome = %q; want failed", envelope.Outcome)
	}
	if !strings.Contains(sprintError(envelope.Error), "supervisor live") {
		t.Fatalf("error = %v; want supervisor live", envelope.Error)
	}
}

func TestSingletonDeadPIDTakesGeneration(t *testing.T) {
	home := testHome(t)
	dead := int64(999999999)
	writeClaim(t, home, records.Claim{Key: "supervisor", Agent: "supervisor-g1", Generation: 1, PID: dead, PIDStart: 111})
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	envelope := agent.StartSupervisor(home, agent.RegisterOptions{Kind: records.AgentKindSupervisor, PaneID: "w1:p0", PID: pid, PIDStart: start})
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome = %q error = %v; want ok", envelope.Outcome, envelope.Error)
	}
	result := envelope.Result.(agent.RegisterResult)
	if result.Claim == nil || result.Claim.Generation != 2 {
		t.Fatalf("claim = %+v; want generation 2", result.Claim)
	}
}

func TestSingletonOwnPIDAnotherStartIsStale(t *testing.T) {
	home := testHome(t)
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	writeClaim(t, home, records.Claim{Key: "supervisor", Agent: "supervisor-g1", Generation: 1, PID: pid, PIDStart: start + 1})
	envelope := agent.StartSupervisor(home, agent.RegisterOptions{Kind: records.AgentKindSupervisor, PaneID: "w1:p0", PID: pid, PIDStart: start})
	if envelope.Outcome != "ok" {
		t.Fatalf("outcome = %q error = %v; want ok", envelope.Outcome, envelope.Error)
	}
	result := envelope.Result.(agent.RegisterResult)
	if result.Claim == nil || result.Claim.Generation != 2 {
		t.Fatalf("claim = %+v; want generation 2", result.Claim)
	}
}

func TestSingletonTwentyStartsGiveOneClaim(t *testing.T) {
	home := testHome(t)
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	const count = 20
	results := make([]output.Envelope, count)
	var wg sync.WaitGroup
	for index := range count {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = agent.StartSupervisor(home, agent.RegisterOptions{Kind: records.AgentKindSupervisor, PaneID: "w1:p0", PID: pid, PIDStart: start})
		}(index)
	}
	wg.Wait()
	ok, live := 0, 0
	for _, envelope := range results {
		switch envelope.Outcome {
		case "ok":
			ok++
		default:
			if strings.Contains(sprintError(envelope.Error), "supervisor live") {
				live++
			}
		}
	}
	if ok != 1 || live != count-1 {
		t.Fatalf("ok = %d live = %d; want one ok and %d live", ok, live, count-1)
	}
}

func TestExitSignalSIGINT(t *testing.T) {
	runSignalExit(t, syscall.SIGINT)
}

func TestExitSignalSIGTERM(t *testing.T) {
	runSignalExit(t, syscall.SIGTERM)
}

func runSignalExit(t *testing.T, sig syscall.Signal) {
	t.Helper()
	home := testHome(t)
	ops := newFakeOps()
	pid := int64(os.Getpid())
	start, err := process.Start(pid)
	if err != nil {
		t.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	done := make(chan output.Envelope, 1)
	go func() {
		_, envelope := Run(ctx, Options{Home: home, Out: io.Discard, PaneID: "w1:p0", PID: pid, PIDStart: start, Operations: ops, Clock: newFakeClock()})
		done <- envelope
	}()
	deadline := time.Now().Add(3 * time.Second)
	for ops.lastStream() == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if ops.lastStream() == nil {
		t.Fatal("no subscription")
	}
	if err := syscall.Kill(os.Getpid(), sig); err != nil {
		t.Fatal(err)
	}
	select {
	case envelope := <-done:
		if envelope.Outcome != "ok" {
			t.Fatalf("outcome = %q error = %v; want ok", envelope.Outcome, envelope.Error)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("run did not exit on the signal")
	}
}

func TestExitSubscriptionClosed(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	clock := newFakeClock()
	cancel, done := startRun(t, home, ops, clock)
	defer cancel()
	stream := ops.lastStream()
	if stream == nil {
		t.Fatal("no subscription")
	}
	_ = stream.Close()
	select {
	case envelope := <-done:
		if envelope.Outcome != "failed" {
			t.Fatalf("outcome = %q; want failed", envelope.Outcome)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("run did not exit when the subscription closed")
	}
}

func TestJSONPrintsOneEnvelope(t *testing.T) {
	home := testHome(t)
	ops := newFakeOps()
	clock := newFakeClock()
	cancel, done := startRun(t, home, ops, clock)
	cancel()
	envelope := <-done
	var buffer bytes.Buffer
	if err := output.Render(&buffer, io.Discard, true, envelope, ""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("json output has %d lines; want one", len(lines))
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("json output is not one object: %v", err)
	}
	if parsed["outcome"] == nil {
		t.Fatalf("json envelope has no outcome: %v", parsed)
	}
}

func sprintError(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(strings.Join(strings.Fields(valueString(value)), " "))
}

func valueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case error:
		return typed.Error()
	default:
		return ""
	}
}
