// Package stop stops one named agent for the agent stop command.
package stop

import (
	"context"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Run resolves one agent name and stops its record.
func Run(ctx context.Context, home, name string) (agentlib.StopResult, output.Envelope) {
	shown := agentlib.Show(home, name)
	if shown.Outcome != "ok" {
		return agentlib.StopResult{}, shown
	}
	record, ok := shown.Result.(records.Agent)
	if !ok {
		return agentlib.StopResult{}, agentlib.Failed("agent", "agent record missing")
	}
	return agentlib.Stop(ctx, home, record)
}

// Text renders the effects and the post-close topology of one stop.
func Text(envelope output.Envelope) string {
	text := agentlib.EffectsText(envelope)
	if envelope.Error != nil {
		return text
	}
	result, ok := envelope.Result.(agentlib.StopResult)
	if !ok {
		return text
	}
	return text + "workspace " + result.Topology + "\n"
}
