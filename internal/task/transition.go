package task

import (
	"context"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// TransitionOptions holds the values of one task transition.
type TransitionOptions struct {
	Task   string
	Agent  string
	Fail   bool
	Reason string
}

// transitions maps a status and a command to the next status.
var transitions = map[string]map[string]string{}

// Assign moves one task to assigned.
func Assign(home string, options TransitionOptions) output.Envelope {
	_ = home
	_ = options
	return output.Envelope{}
}

// Complete moves one task to completed.
func Complete(home string, options TransitionOptions) output.Envelope {
	_ = home
	_ = options
	return output.Envelope{}
}

// Verify records one independent check of a task.
func Verify(home string, options TransitionOptions) output.Envelope {
	_ = home
	_ = options
	return output.Envelope{}
}

// Cancel moves one task to cancelled.
func Cancel(home string, options TransitionOptions) output.Envelope {
	_ = home
	_ = options
	return output.Envelope{}
}

// transition checks the status machine and changes one record.
func transition(locked *store.LockedStore, command string, options TransitionOptions) (records.Task, error) {
	_, _, _ = locked, command, options
	return records.Task{}, nil
}

// runTransition runs one transition command under one lock.
func runTransition(home, command string, options TransitionOptions) output.Envelope {
	_, _, _ = home, command, options
	return output.Envelope{}
}

// failedEnvelope builds one failed outcome.
func failedEnvelope(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}

func _unusedContext() { _ = context.Background() }
