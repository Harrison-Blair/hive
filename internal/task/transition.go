package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
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
func transition(writer transitionWriter, command string, options TransitionOptions) (records.Task, string, error) {
	var result records.Task
	var prior string
	err := writer.UpdateIn([]string{tasksCollection}, options.Task, func(raw json.RawMessage) (json.RawMessage, error) {
		if len(raw) == 0 {
			return nil, fmt.Errorf("task not found: %s", quoteIfControl(options.Task))
		}
		var record records.Task
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", taskRelPath(options.Task))
		}
		prior = record.Status
		result = record
		if command == "assign" && options.Agent != "" {
			known, found, err := identity.FindHeld(writer, options.Agent)
			if err != nil {
				return nil, err
			}
			if found && known.Finished {
				return nil, fmt.Errorf("agent finished: %s: use a fresh agent", quoteIfControl(known.Name))
			}
		}
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

// runFinish runs one complete or verify command under one lock.
func runFinish(home, command string, options TransitionOptions) output.Envelope {
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
	prior := ""
	project := ""
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var current records.Task
		if err := locked.ReadIn([]string{tasksCollection}, options.Task, &current); err == nil {
			prior = current.Status
			project = current.Project
		}
		envelope = finishLocked(locked, command, options)
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	if task, ok := envelope.Result.(records.Task); ok && envelope.Outcome == "ok" {
		log.transition(logCommand, task.Project, task.ID, options.Agent, prior, task.Status)
	} else if envelope.Error != nil {
		log.failure(logCommand, project, options.Task, options.Agent, envelope.Phase, fmt.Sprint(envelope.Error))
	}
	return envelope
}

// finishLocked performs the finish group under one held handle.
func finishLocked(writer finishWriter, command string, options TransitionOptions) output.Envelope {
	var task records.Task
	if err := writer.ReadIn([]string{tasksCollection}, options.Task, &task); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failedEnvelope("transition", "task not found: "+quoteIfControl(options.Task))
		}
		return failedEnvelope("store", err.Error())
	}
	caller, err := identity.ResolveHeld(writer, options.Agent, options.PaneID)
	if err != nil {
		return failedEnvelope("identity", err.Error())
	}
	if caller.Finished {
		return failedEnvelope("finished", advancedFinishMessage(task, caller))
	}
	if command == "complete" {
		if !sameAgent(caller, task.ActingAgent) || caller.Role != "worker" || caller.Project != task.Project || caller.Task != task.ID {
			return failedEnvelope("identity", "not the acting agent: "+quoteIfControl(caller.Name))
		}
	} else {
		if sameAgent(caller, task.ActingAgent) {
			return failedEnvelope("identity", errSelfVerify.Error())
		}
		if caller.Role != "verifier" || caller.Project != task.Project || caller.Task != task.ID {
			return failedEnvelope("identity", "not a verifier of "+task.ID+": "+quoteIfControl(caller.Name))
		}
		if options.Fail && options.Reason == "" {
			return failedEnvelope("transition", errReasonNeeded.Error())
		}
	}
	notification, err := finishNotification(command, task, caller, options)
	if err != nil {
		return failedEnvelope("validate", err.Error())
	}
	transitionOptions := options
	transitionOptions.Agent = caller.Name
	transitionOptions.PaneID = ""
	result, _, err := transition(writer, command, transitionOptions)
	if err != nil {
		return failedEnvelope("task", err.Error())
	}
	effects := []output.Effect{{Resource: "task", Path: statePath("tasks", result.ID), Action: "updated"}}
	appended := inbox.AppendHeld(writer, notification)
	if appended.Outcome != "ok" {
		effects = append(effects, appended.Effects...)
		return output.Envelope{Outcome: "partial", Phase: appended.Phase, Effects: effects, Error: appended.Error, Result: result}
	}
	effects = append(effects, appended.Effects...)
	if err := markFinished(writer, caller.Key); err != nil {
		return output.Envelope{Outcome: "partial", Phase: "finished", Effects: effects, Error: "write finished agent: " + err.Error(), Result: result}
	}
	effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", caller.Key), Action: "updated"})
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: result}
}

func finishNotification(command string, task records.Task, caller records.Agent, options TransitionOptions) (inbox.AppendOptions, error) {
	kind := records.InboxCompleted
	text := "completed"
	if command == "verify" {
		kind = records.InboxVerified
		text = "verified"
		if options.Fail {
			text = "fail: " + options.Reason
		}
	}
	notification := inbox.AppendOptions{
		Role:  task.ControllingRole,
		Kind:  kind,
		Task:  task.ID,
		Agent: caller.Name,
		Text:  text,
	}
	return notification, inbox.ValidateAppend(notification)
}

func sameAgent(record records.Agent, name string) bool {
	return name != "" && (record.Name == name || record.Key == name)
}

func advancedFinishMessage(task records.Task, caller records.Agent) string {
	return fmt.Sprintf("agent finished: %s: run task show %s, agent show %s, and inbox --role %s",
		quoteIfControl(caller.Name), task.ID, quoteIfControl(caller.Name), task.ControllingRole)
}

func statePath(collection, id string) string {
	return ".hive/state/" + collection + "/" + id + ".json"
}

func markFinished(writer finishWriter, key string) error {
	return writer.UpdateIn([]string{"agents"}, key, func(raw json.RawMessage) (json.RawMessage, error) {
		var record records.Agent
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, err
		}
		record.Finished = true
		record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return json.Marshal(record)
	})
}
