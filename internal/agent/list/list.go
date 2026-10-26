// Package list renders the agent list of the parent agent.List API.
package list

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Text renders one line for each agent record.
func Text(envelope output.Envelope) string {
	agents, ok := envelope.Result.([]records.Agent)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(agents) == 0 {
		return "no agents\n"
	}
	lines := make([]string, 0, len(agents))
	for _, record := range agents {
		lines = append(lines, fmt.Sprintf("%s\t%d\t%s\t%s\t%s\t%s\t%s", record.Name, record.Generation, agentLabel(record), record.Status, record.Project, record.Task, record.PaneID))
	}
	return strings.Join(lines, "\n") + "\n"
}

// agentLabel gives the kind of a built-in record, or its profile label.
func agentLabel(record records.Agent) string {
	if record.Kind != "" {
		return record.Kind
	}
	return record.Profile
}
