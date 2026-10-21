package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const coordinatorClaimKey = "coordinator"

// CoordinatorLaunch holds the starting claim of one coordinator start
// (decision 31). Only this launcher can start the harness for the claim.
type CoordinatorLaunch struct {
	home       string
	pid        int64
	pidStart   uint64
	generation uint64
	replaces   string
	logger     *slog.Logger
	// previous is the claim and record that a restart replaced. Release puts
	// them back when the restart fails before the old pane closes.
	previous       *records.Claim
	previousRecord *records.Agent
}

// Home returns the home of the launch.
func (l *CoordinatorLaunch) Home() string { return l.home }

// ClaimCoordinatorStart takes the starting claim for bare hive. It fails when
// a live record holds the claim, when a live launcher holds a starting claim,
// or when the claim names an unresolved record.
func ClaimCoordinatorStart(ctx context.Context, home string, logger *slog.Logger) (*CoordinatorLaunch, error) {
	return claimCoordinator(ctx, home, nil, logger)
}

// claimCoordinator takes the starting claim. With a restart record, the claim
// must name that record or be absent, and the old record becomes stale in the
// same locked write (decision 34).
func claimCoordinator(ctx context.Context, home string, restart *records.Agent, logger *slog.Logger) (launch *CoordinatorLaunch, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	command := "bare hive"
	if restart != nil {
		command = "tui agent restart"
	}
	defer func() {
		if logger == nil {
			return
		}
		attrs := []any{"command", command, "claim_id", coordinatorClaimKey}
		if err != nil {
			attrs = append(attrs, "outcome", "failed", "error", err.Error())
		} else {
			attrs = append(attrs, "outcome", records.ClaimStarting, "generation", launch.generation)
		}
		logger.Debug("coordinator starting claim", attrs...)
	}()
	pid := int64(os.Getpid())
	pidStart, err := process.Start(pid)
	if err != nil {
		return nil, fmt.Errorf("read launcher start: %w", err)
	}
	registry, err := store.Open(home)
	if err != nil {
		return nil, err
	}
	var refusal error
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var claim records.Claim
		claimErr := locked.ReadIn([]string{"claims"}, coordinatorClaimKey, &claim)
		hasClaim := claimErr == nil
		if claimErr != nil && !errors.Is(claimErr, os.ErrNotExist) {
			return fmt.Errorf("read coordinator claim: %w", claimErr)
		}
		if hasClaim && claim.Status == records.ClaimStarting {
			// A launcher that Alive cannot read is dead, as for the supervisor claim.
			if alive, aliveErr := process.Alive(claim.PID, claim.PIDStart); aliveErr == nil && alive {
				refusal = fmt.Errorf("coordinator is starting: launcher pid %d holds the claim", claim.PID)
				return nil
			}
			settled, settleErr := settleDeadLauncher(ctx, locked, claim, logger)
			if settleErr != nil {
				return settleErr
			}
			if settled != "" {
				refusal = fmt.Errorf("coordinator harness runs in pane %s with no launcher: record %s is unresolved; restart the coordinator from Health › Agents", claim.PaneID, settled)
				return nil
			}
			hasClaim = false
		}
		agents, listErr := loadAgents(locked)
		if listErr != nil {
			return listErr
		}
		replaces := ""
		var previousRecord *records.Agent
		if hasClaim {
			holder, found := findAgent(agents, claim.Agent)
			if restart != nil {
				if claim.Agent != restart.Key {
					refusal = fmt.Errorf("coordinator claim names %s, not %s: restart changes nothing", claim.Agent, restart.Key)
					return nil
				}
			} else if found && holder.Status == records.AgentLive {
				refusal = fmt.Errorf("coordinator live: %s", holder.Key)
				return nil
			} else if found && holder.Status == records.AgentUnresolved {
				refusal = fmt.Errorf("coordinator %s is unresolved: restart the coordinator from Health › Agents", holder.Key)
				return nil
			}
			replaces = claim.Agent
		}
		if restart != nil {
			current, found := findAgent(agents, restart.Key)
			if !found || current.Role != "coordinator" || current.PaneID != restart.PaneID {
				refusal = fmt.Errorf("coordinator record %s changed: restart changes nothing", restart.Key)
				return nil
			}
			replaces = current.Key
			if current.Status != records.AgentStale {
				saved := current
				previousRecord = &saved
				current.Status = records.AgentStale
				current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
				if writeErr := locked.WriteIn([]string{"agents"}, current.Key, current); writeErr != nil {
					return fmt.Errorf("mark %s stale: %w", current.Key, writeErr)
				}
			}
		}
		if replaces == "" {
			replaces = newestCoordinatorKey(agents)
		}
		generation := nextCoordinatorGeneration(claim.Generation, agents)
		starting := records.Claim{
			Key:        coordinatorClaimKey,
			Generation: generation,
			PID:        pid,
			PIDStart:   pidStart,
			ClaimedAt:  time.Now().UTC().Format(time.RFC3339Nano),
			Status:     records.ClaimStarting,
		}
		if writeErr := locked.WriteIn([]string{"claims"}, coordinatorClaimKey, starting); writeErr != nil {
			if previousRecord != nil {
				_ = locked.WriteIn([]string{"agents"}, previousRecord.Key, *previousRecord)
			}
			return fmt.Errorf("write starting claim: %w", writeErr)
		}
		launch = &CoordinatorLaunch{home: home, pid: pid, pidStart: pidStart, generation: generation, replaces: replaces, logger: logger, previousRecord: previousRecord}
		if hasClaim {
			saved := claim
			launch.previous = &saved
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if refusal != nil {
		return nil, refusal
	}
	return launch, nil
}

// settleDeadLauncher applies decision 33 under the held store lock. It returns
// the key of the unresolved record when the target pane holds a harness
// process, and the empty string when it removed the claim.
func settleDeadLauncher(ctx context.Context, locked *store.LockedStore, claim records.Claim, logger *slog.Logger) (string, error) {
	found, err := paneHarnessProcess(ctx, claim.PaneID, logger)
	if err != nil {
		return "", fmt.Errorf("check pane %s of the dead launcher: %w", claim.PaneID, err)
	}
	if found == nil {
		if err := locked.RemoveIn([]string{"claims"}, coordinatorClaimKey); err != nil {
			return "", fmt.Errorf("remove the claim of the dead launcher: %w", err)
		}
		if logger != nil {
			logger.Debug("coordinator claim settled", "claim_id", coordinatorClaimKey, "pane_id", claim.PaneID, "launcher_pid", claim.PID, "outcome", "removed")
		}
		return "", nil
	}
	agents, err := loadAgents(locked)
	if err != nil {
		return "", err
	}
	base := claim.Generation
	if base > 0 {
		base--
	}
	generation := nextCoordinatorGeneration(base, agents)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	record := records.Agent{
		Key:           fmt.Sprintf("coordinator-g%d", generation),
		Name:          "coordinator",
		Role:          "coordinator",
		PaneID:        claim.PaneID,
		Status:        records.AgentUnresolved,
		Replaces:      newestCoordinatorKey(agents),
		QuietAfterMin: defaultQuietAfterMin,
		Generation:    generation,
		StartedAt:     now,
		UpdatedAt:     now,
		StartedByHive: true,
		ProcessPID:    int64(found.PID),
		ProcessName:   found.Name,
		LastError:     fmt.Sprintf("launcher pid %d stopped after the harness start", claim.PID),
	}
	if start, startErr := process.Start(record.ProcessPID); startErr == nil {
		record.ProcessStart = start
	}
	if pane, paneErr := herdr.PaneGet(ctx, claim.PaneID, herdr.WithLogger(logger)); paneErr == nil {
		record.WorkspaceID = pane.WorkspaceID
		record.TabID = pane.TabID
		record.TerminalID = pane.TerminalID
	}
	if err := locked.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		return "", fmt.Errorf("write unresolved coordinator: %w", err)
	}
	named := records.Claim{Key: coordinatorClaimKey, Agent: record.Key, Generation: generation, PID: record.ProcessPID, PIDStart: record.ProcessStart, ClaimedAt: now, Status: records.ClaimLive}
	if err := locked.WriteIn([]string{"claims"}, coordinatorClaimKey, named); err != nil {
		return "", fmt.Errorf("write coordinator claim: %w", err)
	}
	if logger != nil {
		logger.Debug("coordinator claim settled", "claim_id", coordinatorClaimKey, "pane_id", claim.PaneID, "launcher_pid", claim.PID, "agent_id", record.Key, "outcome", records.AgentUnresolved)
	}
	return record.Key, nil
}

