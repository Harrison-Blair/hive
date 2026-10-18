package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// RegisterOptions holds explicit values for one agent registration.
type RegisterOptions struct {
	Role          string
	Project       string
	Task          string
	Harness       string
	Model         string
	Name          string
	NameSet       bool
	PaneID        string
	WorkspaceID   string
	TabID         string
	QuietAfterMin int
	QuietAfterSet bool
	PID           int64
}

type lockedWriter interface {
	ReadIn(collection []string, id string, target any) error
	WriteIn(collection []string, id string, value any) error
	UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error
	ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error)
	Next(name string) (uint64, error)
}

// Register validates and records one agent session.
func Register(home string, options RegisterOptions) output.Envelope {
	return failed("stub", "register not implemented")
}

func registerLocked(locked *store.LockedStore, writer lockedWriter, options RegisterOptions) output.Envelope {
	_ = context.Background()
	_ = locked
	_ = writer
	return failed("stub", "register not implemented")
}

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}

var _ = fmt.Sprintf
