package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// StartSupervisor takes the supervisor claim and registers the supervisor
// under one held store lock. It checks the claim with process.Alive before it
// registers. A live claim fails the start.
func StartSupervisor(home string, options RegisterOptions) output.Envelope {
	options.Kind = records.AgentKindSupervisor
	options.Profile = ""
	options.Owns = ""
	options.Project = ""
	options.Task = ""
	if message := validateRegistration(options); message != "" {
		envelope := Failed("validate", message)
		LogRegistration(home, options, envelope)
		return envelope
	}
	registry, err := store.Open(home)
	if err != nil {
		envelope := Failed("store", err.Error())
		LogRegistration(home, options, envelope)
		return envelope
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var claim records.Claim
		claimErr := locked.ReadIn([]string{"claims"}, "supervisor", &claim)
		switch {
		case claimErr == nil:
			alive, aliveErr := process.Alive(claim.PID, claim.PIDStart)
			if aliveErr == nil && alive {
				envelope = Failed("claim", fmt.Sprintf("supervisor live: pid %d", claim.PID))
				return nil
			}
		case errors.Is(claimErr, os.ErrNotExist):
		default:
			envelope = Failed("store", "read supervisor claim: "+claimErr.Error())
			return nil
		}
		envelope = registerLocked(locked, locked, options)
		return nil
	})
	if err != nil {
		envelope = Failed("store", err.Error())
	}
	LogRegistration(home, options, envelope)
	return envelope
}
