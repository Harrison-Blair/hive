package inbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
				envelope = failedRead("validate", "caller unknown: pass --agent or run inside herdr")
			} else {
				envelope = failedRead("identity", err.Error())
			}
			return nil
		}

		ownKey, hasInbox := roleKey(caller)
		if !options.RoleSet && !hasInbox {
			envelope = failedRead("validate", "role has no inbox: "+caller.Role)
			return nil
		}
		key = ownKey
		if options.RoleSet {
			key = options.Role
			if err := inboxlib.ValidateAppend(inboxlib.AppendOptions{Role: key}); err != nil {
				envelope = failedRead("validate", err.Error())
				return nil
			}
		}

		ownerRead := hasInbox && key == ownKey
		coordinatorRead := caller.Role == "coordinator" && options.RoleSet && managerProject(key) != ""
		if !ownerRead && !coordinatorRead {
			envelope = failedRead("validate", fmt.Sprintf("inbox %s belongs to %s", key, inboxOwner(key)))
			return nil
		}

		if options.All || coordinatorRead {
			found, err := inboxlib.AllHeld(locked, key)
			if err != nil {
				envelope = failedRead("read", err.Error())
				return nil
			}
			if coordinatorRead && !options.All {
				unread := found[:0]
				for _, record := range found {
					if record.ReadAt == "" {
						unread = append(unread, record)
					}
				}
				found = unread
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
		envelope = inboxlib.ConsumeHeld(locked, key)
		return nil
	})
	if lockErr != nil {
		return failedRead("store", lockErr.Error())
	}
	inboxlib.LogConsume(home, key, envelope)
	return envelope
}

func managerProject(key string) string {
	project := strings.TrimPrefix(key, "manager-")
	if project == key {
		return ""
	}
	return project
}

func inboxOwner(key string) string {
	if key == "coordinator" {
		return "the coordinator"
	}
	if project := managerProject(key); project != "" {
		return "the manager of project " + project
	}
	return "the owner"
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
