package spawn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
)

const (
	// busyTimeout bounds the wait for the shell of a new pane.
	busyTimeout = 10 * time.Second
	busyPause   = 200 * time.Millisecond
	readyPause  = 250 * time.Millisecond
)

// Options holds the values of one spawn.
type Options struct {
	Name       string
	Harness    string
	Model      string
	Prompt     string
	PromptFile string
	Project    string
	Worktree   string
	// CallerPane is the HERDR_PANE_ID of the caller.
	CallerPane string
	// Cwd is the directory of a new tab when Project is empty.
	Cwd string
	// Guide is the hive guide. The harness gets it in its system prompt.
	Guide  string
	Logger *slog.Logger
}

// Run starts one agent and writes its record. The result is the record.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	var session string
	envelope := run(ctx, home, options, logger, &session)
	attrs := []any{"command", "agent spawn", "target", options.Name, "project_id", options.Project, "outcome", envelope.Outcome, "phase", envelope.Phase}
	if session != "" {
		attrs = append(attrs, "harness_session_id", session)
	}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("agent spawn", attrs...)
	return envelope
}

// run starts the agent. It keeps in session the last harness session ID that
// a herdr reply gives.
func run(ctx context.Context, home string, options Options, logger *slog.Logger, session *string) output.Envelope {
	profile, prompt, message := validate(options)
	if message != "" {
		return output.Failed("validate", message)
	}
	if err := checkName(ctx, home, options.Name, logger); err != nil {
		return output.Failed("name", err.Error())
	}
	parent, err := agent.CallerName(home, options.CallerPane)
	if err != nil {
		return output.Failed("caller", err.Error())
	}

	trust := profile.Trust()
	var extra []string
	var env map[string]string
	if trust.NoMethod == "" {
		extra = trust.Args
		for _, pair := range trust.Env {
			key, value, _ := strings.Cut(pair, "=")
			if env == nil {
				env = map[string]string{}
			}
			env[key] = value
		}
	}

	var place placement
	var phase string
	if options.Project == "" {
		place, phase, err = placeInCaller(ctx, options.CallerPane, options.Cwd, options.Name, env, logger)
	} else {
		var project records.Project
		project, err = loadProject(home, options.Project)
		if err != nil {
			return output.Failed("project", err.Error())
		}
		if options.Worktree == "" {
			place, phase, err = placeInProject(ctx, project, options.Name, env, logger)
		} else {
			place, phase, err = placeInWorktree(ctx, home, project, options.Worktree, logger)
		}
	}
	if err != nil {
		envelope := output.Failed(phase, err.Error())
		envelope.Effects = append(envelope.Effects, place.effects...)
		return envelope
	}
	effects := place.effects
	paneID := place.pane.PaneID
	// A failed rename does not stop the spawn. It makes the outcome partial.
	// A later failure keeps its phase and carries the rename error too.
	var labelErr string
	if place.rootPane {
		if err := herdr.TabRename(ctx, place.pane.TabID, options.Name, herdr.WithLogger(logger)); err != nil {
			labelErr = fmt.Sprintf("tab rename %s to %s: %s", place.pane.TabID, options.Name, err)
		} else {
			effects = append(effects, output.Effect{Resource: "tab", Path: place.pane.TabID, Action: "renamed"})
		}
	}
	withLabel := func(message string) string {
		if labelErr == "" {
			return message
		}
		return message + " (also: tab label: " + labelErr + ")"
	}

	guideArgs := profile.GuideArgs(options.Guide)
	argv, err := start(ctx, herdr.AgentStartParams{
		Name:      options.Name,
		Kind:      profile.Kind(),
		PaneID:    paneID,
		Args:      profile.Args(options.Model, append(extra, guideArgs...)),
		TimeoutMS: uint64(profile.StartTimeout() / time.Millisecond),
	}, logger, session)
	if err != nil {
		envelope := output.Failed("start", withLabel(fmt.Sprintf("agent start in pane %s: %s", paneID, err)))
		envelope.Effects = effects
		if closeErr := herdr.PaneClose(ctx, paneID, herdr.WithLogger(logger)); closeErr == nil {
			envelope.Effects = append(envelope.Effects, output.Effect{Resource: "pane", Path: paneID, Action: "closed"})
		}
		return envelope
	}
	logger.Debug("agent started", "command", "agent spawn", "target", options.Name, "pane_id", paneID, "argv", guide.Redact(argv, guideArgs))
	effects = append(effects, output.Effect{Resource: "agent", Path: paneID, Action: "started"})

	record := agent.Record{
		Name:      options.Name,
		Pane:      paneID,
		Parent:    parent,
		Harness:   profile.Kind(),
		Model:     options.Model,
		Cwd:       place.cwd,
		Worktree:  place.worktree,
		StartedAt: time.Now().UTC(),
	}
	if err := agent.Save(home, record); err != nil {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "record", Effects: effects, Error: withLabel("write record: " + err.Error()), Result: record}
	}
	effects = append(effects, output.Effect{Resource: "record", Path: ".hive/agents/" + record.Name + ".json", Action: "written"})

	partial := func(phase, message string) output.Envelope {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: phase, Effects: effects, Error: withLabel(message), Result: record}
	}
	info, err := waitReady(ctx, paneID, profile.ReadyTimeout(), logger, session)
	if err != nil {
		return partial("wait", fmt.Sprintf("agent %s in pane %s is not ready: %s; the prompt was not sent", record.Name, paneID, err))
	}
	if info.AgentStatus == "blocked" {
		return partial("wait", blockedMessage(record.Name, paneID))
	}
	if prompt != "" {
		prompted, err := herdr.AgentPrompt(ctx, herdr.AgentPromptParams{Target: paneID, Text: prompt}, herdr.WithLogger(logger))
		if agent.IsCode(err, herdr.CodeAgentBlocked) {
			return partial("prompt", blockedMessage(record.Name, paneID))
		}
		if err != nil {
			return partial("prompt", fmt.Sprintf("prompt to agent %s in pane %s: %s", record.Name, paneID, err))
		}
		keepSession(session, prompted.AgentSession)
		logger.Debug("agent prompted", "command", "agent spawn", "target", record.Name, "pane_id", paneID, hivelog.Body(prompt))
		effects = append(effects, output.Effect{Resource: "agent", Path: paneID, Action: "prompted"})
	}
	if labelErr != "" {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "label", Effects: effects, Error: labelErr, Result: record}
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: record}
}

