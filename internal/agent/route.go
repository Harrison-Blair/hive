package agent

import "github.com/Harrison-Blair/hive/internal/lib/records"

// InboxKey returns the inbox key of one agent (decision 26). The coordinator
// keeps coordinator and a manager keeps manager-PROJECT. Each other agent uses
// its agent name.
func InboxKey(record records.Agent) string {
	switch record.Role {
	case "coordinator":
		return "coordinator"
	case "manager":
		return "manager-" + record.Project
	default:
		return record.Name
	}
}

// ParentInbox returns the inbox that gets the nudges of record (decision 28).
// It starts at the recorded parent and follows the replaces chain to the live
// successor. When no record in the chain is live, or the record has no parent,
// it returns coordinator. agents maps each agent key to its record.
func ParentInbox(agents map[string]records.Agent, record records.Agent) string {
	current, ok := agents[record.Parent]
	seen := map[string]bool{}
	for ok && !seen[current.Key] {
		if current.Status == records.AgentLive {
			return InboxKey(current)
		}
		seen[current.Key] = true
		current, ok = successor(agents, current.Key)
	}
	return "coordinator"
}

// successor returns the newest record that replaces key.
func successor(agents map[string]records.Agent, key string) (records.Agent, bool) {
	var newest records.Agent
	found := false
	for _, candidate := range agents {
		if candidate.Replaces != key {
			continue
		}
		if !found || candidate.StartedAt > newest.StartedAt ||
			(candidate.StartedAt == newest.StartedAt && candidate.Key > newest.Key) {
			newest = candidate
			found = true
		}
	}
	return newest, found
}