// paneHarnessProcess returns the foreground process of paneID that has the
// process name of a known harness. A missing pane has no harness process.
func paneHarnessProcess(ctx context.Context, paneID string, logger *slog.Logger) (*herdr.PaneProcessInfoProcess, error) {
	if paneID == "" {
		return nil, nil
	}
	info, err := herdr.PaneProcessInfoGet(ctx, paneID, herdr.WithLogger(logger))
	if isHerdrCode(err, herdr.CodePaneNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, kind := range harness.KnownKinds() {
		profile, lookupErr := harness.Lookup(kind)
		if lookupErr != nil {
			continue
		}
		for _, name := range profile.ProcessNames() {
			names[name] = true
		}
	}
	for _, entry := range info.ForegroundProcesses {
		if names[entry.Name] || names[entryArgvName(entry, true)] || names[entryArgvName(entry, false)] {
			found := entry
			return &found, nil
		}
	}
	return nil, nil
}

// Release removes the starting claim of this launcher. After a restart that
// did not close the old pane, it puts the old claim and record back.
func (l *CoordinatorLaunch) Release() error {
	registry, err := store.Open(l.home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if _, err := l.ownClaim(locked); err != nil {
			return err
		}
		if l.previousRecord != nil {
			if err := locked.WriteIn([]string{"agents"}, l.previousRecord.Key, *l.previousRecord); err != nil {
				return err
			}
		}
		if l.previous != nil {
			return locked.WriteIn([]string{"claims"}, coordinatorClaimKey, *l.previous)
		}
		return locked.RemoveIn([]string{"claims"}, coordinatorClaimKey)
	})
}

