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
	// Role is the wave-A input name of a profile. A new caller passes Profile
	// and Owns instead. Register writes only the facts.
	Role    string
	Profile string
	// Kind is the built-in kind of a record that hive code starts.
	Kind string
	// Owns is the project ID that the agent owns, or empty.
	Owns          string
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
	// TabCreatedByHive and FirstPrompt are the ownership fact and the kept
	// first prompt of a general agent spawn.
	TabCreatedByHive bool
	FirstPrompt      string
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
	options, message := normalizeRegistration(options)
	if message != "" {
		envelope := failed("validate", message)
		logRegistration(home, options, envelope)
		return envelope
	}
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

// RegisterAndRename registers one agent and best-effort labels its tab. It is
// the CLI entry, so it refuses each role and name that hive code registers.
func RegisterAndRename(home string, options RegisterOptions) output.Envelope {
	if message := hiveOwnedRoleError(options.Role, true); message != "" {
		envelope := failed("validate", message)
		logRegistration(home, options, envelope)
		return envelope
	}
	if message := reservedNameError(options.Name, options.NameSet); message != "" {
		envelope := failed("validate", message)
		logRegistration(home, options, envelope)
		return envelope
	}
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
	effect := renameAgentTab(context.Background(), home, logger, pane.TabID, result.Agent.Name, "agent register", result.Agent.Key)
	envelope.Effects = append(envelope.Effects, effect)
	return envelope
}

// hiveOwnedRoleError returns the refusal for a role that only hive code
// registers. Register refuses the manager role too, because only spawn
// registers a manager.
func hiveOwnedRoleError(role string, register bool) string {
	switch role {
	case "coordinator", "supervisor":
		return role + ": hive code registers this role; run bare hive, or the restart in Health › Agents"
	case "manager":
		if register {
			return "manager: hive code registers this role; run hive agent spawn --role manager"
		}
	}
	return ""
}

