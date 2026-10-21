package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	identitylib "github.com/Harrison-Blair/hive/internal/lib/identity"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	defaultQuietAfterMin = 10
	maxFieldBytes        = 512
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// RegisterOptions holds explicit values for one agent registration.
type RegisterOptions struct {
	Role          string
	Project       string
	Task          string
	Harness       string
	Model         string
	Name          string
	NameSet       bool
	PaneID        string
	CallerAgent   string
	CallerPaneID  string
	WorkspaceID   string
	TabID         string
	QuietAfterMin int
	QuietAfterSet bool
	PID           int64
	// PIDStart is the start time of the holder process from /proc/<pid>/stat.
	// A singleton claim stores it so a later start can tell a reused PID apart.
	PIDStart uint64
	// Spawned marks the first record write of a spawn. Manual registration
	// leaves it false. No CLI flag sets it.
	Spawned bool
}

// RegisterResult is the record and optional claim of one registration.
type RegisterResult struct {
	Agent records.Agent  `json:"agent"`
	Claim *records.Claim `json:"claim,omitempty"`
}

type lockedWriter interface {
	ReadIn(collection []string, id string, target any) error
	WriteIn(collection []string, id string, value any) error
	UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error
	ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error)
	Next(name string) (uint64, error)
}

// Register validates and records one agent session.
func Register(home string, options RegisterOptions) output.Envelope {
	if message := validateRegistration(options); message != "" {
		envelope := failed("validate", message)
		logRegistration(home, options, envelope)
		return envelope
	}
	registry, err := store.Open(home)
	if err != nil {
		envelope := failed("store", err.Error())
		logRegistration(home, options, envelope)
		return envelope
	}
	if options.Project != "" {
		var project records.Project
		if err := registry.ReadIn([]string{"projects"}, options.Project, &project); err != nil {
			var envelope output.Envelope
			if errors.Is(err, os.ErrNotExist) {
				envelope = failed("validate", "project not found: "+displayValue(options.Project))
			} else {
				envelope = failed("store", err.Error())
			}
			logRegistration(home, options, envelope)
			return envelope
		}
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = registerLocked(locked, locked, options)
		return nil
	})
	if err != nil {
		envelope = failed("store", err.Error())
	}
	logRegistration(home, options, envelope)
	return envelope
}

// RegisterAndRename registers one agent and best-effort labels its tab.
func RegisterAndRename(home string, options RegisterOptions) output.Envelope {
	envelope := Register(home, options)
	if envelope.Outcome != "ok" {
		return envelope
	}
	result, ok := envelope.Result.(RegisterResult)
	if !ok {
		return envelope
	}

	var logger *slog.Logger
	writer, writerErr := hivelog.NewWriter(home, "agent", time.Now)
	if writerErr == nil {
		defer func() { _ = writer.Close() }()
		logger = slog.New(writer)
	}
	pane, err := herdr.PaneGet(context.Background(), options.PaneID, herdr.WithLogger(logger))
	if err != nil {
		logTabLabel(home, logger, "agent register", result.Agent.Key, "", "", err)
		return envelope
	}
	label := options.Role
	if options.Project != "" {
		label = result.Agent.Name
	}
	effect := renameAgentTab(context.Background(), home, logger, pane.TabID, label, "agent register", result.Agent.Key)
	envelope.Effects = append(envelope.Effects, effect)
	return envelope
}

func renameAgentTab(ctx context.Context, home string, logger *slog.Logger, tabID, label, command, agentID string) output.Effect {
	err := herdr.TabRename(ctx, tabID, label, herdr.WithLogger(logger))
	action := "renamed"
	if err != nil {
		action = "rename failed"
	}
	logTabLabel(home, logger, command, agentID, tabID, label, err)
	return output.Effect{Resource: "tab", Path: tabID, Action: action}
}

func logTabLabel(home string, logger *slog.Logger, command, agentID, tabID, label string, err error) {
	attrs := []any{"command", command, "agent_id", agentID, "tab_id", tabID, "label", label}
	message := "agent tab renamed"
	if err != nil {
		message = "agent tab rename failed"
		attrs = append(attrs, "error", err.Error())
	}
	if logger != nil {
		logger.Debug(message, attrs...)
		return
	}
	logAgentEvent(home, message, attrs...)
}

