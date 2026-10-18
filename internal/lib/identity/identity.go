// Package identity resolves the caller of one command from an agent name or a
// herdr pane ID.
package identity

import (
	"errors"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ErrCallerUnknown reports a caller that no live agent record matches.
var ErrCallerUnknown = errors.New("caller unknown: pass --agent")

// ResolveHeld returns the caller record from one held store handle.
func ResolveHeld(locked *store.LockedStore, agentFlag, paneID string) (records.Agent, error) {
	return records.Agent{}, ErrCallerUnknown
}

// Resolve returns the caller record without a caller-held store handle.
func Resolve(home, agentFlag, paneID string) (records.Agent, error) {
	return records.Agent{}, ErrCallerUnknown
}

// FindHeld returns one agent record by key, name, or claim key.
func FindHeld(locked *store.LockedStore, name string) (records.Agent, bool, error) {
	return records.Agent{}, false, nil
}
