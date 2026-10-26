package stop

import (
	"context"
	"fmt"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestStopUnresolved(t *testing.T) {
	home := agenttest.NewHome(t)
	record := agenttest.ResolveSeedRecord()
	record.Status = records.AgentUnresolved
	agenttest.WriteRecord(t, home, record)

	_, envelope := Run(context.Background(), home, "worker-3")
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "agent unresolved: run hive agent resolve" {
		t.Fatalf("Run() = %+v, want the unresolved refusal", envelope)
	}
	if got := agenttest.ReadAgent(t, home, "worker-3"); got.Status != records.AgentUnresolved {
		t.Fatalf("status = %q, want the unchanged record", got.Status)
	}
}
