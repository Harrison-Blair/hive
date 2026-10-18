package inbox

import (
	"context"
	"errors"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Send appends one inbox record and defaults the sender from the caller.
func Send(home string, options inboxlib.AppendOptions, paneID string) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return failedRead("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if options.Agent == "" {
			caller, err := identity.ResolveHeld(locked, "", paneID)
			if err != nil {
				if errors.Is(err, identity.ErrCallerUnknown) {
					envelope = failedRead("validate", "caller unknown: pass --agent")
				} else {
					envelope = failedRead("identity", err.Error())
				}
				return nil
			}
			options.Agent = caller.Name
		}
		envelope = inboxlib.AppendHeld(locked, options)
		return nil
	})
	if lockErr != nil {
		return failedRead("store", lockErr.Error())
	}
	inboxlib.LogAppend(home, options, envelope)
	return envelope
}
