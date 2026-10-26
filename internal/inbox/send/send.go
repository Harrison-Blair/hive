// Package send appends one inbox message record.
package send

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	inboxlogic "github.com/Harrison-Blair/hive/internal/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/identity"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Run appends one inbox record and defaults the sender from the caller. In a
// herdr pane, the sender is always the agent of herdrPane.
func Run(home string, options inboxlib.AppendOptions, paneID, herdrPane string) output.Envelope {
	if err := inboxlib.ValidateAppend(options); err != nil {
		envelope := inboxlogic.Failed("validate", err.Error())
		inboxlib.LogAppend(home, options, envelope)
		return envelope
	}
	registry, err := store.Open(home)
	if err != nil {
		return inboxlogic.Failed("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if options.Agent == "" || herdrPane != "" {
			caller, err := agent.Caller(locked, agent.CallerOptions{Agent: options.Agent, Pane: paneID, HerdrPane: herdrPane})
			if err != nil {
				if errors.Is(err, identity.ErrCallerUnknown) {
					envelope = inboxlogic.Failed("validate", "caller unknown: pass --agent")
				} else {
					envelope = inboxlogic.Failed("identity", err.Error())
				}
				return nil
			}
			options.Agent = caller.Name
		}
		if err := inboxlib.ValidateAppend(options); err != nil {
			envelope = inboxlogic.Failed("validate", err.Error())
			return nil
		}
		envelope = inboxlib.AppendHeld(locked, options)
		return nil
	})
	if lockErr != nil {
		return inboxlogic.Failed("store", lockErr.Error())
	}
	inboxlib.LogAppend(home, options, envelope)
	return envelope
}

// Text renders the effects of one append.
func Text(envelope output.Envelope) string {
	lines := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
