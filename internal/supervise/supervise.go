package supervise

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// EventStream is one open events.subscribe connection.
type EventStream interface {
	Events() <-chan herdr.Event
	Errors() <-chan error
	Close() error
}

// Operations is the herdr and agent surface of the supervisor.
type Operations interface {
	GetAgent(ctx context.Context, paneID string) (herdr.AgentInfo, error)
	GetPane(ctx context.Context, paneID string) (herdr.PaneInfo, error)
	ListAgents(ctx context.Context) ([]herdr.AgentInfo, error)
	Prompt(ctx context.Context, params herdr.AgentPromptParams) (herdr.AgentInfo, error)
	Stop(ctx context.Context, home string, record records.Agent) (agent.StopResult, output.Envelope)
	Subscribe(ctx context.Context, params herdr.SubscribeParams) (EventStream, error)
}

// Options holds the explicit values of one supervisor run.
type Options struct {
	Home        string
	Logger      *slog.Logger
	Out         io.Writer
	PaneID      string
	WorkspaceID string
	TabID       string
	Harness     string
	Model       string
	PID         int64
	PIDStart    uint64
	// Operations and Clock are test seams. A nil value uses the live
	// implementation.
	Operations Operations
	Clock      Clock
}

// Result is the exit reason of one supervisor run.
type Result struct {
	Generation uint64 `json:"generation"`
	Reason     string `json:"reason"`
}

// Supervisor is one running relay process.
type Supervisor struct {
	options Options
	logger  *slog.Logger
	ops     Operations
	clock   Clock
	out     io.Writer

	claim      records.Claim
	ownKey     string
	generation uint64

	stream  EventStream
	watched map[string]bool

	quiet *quietManager
	// quietDone marks each agent that already wrote a quiet record for the
	// current idle period.
	quietDone map[string]bool
	watch     *watcher

	agents map[string]records.Agent
	claims map[string]records.Claim
}

// Run starts the supervisor and blocks until it exits. The caller cancels ctx
// on SIGINT or SIGTERM.
func Run(ctx context.Context, options Options) (Result, output.Envelope) {
	if ctx == nil {
		ctx = context.Background()
	}
	s := newSupervisor(options)
	return s.run(ctx)
}

func newSupervisor(options Options) *Supervisor {
	s := &Supervisor{options: options}
	s.logger = options.Logger
	if s.logger == nil {
		s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	s.ops = options.Operations
	if s.ops == nil {
		s.ops = liveOperations{logger: s.logger}
	}
	s.clock = options.Clock
	if s.clock == nil {
		s.clock = realClock{}
	}
	s.out = options.Out
	if s.out == nil {
		s.out = io.Discard
	}
	s.agents = map[string]records.Agent{}
	s.claims = map[string]records.Claim{}
	s.watched = map[string]bool{}
	s.quiet = newQuietManager(s.clock)
	s.quietDone = map[string]bool{}
	return s
}

func (s *Supervisor) run(ctx context.Context) (Result, output.Envelope) {
	s.logger.Debug("supervisor start", "command", "supervise", "home", s.options.Home)
	started := agent.StartSupervisor(s.options.Home, agent.RegisterOptions{
		Role:        "supervisor",
		PaneID:      s.options.PaneID,
		WorkspaceID: s.options.WorkspaceID,
		TabID:       s.options.TabID,
		Harness:     s.options.Harness,
		Model:       s.options.Model,
		PID:         s.options.PID,
		PIDStart:    s.options.PIDStart,
	})
	if started.Outcome != "ok" {
		reason := fmt.Sprint(started.Error)
		s.failure("claim", reason)
		return Result{Reason: reason}, output.Envelope{
			Outcome: "failed",
			Phase:   started.Phase,
			Effects: started.Effects,
			Error:   started.Error,
			Result:  Result{Reason: reason},
		}
	}
	result, ok := started.Result.(agent.RegisterResult)
	if !ok || result.Claim == nil {
		reason := "supervisor claim missing"
		s.failure("claim", reason)
		return Result{Reason: reason}, output.Envelope{Outcome: "failed", Phase: "claim", Effects: []output.Effect{}, Error: reason, Result: Result{Reason: reason}}
	}
	s.claim = *result.Claim
	s.ownKey = result.Agent.Key
	s.generation = result.Claim.Generation
	s.writeLine(fmt.Sprintf("supervisor live, generation %d", s.generation))
	s.logger.Debug("supervisor claim", "command", "supervise", "generation", s.generation, "agent_id", s.ownKey, "pid", s.claim.PID)

	if err := s.openWatch(); err != nil {
		reason := err.Error()
		s.failure("watch", reason)
		return Result{Generation: s.generation, Reason: reason}, output.Envelope{Outcome: "failed", Phase: "watch", Effects: []output.Effect{}, Error: reason, Result: Result{Generation: s.generation, Reason: reason}}
	}
	defer s.watch.Close()
	defer s.quiet.close()

	if err := s.refresh(ctx); err != nil {
		return s.exitFailure(ctx, err)
	}
	err := s.loop(ctx)
	s.closeStream()
	// A signal can arrive while a store read is in flight. That read fails
	// with the context error, and it is still a clean signal exit.
	if err != nil && ctx.Err() != nil {
		err = nil
	}
	if err != nil {
		return s.exitFailure(ctx, err)
	}
	s.logger.Debug("supervisor stop", "command", "supervise", "generation", s.generation, "reason", "signal")
	return Result{Generation: s.generation, Reason: "signal"}, output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{},
		Error:   nil,
		Result:  Result{Generation: s.generation, Reason: "signal"},
	}
}

