package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

var (
	// errCallerUnknown reports a transition that needs an agent name.
	errCallerUnknown = errors.New("caller unknown: pass --agent")
	// errSelfVerify reports a verifier that owns the task.
	errSelfVerify = errors.New("verifier must not be the acting agent")
	// errReasonNeeded reports a failed check without a reason.
	errReasonNeeded = errors.New("--fail needs --reason")
)

// TransitionOptions holds the values of one task transition.
type TransitionOptions struct {
	Task   string
	Agent  string
	Fail   bool
	Reason string
}

// transitions maps a status and a command to the next status. The key
// "verify-fail" is the command "verify" with the flag --fail.
var transitions = map[string]map[string]string{
	"created":   {"assign": "assigned", "cancel": "cancelled"},
	"assigned":  {"assign": "assigned", "complete": "completed", "cancel": "cancelled"},
	"completed": {"verify": "verified", "verify-fail": "assigned", "cancel": "cancelled"},
	"verified":  {},
	"cancelled": {},
}

// transition checks the status machine and changes one record under the lock.
// A failure changes nothing.
func transition(locked *store.LockedStore, command string, options TransitionOptions) (records.Task, error) {
	var result records.Task
	err := locked.UpdateIn([]string{tasksCollection}, options.Task, func(raw json.RawMessage) (json.RawMessage, error) {
		if len(raw) == 0 {
			return nil, fmt.Errorf("task not found: %s", options.Task)
		}
		var record records.Task
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", taskRelPath(options.Task))
		}
		if err := applyTransition(&record, command, options); err != nil {
			return nil, err
		}
		record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		result = record
		return json.Marshal(record)
	})
	return result, err
}

// applyTransition changes one record for one command.
func applyTransition(record *records.Task, command string, options TransitionOptions) error {
	key := command
	if command == "verify" && options.Fail {
		key = "verify-fail"
	}
	next, ok := transitions[record.Status][key]
	if !ok {
		return fmt.Errorf("invalid transition: %s -> %s", record.Status, command)
	}
	switch command {
	case "assign":
		if options.Agent == "" {
			return errCallerUnknown
		}
		record.ActingAgent = options.Agent
	case "complete":
		if options.Agent == "" {
			return errCallerUnknown
		}
		if options.Agent != record.ActingAgent {
			return fmt.Errorf("not the acting agent: %s", options.Agent)
		}
	case "verify":
		if options.Agent == "" {
			return errCallerUnknown
		}
		if options.Agent == record.ActingAgent {
			return errSelfVerify
		}
		if options.Fail && options.Reason == "" {
			return errReasonNeeded
		}
		result := "pass"
		if options.Fail {
			result = "fail"
		}
		record.Verification = &records.Verification{
			Agent:  options.Agent,
			Result: result,
			Reason: options.Reason,
			At:     time.Now().UTC().Format(time.RFC3339Nano),
		}
	case "cancel":
		record.CancelReason = options.Reason
	}
	record.Status = next
	return nil
}

// runTransition runs one transition command under one lock.
func runTransition(home, command string, options TransitionOptions) output.Envelope {
	if !records.ValidTaskID(options.Task) {
		return failedEnvelope("validate", "task not found: "+options.Task)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		record, innerErr := transition(locked, command, options)
		if innerErr != nil {
			envelope = failedEnvelope("transition", innerErr.Error())
			return nil
		}
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: []output.Effect{},
			Error:   nil,
			Result:  record,
		}
		return nil
	})
	if lockErr != nil {
		return failedEnvelope("store", lockErr.Error())
	}
	return envelope
}

// failedEnvelope builds one failed outcome.
func failedEnvelope(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
