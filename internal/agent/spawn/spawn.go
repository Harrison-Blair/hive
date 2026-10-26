// Package spawn renders the result of the parent agent.SpawnFromCLI API.
package spawn

import (
	"fmt"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Text renders the effects and the display name of one spawn.
func Text(envelope output.Envelope) string {
	text := agentlib.EffectsText(envelope)
	if envelope.Error != nil {
		return text
	}
	result, ok := envelope.Result.(agentlib.SpawnResult)
	if !ok {
		return text
	}
	return text + fmt.Sprintf("agent %s %s\n", result.Agent.Name, result.Agent.Status)
}
