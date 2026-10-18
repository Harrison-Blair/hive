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
		return failed("read", "agent not found: "+name)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error())
	}
	var record records.Agent
	found := false
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		err := locked.ReadIn([]string{"agents"}, name, &record)
		if err == nil {
			found = true
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		var claim records.Claim
		err = locked.ReadIn([]string{"claims"}, name, &claim)
		if err == nil {
			err = locked.ReadIn([]string{"agents"}, claim.Agent, &record)
			if err == nil {
				found = true
				return nil
			}
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return findAgentByName(locked, name, &record, &found)
	})
	if err != nil {
		return failed("store", err.Error())
	}
	if !found {
		return failed("read", "agent not found: "+name)
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}

func findAgentByName(locked *store.LockedStore, name string, target *records.Agent, found *bool) error {
	raw, err := locked.ListIn([]string{"agents"}, nil)
	if err != nil {
		return err
	}
	candidates := make([]records.Agent, 0)
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}
		if record.Name == name {
			candidates = append(candidates, record)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		if (candidates[i].Status == records.AgentLive) != (candidates[j].Status == records.AgentLive) {
			return candidates[i].Status == records.AgentLive
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