func validateRegistration(options RegisterOptions) string {
	switch options.Role {
	case "coordinator", "supervisor":
		if options.Project != "" || options.Task != "" {
			return options.Role + " takes no --project or --task"
		}
	case "manager":
		if options.Project == "" {
			return "manager requires --project"
		}
	case "worker", "verifier":
		if options.Project == "" || options.Task == "" {
			return options.Role + " requires --project and --task"
		}
	default:
		return "invalid role: " + displayValue(options.Role)
	}
	if options.PaneID == "" {
		return "no pane: run inside herdr or pass --pane"
	}
	if options.Project != "" && !records.ValidID(options.Project) {
		return "project not found: " + displayValue(options.Project)
	}
	for _, field := range []struct {
		label string
		value string
	}{
		{"pane", options.PaneID},
		{"workspace", options.WorkspaceID},
		{"tab", options.TabID},
		{"project", options.Project},
		{"task", options.Task},
		{"harness", options.Harness},
		{"model", options.Model},
	} {
		if message := validateField(field.label, field.value); message != "" {
			return message
		}
	}
	if options.NameSet {
		if !namePattern.MatchString(options.Name) {
			return "invalid agent name: " + displayValue(options.Name)
		}
	} else if claimKeyFor(options.Role, options.Project) != "" {
		name := defaultSingletonName(options.Role, options.Project)
		if !namePattern.MatchString(name) {
			return "invalid agent name: " + displayValue(name)
		}
	}
	if options.QuietAfterSet && options.QuietAfterMin < 0 {
		return fmt.Sprintf("invalid --quiet-after: %d", options.QuietAfterMin)
	}
	if options.PID < 0 {
		return fmt.Sprintf("invalid --pid: %d", options.PID)
	}
	return ""
}

func validateField(label, value string) string {
	if len(value) > maxFieldBytes {
		return fmt.Sprintf("invalid %s: value is longer than %d bytes", label, maxFieldBytes)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Sprintf("invalid %s: %q", label, value)
		}
	}
	return ""
}

func defaultSingletonName(role, project string) string {
	switch role {
	case "manager":
		return "manager-" + project
	default:
		return role
	}
}

func claimKeyFor(role, project string) string {
	switch role {
	case "coordinator":
		return "coordinator"
	case "manager":
		return "manager-" + project
	case "supervisor":
		return "supervisor"
	default:
		return ""
	}
}

func taskRegistrationError(project, task string) string {
	return fmt.Sprintf("task not found in %s: %s", displayValue(project), displayValue(task))
}

