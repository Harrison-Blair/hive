package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestStopUnresolved(t *testing.T) {
	home := testHome(t)
	record := resolveSeedRecord()
	record.Status = records.AgentUnresolved
	writeRecord(t, home, record)

	_, envelope := StopByName(context.Background(), home, "worker-3")
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "agent unresolved: run hive agent resolve" {
		t.Fatalf("StopByName() = %+v, want the unresolved refusal", envelope)
	}
	if got := readResolveRecord(t, home, "worker-3"); got.Status != records.AgentUnresolved {
		t.Fatalf("status = %q, want the unchanged record", got.Status)
	}
}
