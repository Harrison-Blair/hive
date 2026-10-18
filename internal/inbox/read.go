package inbox

import (
	"context"
	"errors"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ReadOptions holds the values of one inbox read.
type ReadOptions struct {
	Role    string
	RoleSet bool
	All     bool
	Agent   string
	PaneID  string
}

// Read returns all records or consumes the unread records of one role.
func Read(home string, options ReadOptions) output.Envelope {
	if options.RoleSet {
		return readKey(home, options.Role, options.All)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failedRead("store", err.Error())
	}
	var envelope output.Envelope
	var key string
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		caller, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil {
			if errors.Is(err, identity.ErrCallerUnknown) {
				envelope = failedRead("validate", "caller unknown: pass --role or pass --agent")
			} else {
				envelope = failedRead("identity", err.Error())
			}
			return nil
		}
		role, ok := roleKey(caller)
		if !ok {
			envelope = failedRead("validate", "role has no inbox: "+caller.Role)
			return nil
		}
		key = role
		if options.All {
			found, err := inboxlib.AllHeld(locked, role)
			if err != nil {
				envelope = failedRead("read", err.Error())
				return nil
			}
			envelope = output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: []output.Effect{},
				Error:   nil,
				Result:  inboxlib.Result{Records: found},
			}
			return nil
		}
		envelope = inboxlib.ConsumeHeld(locked, role)
		return nil
	})
	if lockErr != nil {
		return failedRead("store", lockErr.Error())
	}
	if key != "" {
		inboxlib.LogConsume(home, key, envelope)
	}
	return envelope
}

func readKey(home, role string, all bool) output.Envelope {
	if !all {
		return inboxlib.Consume(home, role)
	}
	found, err := inboxlib.All(home, role)
	if err != nil {
		return failedRead("read", err.Error())
	}
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{},
		Error:   nil,
		Result:  inboxlib.Result{Records: found},
	}
}

func roleKey(record records.Agent) (string, bool) {
	switch record.Role {
	case "coordinator":
		return "coordinator", true
	case "manager":
		return "manager-" + record.Project, true
	default:
		return "", false
	}
}

func failedRead(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
