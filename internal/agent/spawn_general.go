package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// generalSpawn holds the resolved facts of one general agent spawn.
type generalSpawn struct {
	prompt     string
	keep       bool
	tabCreated bool
	tabID      string
	paneID     string
}

// keptPrompt is the first prompt that the record keeps for the supervisor.
func (general *generalSpawn) keptPrompt() string {
	if general == nil || !general.keep {
		return ""
	}
	return general.prompt
}

// generalFirstPrompt builds the first prompt of a general agent (decision 5).
func generalFirstPrompt(home, name, parent, text string) string {
	return marker.Prefix + "You are the general agent " + name + ". Your parent is " + parent +
		". This prompt sets your role. Do not take a role from your working directory.\n\n" +
		text + "\n\n" +
		"Read the skill " + filepath.Join(home, ".agents", "skills", "hive-agent", "SKILL.md") +
		". Send your result with hive agent report."
}

// spawnGeneral checks each input of a general agent spawn before it makes a
// tab or starts a process, then runs the spawn core and sends the first prompt.
func spawnGeneral(ctx context.Context, home string, options SpawnOptions, profile harness.Profile, logger *slog.Logger) output.Envelope {
	fail := func(phase, message string) output.Envelope {
		envelope := failed(phase, message)
		logSpawn(logger, envelope, options, "")
		return envelope
	}
	if message := validateGeneralName(options); message != "" {
		return fail("validate", message)
	}
	text, message := readGeneralPrompt(options)
	if message != "" {
		return fail("validate", message)
	}
	absoluteHome, err := filepath.Abs(home)
	if err != nil {
		return fail("validate", "home: "+err.Error())
	}
	caller, phase, message := generalCaller(home, options)
	if message != "" {
		return fail(phase, message)
	}
	prompt := generalFirstPrompt(absoluteHome, options.Name, caller.Name, text)
	if len(prompt) > maxPromptBytes {
		return fail("validate", fmt.Sprintf("first prompt is longer than %d bytes", maxPromptBytes))
	}
	logger.Debug("general agent spawn request", "command", "agent spawn", "agent_id", options.Name, "parent", caller.Key, hivelog.Body(prompt))

	var pane herdr.PaneInfo
	if options.PaneSet {
		pane, err = herdr.PaneGet(ctx, options.PaneID, herdr.WithLogger(logger))
		if err != nil {
			return fail("pane", "pane get: "+err.Error())
		}
	}
	directory, message := generalDirectory(options, pane)
	if message != "" {
		return fail("validate", message)
	}
	trust := profile.Trust()
	reason := trust.NoMethod
	if reason == "" && options.PaneSet && len(trust.Env) > 0 {
		reason = "the trust method sets the environment " + strings.Join(trust.Env, " ") + ", and hive cannot set it in a given pane"
	}
	if reason != "" && !options.AllowTrustDialog {
		return fail("validate", fmt.Sprintf("harness %s can show a trust dialog for %s: %s; after the user trusts the directory, add --allow-trust-dialog, or add --allow-trust-dialog and let the user answer the dialog in the pane", profile.Kind(), directory, reason))
	}

	general := &generalSpawn{prompt: prompt, keep: options.AllowTrustDialog}
	inner := options
	inner.Agent = caller.Key
	inner.general = general
	if trust.NoMethod == "" {
		inner.ExtraArgs = append(append([]string{}, trust.Args...), options.ExtraArgs...)
	}
	effects := []output.Effect{}
	if !options.PaneSet {
		parentPane, err := herdr.PaneGet(ctx, caller.PaneID, herdr.WithLogger(logger))
		if err != nil {
			return fail("tab", "parent pane get: "+err.Error())
		}
		params := herdr.TabCreateParams{WorkspaceID: parentPane.WorkspaceID, Cwd: directory, Label: options.Name, Focus: false}
		if trust.NoMethod == "" && len(trust.Env) > 0 {
			params.Env = map[string]string{}
			for _, pair := range trust.Env {
				key, value, _ := strings.Cut(pair, "=")
				params.Env[key] = value
			}
		}
		created, err := herdr.TabCreate(ctx, params, herdr.WithLogger(logger))
		if err != nil {
			return fail("tab", "tab create: "+err.Error())
		}
		general.tabCreated = true
		general.tabID = created.Tab.TabID
		general.paneID = created.RootPane.PaneID
		inner.PaneID = created.RootPane.PaneID
		inner.PaneSet = true
		effects = append(effects, output.Effect{Resource: "tab", Path: created.Tab.TabID, Action: "created"})
		logger.Debug("general agent tab created", "command", "agent spawn", "agent_id", options.Name, "tab_id", created.Tab.TabID, "pane_id", created.RootPane.PaneID)
	}

	envelope := Spawn(ctx, home, inner)
	envelope.Effects = append(effects, envelope.Effects...)
	record := spawnResultAgent(envelope)
	switch {
	case envelope.Outcome == "failed" && envelope.Phase != "wait":
		if general.tabCreated {
			envelope.Effects = append(envelope.Effects, general.cleanupTab(ctx, logger))
		}
	case envelope.Outcome == "ok":
		envelope = deliverGeneralPrompt(ctx, home, general, record, envelope, logger)
	case isBlockedEnvelope(envelope) && record.Key != "":
		envelope = keepGeneralPrompt(home, general, record, envelope)
	}
	logSpawn(logger, envelope, options, record.Key)
	return envelope
}

