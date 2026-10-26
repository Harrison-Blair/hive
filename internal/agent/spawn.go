package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/gitx"
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
	Profile string
	// ProfileSet marks an explicit --profile.
	ProfileSet bool
	// Kind is the built-in kind of a record that hive code starts.
	Kind string
	// Owner marks the one live owner of the project. Owns is its project.
	Owner         bool
	Owns          string
	Agent         string
	Project       string
	Task          string
	Harness       string
	Model         string
	Name          string
	NameSet       bool
	PaneID        string
	PaneSet       bool
	Adopt         string
	QuietAfterMin int
	QuietAfterSet bool
	ExtraArgs     []string
	PID           int64
	// Prompt and PromptFile are the spawner text. Cwd and AllowTrustDialog
	// are the placement and trust inputs of a spawn with no task or project.
	Prompt            string
	PromptSet         bool
	PromptFile        string
	PromptFileSet     bool
	Cwd               string
	CwdSet            bool
	AllowTrustDialog  bool
	shellReadyTimeout time.Duration
	general           *generalSpawn
	taskRecord        *records.Task
	projectRecord     *records.Project
	caller            records.Agent
	callerKnown       bool
	callerWorkspace   string
	firstPrompt       string
	instructions      string
	fromCLI           bool
}

func (options SpawnOptions) callerKey() string {
	if options.callerKnown {
		return options.caller.Key
	}
	return ""
}

// SpawnResult is the record and optional claim of one spawn.
type SpawnResult struct {
	Agent records.Agent  `json:"agent"`
	Claim *records.Claim `json:"claim,omitempty"`
}

// SpawnFromCLI is the CLI entry of spawn. It applies the caller rules of the
// CLI, such as the rule for --owner.
func SpawnFromCLI(ctx context.Context, home string, options SpawnOptions) output.Envelope {
	options.fromCLI = true
	return Spawn(ctx, home, options)
}