// release removes the starting claim and keeps the old record stale. Hive
// uses it after the old pane closed.
func (l *CoordinatorLaunch) release() error {
	l.previous = nil
	l.previousRecord = nil
	return l.Release()
}

func (l *CoordinatorLaunch) ownClaim(locked *store.LockedStore) (records.Claim, error) {
	var claim records.Claim
	if err := locked.ReadIn([]string{"claims"}, coordinatorClaimKey, &claim); err != nil {
		return claim, fmt.Errorf("read coordinator claim: %w", err)
	}
	if claim.Status != records.ClaimStarting || claim.PID != l.pid || claim.PIDStart != l.pidStart || claim.Generation != l.generation {
		return claim, errors.New("coordinator claim changed")
	}
	return claim, nil
}

// setPane writes the target pane into the starting claim before the harness
// starts.
func (l *CoordinatorLaunch) setPane(paneID string) error {
	registry, err := store.Open(l.home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		claim, err := l.ownClaim(locked)
		if err != nil {
			return err
		}
		claim.PaneID = paneID
		return locked.WriteIn([]string{"claims"}, coordinatorClaimKey, claim)
	})
}

// commit writes the coordinator record and sets the claim live in one locked
// write.
func (l *CoordinatorLaunch) commit(record records.Agent) error {
	registry, err := store.Open(l.home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		claim, err := l.ownClaim(locked)
		if err != nil {
			return err
		}
		if err := locked.WriteIn([]string{"agents"}, record.Key, record); err != nil {
			return fmt.Errorf("write coordinator record: %w", err)
		}
		claim.Agent = record.Key
		claim.Status = records.ClaimLive
		claim.PID = record.ProcessPID
		claim.PIDStart = record.ProcessStart
		claim.PaneID = ""
		claim.ClaimedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn([]string{"claims"}, coordinatorClaimKey, claim); err != nil {
			return fmt.Errorf("write coordinator claim: %w", err)
		}
		return nil
	})
}

// markUnresolved sets the committed record unresolved with message.
func (l *CoordinatorLaunch) markUnresolved(key, message string) error {
	registry, err := store.Open(l.home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		return locked.UpdateIn([]string{"agents"}, key, func(raw json.RawMessage) (json.RawMessage, error) {
			var record records.Agent
			if err := json.Unmarshal(raw, &record); err != nil {
				return nil, err
			}
			record.Status = records.AgentUnresolved
			record.LastError = message
			record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			return json.Marshal(record)
		})
	})
}

func findAgent(agents []records.Agent, key string) (records.Agent, bool) {
	for _, record := range agents {
		if record.Key == key {
			return record, true
		}
	}
	return records.Agent{}, false
}

func newestCoordinatorKey(agents []records.Agent) string {
	var newest records.Agent
	for _, record := range agents {
		if record.Role != "coordinator" {
			continue
		}
		if newest.Key == "" || record.Generation > newest.Generation {
			newest = record
		}
	}
	return newest.Key
}

// nextCoordinatorGeneration returns the first generation above the claim and
// each coordinator record whose key is free.
func nextCoordinatorGeneration(claimGeneration uint64, agents []records.Agent) uint64 {
	generation := claimGeneration
	for _, record := range agents {
		if record.Role == "coordinator" && record.Generation > generation {
			generation = record.Generation
		}
	}
	for {
		generation++
		if _, taken := findAgent(agents, fmt.Sprintf("coordinator-g%d", generation)); !taken {
			return generation
		}
	}
}