func registerLocked(locked *store.LockedStore, writer lockedWriter, options RegisterOptions) output.Envelope {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	quietAfter := defaultQuietAfterMin
	if options.QuietAfterSet {
		quietAfter = options.QuietAfterMin
	}
	claimKey := claimKeyFor(options.Role, options.Project)
	singleton := claimKey != ""

	if options.Task != "" {
		message := taskRegistrationError(options.Project, options.Task)
		if !records.ValidTaskID(options.Task) {
			return failed("validate", message)
		}
		var task records.Task
		err := writer.ReadIn([]string{"tasks"}, options.Task, &task)
		if errors.Is(err, os.ErrNotExist) {
			return failed("validate", message)
		}
		if err != nil {
			return failed("store", err.Error())
		}
		if task.Project != options.Project {
			return failed("validate", message)
		}
	}

	effects := []output.Effect{}
	counterEffect := output.Effect{Resource: "counter", Path: ".hive/state/counters/agent.json", Action: "updated"}
	name := options.Name
	generated := false
	taskName := !options.NameSet && options.Spawned && options.Task != "" && (options.Role == "worker" || options.Role == "verifier")
	if !options.NameSet && !taskName {
		if options.Role == "worker" || options.Role == "verifier" {
			var counter struct {
				Value uint64 `json:"value"`
			}
			err := locked.ReadIn([]string{"counters"}, "agent", &counter)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return failed("agent", "read agent counter: "+err.Error())
			}
			candidate := counter.Value + 1
			name = fmt.Sprintf("%s-%s-%d", options.Role, options.Project, candidate)
			generated = true
		} else {
			name = defaultSingletonName(options.Role, options.Project)
		}
	}
	if !taskName && !namePattern.MatchString(name) {
		return failed("validate", "invalid agent name: "+displayValue(name))
	}

	agentRecords, err := loadAgents(writer)
	if err != nil {
		return failed("store", err.Error())
	}
	claims, err := loadClaims(writer)
	if err != nil {
		return failed("store", err.Error())
	}
	if taskName {
		name = nextTaskAgentName(options.Task, options.Role, agentRecords, claims)
	}
	if !namePattern.MatchString(name) {
		return failed("validate", "invalid agent name: "+displayValue(name))
	}
	agentByKey := make(map[string]records.Agent, len(agentRecords))
	for _, record := range agentRecords {
		agentByKey[record.Key] = record
	}
	claimByKey := make(map[string]records.Claim, len(claims))
	for _, claim := range claims {
		claimByKey[claim.Key] = claim
	}

	var currentClaim *records.Claim
	generation := uint64(1)
	recordKey := name
	previousHolderKey := ""
	if singleton {
		if claim, ok := claimByKey[claimKey]; ok {
			currentClaim = &claim
			generation = claim.Generation + 1
			previousHolderKey = claim.Agent
		} else {
			generation = nextSingletonGeneration(name, options, agentRecords)
		}
		recordKey = fmt.Sprintf("%s-g%d", name, generation)
	}

	existing, hasExisting := agentByKey[recordKey]
	replaceOrphan := false
	if hasExisting {
		replaceOrphan = singleton && existing.Status == records.AgentLive &&
			existing.Name == name && existing.Role == options.Role &&
			existing.Project == options.Project && existing.Task == options.Task &&
			existing.Generation == generation && existing.PaneID == options.PaneID
		if !replaceOrphan {
			return failed("agent", "agent exists: "+recordKey)
		}
	}

	for _, claim := range claims {
		if claim.Key == claimKey {
			continue
		}
		switch {
		case claim.Key == name:
			return failed("agent", "agent exists: "+name)
		case claim.Key == recordKey || claim.Agent == recordKey:
			return failed("agent", "agent exists: "+recordKey)
		}
	}
	if singleton {
		if _, ok := agentByKey[claimKey]; ok {
			return failed("agent", "agent exists: "+claimKey)
		}
		if currentClaim != nil {
			if holder, ok := agentByKey[currentClaim.Agent]; ok && holder.Status == records.AgentStarting {
				return failed("agent", "agent is starting: "+holder.Name)
			}
		}
	}
	for _, record := range agentRecords {
		if record.Key == recordKey {
			continue
		}
		if singleton && currentClaim != nil && record.Key == previousHolderKey {
			continue
		}
		if record.Key == name {
			return failed("agent", "agent exists: "+name)
		}
		if isActiveAgentStatus(record.Status) && record.Name == name {
			return failed("agent", "agent exists: "+name)
		}
		if singleton && isActiveAgentStatus(record.Status) && record.Name == claimKey {
			return failed("agent", "agent exists: "+claimKey)
		}
		if isActiveAgentStatus(record.Status) && record.Name == recordKey {
			return failed("agent", "agent exists: "+recordKey)
		}
	}
	for _, record := range agentRecords {
		if record.Status != records.AgentLive && record.Status != records.AgentStarting && record.Status != records.AgentUnresolved {
			continue
		}
		if record.PaneID != options.PaneID {
			continue
		}
		allowed := false
		if singleton {
			if record.Key == recordKey {
				allowed = true
			}
			if currentClaim != nil && record.Key == previousHolderKey {
				allowed = true
			}
		}
		if !allowed {
			return failed("agent", "pane busy: "+options.PaneID+" holds "+displayValue(record.Name))
		}
	}

	if generated {
		if _, err := writer.Next("agent"); err != nil {
			return failed("agent", "next agent counter: "+err.Error())
		}
		effects = append(effects, counterEffect)
	}

	if currentClaim != nil {
		if holder, ok := agentByKey[currentClaim.Agent]; ok && holder.Status != records.AgentStale {
			err := writer.UpdateIn([]string{"agents"}, holder.Key, func(raw json.RawMessage) (json.RawMessage, error) {
				var record records.Agent
				if err := json.Unmarshal(raw, &record); err != nil {
					return nil, err
				}
				record.Status = records.AgentStale
				record.UpdatedAt = now
				return json.Marshal(record)
			})
			if err != nil {
				return failed("agent", "write stale agent: "+err.Error())
			}
			effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", holder.Key), Action: "stale"})
		}
	}

	parent := ""
	replaces := ""
	if options.Spawned {
		caller, err := identitylib.ResolveHeld(writer, options.CallerAgent, options.CallerPaneID)
		if err == nil {
			parent = caller.Key
		} else if !errors.Is(err, identitylib.ErrCallerUnknown) &&
			!strings.HasPrefix(err.Error(), "agent not found:") &&
			!strings.HasPrefix(err.Error(), "caller ambiguous:") {
			return failed("identity", err.Error())
		}
		replaces = replacementKey(options.Role, options.Project, options.Task, recordKey, previousHolderKey, agentRecords)
	}

	startedAt := now
	if replaceOrphan && !options.Spawned {
		startedAt = existing.StartedAt
	}
	record := records.Agent{
		Key:           recordKey,
		Name:          name,
		Role:          options.Role,
		Project:       options.Project,
		Task:          options.Task,
		Harness:       options.Harness,
		Model:         options.Model,
		PaneID:        options.PaneID,
		WorkspaceID:   options.WorkspaceID,
		TabID:         options.TabID,
		TerminalID:    "",
		Cwd:           "",
		Status:        initialAgentStatus(options.Spawned),
		Parent:        parent,
		Replaces:      replaces,
		Finished:      false,
		QuietAfterMin: quietAfter,
		Generation:    generation,
		StartedAt:     startedAt,
		UpdatedAt:     now,
		Spawned:       options.Spawned,
		StartedByHive: startupStartedByHive(options),
	}
	action := "created"
	if replaceOrphan {
		action = "updated"
	}
	if err := writer.WriteIn([]string{"agents"}, recordKey, record); err != nil {
		if len(effects) == 0 {
			return failed("agent", "write agent: "+err.Error())
		}
		return output.Envelope{Outcome: "partial", Phase: "agent", Effects: effects, Error: "write agent: " + err.Error(), Result: nil}
	}
	effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", recordKey), Action: action})

	result := RegisterResult{Agent: record}
	if !singleton {
		return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: result}
	}

	claim := records.Claim{
		Key:        claimKey,
		Agent:      recordKey,
		Generation: generation,
		PID:        options.PID,
		PIDStart:   options.PIDStart,
		ClaimedAt:  now,
	}
	if err := writer.WriteIn([]string{"claims"}, claimKey, claim); err != nil {
		message := "write claim: " + err.Error()
		if currentClaim != nil {
			message += fmt.Sprintf("; claim unchanged at generation %d", currentClaim.Generation)
		}
		if options.Spawned {
			record.Status = records.AgentUnresolved
			record.LastError = message
			record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if settleErr := writer.WriteIn([]string{"agents"}, recordKey, record); settleErr != nil {
				message += "; settle agent record: " + settleErr.Error()
			} else {
				effects = append(effects, output.Effect{Resource: "agent", Path: statePath("agents", recordKey), Action: "unresolved"})
			}
		}
		return output.Envelope{Outcome: "partial", Phase: "claim", Effects: effects, Error: message, Result: nil}
	}
	effects = append(effects, output.Effect{Resource: "claim", Path: statePath("claims", claimKey), Action: "updated"})
	result.Claim = &claim
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: result}
}

