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
	// ResolvedActor receives the caller that a finish command resolved. A
	// finish command uses this one record for the identity and the transition
	// checks. It adds no record field.
	ResolvedActor *records.Agent
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
		if !hasCaller(options) {
			return errCallerUnknown
		}
		if !isActingAgent(record.ActingAgent, options) {
			return fmt.Errorf("not the acting agent: %s", quoteIfControl(callerName(options)))
		}
	case "verify":
		if !hasCaller(options) {
			return errCallerUnknown
		}
		if isActingAgent(record.ActingAgent, options) {
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
			Agent:  callerName(options),
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
	// One fresh actor belongs to this invocation. The finish group writes it
	// for the logs. The optional caller output receives the same record after
	// the group.
	var resolved records.Agent
	callerOutput := options.ResolvedActor
	options.ResolvedActor = &resolved
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
	if callerOutput != nil && resolved.Key != "" {
		*callerOutput = resolved
	}
	actorID := options.Agent
	if resolved.Key != "" {
		actorID = resolved.Key
	}
	if task, ok := envelope.Result.(records.Task); ok && hasEffect(envelope.Effects, "task") {
		log.transition(logCommand, task.Project, task.ID, actorID, prior, task.Status)
		for _, effect := range envelope.Effects {
			role := ""
			if effect.Resource == "counter" || effect.Resource == "inbox" {
				role = task.ControllingRole
			}
			log.finishEffect(logCommand, task.Project, task.ID, actorID, role, effect)
		}
	}
	if envelope.Error != nil {
		log.failure(logCommand, project, options.Task, actorID, envelope.Phase, fmt.Sprint(envelope.Error))
	}
	return envelope
}

func hasEffect(effects []output.Effect, resource string) bool {
	for _, effect := range effects {
		if effect.Resource == resource {
			return true
		}
	}
	return false
}

// finishEffect writes one completed write of a finish group.
func (l operationLog) finishEffect(command, projectID, taskID, agentID, role string, effect output.Effect) {
	if l.logger == nil {
		return
	}
	fields := l.fields(command, projectID, taskID, agentID)
	if role != "" {
		fields = append(fields, "role", role)
	}
	fields = append(fields, "resource", effect.Resource, "path", effect.Path, "action", effect.Action)
	l.logger.Debug("task effect", fields...)
}

// LogOutputFailure writes one output stream error after a durable finish. It
// separates the successful durable effects from the failed output.
func LogOutputFailure(home, command, projectID, taskID, agentID string, outputErr error) {
	log := openOperationLog(home, nil)
	defer log.close()
	if log.logger == nil {
		return
	}
	fields := log.fields(command, projectID, taskID, agentID)
	fields = append(fields, "phase", "output", "error", outputErr.Error())
	log.logger.Error("task output error", fields...)
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
	if options.ResolvedActor != nil {
		*options.ResolvedActor = caller
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
	resolved := caller
	transitionOptions := options
	transitionOptions.ResolvedActor = &resolved
	transitionOptions.Agent = caller.Name
	transitionOptions.PaneID = ""
	result, _, err := transition(writer, command, transitionOptions)
	if err != nil {
		if task.Status == "created" {
			return failedEnvelope("task", err.Error())
		}
		return failedEnvelope("task", finishInspectionMessage(task, caller, err.Error()))
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

func hasCaller(options TransitionOptions) bool {
	return options.ResolvedActor != nil || options.Agent != ""
}

func callerName(options TransitionOptions) string {
	if options.ResolvedActor != nil {
		return options.ResolvedActor.Name
	}
	return options.Agent
}

func isActingAgent(acting string, options TransitionOptions) bool {
	if options.ResolvedActor != nil {
		return sameAgent(*options.ResolvedActor, acting)
	}
	return acting != "" && acting == options.Agent
}

func advancedFinishMessage(task records.Task, caller records.Agent) string {
	return fmt.Sprintf("agent finished: %s: %s",
		quoteIfControl(caller.Name), finishInspectionCommands(task, caller))
}

func finishInspectionMessage(task records.Task, caller records.Agent, message string) string {
	return message + "; " + finishInspectionCommands(task, caller)
}

func finishInspectionCommands(task records.Task, caller records.Agent) string {
	return fmt.Sprintf("run task show %s, agent show %s, and inbox --role %s",
		task.ID, quoteIfControl(caller.Name), task.ControllingRole)
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
