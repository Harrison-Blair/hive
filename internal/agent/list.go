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
	All     bool
}

// List returns agent records in key order.
func List(home string, options ListOptions) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error())
	}
	raw, err := registry.ListIn([]string{"agents"}, nil)
	if err != nil {
		return failed("store", err.Error())
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return failed("store", err.Error())
		}
		if !options.All && hiddenStatus(record.Status) {
			continue
		}
		if options.Project != "" && record.Project != options.Project {
			continue
		}
		if options.Role != "" && record.Role != options.Role {
			continue
		}
		found = append(found, record)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Key < found[j].Key })
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: found}
}

func hiddenStatus(status string) bool {
	return status == records.AgentStopped || status == records.AgentExited || status == records.AgentStale
}
