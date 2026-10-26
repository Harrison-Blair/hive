// Package reply saves one coordinator reply in the chat history.
package reply

import (
	"context"
	"fmt"
	"time"

	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Options holds the caller and text of one coordinator reply.
type Options struct {
	Agent  string
	PaneID string
	To     string
	Text   string
}

// Result reports one coordinator reply.
type Result struct {
	Entry chatlib.Entry `json:"entry"`
}

// Reply appends one message from the live coordinator.
func Reply(home string, options Options) output.Envelope {
	if options.Text == "" {
		return chatlib.Failed("validate", "chat reply text is empty", nil)
	}
	registry, err := store.Open(home)
	if err != nil {
		return chatlib.Failed("store", err.Error(), nil)
	}
	var entry chatlib.Entry
	var phase, message string
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		caller, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil || caller.Kind != records.AgentKindCoordinator {
			phase = "identity"
			message = "only the coordinator can run hive chat reply"
			if err != nil {
				message += ": " + err.Error()
			}
			return nil
		}
		if options.To != "" {
			if !chatlib.IDPattern.MatchString(options.To) {
				phase, message = "validate", "chat user message not found: "+options.To
				return nil
			}
			existing, _, err := chatlib.ReadEntriesLocked(home, "")
			if err != nil {
				phase, message = "read", "read chat history: "+err.Error()
				return nil
			}
			found := false
			for _, candidate := range existing {
				if candidate.ID == options.To && candidate.From == "user" {
					found = true
					break
				}
			}
			if !found {
				phase, message = "validate", "chat user message not found: "+options.To
				return nil
			}
		}
		at := time.Now().UTC().Format(time.RFC3339Nano)
		seq, err := locked.Next("chat")
		if err != nil {
			phase, message = "counter", "next chat ID: "+err.Error()
			return nil
		}
		entry = chatlib.Entry{ID: fmt.Sprintf("c%d", seq), At: at, From: "coordinator", Text: options.Text, ReplyTo: options.To}
		if err := chatlib.AppendLine(home, at[:10], entry); err != nil {
			phase, message = "write", "append chat reply: "+err.Error()
		}
		return nil
	})
	if lockErr != nil {
		return chatlib.Failed("store", lockErr.Error(), nil)
	}
	if message != "" {
		return chatlib.Failed(phase, message, nil)
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: Result{Entry: entry}}
}

// Text renders the ID of one coordinator reply.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok {
		return ""
	}
	return fmt.Sprintf("chat %s reply saved\n", result.Entry.ID)
}
