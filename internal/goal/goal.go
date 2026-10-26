package goal

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	// GoalsCollection is the store collection of goal records.
	GoalsCollection = "goals"
	// TasksCollection is the store collection of task records.
	TasksCollection = "tasks"
)

// Authorize resolves the caller and accepts the project owner or the coordinator.
func Authorize(locked *store.LockedStore, agent, paneID, project string) (records.Agent, error) {
	caller, err := identity.ResolveHeld(locked, agent, paneID)
	if err != nil {
		return records.Agent{}, err
	}
	if caller.Kind == records.AgentKindCoordinator || caller.Owns == project {
		return caller, nil
	}
	return caller, fmt.Errorf("not the owner of project %s or the coordinator: %s", project, OneLine(caller.Name))
}

// GoalPath returns the home-relative path of one goal record.
func GoalPath(id string) string { return ".hive/state/goals/" + id + ".json" }

// Failed returns a failed envelope with no effects.
func Failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