func (s *Supervisor) exitFailure(ctx context.Context, err error) (Result, output.Envelope) {
	reason := err.Error()
	var storage *storageError
	outcome := "failed"
	if errors.As(err, &storage) {
		outcome = storage.envelope.Outcome
	}
	s.failure("run", reason)
	return Result{Generation: s.generation, Reason: reason}, output.Envelope{
		Outcome: outcome,
		Phase:   "run",
		Effects: []output.Effect{},
		Error:   reason,
		Result:  Result{Generation: s.generation, Reason: reason},
	}
}

func (s *Supervisor) failure(phase, reason string) {
	s.logger.Debug("supervisor failed", "command", "supervise", "phase", phase, "error", reason)
}

func (s *Supervisor) openWatch() error {
	state := filepath.Join(s.options.Home, ".hive", "state")
	inboxDir := filepath.Join(state, "inbox")
	agentsDir := filepath.Join(state, "agents")
	for _, dir := range []string{inboxDir, agentsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	watch, err := newWatcher(inboxDir, agentsDir)
	if err != nil {
		return err
	}
	s.watch = watch
	return nil
}

func (s *Supervisor) closeStream() {
	if s.stream != nil {
		_ = s.stream.Close()
		s.stream = nil
	}
}

func (s *Supervisor) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.watch.Events():
			if err := s.refresh(ctx); err != nil {
				return err
			}
		case event, ok := <-s.stream.Events():
			if !ok {
				return s.streamEnd()
			}
			if err := s.handleEvent(ctx, event); err != nil {
				return err
			}
		case err := <-s.stream.Errors():
			if err != nil {
				return err
			}
			return errors.New("supervisor subscription closed")
		case key := <-s.quiet.fired:
			if err := s.handleQuiet(ctx, key); err != nil {
				return err
			}
		}
	}
}

func (s *Supervisor) streamEnd() error {
	select {
	case err := <-s.stream.Errors():
		if err != nil {
			return err
		}
	default:
	}
	return errors.New("supervisor subscription closed")
}

func (s *Supervisor) writeLine(line string) {
	_, _ = fmt.Fprintln(s.out, line)
}

// storageError carries the envelope of one failed notification group.
type storageError struct {
	envelope output.Envelope
}

func (e *storageError) Error() string {
	return fmt.Sprint(e.envelope.Error)
}

func newStorageError(envelope output.Envelope) error {
	if envelope.Outcome == "ok" {
		return nil
	}
	return &storageError{envelope: envelope}
}

// logEvent writes one event line to the supervisor log.
func (s *Supervisor) logEvent(event herdr.Event, detail string) {
	body := ""
	if len(event.Data) > 0 {
		body = string(event.Data)
	}
	s.logger.Debug("supervisor event", "command", "supervise", "event", event.Event, "detail", detail, hivelog.Body(body))
}

