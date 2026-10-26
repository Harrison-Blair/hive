package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sort"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

var lookupPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,199}$`)

// Show returns one agent record by name, singleton name, or record key.
func Show(home, name string) output.Envelope {
	if !lookupPattern.MatchString(name) {
		return showFailed(home, name, "read", "agent not found: "+DisplayValue(name))
	}
	registry, err := store.Open(home)
	if err != nil {
		return showFailed(home, name, "store", err.Error())
	}
	var record records.Agent
	found := false
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var claim records.Claim
		claimErr := locked.ReadIn([]string{"claims"}, name, &claim)
		if claimErr == nil {
			err := locked.ReadIn([]string{"agents"}, claim.Agent, &record)
			if err == nil && claimMatchesRecord(claim, record) {
				found = true
				return nil
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			return nil
		}
		if !errors.Is(claimErr, os.ErrNotExist) {
			return claimErr
		}
		err := locked.ReadIn([]string{"agents"}, name, &record)
		if err == nil {
			found = true
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return findAgentByName(locked, name, &record, &found)
	})
	if err != nil {
		return showFailed(home, name, "store", err.Error())
	}
	if !found {
		return showFailed(home, name, "read", "agent not found: "+DisplayValue(name))
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}

func showFailed(home, name, phase, message string) output.Envelope {
	LogAgentEvent(home, "agent show failed", "name", DisplayValue(name), "phase", phase, "error", message)
	return Failed(phase, message)
}

func claimMatchesRecord(claim records.Claim, record records.Agent) bool {
	return record.Key == claim.Agent &&
		record.Generation == claim.Generation &&
		claim.Key == recordClaimKey(record)
}

// recordClaimKey returns the wave-A claim key of one record from its facts.
func recordClaimKey(record records.Agent) string {
	switch {
	case record.Kind == records.AgentKindCoordinator:
		return coordinatorClaimKey
	case record.Kind == records.AgentKindSupervisor:
		return records.AgentKindSupervisor
	case record.Owns != "":
		return "owner-" + record.Owns
	default:
		return ""
	}
}

func findAgentByName(locked *store.LockedStore, name string, target *records.Agent, found *bool) error {
	raw, err := locked.ListIn([]string{"agents"}, nil)
	if err != nil {
		return err
	}
	claimRaw, err := locked.ListIn([]string{"claims"}, nil)
	if err != nil {
		return err
	}
	claims := make([]records.Claim, 0, len(claimRaw))
	for _, data := range claimRaw {
		var claim records.Claim
		if err := json.Unmarshal(data, &claim); err != nil {
			return err
		}
		claims = append(claims, claim)
	}
	candidates := make([]records.Agent, 0)
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}
		if record.Name != name {
			continue
		}
		if recordClaimKey(record) != "" && !heldByClaim(claims, record) {
			continue
		}
		candidates = append(candidates, record)
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		if isActiveAgentStatus(candidates[i].Status) != isActiveAgentStatus(candidates[j].Status) {
			return isActiveAgentStatus(candidates[i].Status)
		}
		if candidates[i].Generation != candidates[j].Generation {
			return candidates[i].Generation > candidates[j].Generation
		}
		return candidates[i].Key < candidates[j].Key
	})
	*target = candidates[0]
	*found = true
	return nil
}

func heldByClaim(claims []records.Claim, record records.Agent) bool {
	key := recordClaimKey(record)
	for _, claim := range claims {
		if claim.Key == key && claim.Agent == record.Key && claim.Generation == record.Generation {
			return true
		}
	}
	return false
}