// Spawn validates one spawn, then starts the agent.
func Spawn(ctx context.Context, home string, options SpawnOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return failed("log", err.Error())
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)

	profile, err := harness.Lookup(options.Harness)
	if err != nil {
		envelope := failed("validate", err.Error())
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if err := harness.Check(profile); err != nil {
		envelope := failed("validate", err.Error())
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	options, message := normalizeSpawn(ctx, home, options, profile)
	if message != "" {
		envelope := failed("validate", message)
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if message := validateExtraArgs(options.ExtraArgs); message != "" {
		envelope := failed("validate", message)
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if options.Kind != "" {
		if !options.PaneSet {
			envelope := failed("validate", "pane required: pass --pane")
			logSpawn(logger, envelope, options, "")
			return envelope
		}
		return spawnInPane(ctx, home, options, profile, options.PaneID, false, logger)
	}
	if options.Task == "" && options.Project == "" {
		return spawnGeneral(ctx, home, options, profile, logger)
	}
	choice, paneError := choosePane(ctx, home, options, logger)
	if paneError != "" {
		envelope := failed("validate", paneError)
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if options.Task != "" {
		if choice.Task == nil {
			choice.Task = givenPaneTaskChoice(ctx, home, options, choice.PaneID)
		}
		return spawnTaskAgent(ctx, home, options, profile, choice.Task, logger)
	}
	return spawnProjectAgent(ctx, home, options, profile, choice.ProjectChoice, logger)
}

// spawnInPane registers one agent, starts it in paneID, and records its
// identity. The caller sends the first prompt.
func spawnInPane(ctx context.Context, home string, options SpawnOptions, profile harness.Profile, paneID string, renameTab bool, logger *slog.Logger) (envelope output.Envelope) {
	systemArgs, file, err := prepareInstructions(home, profile, options.instructions)
	if err != nil {
		envelope := failed("instructions", err.Error())
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	startSent := false
	if file != nil {
		defer func() { envelope = settleInstructionFile(logger, file, startSent, envelope) }()
	}
	registered := Register(home, RegisterOptions{
		Profile:             options.Profile,
		Kind:                options.Kind,
		Owns:                options.Owns,
		Project:             options.Project,
		Task:                options.Task,
		Harness:             options.Harness,
		Model:               options.Model,
		Name:                options.Name,
		NameSet:             options.NameSet,
		PaneID:              paneID,
		CallerAgent:         options.Agent,
		CallerPaneID:        os.Getenv("HERDR_PANE_ID"),
		CallerPaneWorkspace: options.callerWorkspace,
		QuietAfterMin:       options.QuietAfterMin,
		QuietAfterSet:       options.QuietAfterSet,
		PID:                 options.PID,
		Spawned:             true,

		TabCreatedByHive: options.general != nil && options.general.tabCreated,
		FirstPrompt:      options.general.keptPrompt(),
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
	logger.Debug("agent spawn intent", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "profile", options.Profile)

	readyTimeout := options.shellReadyTimeout
	if readyTimeout <= 0 {
		readyTimeout = defaultShellReadyTimeout
	}
	if waitErr := waitForShell(ctx, paneID, readyTimeout, logger); waitErr != nil {
		override := ""
		if isHerdrCode(waitErr, "agent_pane_busy") {
			override = "first tab busy: stop the agent or pass --pane"
		}
		envelope := spawnStartFailure(home, logger, options, record, result.Claim, effects, waitErr, override)
		return envelope
	}

	args := profile.Args(options.Model, append(systemArgs, options.ExtraArgs...))
	startSent = true
	started, startErr := herdr.AgentStart(ctx, herdr.AgentStartParams{
		Name:      record.Key,
		Kind:      profile.Kind(),
		PaneID:    paneID,
		Args:      args,
		TimeoutMS: uint64(profile.StartTimeout() / time.Millisecond),
	}, herdr.WithLogger(logger))
	if startErr != nil {
		override := ""
		if isHerdrCode(startErr, "agent_pane_busy") {
			override = "first tab busy: stop the agent or pass --pane"
		}
		envelope := spawnStartFailure(home, logger, options, record, result.Claim, effects, startErr, override)
		return envelope
	}
	effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "started"})
	if options.general != nil {
		updated, effect, updateErr := mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
			current.StartedByHive = true
		}, "updated")
		effects = appendEffect(effects, effect)
		if updateErr != nil {
			envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
			logSpawn(logger, envelope, options, record.Key)
			return envelope
		}
		record = updated
	}
	if renameTab {
		effects = append(effects, renameAgentTab(ctx, home, logger, started.Agent.TabID, record.Name, "agent spawn", record.Key))
	}
	logArgv(logger, record.Key, started.Argv)

	session := &sessionTrack{first: started.Agent.AgentSession}
	updated, effect, updateErr := persistSession(home, record, session.first)
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	record = updated

	paneInfo, paneErr := herdr.PaneGet(ctx, paneID, herdr.WithLogger(logger))
	if paneErr != nil {
		return spawnUnresolved(home, logger, options, record, result.Claim, effects, session.first, "pane", "pane get: "+paneErr.Error())
	}
	cwd := paneInfo.Cwd
	if cwd == "" {
		cwd = paneInfo.ForegroundCwd
	}
	paneReason := session.observe(paneInfo.AgentSession)
	updated, effect, updateErr = mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
		current.TerminalID = paneInfo.TerminalID
		current.WorkspaceID = paneInfo.WorkspaceID
		current.TabID = paneInfo.TabID
		current.Cwd = cwd
		setSessionRef(current, session.first)
	}, "updated")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	record = updated
	if paneReason != "" {
		return spawnUnresolved(home, logger, options, record, result.Claim, effects, session.first, "pane", paneReason)
	}

	waitInfo, waitErr := herdr.AgentWait(ctx, herdr.AgentWaitParams{
		Target:    paneID,
		Until:     settledStatuses,
		TimeoutMS: uint64(profile.ReadyTimeout() / time.Millisecond),
	}, herdr.WithDeadline(profile.ReadyTimeout()+5*time.Second), herdr.WithLogger(logger))

	var sessionReason string
	if waitErr == nil {
		sessionReason = session.observe(waitInfo.AgentSession)
	}
	if sessionReason == "" && waitErr == nil {
		updated, effect, updateErr := persistSession(home, record, session.first)
		effects = appendEffect(effects, effect)
		if updateErr != nil {
			envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
			logSpawn(logger, envelope, options, record.Key)
			return envelope
		}
		record = updated
	}

	getInfo, getErr := herdr.AgentGet(ctx, paneID, herdr.WithLogger(logger))
	if getErr != nil {
		logger.Debug("agent reply has no name", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "error", getErr.Error())
	} else if getInfo.Name == "" {
		logger.Debug("agent reply has a missing name", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID)
	} else {
		logger.Debug("agent named", "command", "agent spawn", "agent_id", record.Key, "pane_id", paneID, "name", getInfo.Name)
	}

	if waitErr != nil {
		return spawnWaitFailure(ctx, home, logger, options, record, result.Claim, effects, profile, waitInfo, waitErr, getInfo, getErr, session, sessionReason)
	}
	if sessionReason == "" && getErr == nil {
		sessionReason = session.observe(getInfo.AgentSession)
	}
	if sessionReason != "" {
		return spawnUnresolved(home, logger, options, record, result.Claim, effects, session.first, "wait", sessionReason)
	}
	updated, effect, updateErr = persistSession(home, record, session.first)
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	record = updated
	status := waitInfo.AgentStatus
	if status != "idle" && status != "done" && status != "blocked" {
		updated, effect, updateErr := setRecordStatus(home, record, record.Status, records.AgentUnresolved, "unexpected agent status: "+status, "unresolved")
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
		return spawnUnresolved(home, logger, options, record, result.Claim, effects, session.first, "process", "agent process identity unresolved: "+identityError)
	}
	updated, effect, updateErr = mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
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
		if options.general != nil {
			var blockedEffects []output.Effect
			updated, blockedEffects, _ = blockGeneralRecord(home, record, message)
			effects = append(effects, blockedEffects...)
		} else {
			updated, effect, _ = setRecordStatus(home, record, record.Status, records.AgentLive, message, "updated")
			effects = appendEffect(effects, effect)
		}
		envelope := spawnEnvelope("partial", "wait", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	updated, effect, updateErr = setRecordStatus(home, record, record.Status, records.AgentLive, "", "updated")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		envelope := spawnEnvelope("partial", "agent", effects, updateErr.Error(), SpawnResult{Agent: record, Claim: result.Claim})
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	envelope = spawnEnvelope("ok", "complete", effects, nil, SpawnResult{Agent: chooseRecord(updated, record), Claim: result.Claim})
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

// paneChoice is the pane and the placement of one spawn.
type paneChoice struct {
	PaneID        string
	Project       string
	Task          *taskChoice
	ProjectChoice *projectChoice
}

func choosePane(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (paneChoice, string) {
	if options.Task != "" {
		if options.PaneSet {
			if options.taskRecord != nil && options.taskRecord.Worktree != nil {
				worktree := *options.taskRecord.Worktree
				worktree.RootPane = options.PaneID
				pane, message, _ := checkWorktreePane(ctx, home, worktree, *options.projectRecord, logger)
				if message != "" {
					return paneChoice{}, message
				}
				return paneChoice{PaneID: pane.PaneID, Project: options.Project}, ""
			}
			return paneChoice{PaneID: options.PaneID, Project: options.Project}, ""
		}
		choice, message := resolveTaskPane(ctx, home, options, logger)
		if message != "" {
			return paneChoice{}, message
		}
		return paneChoice{PaneID: choice.PaneID, Project: options.Project, Task: &choice}, ""
	}
	if options.PaneSet {
		return paneChoice{PaneID: options.PaneID, Project: options.Project, ProjectChoice: &projectChoice{PaneID: options.PaneID, ProjectID: options.Project}}, ""
	}
	choice, message := resolveProjectPane(ctx, home, options, logger)
	if message != "" {
		return paneChoice{}, message
	}
	return paneChoice{PaneID: choice.PaneID, Project: options.Project, ProjectChoice: &choice}, ""
}

// givenPaneTaskChoice builds the task facts of a spawn in a given pane.
func givenPaneTaskChoice(ctx context.Context, home string, options SpawnOptions, paneID string) *taskChoice {
	choice := &taskChoice{PaneID: paneID}
	if options.taskRecord != nil {
		choice.Task = *options.taskRecord
		choice.ShouldAssign = taskNeedsActingAgent(home, options.taskRecord)
		if options.taskRecord.Worktree != nil {
			choice.Worktree = *options.taskRecord.Worktree
			if branch, err := gitx.CurrentBranch(ctx, choice.Worktree.Path); err == nil {
				choice.Branch = branch
			}
		}
	}
	if options.projectRecord != nil {
		choice.Project = *options.projectRecord
	}
	return choice
}

// spawnProjectAgent starts one project agent, writes the project
// instructions, and sends the first prompt.
func spawnProjectAgent(ctx context.Context, home string, options SpawnOptions, profile harness.Profile, choice *projectChoice, logger *slog.Logger) output.Envelope {
	if choice == nil {
		envelope := failed("validate", "pane required: pass --pane")
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if options.projectRecord != nil {
		if message := writeProjectInstructions(home, *options.projectRecord, logger); message != "" {
			choice.cleanup(ctx, home, logger)
			envelope := failed("instructions", message)
			logSpawn(logger, envelope, options, "")
			return envelope
		}
	}
	renameTab := false
	if !options.PaneSet {
		renameTab = !choice.Workspace.TabCreatedByHive
	}
	envelope := spawnInPane(ctx, home, options, profile, choice.PaneID, renameTab, logger)
	if envelope.Outcome == "failed" {
		choice.cleanup(ctx, home, logger)
	}
	record := spawnResultAgent(envelope)
	if record.Key == "" {
		return envelope
	}
	switch {
	case envelope.Outcome == "ok":
		return promptPlacedAgent(ctx, options.firstPrompt, record, envelope, logger)
	case envelope.Outcome == "partial" && isBlockedEnvelope(envelope):
		hint := "hive agent send --agent " + record.Name + " " + shellQuote(options.firstPrompt)
		envelope.Error = appendHint(envelope.Error, hint)
	}
	return envelope
}

// normalizeSpawn checks each input of one spawn and sets its facts. It writes
// no agent record, starts no process, and changes no task.
func normalizeSpawn(ctx context.Context, home string, options SpawnOptions, profile harness.Profile) (SpawnOptions, string) {
	if options.Profile == "" && options.Kind != records.AgentKindSupervisor {
		options.Profile = records.ProfileGeneral
	}
	if options.fromCLI && options.Profile == records.ProfileCoordinator {
		return options, "profile coordinator is built in: run bare hive, or the restart in Health › Agents"
	}
	if options.Profile != "" && (!namePattern.MatchString(options.Profile) || (options.ProfileSet && !profileFileExists(home, options.Profile))) {
		return options, "profile not found: profiles/" + displayValue(options.Profile) + ".md"
	}
	instructions, message := instructionsText(home, options.Profile)
	if message != "" {
		return options, message
	}
	options.instructions = instructions
	if options.Task != "" && options.CwdSet {
		return options, "--cwd conflicts with --task: a task agent works in its worktree"
	}
	if options.Project != "" && options.CwdSet {
		return options, "--cwd conflicts with --project: a project agent works in projects/" + options.Project + "/"
	}
	if options.PaneSet && options.CwdSet {
		return options, "--cwd conflicts with --pane"
	}
	text, message := readSpawnPrompt(options)
	if message != "" {
		return options, message
	}
	if (options.PromptSet || options.PromptFileSet) && strings.TrimSpace(text) == "" {
		return options, "prompt text is empty"
	}
	if isUserTagged(text) {
		return options, "prompt text cannot use the user: tag"
	}
	if options.NameSet {
		if message := reservedNameError(options.Name, true); message != "" {
			return options, message
		}
		if options.Name == "" || !namePattern.MatchString(options.Name) {
			return options, "invalid agent name: " + displayValue(options.Name)
		}
	}
	if options.Owner && options.Project == "" && options.Task == "" {
		return options, "--owner needs --project or --task"
	}

	resolveSpawnCaller(ctx, home, &options)

	if options.Task != "" {
		task, message := readSpawnTask(home, options.Task)
		if message != "" {
			return options, message
		}
		options.taskRecord = task
		if options.Project != "" && options.Project != task.Project {
			return options, fmt.Sprintf("task %s is in project %s, not %s", task.ID, task.Project, options.Project)
		}
		options.Project = task.Project
	}
	if options.Project != "" {
		project, message := readProjectRecord(home, options.Project)
		if message != "" {
			return options, message
		}
		options.projectRecord = &project
	}
	if options.Owner {
		if options.fromCLI {
			if message, workspace := checkOwnerCaller(ctx, home, options); message != "" {
				return options, message
			} else {
				options.callerWorkspace = workspace
			}
		} else if options.callerKnown && !isCoordinator(options.caller) {
			return options, "--owner needs the coordinator or the user: " + options.caller.Name
		}
		if name, found := liveOwner(home, options.Project); found {
			return options, fmt.Sprintf("project %s has an owner: %s", options.Project, name)
		}
		options.Owns = options.Project
	}
	name, message := spawnName(home, options)
	if message != "" {
		return options, message
	}
	options.Name = name
	options.NameSet = true
	plan := records.Agent{Name: options.Name, Parent: options.callerKey(), Project: options.Project, Task: options.Task}
	if options.taskRecord != nil && options.taskRecord.Worktree != nil {
		plan.Worktree = options.taskRecord.Worktree.Path
	}
	options.firstPrompt = firstPromptWithInstructions(plan, options.taskRecord, text, options.instructions, profile.SystemPromptArg() == "")
	if len(options.firstPrompt) > maxPromptBytes {
		return options, fmt.Sprintf("first prompt is longer than %d bytes", maxPromptBytes)
	}
	return options, ""
}

func profileFileExists(home, profile string) bool {
	info, err := os.Stat(filepath.Join(home, "profiles", profile+".md"))
	return err == nil && !info.IsDir()
}

func readSpawnPrompt(options SpawnOptions) (string, string) {
	if options.PromptSet && options.PromptFileSet {
		return "", "pass one of --prompt or --prompt-file, not both"
	}
	text := options.Prompt
	if options.PromptFileSet {
		data, err := os.ReadFile(options.PromptFile)
		if err != nil {
			return "", "prompt file: " + err.Error()
		}
		text = string(data)
	}
	return text, ""
}

func resolveSpawnCaller(ctx context.Context, home string, options *SpawnOptions) {
	registry, err := store.Open(home)
	if err != nil {
		return
	}
	_ = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		caller, callerErr := Caller(locked, CallerOptions{Agent: options.Agent, HerdrPane: os.Getenv("HERDR_PANE_ID")})
		if callerErr == nil {
			options.caller = caller
			options.callerKnown = true
			options.Agent = caller.Key
		}
		return nil
	})
}

func checkOwnerCaller(ctx context.Context, home string, options SpawnOptions) (string, string) {
	if options.callerKnown {
		if isCoordinator(options.caller) {
			return "", ""
		}
		return "--owner needs the coordinator or the user: " + options.caller.Name, ""
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	workspace, err := chat.ValidateCallerWorkspace(ctx, home, options.Agent, os.Getenv("HERDR_PANE_ID"), cwd)
	if err != nil {
		name := options.Agent
		if name == "" {
			name = "the caller"
		}
		return "--owner needs the coordinator or the user: " + displayValue(name), ""
	}
	return "", workspace
}

func isCoordinator(record records.Agent) bool {
	return record.Kind == records.AgentKindCoordinator
}

func liveOwner(home, projectID string) (string, bool) {
	registry, err := store.Open(home)
	if err != nil {
		return "", false
	}
	agents, err := loadAgentRecords(registry)
	if err != nil {
		return "", false
	}
	for _, record := range agents {
		if record.Owns != projectID {
			continue
		}
		switch record.Status {
		case records.AgentStarting, records.AgentLive, records.AgentUnresolved:
			return record.Name, true
		}
	}
	return "", false
}

func spawnName(home string, options SpawnOptions) (string, string) {
	if options.NameSet {
		return options.Name, ""
	}
	if key := claimKeyForFacts(options.Kind, options.Owns); key != "" {
		return defaultSingletonName(options.Profile, options.Kind, options.Owns), ""
	}
	registry, err := store.Open(home)
	if err != nil {
		return "", "store: " + err.Error()
	}
	agents, err := loadAgentRecords(registry)
	if err != nil {
		return "", "store: " + err.Error()
	}
	claims, err := loadClaims(registry)
	if err != nil {
		return "", "store: " + err.Error()
	}
	if options.Task != "" {
		return nextTaskAgentName(options.Task, options.Profile, agents, claims), ""
	}
	occupied := map[string]bool{}
	for _, record := range agents {
		occupied[record.Key] = true
		occupied[record.Name] = true
	}
	for _, claim := range claims {
		occupied[claim.Key] = true
		occupied[claim.Agent] = true
	}
	for index := uint64(1); ; index++ {
		candidate := fmt.Sprintf("%s-%d", options.Profile, index)
		if !occupied[candidate] {
			return candidate, ""
		}
	}
}

func readSpawnTask(home, taskID string) (*records.Task, string) {
	if !records.ValidTaskID(taskID) {
		return nil, "task not found: " + displayValue(taskID)
	}
	registry, err := store.Open(home)
	if err != nil {
		return nil, "store: " + err.Error()
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "task not found: " + displayValue(taskID)
		}
		return nil, "store: " + err.Error()
	}
	return &task, ""
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

// settleInstructionFile removes the instruction file of one spawn after a
// known outcome: no start request, a ready start, or an exited agent. Claude
// Code reads the file before it draws its prompt, so a ready start with a
// checked process identity has read it. For each other outcome the harness
// can still read the file, so the file stays until the owner settles the
// agent.
func settleInstructionFile(logger *slog.Logger, file *instructionFile, startSent bool, envelope output.Envelope) output.Envelope {
	agent := spawnResultAgent(envelope)
	if !startSent || envelope.Outcome == "ok" || agent.Status == records.AgentExited {
		err := file.remove()
		if err == nil {
			return envelope
		}
		logger.Error("instruction file remove", "command", "agent spawn", "agent_id", agent.Key, "path", file.path, "error", err.Error())
	}
	envelope.Effects = append(envelope.Effects, file.keptEffect())
	logger.Debug("instruction file kept", "command", "agent spawn", "agent_id", agent.Key, "path", file.path, "cleanup", "remove the file after the owner settles agent "+agent.Key)
	return envelope
}

func spawnStartFailure(home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, startErr error, override string) output.Envelope {
	message := startErr.Error()
	if override != "" {
		message = override
	}
	var serverError *herdr.Error
	if errors.As(startErr, &serverError) {
		updated, effect, updateErr := setRecordStatus(home, record, record.Status, records.AgentExited, message, "exited")
		effects = appendEffect(effects, effect)
		envelope := spawnEnvelope("failed", "start", effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
		if updateErr != nil {
			envelope.Error = message + "; " + updateErr.Error()
		}
		logSpawn(logger, envelope, options, record.Key)
		return envelope
	}
	updated, effect, updateErr := setRecordStatus(home, record, record.Status, records.AgentUnresolved, message, "unresolved")
	effects = appendEffect(effects, effect)
	envelope := spawnEnvelope("partial", "start", effects, "agent start: "+message, SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
	if updateErr != nil {
		envelope.Error = "agent start: " + message + "; " + updateErr.Error()
	}
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

func spawnWaitFailure(ctx context.Context, home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, profile harness.Profile, waitInfo herdr.AgentInfo, waitErr error, getInfo herdr.AgentInfo, getErr error, session *sessionTrack, sessionReason string) output.Envelope {
	var serverError *herdr.Error
	if errors.As(waitErr, &serverError) {
		switch serverError.Code {
		case herdr.CodeTimeout:
			message := fmt.Sprintf("agent not ready after %d s", int(profile.ReadyTimeout()/time.Second))
			if sessionReason == "" && getErr == nil {
				sessionReason = session.observe(getInfo.AgentSession)
			}
			processInfo, processErr := herdr.PaneProcessInfoGet(ctx, record.PaneID, herdr.WithLogger(logger))
			identity, identityError := captureIdentity(profile, record, processInfo, processErr, waitInfo.AgentSession, getInfo.AgentSession)
			if getErr == nil && identityError == "" && sessionReason == "" {
				updated, effect, updateErr := mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
					current.Status = records.AgentLive
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
			} else if sessionReason != "" {
				reason = "agent process identity unresolved: " + sessionReason
			} else if getErr != nil {
				reason = "agent get: " + getErr.Error()
			}
			return spawnUnresolved(home, logger, options, record, claim, effects, session.first, "wait", reason)
		case herdr.CodeAgentNotRunning, herdr.CodeAgentNotFound:
			inspection := inspectAgent(ctx, home, record, logger)
			effects = appendEffect(effects, inspection.effect)
			envelope := spawnEnvelope("failed", "wait", effects, inspection.messageForStart(), SpawnResult{Agent: chooseRecord(inspection.record, record), Claim: claim})
			logSpawn(logger, envelope, options, record.Key)
			return envelope
		}
	}
	return spawnUnresolved(home, logger, options, record, claim, effects, session.first, "wait", "agent wait: "+waitErr.Error())
}

func spawnUnresolved(home string, logger *slog.Logger, options SpawnOptions, record records.Agent, claim *records.Claim, effects []output.Effect, first *herdr.AgentSessionInfo, phase, message string) output.Envelope {
	updated, effect, updateErr := mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
		current.Status = records.AgentUnresolved
		current.LastError = message
		setSessionRef(current, first)
	}, "unresolved")
	effects = appendEffect(effects, effect)
	envelope := spawnEnvelope("partial", phase, effects, message, SpawnResult{Agent: chooseRecord(updated, record), Claim: claim})
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
		if names[entry.Name] || names[entryArgvName(entry, true)] || names[entryArgvName(entry, false)] {
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

// entryArgvName returns the base name of the argv0 value of one process entry
// when useArgv0 is true, and the base name of argv[0] when it is false. A
// wrapper launcher replaces itself, so the process name differs from the
// executable in argv. The empty string is not a profile name.
func entryArgvName(entry herdr.PaneProcessInfoProcess, useArgv0 bool) string {
	if useArgv0 {
		if entry.Argv0 == nil {
			return ""
		}
		return filepath.Base(*entry.Argv0)
	}
	if len(entry.Argv) == 0 {
		return ""
	}
	return filepath.Base(entry.Argv[0])
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

// sessionTrack keeps the earliest present session reference of one spawn.
type sessionTrack struct {
	first *herdr.AgentSessionInfo
}

// observe records one session reference in start, pane, wait, get order.
func (track *sessionTrack) observe(observed *herdr.AgentSessionInfo) string {
	if observed == nil {
		if track.first != nil {
			return "the agent session is missing"
		}
		return ""
	}
	if track.first == nil {
		track.first = observed
		return ""
	}
	if sessionRef(track.first) != sessionRef(observed) {
		return "the agent session references differ"
	}
	return ""
}

func persistSession(home string, record records.Agent, first *herdr.AgentSessionInfo) (records.Agent, output.Effect, error) {
	if first == nil || record.AgentSessionRef == sessionRef(first) {
		return record, output.Effect{}, nil
	}
	return mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
		setSessionRef(current, first)
	}, "updated")
}

func setSessionRef(record *records.Agent, session *herdr.AgentSessionInfo) {
	if session != nil {
		record.AgentSessionRef = sessionRef(session)
	}
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
	return mutateAgentRecordChecked(home, expected, allowedStatus, true, apply, action)
}

// mutateAgentRecordChecked writes a record mutation. A false requireSpawned
// settles a record that a person registered by hand.
func mutateAgentRecordChecked(home string, expected records.Agent, allowedStatus string, requireSpawned bool, apply func(*records.Agent), action string) (records.Agent, output.Effect, error) {
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
		if err := checkCurrentRecordChecked(locked, current, expected, allowedStatus, requireSpawned); err != nil {
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
	return checkCurrentRecordChecked(locked, current, expected, allowedStatus, true)
}

func checkCurrentRecordChecked(locked *store.LockedStore, current, expected records.Agent, allowedStatus string, requireSpawned bool) error {
	if current.Key != expected.Key || current.Name != expected.Name ||
		current.Project != expected.Project || current.Task != expected.Task ||
		current.PaneID != expected.PaneID || current.Generation != expected.Generation {
		return errRecordChanged
	}
	if current.Status != allowedStatus {
		return errRecordChanged
	}
	if requireSpawned && !current.Spawned {
		return errRecordChanged
	}
	claimKey := claimKeyForFacts(current.Kind, current.Owns)
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
	attrs := []any{"command", "agent spawn", "profile", options.Profile, "outcome", envelope.Outcome, "phase", envelope.Phase}
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
