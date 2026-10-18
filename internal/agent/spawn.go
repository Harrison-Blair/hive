package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	maxExtraArgs  = 64
	maxExtraBytes = 512
	waitSettled   = "idle, blocked, done"
)

var (
	errRecordChanged = errors.New("agent record changed")
	settledStatuses  = []string{"idle", "blocked", "done"}
)

// SpawnOptions holds explicit values for one agent spawn.
type SpawnOptions struct {
	Role          string
	Project       string
	Task          string
	Harness       string
	Model         string
	Name          string
	NameSet       bool
	PaneID        string
	PaneSet       bool
	QuietAfterMin int
	QuietAfterSet bool
	ExtraArgs     []string
	PID           int64
}

// SpawnResult is the record and optional claim of one spawn.
type SpawnResult struct {
	Agent records.Agent  `json:"agent"`
	Claim *records.Claim `json:"claim,omitempty"`
}

// Spawn writes the intent record, starts one agent, and records its identity.
func Spawn(ctx context.Context, home string, options SpawnOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	paneID, paneError := choosePane(options)
	if paneError != "" {
		return failed("validate", paneError)
	}
	profile, err := harness.Lookup(options.Harness)
	if err != nil {
		return failed("validate", err.Error())
	}
	if err := harness.Check(profile); err != nil {
		return failed("validate", err.Error())
	}
	if message := validateExtraArgs(options.ExtraArgs); message != "" {
		return failed("validate", message)
	}
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return failed("log", err.Error())
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)

	registered := Register(home, RegisterOptions{
		Role:          options.Role,
		Project:       options.Project,
		Task:          options.Task,
		Harness:       options.Harness,
		Model:         options.Model,
		Name:          options.Name,
		NameSet:       options.NameSet,
		PaneID:        paneID,
		QuietAfterMin: options.QuietAfterMin,
		QuietAfterSet: options.QuietAfterSet,
		PID:           options.PID,
		Spawned:       true,
	})
	if registered.Outcome != "ok" {
		logSpawn(logger, registered, options, "")
		return registered
	}
	result, ok := registered.Result.(RegisterResult)
	if !ok {
		envelope := output.Envelope{Outcome: "failed", Phase: "agent", Effects: registered.Effects, Error: "register result missing", Result: nil}
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	record := result.Agent
	effects := append([]output.Effect{}, registered.Effects...)
	logger.Debug("agent spawn intent", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "role", options.Role)

	args := profile.Args(options.Model, options.ExtraArgs)
	started, startErr := herdr.AgentStart(ctx, herdr.AgentStartParams{
		Name:      record.Key,
		Kind:      profile.Kind(),
		PaneID:    paneID,
		Args:      args,
		TimeoutMS: uint64(profile.StartTimeout() / time.Millisecond),
	}, herdr.WithLogger(logger))
	if startErr != nil {
		return spawnStartFailure(home, logger, options, record, result.Claim, effects, startErr)
	}
	effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "started"})
	logArgv(logger, record.Key, started.Argv)

	paneInfo, paneErr := herdr.PaneGet(ctx, paneID, herdr.WithLogger(logger))
	if paneErr != nil {
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, "pane get: "+paneErr.Error(), "unresolved")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("partial", "pane", effects, "pane get: "+paneErr.Error(), SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		if updateErr != nil {
			envelope.Error = "pane get: " + paneErr.Error() + "; " + updateErr.Error()
		}
		return envelope
	}
	cwd := paneInfo.Cwd
	if cwd == "" {
		cwd = paneInfo.ForegroundCwd
	}
	startSession := started.Agent.AgentSession
	if startSession != nil && paneInfo.AgentSession != nil && sessionRef(startSession) != sessionRef(paneInfo.AgentSession) {
		message := "the agent session references differ"
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, message, "unresolved")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("partial", "pane", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		if updateErr != nil {
			envelope.Error = message + "; " + updateErr.Error()
		}
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	firstSession := startSession
	if firstSession == nil {
		firstSession = paneInfo.AgentSession
	}
	updated, effect, updateErr := mutateAgentRecord(home, record, records.AgentLive, func(current *records.Agent) {
		current.TerminalID = paneInfo.TerminalID
		current.WorkspaceID = paneInfo.WorkspaceID
		current.TabID = paneInfo.TabID
		current.Cwd = cwd
		if firstSession != nil {
			current.AgentSessionRef = sessionRef(firstSession)
		}
	}, "updated")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	record = updated

	waitInfo, waitErr := herdr.AgentWait(ctx, herdr.AgentWaitParams{
		Target:    paneID,
		Until:     settledStatuses,
		TimeoutMS: uint64(profile.ReadyTimeout() / time.Millisecond),
	}, herdr.WithDeadline(profile.ReadyTimeout()+5*time.Second), herdr.WithLogger(logger))

	getInfo, getErr := herdr.AgentGet(ctx, paneID, herdr.WithLogger(logger))
	if getErr != nil {
		logger.Debug("agent reply has no name", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "error", getErr.Error())
	} else if getInfo.Name == "" {
		logger.Debug("agent reply has a missing name", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID)
	} else {
		logger.Debug("agent named", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "name", getInfo.Name)
	}

	if waitErr != nil {
		return spawnWaitFailure(ctx, home, logger, options, record, result.Claim, effects, profile, waitInfo, waitErr, getInfo, getErr)
	}
	status := waitInfo.AgentStatus
	if status != "idle" && status != "done" && status != "blocked" {
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, "unexpected agent status: "+status, "unresolved")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("partial", "wait", effects, "unexpected agent status: "+status, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		if updateErr != nil {
			envelope.Error = updateErr.Error()
		}
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	processInfo, processErr := herdr.PaneProcessInfoGet(ctx, paneID, herdr.WithLogger(logger))
	identity, identityError := captureIdentity(profile, record, processInfo, processErr, waitInfo.AgentSession, getInfo.AgentSession)
	if identityError != "" {
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, "agent process identity unresolved: "+identityError, "unresolved")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("partial", "process", effects, "agent process identity unresolved: "+identityError, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		if updateErr != nil {
			envelope.Error = "agent process identity unresolved: " + identityError + "; " + updateErr.Error()
		}
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	updated, effect, updateErr = mutateAgentRecord(home, record, records.AgentLive, func(current *records.Agent) {
		current.ProcessPID = identity.pid
		current.ProcessName = identity.name
		current.ProcessStart = identity.start
		if identity.session != nil {
			current.AgentSessionRef = sessionRef(identity.session)
		}
	}, "updated")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	record = updated
	logger.Debug("agent process identity", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "process_pid", record.ProcessPID, "process_name", record.ProcessName, "process_start", record.ProcessStart)

	if status == "blocked" {
		message := "agent blocked: accept the dialog in " + paneID
		updated, effect, _ = setRecordStatus(home, record, records.AgentLive, records.AgentLive, message, "updated")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("partial", "wait", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	updated, effect, updateErr = setRecordStatus(home, record, records.AgentLive, records.AgentLive, "", "updated")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	envelope := spawnEnvelope("ok", "complete", effects, nil, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

func choosePane(options SpawnOptions) (string, string) {
	if options.PaneSet && options.PaneID != "" {
		return options.PaneID, ""
	}
	switch options.Role {
	default:
		return "", "pane required: pass --pane"
	}
}

func validateExtraArgs(args []string) string {
	if len(args) > maxExtraArgs {
		return fmt.Sprintf("too many --arg values: %d", len(args))
	}
	for _, arg := range args {
		if len(arg) > maxExtraBytes {
			return fmt.Sprintf("--arg is longer than %d bytes", maxExtraBytes)
		}
		if !utf8.ValidString(arg) {
			return "invalid --arg: not UTF-8"
		}
		for _, value := range arg {
			if unicode.IsControl(value) {
				return fmt.Sprintf("invalid --arg: %q", arg)
			}
		}
	}
	return ""
}

func spawnStartFailure(home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, startErr error) output.Envelope {
	var serverError *herdr.Error
	if errors.As(startErr, &serverError) {
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentExited, startErr.Error(), "exited")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("failed", "start", effects, startErr.Error(), SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
		if updateErr != nil {
			envelope.Error = startErr.Error() + "; " + updateErr.Error()
		}
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, startErr.Error(), "unresolved")
	effects = appendEffect(effects, effect)
	envelope := spawnEnvelope("partial", "start", effects, "agent start: "+startErr.Error(), SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
	if updateErr != nil {
		envelope.Error = "agent start: " + startErr.Error() + "; " + updateErr.Error()
	}
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

func spawnWaitFailure(ctx context.Context, home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, profile harness.Profile, waitInfo herdr.AgentInfo, waitErr error, getInfo herdr.AgentInfo, getErr error) output.Envelope {
	var serverError *herdr.Error
	if errors.As(waitErr, &serverError) {
		switch serverError.Code {
		case herdr.CodeTimeout:
			message := fmt.Sprintf("agent not ready after %d s", int(profile.ReadyTimeout()/time.Second))
			processInfo, processErr := herdr.PaneProcessInfoGet(ctx, record.PaneID, herdr.WithLogger(logger))
			identity, identityError := captureIdentity(profile, record, processInfo, processErr, waitInfo.AgentSession, getInfo.AgentSession)
			if getErr == nil && identityError == "" {
				updated, effect, updateErr := mutateAgentRecord(home, record, records.AgentLive, func(current *records.Agent) {
					current.ProcessPID = identity.pid
					current.ProcessName = identity.name
					current.ProcessStart = identity.start
					current.LastError = message
					if identity.session != nil {
						current.AgentSessionRef = sessionRef(identity.session)
					}
				}, "updated")
				effects = appendEffect(effects, effect)
				envelope := spawnEnvelope("partial", "wait", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
				if updateErr != nil {
					envelope.Error = message + "; " + updateErr.Error()
				}
				logSpawn(logger, envelope, options, record.Key)
				return envelope
			}
			reason := "independent ownership proof failed"
			if identityError != "" {
				reason = "agent process identity unresolved: " + identityError
			} else if getErr != nil {
				reason = "agent get: " + getErr.Error()
			}
			return spawnUnresolved(home, logger, options, record, claim, effects, reason)
		case herdr.CodeAgentNotRunning, herdr.CodeAgentNotFound:
			inspection := inspectAgent(ctx, home, record, logger)
			effects = appendEffect(effects, inspection.effect)
			envelope := spawnEnvelope("failed", "wait", effects, inspection.messageForStart(), SpawnResult{Agent: chooseRecord(inspection.record, record), Claim: claim})
			logSpawn(logger, envelope, options, record.Key)
			return envelope
		}
	}
	return spawnUnresolved(home, logger, options, record, claim, effects, "agent wait: "+waitErr.Error())
}

func spawnUnresolved(home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, message string) output.Envelope {
	updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, message, "unresolved")
	effects = appendEffect(effects, effect)
	envelope := spawnEnvelope("partial", "wait", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
	if updateErr != nil {
		envelope.Error = message + "; " + updateErr.Error()
	}
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

type identity struct {
	pid     int64
	name    string
	start   uint64
	session *herdr.AgentSessionInfo
}

func captureIdentity(profile harness.Profile, record records.Agent, info herdr.PaneProcessInfo, processErr error, sources ...*herdr.AgentSessionInfo) (identity, string) {
	if processErr != nil {
		return identity{}, "process info: " + processErr.Error()
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	candidates := []herdr.PaneProcessInfoProcess{}
	for _, entry := range info.ForegroundProcesses {
		if names[entry.Name] {
			candidates = append(candidates, entry)
		}
	}
	if len(candidates) == 0 {
		return identity{}, "no foreground process has a profile name"
	}
	if len(candidates) > 1 {
		return identity{}, "multiple foreground processes have a profile name"
	}
	candidate := candidates[0]
	members := 0
	for _, entry := range info.ForegroundProcesses {
		if entry.PID == candidate.PID {
			members++
		}
	}
	if members != 1 {
		return identity{}, "the candidate PID is not in the foreground process set"
	}
	before, err := process.Start(int64(candidate.PID))
	if err != nil {
		return identity{}, err.Error()
	}
	alive, err := process.Alive(int64(candidate.PID), before)
	if err != nil {
		return identity{}, err.Error()
	}
	if !alive {
		return identity{}, "the process start is not stable"
	}
	after, err := process.Start(int64(candidate.PID))
	if err != nil {
		return identity{}, err.Error()
	}
	if after != before {
		return identity{}, "the process start is not stable"
	}
	session, reason := mergeSessions(record, sources...)
	if reason != "" {
		return identity{}, reason
	}
	return identity{pid: int64(candidate.PID), name: candidate.Name, start: before, session: session}, ""
}

func mergeSessions(record records.Agent, sources ...*herdr.AgentSessionInfo) (*herdr.AgentSessionInfo, string) {
	var found *herdr.AgentSessionInfo
	for _, source := range sources {
		if source == nil {
			continue
		}
		if found == nil {
			found = source
			continue
		}
		if sessionRef(found) != sessionRef(source) {
			return nil, "the agent session references differ"
		}
	}
	if record.AgentSessionRef == "" {
		return found, ""
	}
	if found == nil {
		return nil, "the recorded agent session is missing"
	}
	if record.AgentSessionRef != sessionRef(found) {
		return nil, "the recorded agent session differs"
	}
	return found, ""
}

func sessionRef(session *herdr.AgentSessionInfo) string {
	if session == nil {
		return ""
	}
	data, err := json.Marshal(session)
	if err != nil {
		return ""
	}
	return string(data)
}

func setRecordStatus(home string, record records.Agent, allowedStatus, status, message, action string) (records.Agent, output.Effect, error) {
	return mutateAgentRecord(home, record, allowedStatus, func(current *records.Agent) {
		current.Status = status
		current.LastError = message
	}, action)
}

func mutateAgentRecord(home string, expected records.Agent, allowedStatus string, apply func(*records.Agent), action string) (records.Agent, output.Effect, error) {
	effect := output.Effect{Resource: "agent", Path: statePath("agents", expected.Key), Action: action}
	registry, err := store.Open(home)
	if err != nil {
		return expected, output.Effect{}, err
	}
	var updated records.Agent
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			return err
		}
		if err := checkCurrentRecord(locked, current, expected, allowedStatus); err != nil {
			return err
		}
		apply(&current)
		current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			return err
		}
		updated = current
		return nil
	})
	if err != nil {
		return expected, output.Effect{}, err
	}
	return updated, effect, nil
}

func checkCurrentRecord(locked *store.LockedStore, current, expected records.Agent, allowedStatus string) error {
	if current.Key != expected.Key || current.Name != expected.Name || current.Role != expected.Role ||
		current.Project != expected.Project || current.Task != expected.Task ||
		current.PaneID != expected.PaneID || current.Generation != expected.Generation {
		return errRecordChanged
	}
	if current.Status != allowedStatus || !current.Spawned {
		return errRecordChanged
	}
	claimKey := claimKeyFor(current.Role, current.Project)
	if claimKey == "" {
		return nil
	}
	var claim records.Claim
	if err := locked.ReadIn([]string{"claims"}, claimKey, &claim); err != nil {
		return errRecordChanged
	}
	if claim.Agent != current.Key || claim.Generation != current.Generation || claim.Key != claimKey {
		return errRecordChanged
	}
	return nil
}

func spawnEnvelope(outcome, phase string, effects []output.Effect, err any, result SpawnResult) output.Envelope {
	return output.Envelope{Outcome: outcome, Phase: phase, Effects: effects, Error: err, Result: result}
}

func appendEffect(effects []output.Effect, effect output.Effect) []output.Effect {
	if effect.Resource == "" {
		return effects
	}
	return append(effects, effect)
}

func chooseRecord(updated, fallback records.Agent) records.Agent {
	if updated.Key != "" {
		return updated
	}
	return fallback
}

func logSpawn(logger *slog.Logger, envelope output.Envelope, options SpawnOptions, key string) {
	if logger == nil {
		return
	}
	attrs := []any{"command", "agent spawn", "role", options.Role, "outcome", envelope.Outcome, "phase", envelope.Phase}
	if key != "" {
		attrs = append(attrs, "agent_id", key)
	}
	if options.Project != "" {
		attrs = append(attrs, "project_id", options.Project)
	}
	if options.Task != "" {
		attrs = append(attrs, "task_id", options.Task)
	}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("agent spawn end", attrs...)
}

func logArgv(logger *slog.Logger, key string, argv []string) {
	if logger == nil {
		return
	}
	data, err := json.Marshal(argv)
	if err != nil {
		return
	}
	logger.Debug("agent start argv", "command", "agent spawn", "agent_id", key, hivelog.Body(string(data)))
}

// SpawnText renders the effects and the display name of one spawn.
func SpawnText(envelope output.Envelope) string {
	text := effectsText(envelope)
	if envelope.Error != nil {
		return text
	}
	result, ok := envelope.Result.(SpawnResult)
	if !ok {
		return text
	}
	return text + fmt.Sprintf("agent %s %s\n", result.Agent.Name, result.Agent.Status)
}
