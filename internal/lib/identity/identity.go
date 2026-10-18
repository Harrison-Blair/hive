// Package identity resolves the caller of one command from an agent name or a
// herdr pane ID.
package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ErrCallerUnknown reports a caller that no live agent record matches.
var ErrCallerUnknown = errors.New("caller unknown: pass --agent")

type reader interface {
	ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error)
}

// Resolve returns the caller record without a caller-held store handle.
func Resolve(home, agentFlag, paneID string) (records.Agent, error) {
	registry, err := store.Open(home)
	if err != nil {
		return records.Agent{}, err
	}
	var record records.Agent
	var inner error
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		record, inner = ResolveHeld(locked, agentFlag, paneID)
		return nil
	})
	if lockErr != nil {
		return records.Agent{}, lockErr
	}
	return record, inner
}

// ResolveHeld returns the caller record from one held store handle.
func ResolveHeld(locked *store.LockedStore, agentFlag, paneID string) (records.Agent, error) {
	agents, err := loadAgents(locked)
	if err != nil {
		return records.Agent{}, err
	}
	claims, err := loadClaims(locked)
	if err != nil {
		return records.Agent{}, err
	}
	claimByKey := make(map[string]records.Claim, len(claims))
	for _, claim := range claims {
		claimByKey[claim.Key] = claim
	}
	if agentFlag != "" {
		return resolveFlag(agents, claimByKey, agentFlag)
	}
	return resolvePane(agents, claimByKey, paneID)
}

func resolveFlag(agents []records.Agent, claims map[string]records.Claim, flag string) (records.Agent, error) {
	if claim, ok := claims[flag]; ok {
		for _, record := range agents {
			if record.Key == claim.Agent && validRecord(record, claims) {
				return record, nil
			}
		}
		return records.Agent{}, fmt.Errorf("agent not found: %s", displayValue(flag))
	}
	matches := make([]records.Agent, 0, 1)
	for _, record := range agents {
		if record.Key != flag && record.Name != flag {
			continue
		}
		if !validRecord(record, claims) {
			continue
		}
		matches = append(matches, record)
	}
	switch len(matches) {
	case 0:
		return records.Agent{}, fmt.Errorf("agent not found: %s", displayValue(flag))
	case 1:
		return matches[0], nil
	default:
		return records.Agent{}, ambiguousError(matches)
	}
}

func resolvePane(agents []records.Agent, claims map[string]records.Claim, paneID string) (records.Agent, error) {
	if paneID == "" {
		return records.Agent{}, ErrCallerUnknown
	}
	matches := make([]records.Agent, 0, 1)
	for _, record := range agents {
		if !validRecord(record, claims) {
			continue
		}
		if record.PaneID != paneID && !hasPane(record.PaneHistory, paneID) {
			continue
		}
		matches = append(matches, record)
	}
	switch len(matches) {
	case 0:
		return records.Agent{}, ErrCallerUnknown
	case 1:
		return matches[0], nil
	default:
		return records.Agent{}, ambiguousError(matches)
	}
}

func validRecord(record records.Agent, claims map[string]records.Claim) bool {
	if record.Status != records.AgentLive {
		return false
	}
	key := claimKey(record.Role, record.Project)
	if key == "" {
		return true
	}
	claim, ok := claims[key]
	if !ok {
		return false
	}
	return claim.Key == key && claim.Agent == record.Key && claim.Generation == record.Generation
}

func hasPane(history []string, paneID string) bool {
	for _, candidate := range history {
		if candidate == paneID {
			return true
		}
	}
	return false
}

func ambiguousError(matches []records.Agent) error {
	names := make([]string, 0, len(matches))
	for _, record := range matches {
		names = append(names, displayValue(record.Name))
	}
	sort.Strings(names)
	return fmt.Errorf("caller ambiguous: %s", strings.Join(names, ", "))
}

func claimKey(role, project string) string {
	switch role {
	case "coordinator":
		return "coordinator"
	case "manager":
		return "manager-" + project
	case "supervisor":
		return "supervisor"
	default:
		return ""
	}
}

func loadAgents(source reader) ([]records.Agent, error) {
	raw, err := source.ListIn([]string{"agents"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
		found = append(found, record)
	}
	return found, nil
}

func loadClaims(source reader) ([]records.Claim, error) {
	raw, err := source.ListIn([]string{"claims"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Claim, 0, len(raw))
	for _, data := range raw {
		var claim records.Claim
		if err := json.Unmarshal(data, &claim); err != nil {
			return nil, err
		}
		found = append(found, claim)
	}
	return found, nil
}

// FindHeld returns one agent record by key, name, or claim key. It does not
// require a live status, because a task can name an agent that has finished.
func FindHeld(locked *store.LockedStore, name string) (records.Agent, bool, error) {
	agents, err := loadAgents(locked)
	if err != nil {
		return records.Agent{}, false, err
	}
	claims, err := loadClaims(locked)
	if err != nil {
		return records.Agent{}, false, err
	}
	byKey := make(map[string]records.Agent, len(agents))
	for _, record := range agents {
		byKey[record.Key] = record
	}
	for _, claim := range claims {
		if claim.Key != name {
			continue
		}
		record, ok := byKey[claim.Agent]
		if !ok {
			return records.Agent{}, false, nil
		}
		if record.Generation != claim.Generation || claimKey(record.Role, record.Project) != claim.Key {
			return records.Agent{}, false, nil
		}
		return record, true, nil
	}
	if record, ok := byKey[name]; ok {
		return record, true, nil
	}
	candidates := make([]records.Agent, 0, 1)
	for _, record := range agents {
		if record.Name != name {
			continue
		}
		candidates = append(candidates, record)
	}
	if len(candidates) == 0 {
		return records.Agent{}, false, nil
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
	return candidates[0], true, nil
}

func displayValue(value string) string {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return strconv.Quote(value)
		}
	}
	return value
}
