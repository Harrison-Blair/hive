package spawn

import (
	"strings"
	"testing"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestTextReportsTheLiveAgent(t *testing.T) {
	envelope := output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Result: agentlib.SpawnResult{Agent: records.Agent{Name: "worker-1", Status: records.AgentLive}}}
	if text := Text(envelope); !strings.Contains(text, "agent worker-1 live\n") {
		t.Fatalf("spawn text = %q; want live after confirmed prompt", text)
	}
}

func TestTextOmitsTheAgentOnError(t *testing.T) {
	envelope := output.Envelope{Outcome: "failed", Phase: "prompt", Effects: []output.Effect{}, Error: "worker-1: agent_prompt_stalled", Result: agentlib.SpawnResult{Agent: records.Agent{Name: "worker-1", Status: records.AgentLive}}}
	if text := Text(envelope); strings.Contains(text, "agent worker-1 live") {
		t.Fatalf("spawn text = %q; must not report live without a confirmed prompt", text)
	}
}
