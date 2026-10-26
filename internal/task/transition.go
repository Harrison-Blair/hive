package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	// errReasonNeeded reports a failed check without a reason.
	errReasonNeeded = errors.New("--fail needs --reason")
	// errReopenReasonNeeded reports a reopen without a reason.
	errReopenReasonNeeded = errors.New("--reason is required")
)

// TransitionOptions holds the values of one task transition.
type TransitionOptions struct {
	Task   string
	Agent  string
	PaneID string
	Fail   bool
	Reason string
	Logger *slog.Logger
	// ResolvedActor receives the caller that a command resolved. A finish
	// command uses this record for identity and transition checks. It adds no
	// task record field.
	ResolvedActor *records.Agent
}

// transitions maps a status and a command to the next status. The key
// "verify-fail" is the command "verify" with the flag --fail.
var transitions = map[string]map[string]string{
	"created":   {"assign": "assigned", "cancel": "cancelled"},
	"assigned":  {"assign": "assigned", "complete": "completed", "cancel": "cancelled"},
	"completed": {"verify": "verified", "verify-fail": "assigned", "cancel": "cancelled"},
	"verified":  {"cancel": "cancelled", "reopen": "assigned"},
	"cancelled": {},
}

// transition checks the status machine and changes one record under the lock.
// A failure changes nothing.
func transition(writer transitionWriter, command string, options TransitionOptions) (records.Task, string, error) {
	var result records.Task
	var prior string
	err := writer.UpdateIn([]string{TasksCollection}, options.Task, func(raw json.RawMessage) (json.RawMessage, error) {
		if len(raw) == 0 {
			return nil, fmt.Errorf("task not found: %s", QuoteIfControl(options.Task))
		}
		var record records.Task
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", TaskRelPath(options.Task))
		}
		prior = record.Status
		result = record
		if command == "reopen" {
			if _, ok := transitions[record.Status][command]; ok {
				if options.ResolvedActor == nil {
					return nil, errCallerUnknown
				}
				caller, err := identity.ResolveHeld(writer, options.Agent, options.PaneID)
				if err != nil {
					return nil, err
				}
				*options.ResolvedActor = caller
			}
		}
		if command == "assign" && options.Agent != "" {
			known, found, err := identity.FindHeld(writer, options.Agent)
			if err != nil {
				return nil, err
			}
			if found && known.Finished {
				return nil, fmt.Errorf("agent finished: %s: use a fresh agent", QuoteIfControl(known.Name))
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

// CheckOwner refuses a caller that is not the owner of the project of the
// task and not the coordinator (D7).
func CheckOwner(task records.Task, caller records.Agent) error {
	if caller.Kind == records.AgentKindCoordinator || caller.Owns == task.Project {
		return nil
	}
	return fmt.Errorf("not the owner of project %s or the coordinator: %s", task.Project, QuoteIfControl(caller.Name))
}

// applyTransition changes one record for one command.
func applyTransition(record *records.Task, command string, options TransitionOptions) error {
	key := command
	if command == "verify" && options.Fail {
		key = "verify-fail"
	}
	next, ok := transitions[record.Status][key]
	if !ok {
		return fmt.Errorf("invalid transition: %s -> %s", QuoteIfControl(record.Status), QuoteIfControl(command))
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
		if !isActingAgent(record.ActingAgent, options) || (options.ResolvedActor != nil && options.ResolvedActor.Task != record.ID) {
			return fmt.Errorf("not the acting agent of %s: %s", QuoteIfControl(record.ID), QuoteIfControl(callerName(options)))
		}
	case "verify":
		if !hasCaller(options) {
			return errCallerUnknown
		}
		if options.ResolvedActor != nil && options.ResolvedActor.Task != record.ID {
			return fmt.Errorf("not bound to %s: %s", QuoteIfControl(record.ID), QuoteIfControl(callerName(options)))
		}
		if isActingAgent(record.ActingAgent, options) {
			return fmt.Errorf("self verify refused: %s is the acting agent of %s", QuoteIfControl(callerName(options)), QuoteIfControl(record.ID))
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
		if record.Status == "verified" && record.Delivery != nil && record.Delivery.Pushed {
			return errors.New("task is landed: cancel refused")
		}
		record.CancelReason = options.Reason
	case "reopen":
		if strings.TrimSpace(options.Reason) == "" {
			return errReopenReasonNeeded
		}
		if options.ResolvedActor == nil {
			return errCallerUnknown
		}
		if err := CheckOwner(*record, *options.ResolvedActor); err != nil {
			return err
		}
		record.VerificationHistory = append(record.VerificationHistory, records.VerificationHistory{
			Verification: record.Verification,
			Reason:       options.Reason,
			Caller:       options.ResolvedActor.Name,
			At:           time.Now().UTC().Format(time.RFC3339Nano),
		})
		record.Verification = nil
	}
	record.Status = next
	return nil
}

// RunTransition runs one transition command under one lock.
func RunTransition(home, command string, options TransitionOptions) output.Envelope {
	log := OpenLog(home, options.Logger)
	defer log.Close()
	logCommand := "task " + command
	actorID := options.Agent
	if options.ResolvedActor != nil && options.ResolvedActor.Key != "" {
		actorID = options.ResolvedActor.Key
	}
	fail := func(phase, message string) output.Envelope {
		log.Failure(logCommand, "", options.Task, actorID, phase, message)
		return FailedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+QuoteIfControl(options.Task))
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		record, prior, innerErr := transition(locked, command, options)
		if innerErr != nil {
			envelope = FailedEnvelope("transition", innerErr.Error())
			if options.ResolvedActor != nil && options.ResolvedActor.Key != "" {
				actorID = options.ResolvedActor.Key
			}
			log.Failure(logCommand, record.Project, options.Task, actorID, "transition", innerErr.Error())
			return nil
		}
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: []output.Effect{},
			Error:   nil,
			Result:  record,
		}
		if options.ResolvedActor != nil && options.ResolvedActor.Key != "" {
			actorID = options.ResolvedActor.Key
		}
		if command == "reopen" {
			log.Transition(logCommand, record.Project, record.ID, actorID, prior, record.Status, "reason", options.Reason)
		} else {
			log.Transition(logCommand, record.Project, record.ID, actorID, prior, record.Status)
		}
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	return envelope
}

// FailedEnvelope builds one failed outcome.
func FailedEnvelope(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: OneLine(message), Result: nil}
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

// RunFinish runs one complete or verify command under one lock.
func RunFinish(home, command string, options TransitionOptions) output.Envelope {
	log := OpenLog(home, options.Logger)
	defer log.Close()
	logCommand := "task " + command
	fail := func(phase, message string) output.Envelope {
		log.Failure(logCommand, "", options.Task, options.Agent, phase, message)
		return FailedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+QuoteIfControl(options.Task))
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
		if err := locked.ReadIn([]string{TasksCollection}, options.Task, &current); err == nil {
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
	if task, ok := envelope.Result.(records.Task); ok && envelope.Outcome == "ok" {
		copied, err := copyEvidenceOut(home, command, task)
		envelope.Effects = append(envelope.Effects, copied...)
		if err != nil {
			envelope.Outcome = "partial"
			envelope.Phase = "evidence"
			envelope.Error = OneLine(err.Error())
		}
	}
	if task, ok := envelope.Result.(records.Task); ok && hasEffect(envelope.Effects, "task") {
		log.Transition(logCommand, task.Project, task.ID, actorID, prior, task.Status)
		for _, effect := range envelope.Effects {
			inbox := ""
			if effect.Resource == "counter" || effect.Resource == "inbox" {
				inbox = ProjectAddress(task.Project)
			}
			log.finishEffect(logCommand, task.Project, task.ID, actorID, inbox, effect)
		}
	}
	if envelope.Error != nil {
		log.Failure(logCommand, project, options.Task, actorID, envelope.Phase, fmt.Sprint(envelope.Error))
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
func (l Log) finishEffect(command, projectID, taskID, agentID, inbox string, effect output.Effect) {
	if l.Logger == nil {
		return
	}
	fields := l.Fields(command, projectID, taskID, agentID)
	if inbox != "" {
		fields = append(fields, "inbox", inbox)
	}
	fields = append(fields, "resource", effect.Resource, "path", effect.Path, "action", effect.Action)
	l.Logger.Debug("task effect", fields...)
}

// LogOutputFailure writes one output stream error after a durable finish. It
// separates the successful durable effects from the failed output.
func LogOutputFailure(home, command, projectID, taskID, agentID string, outputErr error) {
	log := OpenLog(home, nil)
	defer log.Close()
	if log.Logger == nil {
		return
	}
	fields := log.Fields(command, projectID, taskID, agentID)
	fields = append(fields, "phase", "output", "error", outputErr.Error())
	log.Logger.Error("task output error", fields...)
}

// finishLocked performs the finish group under one held handle.
func finishLocked(writer finishWriter, command string, options TransitionOptions) output.Envelope {
	var task records.Task
	if err := writer.ReadIn([]string{TasksCollection}, options.Task, &task); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FailedEnvelope("transition", "task not found: "+QuoteIfControl(options.Task))
		}
		return FailedEnvelope("store", err.Error())
	}
	caller, err := identity.ResolveHeld(writer, options.Agent, options.PaneID)
	if err != nil {
		return FailedEnvelope("identity", err.Error())
	}
	if options.ResolvedActor != nil {
		*options.ResolvedActor = caller
	}
	if caller.Finished {
		return FailedEnvelope("finished", advancedFinishMessage(task, caller))
	}
	if command == "complete" {
		if !sameAgent(caller, task.ActingAgent) || caller.Task != task.ID {
			return FailedEnvelope("identity", fmt.Sprintf("not the acting agent of %s: %s", QuoteIfControl(task.ID), QuoteIfControl(caller.Name)))
		}
	} else {
		if caller.Task != task.ID {
			return FailedEnvelope("identity", fmt.Sprintf("not bound to %s: %s", QuoteIfControl(task.ID), QuoteIfControl(caller.Name)))
		}
		if sameAgent(caller, task.ActingAgent) {
			return FailedEnvelope("identity", fmt.Sprintf("self verify refused: %s is the acting agent of %s", QuoteIfControl(caller.Name), QuoteIfControl(task.ID)))
		}
		if options.Fail && options.Reason == "" {
			return FailedEnvelope("transition", errReasonNeeded.Error())
		}
	}
	notification, err := finishNotification(command, task, caller, options)
	if err != nil {
		return FailedEnvelope("validate", err.Error())
	}
	resolved := caller
	transitionOptions := options
	transitionOptions.ResolvedActor = &resolved
	transitionOptions.Agent = caller.Name
	transitionOptions.PaneID = ""
	result, _, err := transition(writer, command, transitionOptions)
	if err != nil {
		if task.Status == "created" {
			return FailedEnvelope("task", err.Error())
		}
		return FailedEnvelope("task", finishInspectionMessage(task, caller, err.Error()))
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
		Inbox: ProjectAddress(task.Project),
		Kind:  kind,
		Task:  task.ID,
		Agent: caller.Name,
		Text:  text,
	}
	return notification, inbox.ValidateAppend(notification)
}

// ProjectAddress returns the address that routes mail to the owner of one
// project (D7).
func ProjectAddress(project string) string {
	return "project:" + project
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
		QuoteIfControl(caller.Name), finishInspectionCommands(task, caller))
}

func finishInspectionMessage(task records.Task, caller records.Agent, message string) string {
	return message + "; " + finishInspectionCommands(task, caller)
}

func finishInspectionCommands(task records.Task, caller records.Agent) string {
	return fmt.Sprintf("run task show %s, agent show %s, and inbox --inbox %s --all",
		task.ID, QuoteIfControl(caller.Name), ProjectAddress(task.Project))
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