// reservedNameError returns the refusal for a CLI name that only hive code
// registers. Internal callers can use a custom singleton name.
func reservedNameError(name string, nameSet bool) string {
	if !nameSet {
		return ""
	}
	if name == records.AgentKindCoordinator || name == records.AgentKindSupervisor {
		return "name is reserved: " + name
	}
	return ""
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

// normalizeRegistration sets the facts of one registration from the wave-A
// role, or defaults the profile. It returns the refusal of an old role that
// hive code starts.
func normalizeRegistration(options RegisterOptions) (RegisterOptions, string) {
	if options.Role == "" {
		if options.Kind == records.AgentKindCoordinator {
			options.Profile = records.ProfileCoordinator
		}
		return options, ""
	}
	switch options.Role {
	case records.AgentKindCoordinator:
		options.Kind = records.AgentKindCoordinator
		options.Profile = records.ProfileCoordinator
	case records.AgentKindSupervisor:
		options.Kind = records.AgentKindSupervisor
		options.Profile = ""
	case "manager":
		options.Profile = records.ProfilePlanner
		options.Owns = options.Project
	case "worker":
		options.Profile = records.ProfileBuilder
	case "verifier":
		options.Profile = records.ProfileVerifier
	case "agent":
		options.Profile = records.ProfileGeneral
	default:
		return options, "invalid role: " + displayValue(options.Role)
	}
	return options, ""
}

func validateRegistration(options RegisterOptions) string {
	options, _ = normalizeRegistration(options)
	if message := validateLegacyRole(options); message != "" {
		return message
	}
	if options.Profile == "" && options.Kind == "" {
		return "invalid role: " + displayValue(options.Role)
	}
	if options.Owns != "" && options.Project == "" {
		return "owner requires --project"
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
	} else if claimKeyForFacts(options.Kind, options.Owns) != "" {
		name := defaultSingletonName(options.Profile, options.Kind, options.Owns)
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

// validateLegacyRole keeps the rules of the wave-A role input until the role
// flag goes away.
func validateLegacyRole(options RegisterOptions) string {
	switch options.Role {
	case records.AgentKindCoordinator, records.AgentKindSupervisor:
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
	case "agent":
		if options.Task != "" {
			return "agent takes no --task"
		}
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

func defaultSingletonName(profile, kind, owns string) string {
	switch {
	case owns != "":
		return "manager-" + owns
	case kind == records.AgentKindCoordinator:
		return records.AgentKindCoordinator
	case kind == records.AgentKindSupervisor:
		return records.AgentKindSupervisor
	default:
		return profile
	}
}

func claimKeyForFacts(kind, owns string) string {
	switch {
	case kind == records.AgentKindCoordinator:
		return records.AgentKindCoordinator
	case kind == records.AgentKindSupervisor:
		return records.AgentKindSupervisor
	case owns != "":
		return "owner-" + owns
	default:
		return ""
	}
}

func taskRegistrationError(project, task string) string {
	return fmt.Sprintf("task not found in %s: %s", displayValue(project), displayValue(task))
}

func registerLocked(locked *store.LockedStore, writer lockedWriter, options RegisterOptions) output.Envelope {
	options, message := normalizeRegistration(options)
	if message != "" {
		return failed("validate", message)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	quietAfter := defaultQuietAfterMin
	if options.QuietAfterSet {
		quietAfter = options.QuietAfterMin
	}
	legacy := options.Role != ""
	claimKey := claimKeyForFacts(options.Kind, options.Owns)
	if legacy && options.Role == "manager" {
		claimKey = "manager-" + options.Project
	}
	singleton := claimKey != ""

	worktree := ""
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
		if task.Worktree != nil {
			worktree = task.Worktree.Path
		}
	}

	effects := []output.Effect{}
	counterEffect := output.Effect{Resource: "counter", Path: ".hive/state/counters/agent.json", Action: "updated"}
	name := options.Name
	generated := false
	taskName := !options.NameSet && options.Spawned && options.Task != ""
	if !options.NameSet && !taskName {
		switch {
		case singleton:
			name = defaultSingletonName(options.Profile, options.Kind, options.Owns)
		case legacy && options.Role == "agent":
			name = "agent"
		case legacy && (options.Role == "worker" || options.Role == "verifier"):
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
		default:
			var counter struct {
				Value uint64 `json:"value"`
			}
			err := locked.ReadIn([]string{"counters"}, "agent", &counter)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return failed("agent", "read agent counter: "+err.Error())
			}
			candidate := counter.Value + 1
			name = fmt.Sprintf("%s-%d", options.Profile, candidate)
			generated = true
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
		label := options.Profile
		if legacy && (options.Role == "worker" || options.Role == "verifier") {
			label = options.Role
		}
		name = nextTaskAgentName(options.Task, label, agentRecords, claims)
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
			existing.Name == name && sameRegisteredFacts(existing, options) &&
			existing.Project == options.Project && existing.Task == options.Task &&
			existing.Generation == generation && existing.PaneID == options.PaneID
		if !replaceOrphan {
			return failed("agent", "agent exists: "+recordKey)
		}
	}
	if options.Owns != "" && currentClaim != nil {
		if holder, ok := agentByKey[currentClaim.Agent]; ok && !(hasExisting && replaceOrphan && existing.Key == currentClaim.Agent) {
			switch holder.Status {
			case records.AgentStarting:
				return failed("agent", "agent is starting: "+holder.Name)
			case records.AgentLive, records.AgentUnresolved:
				return failed("validate", "project "+options.Owns+" has an owner: "+holder.Name)
			}
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
		replaces = replacementKey(options.Profile, options.Kind, options.Owns, options.Project, options.Task, recordKey, previousHolderKey, agentRecords)
	}

	startedAt := now
	if replaceOrphan && !options.Spawned {
		startedAt = existing.StartedAt
	}
	record := records.Agent{
		Key:           recordKey,
		Name:          name,
		Profile:       options.Profile,
		Kind:          options.Kind,
		Owns:          options.Owns,
		Worktree:      worktree,
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

		TabCreatedByHive: options.TabCreatedByHive,
		FirstPrompt:      options.FirstPrompt,
	}
	record.SyncRole()
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
	if options.Kind == "" {
		return false
	}
	return os.Getenv("HIVE_STARTUP_ROLE") == options.Kind && os.Getenv("HERDR_PANE_ID") == options.PaneID
}

// sameRegisteredFacts reports whether one record has the same facts as a
// registration.
func sameRegisteredFacts(record records.Agent, options RegisterOptions) bool {
	return record.Profile == options.Profile && record.Kind == options.Kind && record.Owns == options.Owns
}

func nextSingletonGeneration(name string, options RegisterOptions, agents []records.Agent) uint64 {
	highest := uint64(0)
	retryGeneration := uint64(0)
	for _, record := range agents {
		if !sameRegisteredFacts(record, options) || record.Name != name || record.Project != options.Project || record.Task != options.Task {
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

func replacementKey(profile, kind, owns, project, task, currentKey, previousHolderKey string, agents []records.Agent) string {
	matches := func(candidate records.Agent) bool {
		switch {
		case kind == records.AgentKindCoordinator:
			return candidate.Kind == records.AgentKindCoordinator
		case kind == records.AgentKindSupervisor:
			return candidate.Kind == records.AgentKindSupervisor
		case owns != "":
			return candidate.Owns == owns
		case task != "":
			return candidate.Task == task && candidate.Profile == profile
		default:
			return false
		}
	}
	var newest records.Agent
	for _, candidate := range agents {
		if candidate.Key == currentKey || !matches(candidate) {
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

func nextTaskAgentName(taskID, profile string, agentRecords []records.Agent, claims []records.Claim) string {
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
		candidate := fmt.Sprintf("t%s-%s-%d", prefix, profile, index)
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
	attrs := []any{"profile", options.Profile}
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