func startupStartedByHive(options RegisterOptions) bool {
	if options.Role != "coordinator" && options.Role != "supervisor" {
		return false
	}
	return os.Getenv("HIVE_STARTUP_ROLE") == options.Role && os.Getenv("HERDR_PANE_ID") == options.PaneID
}

func nextSingletonGeneration(name string, options RegisterOptions, agents []records.Agent) uint64 {
	highest := uint64(0)
	retryGeneration := uint64(0)
	for _, record := range agents {
		if record.Role != options.Role || record.Name != name || record.Project != options.Project || record.Task != options.Task {
			continue
		}
		if record.Generation > highest {
			highest = record.Generation
		}
		if record.Status == records.AgentLive && record.PaneID == options.PaneID &&
			record.Key == fmt.Sprintf("%s-g%d", name, record.Generation) && record.Generation > retryGeneration {
			retryGeneration = record.Generation
		}
	}
	if retryGeneration != 0 && retryGeneration == highest {
		return retryGeneration
	}
	generation := highest + 1
	for {
		key := fmt.Sprintf("%s-g%d", name, generation)
		occupied := false
		for _, record := range agents {
			if record.Key == key {
				occupied = true
				break
			}
		}
		if !occupied {
			return generation
		}
		generation++
	}
}

func initialAgentStatus(spawned bool) string {
	if spawned {
		return records.AgentStarting
	}
	return records.AgentLive
}