func (s *Supervisor) snapshot(ctx context.Context) error {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return err
	}
	agents := map[string]records.Agent{}
	claims := map[string]records.Claim{}
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		agentRaw, err := locked.ListIn([]string{"agents"}, nil)
		if err != nil {
			return err
		}
		for _, data := range agentRaw {
			var record records.Agent
			if err := json.Unmarshal(data, &record); err != nil {
				return fmt.Errorf("record invalid: agents: %w", err)
			}
			if record.Key != "" {
				agents[record.Key] = record
			}
		}
		claimRaw, err := locked.ListIn([]string{"claims"}, nil)
		if err != nil {
			return err
		}
		for _, data := range claimRaw {
			var claim records.Claim
			if err := json.Unmarshal(data, &claim); err != nil {
				return fmt.Errorf("record invalid: claims: %w", err)
			}
			if claim.Key != "" {
				claims[claim.Key] = claim
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.agents = agents
	s.claims = claims
	return nil
}

// liveByPane returns the live record that holds one pane.
func (s *Supervisor) liveByPane(paneID string) (records.Agent, bool) {
	for _, record := range s.agents {
		if record.Status == records.AgentLive && record.PaneID == paneID {
			return record, true
		}
	}
	return records.Agent{}, false
}

// liveByField returns the live record that matches one topology field.
func (s *Supervisor) liveByField(field, value string) (records.Agent, bool) {
	if value == "" {
		return records.Agent{}, false
	}
	for _, record := range s.agents {
		if record.Status != records.AgentLive {
			continue
		}
		switch field {
		case "pane_id":
			if record.PaneID == value {
				return record, true
			}
		case "tab_id":
			if record.TabID == value {
				return record, true
			}
		case "workspace_id":
			if record.WorkspaceID == value {
				return record, true
			}
		}
	}
	return records.Agent{}, false
}

// controllingRole returns the inbox role that receives the notifications of one
// agent. A coordinator or a supervisor record has no controlling role.
func controllingRole(record records.Agent) string {
	switch record.Role {
	case "worker", "verifier":
		return "manager-" + record.Project
	case "manager":
		return "coordinator"
	default:
		return ""
	}
}

// roleForAgent is the inbox role of one agent.
func roleForAgent(record records.Agent) string {
	switch record.Role {
	case "coordinator":
		return "coordinator"
	case "supervisor":
		return "supervisor"
	case "manager":
		return "manager-" + record.Project
	default:
		return ""
	}
}

func quietAfter(record records.Agent) time.Duration {
	minutes := record.QuietAfterMin
	if minutes <= 0 {
		minutes = defaultQuietAfterMin
	}
	return time.Duration(minutes) * time.Minute
}

func statePath(collection, id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", collection, id+".json"))
}

// liveOperations is the production Operations implementation.
type liveOperations struct {
	logger *slog.Logger
}

func (o liveOperations) GetAgent(ctx context.Context, paneID string) (herdr.AgentInfo, error) {
	return herdr.AgentGet(ctx, paneID, herdr.WithLogger(o.logger))
}

func (o liveOperations) GetPane(ctx context.Context, paneID string) (herdr.PaneInfo, error) {
	return herdr.PaneGet(ctx, paneID, herdr.WithLogger(o.logger))
}

func (o liveOperations) ListAgents(ctx context.Context) ([]herdr.AgentInfo, error) {
	return herdr.AgentList(ctx, herdr.WithLogger(o.logger))
}

func (o liveOperations) Prompt(ctx context.Context, params herdr.AgentPromptParams) (herdr.AgentInfo, error) {
	return herdr.AgentPrompt(ctx, params, herdr.WithLogger(o.logger))
}

func (o liveOperations) Stop(ctx context.Context, home string, record records.Agent) (agent.StopResult, output.Envelope) {
	return agent.Stop(ctx, home, record)
}

func (o liveOperations) Subscribe(ctx context.Context, params herdr.SubscribeParams) (EventStream, error) {
	return herdr.Subscribe(ctx, params, herdr.WithLogger(o.logger))
}

// appendAndStatus writes one inbox counter, one inbox record, then the final
// agent status under one held store handle.
func (s *Supervisor) appendAndStatus(ctx context.Context, n notification) output.Envelope {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	envelope := output.Envelope{}
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		appended := inbox.AppendHeld(locked, inbox.AppendOptions{Role: n.role, Kind: n.kind, Task: n.task, Agent: n.agent, Text: n.text})
		if appended.Outcome != "ok" {
			envelope = appended
			return nil
		}
		if n.agentKey == "" {
			envelope = appended
			return nil
		}
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, n.agentKey, &current); err != nil {
			envelope = partialEnvelope("status", appended.Effects, "read agent: "+err.Error())
			return nil
		}
		if n.lifecycle != "" {
			current.Status = n.lifecycle
		}
		if n.agentStatus != "" {
			current.AgentStatus = n.agentStatus
		}
		current.UpdatedAt = now
		if err := locked.WriteIn([]string{"agents"}, n.agentKey, current); err != nil {
			envelope = partialEnvelope("status", appended.Effects, "write agent: "+err.Error())
			return nil
		}
		effects := append([]output.Effect{}, appended.Effects...)
		effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", n.agentKey), Action: "updated"})
		envelope = output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects}
		return nil
	})
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	return envelope
}

// statusOnly writes one agent status and no inbox record.
func (s *Supervisor) statusOnly(ctx context.Context, record records.Agent, status string) output.Envelope {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	envelope := output.Envelope{}
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, record.Key, &current); err != nil {
			envelope = failedEnvelope("status", "read agent: "+err.Error())
			return nil
		}
		if current.Status != records.AgentLive || current.Generation != record.Generation {
			envelope = failedEnvelope("status", "agent record changed")
			return nil
		}
		current.AgentStatus = status
		current.UpdatedAt = now
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			envelope = failedEnvelope("status", "write agent: "+err.Error())
			return nil
		}
		envelope = output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{{Resource: "agent", Path: statePath("agents", current.Key), Action: "updated"}}}
		return nil
	})
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	return envelope
}

func (s *Supervisor) logNotification(n notification, envelope output.Envelope) {
	inbox.LogAppend(s.options.Home, inbox.AppendOptions{Role: n.role, Kind: n.kind, Task: n.task, Agent: n.agent, Text: n.text}, envelope)
}

func failedEnvelope(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message}
}

func partialEnvelope(phase string, effects []output.Effect, message string) output.Envelope {
	return output.Envelope{Outcome: "partial", Phase: phase, Effects: effects, Error: message}
}