func blockedMessage(name, paneID string) string {
	return fmt.Sprintf("agent %s is blocked by a dialog in pane %s: answer it in the pane; the prompt was not sent", name, paneID)
}

// validate checks the options before any herdr call. It returns the harness
// profile and the first prompt text.
func validate(options Options) (harness.Profile, string, string) {
	for _, required := range []struct{ flag, value string }{
		{"--name", options.Name},
		{"--harness", options.Harness},
		{"--model", options.Model},
	} {
		if required.value == "" {
			return nil, "", required.flag + " is required"
		}
	}
	if err := agent.ValidateName(options.Name); err != nil {
		return nil, "", err.Error()
	}
	profile, err := harness.Lookup(options.Harness)
	if err != nil {
		return nil, "", err.Error()
	}
	if options.Prompt != "" && options.PromptFile != "" {
		return nil, "", "give at most one of --prompt and --prompt-file"
	}
	if options.Worktree != "" {
		if options.Project == "" {
			return nil, "", "--worktree needs --project"
		}
		if err := agent.ValidateName(options.Worktree); err != nil {
			return nil, "", "--worktree: " + err.Error()
		}
	}
	prompt := options.Prompt
	if options.PromptFile != "" {
		data, err := os.ReadFile(options.PromptFile)
		if err != nil {
			return nil, "", "--prompt-file: " + err.Error()
		}
		prompt = string(data)
	}
	return profile, prompt, ""
}

// checkName fails when a record has name and herdr still has its pane.
func checkName(ctx context.Context, home, name string, logger *slog.Logger) error {
	record, err := agent.Load(home, name)
	if errors.Is(err, agent.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
	if agent.IsCode(err, herdr.CodePaneNotFound) {
		logger.Debug("agent record replaced", "command", "agent spawn", "target", name, "pane_id", record.Pane)
		return nil
	}
	if err != nil {
		return fmt.Errorf("check agent %s: %w", name, err)
	}
	return fmt.Errorf("agent %s exists in pane %s", name, record.Pane)
}

func loadProject(home, id string) (records.Project, error) {
	if !records.ValidID(id) {
		return records.Project{}, errors.New(projectlib.InvalidIDMessage(id))
	}
	registry, err := store.Open(home)
	if err != nil {
		return records.Project{}, err
	}
	var project records.Project
	if err := registry.Read("projects", id, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return records.Project{}, fmt.Errorf("project not registered: %s", id)
		}
		return records.Project{}, err
	}
	return project, nil
}

// waitReady polls agent.get until herdr reports the agent interactive_ready
// or blocked. agent.wait does not serve: right after agent.start it can match
// the idle status of the shell before the agent runs (herdr 0.9.5).
func waitReady(ctx context.Context, paneID string, timeout time.Duration, logger *slog.Logger, session *string) (herdr.AgentInfo, error) {
	deadline := time.Now().Add(timeout)
	for {
		info, err := herdr.AgentGet(ctx, paneID, herdr.WithLogger(logger))
		keepSession(session, info.AgentSession)
		if err == nil && (info.Interactive || info.AgentStatus == "blocked") {
			return info, nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return herdr.AgentInfo{}, err
			}
			return herdr.AgentInfo{}, fmt.Errorf("no interactive agent after %s (status %s)", timeout, info.AgentStatus)
		}
		select {
		case <-ctx.Done():
			return herdr.AgentInfo{}, ctx.Err()
		case <-time.After(readyPause):
		}
	}
}

// start sends agent.start. A new pane can hold no shell prompt yet, so start
// tries again while herdr reports the pane busy.
func start(ctx context.Context, params herdr.AgentStartParams, logger *slog.Logger, session *string) ([]string, error) {
	deadline := time.Now().Add(busyTimeout)
	for {
		started, err := herdr.AgentStart(ctx, params, herdr.WithLogger(logger))
		if err == nil {
			keepSession(session, started.Agent.AgentSession)
			return started.Argv, nil
		}
		if !agent.IsCode(err, "agent_pane_busy") || time.Now().After(deadline) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(busyPause):
		}
	}
}

// keepSession sets session to the ID of reported when herdr reports one.
func keepSession(session *string, reported *herdr.AgentSessionInfo) {
	if id := agent.SessionID(reported); id != "" {
		*session = id
	}
}

// Text renders the effects of one spawn.
func Text(envelope output.Envelope) string {
	var lines []string
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if record, ok := envelope.Result.(agent.Record); ok && envelope.Outcome == output.OutcomeOK {
		lines = append(lines, fmt.Sprintf("agent %s runs in pane %s", record.Name, record.Pane))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
