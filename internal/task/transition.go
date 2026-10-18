package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	PaneID string
	Fail   bool
	Reason string
	Logger *slog.Logger
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
func transition(locked *store.LockedStore, command string, options TransitionOptions) (records.Task, string, error) {
	var result records.Task
	var prior string
	err := locked.UpdateIn([]string{tasksCollection}, options.Task, func(raw json.RawMessage) (json.RawMessage, error) {
		if len(raw) == 0 {
			return nil, fmt.Errorf("task not found: %s", quoteIfControl(options.Task))
		}
		var record records.Task
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", taskRelPath(options.Task))
		}
		prior = record.Status
		result = record
		if err := applyTransition(&record, command, options); err != nil {
			return nil, err
		}
		record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		result = record
		return json.Marshal(record)
	})
	return result, prior, err
}

// applyTransition changes one record for one command.
func applyTransition(record *records.Task, command string, options TransitionOptions) error {
	key := command
	if command == "verify" && options.Fail {
		key = "verify-fail"
	}
	next, ok := transitions[record.Status][key]
	if !ok {
		return fmt.Errorf("invalid transition: %s -> %s", quoteIfControl(record.Status), quoteIfControl(command))
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
			return fmt.Errorf("not the acting agent: %s", quoteIfControl(options.Agent))
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
	log := openOperationLog(home, options.Logger)
	defer log.close()
	logCommand := "task " + command
	fail := func(phase, message string) output.Envelope {
		log.failure(logCommand, "", options.Task, options.Agent, phase, message)
		return failedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+quoteIfControl(options.Task))
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		record, prior, innerErr := transition(locked, command, options)
		if innerErr != nil {
			envelope = failedEnvelope("transition", innerErr.Error())
			log.failure(logCommand, record.Project, options.Task, options.Agent, "transition", innerErr.Error())
			return nil
		}
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: []output.Effect{},
			Error:   nil,
			Result:  record,
		}
		log.transition(logCommand, record.Project, record.ID, options.Agent, prior, record.Status)
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	return envelope
}

// failedEnvelope builds one failed outcome.
func failedEnvelope(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: oneLine(message), Result: nil}
}

// transitionWriter holds the store operations of one transition under a held handle.
type transitionWriter interface {
	UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error
	ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error)
}

// finishWriter holds each store operation of one finish group under a held handle.
type finishWriter interface {
	transitionWriter
	ReadIn(collection []string, id string, target any) error
	WriteIn(collection []string, id string, value any) error
	Next(name string) (uint64, error)
}

// finishLocked performs the finish group under one held handle.
func finishLocked(writer finishWriter, command string, options TransitionOptions) output.Envelope {
	return failedEnvelope("stub", "finish not implemented")
}