func isActiveAgentStatus(status string) bool {
	return status == records.AgentStarting || status == records.AgentLive
}

func replacementKey(role, project, task, currentKey, previousHolderKey string, agents []records.Agent) string {
	var newest records.Agent
	for _, candidate := range agents {
		if candidate.Key == currentKey || candidate.Role != role {
			continue
		}
		switch role {
		case "worker", "verifier":
			if candidate.Task != task {
				continue
			}
		case "manager":
			if candidate.Project != project {
				continue
			}
		case "coordinator":
		default:
			continue
		}
		if candidate.Status == records.AgentLive && candidate.Key != previousHolderKey {
			continue
		}
		if newest.Key == "" || candidate.StartedAt > newest.StartedAt ||
			(candidate.StartedAt == newest.StartedAt && candidate.Key > newest.Key) {
			newest = candidate
		}
	}
	return newest.Key
}

func nextTaskAgentName(taskID, role string, agentRecords []records.Agent, claims []records.Claim) string {
	prefix := taskID[:strings.IndexByte(taskID, '-')]
	occupied := make(map[string]bool, len(agentRecords)*2+len(claims)*2)
	for _, record := range agentRecords {
		occupied[record.Key] = true
		occupied[record.Name] = true
	}
	for _, claim := range claims {
		occupied[claim.Key] = true
		occupied[claim.Agent] = true
	}
	for index := uint64(1); ; index++ {
		candidate := fmt.Sprintf("t%s-%s-%d", prefix, role, index)
		if !occupied[candidate] {
			return candidate
		}
	}
}

func loadAgents(writer lockedWriter) ([]records.Agent, error) {
	raw, err := writer.ListIn([]string{"agents"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
		found = append(found, record)
	}
	return found, nil
}

func loadClaims(writer lockedWriter) ([]records.Claim, error) {
	raw, err := writer.ListIn([]string{"claims"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Claim, 0, len(raw))
	for _, data := range raw {
		var claim records.Claim
		if err := json.Unmarshal(data, &claim); err != nil {
			return nil, err
		}
		found = append(found, claim)
	}
	return found, nil
}

func statePath(collection, id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", collection, id+".json"))
}

func displayValue(value string) string {
	if !utf8.ValidString(value) {
		return strconv.Quote(value)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return strconv.Quote(value)
		}
	}
	return value
}

func logAgentEvent(home, message string, attrs ...any) {
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return
	}
	defer func() { _ = writer.Close() }()
	slog.New(writer).Debug(message, attrs...)
}

func logRegistration(home string, options RegisterOptions, envelope output.Envelope) {
	attrs := []any{"role", options.Role}
	if options.Project != "" {
		attrs = append(attrs, "project_id", options.Project)
	}
	if options.Task != "" {
		attrs = append(attrs, "task_id", options.Task)
	}
	holderID := ""
	for _, effect := range envelope.Effects {
		switch {
		case effect.Resource == "counter":
			logAgentEvent(home, "agent counter updated", append(attrs, "path", effect.Path)...)
		case effect.Resource == "agent" && effect.Action == "stale":
			id := effectID(effect.Path)
			logAgentEvent(home, "agent holder marked stale", append(attrs, "agent_id", id)...)
		case effect.Resource == "agent":
			holderID = effectID(effect.Path)
			logAgentEvent(home, "agent record written", append(attrs, "agent_id", holderID)...)
		case effect.Resource == "claim":
			claimAttrs := append(attrs, "agent_id", holderID)
			claimAttrs = append(claimAttrs, "claim_id", effectID(effect.Path))
			logAgentEvent(home, "agent claim written", claimAttrs...)
		}
	}
	if envelope.Error == nil {
		return
	}
	errorAttrs := append([]any{}, attrs...)
	if holderID != "" {
		errorAttrs = append(errorAttrs, "agent_id", holderID)
	}
	errorAttrs = append(errorAttrs, "phase", envelope.Phase, "error", fmt.Sprint(envelope.Error))
	logAgentEvent(home, "agent registration failed", errorAttrs...)
}

func effectID(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".json")
}

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