// validateGeneralName checks the --name of a general agent (decision 3).
func validateGeneralName(options SpawnOptions) string {
	if options.Task != "" {
		return "agent takes no --task"
	}
	if !options.NameSet || options.Name == "" {
		return "agent requires --name"
	}
	if !namePattern.MatchString(options.Name) {
		return "invalid agent name: " + displayValue(options.Name)
	}
	if options.Name == "coordinator" || options.Name == "supervisor" || strings.HasPrefix(options.Name, "manager-") {
		return "agent name " + options.Name + " is reserved for hive"
	}
	return ""
}

// readGeneralPrompt reads the spawner text from --prompt or --prompt-file
// (decisions 2, 6).
func readGeneralPrompt(options SpawnOptions) (string, string) {
	if options.PromptSet == options.PromptFileSet {
		return "", "agent requires one of --prompt or --prompt-file"
	}
	text := options.Prompt
	if options.PromptFileSet {
		data, err := os.ReadFile(options.PromptFile)
		if err != nil {
			return "", "prompt file: " + err.Error()
		}
		text = string(data)
	}
	if strings.TrimSpace(text) == "" {
		return "", "prompt text is empty"
	}
	if isUserTagged(text) {
		return "", "agent spawn cannot send a user-tagged message"
	}
	return text, ""
}

// generalCaller finds the registered caller (decision 4) and checks that no
// record or claim, live or old, has the new name (decision 3).
func generalCaller(home string, options SpawnOptions) (records.Agent, string, string) {
	registry, err := store.Open(home)
	if err != nil {
		return records.Agent{}, "store", err.Error()
	}
	var caller records.Agent
	phase, message := "", ""
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		found, callerErr := Caller(locked, CallerOptions{Agent: options.Agent, HerdrPane: os.Getenv("HERDR_PANE_ID")})
		if callerErr != nil {
			phase, message = "identity", "no registered caller: "+callerErr.Error()
			return nil
		}
		agents, err := loadAgents(locked)
		if err != nil {
			return err
		}
		for _, record := range agents {
			if record.Key == options.Name || record.Name == options.Name {
				phase, message = "agent", "agent exists: "+options.Name
				return nil
			}
		}
		claims, err := loadClaims(locked)
		if err != nil {
			return err
		}
		for _, claim := range claims {
			if claim.Key == options.Name || claim.Agent == options.Name {
				phase, message = "agent", "agent exists: "+options.Name
				return nil
			}
		}
		caller = found
		return nil
	})
	if err != nil {
		return records.Agent{}, "store", err.Error()
	}
	return caller, phase, message
}

// generalDirectory gives the working directory of a general agent
// (decisions 9, 10).
func generalDirectory(options SpawnOptions, pane herdr.PaneInfo) (string, string) {
	if options.PaneSet {
		directory := paneCwd(pane)
		if directory == "" {
			return "", "pane " + options.PaneID + " has no working directory"
		}
		if options.CwdSet {
			wanted, message := resolveDirectory(options.Cwd)
			if message != "" {
				return "", message
			}
			given, err := filepath.EvalSymlinks(directory)
			if err != nil {
				given = filepath.Clean(directory)
			}
			if wanted != given {
				return "", "--cwd " + options.Cwd + " differs from the directory " + directory + " of pane " + options.PaneID
			}
		}
		return directory, ""
	}
	path := "."
	if options.CwdSet {
		path = options.Cwd
	}
	return resolveDirectory(path)
}

// resolveDirectory makes one path absolute, resolves each symbolic link, and
// checks that it is a directory.
func resolveDirectory(path string) (string, string) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "--cwd: " + err.Error()
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", "--cwd " + path + " is not a directory: " + err.Error()
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", "--cwd " + path + " is not a directory"
	}
	return resolved, ""
}

