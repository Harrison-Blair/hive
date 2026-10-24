package agent

import (
	"encoding/json"
	"sort"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ListOptions filters the agent list.
type ListOptions struct {
	Project string
	Role    string
	Profile string
	All     bool
}

// List returns agent records in key order.
func List(home string, options ListOptions) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return listFailed(home, options, "store", err.Error())
	}
	raw, err := registry.ListIn([]string{"agents"}, nil)
	if err != nil {
		return listFailed(home, options, "store", err.Error())
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return listFailed(home, options, "store", err.Error())
		}
		if !options.All && hiddenStatus(record.Status) {
			continue
		}
		if options.Project != "" && record.Project != options.Project {
			continue
		}
		if options.Role != "" && record.MirrorRole() != options.Role {
			continue
		}
		if options.Profile != "" && record.Profile != options.Profile {
			continue
		}
		found = append(found, record)
	}
	sort.Slice(found, func(i, j int) bool {
		unresolvedI := found[i].Status == records.AgentUnresolved
		unresolvedJ := found[j].Status == records.AgentUnresolved
		if unresolvedI != unresolvedJ {
			return unresolvedI
		}
		return found[i].Key < found[j].Key
	})
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: found}
}

func hiddenStatus(status string) bool {
	return status == records.AgentStopped || status == records.AgentExited || status == records.AgentStale
}

func listFailed(home string, options ListOptions, phase, message string) output.Envelope {
	attrs := []any{}
	if options.Project != "" {
		attrs = append(attrs, "project_id", options.Project)
	}
	if options.Role != "" {
		attrs = append(attrs, "role", options.Role)
	}
	if options.Profile != "" {
		attrs = append(attrs, "profile", options.Profile)
	}
	attrs = append(attrs, "phase", phase, "error", message)
	logAgentEvent(home, "agent list failed", attrs...)
	return failed(phase, message)
}