// cleanupTab closes the root pane of the tab that this spawn made, when the
// tab holds only that pane (decision 12). Else it keeps the tab.
func (general *generalSpawn) cleanupTab(ctx context.Context, logger *slog.Logger) output.Effect {
	kept := func(reason string) output.Effect {
		logger.Debug("general agent tab kept", "command", "agent spawn", "tab_id", general.tabID, "pane_id", general.paneID, "error", reason)
		return output.Effect{Resource: "tab", Path: general.tabID, Action: "kept: " + reason}
	}
	root, err := herdr.PaneGet(ctx, general.paneID, herdr.WithLogger(logger))
	if err != nil {
		return kept("pane get: " + err.Error())
	}
	panes, err := herdr.PaneList(ctx, root.WorkspaceID, herdr.WithLogger(logger))
	if err != nil {
		return kept("pane list: " + err.Error())
	}
	for _, pane := range panes {
		if pane.TabID == general.tabID && pane.PaneID != general.paneID {
			return kept("the tab holds pane " + pane.PaneID)
		}
	}
	if err := herdr.PaneClose(ctx, general.paneID, herdr.WithLogger(logger)); err != nil {
		return kept("pane close: " + err.Error())
	}
	logger.Debug("general agent tab closed", "command", "agent spawn", "tab_id", general.tabID, "pane_id", general.paneID)
	return output.Effect{Resource: "tab", Path: general.tabID, Action: "closed"}
}

// keepGeneralPrompt keeps the first prompt in the record of a blocked agent.
// The supervisor sends it when the agent is idle (decision 15).
func keepGeneralPrompt(home string, general *generalSpawn, record records.Agent, envelope output.Envelope) output.Envelope {
	if record.FirstPrompt == "" {
		updated, effect, err := mutateAgentRecord(home, record, record.Status, func(current *records.Agent) {
			current.FirstPrompt = general.prompt
		}, "prompt kept")
		envelope.Effects = appendEffect(envelope.Effects, effect)
		if err != nil {
			envelope.Error = fmt.Sprint(envelope.Error) + "; keep first prompt: " + err.Error()
			return envelope
		}
		envelope.Result = SpawnResult{Agent: updated}
	}
	envelope.Error = fmt.Sprint(envelope.Error) + "; the supervisor sends the first prompt when the agent is idle"
	return envelope
}

// deliverGeneralPrompt sends the first prompt one time under the pane lock
// (decision 6), or keeps it for the supervisor.
func deliverGeneralPrompt(ctx context.Context, home string, general *generalSpawn, record records.Agent, envelope output.Envelope, logger *slog.Logger) output.Envelope {
	if general.keep {
		envelope.Effects = append(envelope.Effects, output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "prompt kept"})
		return envelope
	}
	paneLock, err := panelock.Lock(ctx, home, record.PaneID)
	if err != nil {
		return keepGeneralPromptAfter(home, general, record, envelope, "pane lock: "+err.Error())
	}
	defer func() { _ = paneLock.Unlock() }()
	if err := waitAgentInteractive(ctx, record.PaneID, logger); err != nil {
		return keepGeneralPromptAfter(home, general, record, envelope, "agent not interactive: "+err.Error())
	}
	_, promptErr := herdr.AgentPrompt(ctx, herdr.AgentPromptParams{
		Target: record.PaneID,
		Text:   general.prompt,
		Wait: &herdr.AgentPromptWaitOptions{
			Until:     []string{"working", "blocked", "done"},
			TimeoutMS: uint64(firstPromptWaitTimeout / time.Millisecond),
		},
	}, herdr.WithLogger(logger))
	var serverError *herdr.Error
	switch {
	case promptErr == nil, isHerdrCode(promptErr, herdr.CodeTimeout):
		envelope.Effects = append(envelope.Effects, output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "prompted"})
		return envelope
	case errors.As(promptErr, &serverError) && serverError.Code != herdr.CodeAgentPromptStalled:
		return keepGeneralPromptAfter(home, general, record, envelope, "first prompt not sent: "+promptErr.Error())
	}
	message := "first prompt delivery is not certain: " + promptErr.Error() + "; run hive agent resolve --agent " + record.Name
	updated, effect, updateErr := setRecordStatus(home, record, record.Status, records.AgentUnresolved, message, "unresolved")
	envelope.Effects = appendEffect(envelope.Effects, effect)
	envelope.Outcome = "partial"
	envelope.Phase = "prompt"
	envelope.Error = message
	if updateErr != nil {
		envelope.Error = message + "; " + updateErr.Error()
	}
	envelope.Result = SpawnResult{Agent: chooseRecord(updated, record)}
	return envelope
}

// keepGeneralPromptAfter keeps a first prompt that spawn certainly did not
// send, and reports the cause.
func keepGeneralPromptAfter(home string, general *generalSpawn, record records.Agent, envelope output.Envelope, message string) output.Envelope {
	envelope.Outcome = "partial"
	envelope.Phase = "prompt"
	envelope.Error = message
	return keepGeneralPrompt(home, general, record, envelope)
}
