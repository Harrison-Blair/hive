package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	homeconfig "github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// RestartResult reports the tab opened for one restarted system agent.
type RestartResult struct {
	Tab     herdr.TabCreated `json:"tab"`
	Warning string           `json:"warning,omitempty"`
}

// Restart closes an owned coordinator or supervisor pane and starts a new one.
func Restart(ctx context.Context, home string, record records.Agent, logger *slog.Logger) (result RestartResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger != nil {
		logger.Debug("agent restart start", "command", "tui agent restart", "agent_id", record.Key, "role", record.Role)
		defer func() {
			attrs := []any{"command", "tui agent restart", "agent_id", record.Key, "role", record.Role}
			if err != nil {
				attrs = append(attrs, "outcome", "failed", "error", err.Error())
			} else {
				attrs = append(attrs, "outcome", "started", "tab_id", result.Tab.Tab.TabID, "pane_id", result.Tab.RootPane.PaneID)
				if result.Warning != "" {
					attrs = append(attrs, "warning", result.Warning)
				}
			}
			logger.Debug("agent restart result", attrs...)
		}()
	}
	if record.Role != "coordinator" && record.Role != "supervisor" {
		return RestartResult{}, fmt.Errorf("restart applies only to the coordinator and the supervisor")
	}
	if record.Role == "coordinator" {
		return restartCoordinator(ctx, home, record, logger)
	}
	if record.Status != records.AgentLive && record.Status != records.AgentStale && record.Status != records.AgentStopped && record.Status != records.AgentExited {
		return RestartResult{}, fmt.Errorf("cannot restart %s with status %s", record.Name, record.Status)
	}
	workspaceID, err := restartWorkspace(ctx, record, logger)
	if err != nil {
		return RestartResult{}, err
	}

	newStatus := record.Status
	if record.Status == records.AgentLive || record.Status == records.AgentStale {
		if record.StartedByHive {
			if record.PaneID != "" {
				closeErr := herdr.PaneClose(ctx, record.PaneID, herdr.WithLogger(logger))
				if closeErr != nil && !isHerdrCode(closeErr, herdr.CodePaneNotFound) {
					return RestartResult{}, fmt.Errorf("close old %s pane: %w", record.Role, closeErr)
				}
			}
			newStatus = records.AgentStopped
		} else {
			newStatus = records.AgentStale
			result.Warning = "old pane kept because Hive did not start it"
		}
	}
	if newStatus != record.Status {
		if err := updateRestartedRecord(home, record, newStatus); err != nil {
			return RestartResult{}, fmt.Errorf("update old %s record: %w", record.Role, err)
		}
	}

	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{
		WorkspaceID: workspaceID,
		Cwd:         home,
		Label:       record.Role,
		Focus:       false,
		Env:         StartupTabEnv(home, record.Role),
	}, herdr.WithLogger(logger))
	if err != nil {
		return RestartResult{}, fmt.Errorf("create %s tab: %w", record.Role, err)
	}
	if created.Tab.WorkspaceID != workspaceID || created.RootPane.WorkspaceID != workspaceID || created.RootPane.TabID != created.Tab.TabID {
		return RestartResult{}, errors.New("tab create returned a different workspace")
	}
	result.Tab = created
	if err := herdr.PaneSendInput(ctx, created.RootPane.PaneID, "hive supervise", herdr.WithLogger(logger)); err != nil {
		return result, fmt.Errorf("start supervisor: %w", err)
	}
	return result, nil
}

// restartCoordinator follows decision 34. It takes a new starting claim and
// marks the old record stale in one locked write. It closes the old pane when
// hive started the old harness, then starts the new harness under the claim.
func restartCoordinator(ctx context.Context, home string, record records.Agent, logger *slog.Logger) (RestartResult, error) {
	var result RestartResult
	if record.Status == records.AgentStarting {
		return result, fmt.Errorf("cannot restart %s with status %s", record.Name, record.Status)
	}
	workspaceID, err := restartWorkspace(ctx, record, logger)
	if err != nil {
		return result, err
	}
	settings := homeconfig.Settings{Coordinator: homeconfig.CoordinatorConfig{Harness: "claude", Model: "claude-opus-5-5"}}
	configured, configErr := homeconfig.ReadSettings()
	if configErr == nil {
		settings = configured
	} else if !errors.Is(configErr, os.ErrNotExist) {
		return result, fmt.Errorf("read coordinator config: %w", configErr)
	}
	launch, err := claimCoordinator(ctx, home, &record, logger)
	if err != nil {
		return result, err
	}
	closed := false
	if record.Status == records.AgentLive || record.Status == records.AgentUnresolved {
		if record.StartedByHive && record.PaneID != "" {
			if err := closeCoordinatorPane(ctx, record, logger); err != nil {
				if releaseErr := launch.Release(); releaseErr != nil {
					err = fmt.Errorf("%w; restore coordinator claim: %v", err, releaseErr)
				}
				return result, fmt.Errorf("close old coordinator pane: %w", err)
			}
			closed = true
		} else {
			result.Warning = "old pane kept because Hive did not start it"
		}
	}
	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{
		WorkspaceID: workspaceID,
		Cwd:         home,
		Label:       record.Role,
		Focus:       false,
		Env:         StartupTabEnv(home, record.Role),
	}, herdr.WithLogger(logger))
	if err == nil && (created.Tab.WorkspaceID != workspaceID || created.RootPane.WorkspaceID != workspaceID || created.RootPane.TabID != created.Tab.TabID) {
		err = errors.New("tab create returned a different workspace")
	}
	if err != nil {
		release := launch.Release
		if closed {
			release = launch.release
		}
		if releaseErr := release(); releaseErr != nil {
			err = fmt.Errorf("%w; remove starting claim: %v", err, releaseErr)
		}
		return result, fmt.Errorf("create %s tab: %w", record.Role, err)
	}
	result.Tab = created
	if err := launch.Start(ctx, created.RootPane.PaneID, settings.Coordinator.Harness, settings.Coordinator.Model); err != nil {
		return result, fmt.Errorf("start coordinator: %w", err)
	}
	return result, nil
}

// closeCoordinatorPane closes the pane of an old coordinator that hive
// started, with the identity rules of hive agent stop. It closes no pane that
// holds another terminal or another harness process.
func closeCoordinatorPane(ctx context.Context, record records.Agent, logger *slog.Logger) error {
	pane, err := herdr.PaneGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if isHerdrCode(err, herdr.CodePaneNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if record.TerminalID != "" && pane.TerminalID != record.TerminalID {
		return nil
	}
	if record.ProcessPID > 0 {
		info, err := herdr.PaneProcessInfoGet(ctx, record.PaneID, herdr.WithLogger(logger))
		if err != nil {
			return fmt.Errorf("process info: %w", err)
		}
		reason := checkProcessIdentity(record, info, pane.AgentSession)
		if reason == "process_pid is not in the foreground process set" {
			if other, otherErr := paneHarnessProcess(ctx, record.PaneID, logger); otherErr != nil || other != nil {
				reason = "pane " + record.PaneID + " holds another harness process"
			} else {
				reason = ""
			}
		}
		if reason != "" {
			return errors.New("old coordinator identity unresolved: " + reason)
		}
	}
	err = herdr.PaneClose(ctx, record.PaneID, herdr.WithLogger(logger))
	if err != nil && !isHerdrCode(err, herdr.CodePaneNotFound) {
		return err
	}
	return nil
}

// StartupTabEnv sets the environment that a bare hive startup tab needs.
func StartupTabEnv(homePath, role string) map[string]string {
	env := map[string]string{"HIVE_HOME": homePath, "HIVE_STARTUP_ROLE": role}
	for _, key := range []string{"XDG_CONFIG_HOME", "PATH"} {
		if value := os.Getenv(key); value != "" {
			env[key] = value
		}
	}
	return env
}

func restartWorkspace(ctx context.Context, record records.Agent, logger *slog.Logger) (string, error) {
	if record.WorkspaceID != "" {
		return record.WorkspaceID, nil
	}
	paneID := os.Getenv("HERDR_PANE_ID")
	if paneID == "" {
		return "", errors.New("restart has no workspace ID")
	}
	pane, err := herdr.PaneGet(ctx, paneID, herdr.WithLogger(logger))
	if err != nil {
		return "", fmt.Errorf("read current pane: %w", err)
	}
	if pane.WorkspaceID == "" {
		return "", errors.New("current pane has no workspace ID")
	}
	return pane.WorkspaceID, nil
}

func updateRestartedRecord(home string, expected records.Agent, status string) error {
	registry, err := store.Open(home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			return err
		}
		if current.Key != expected.Key || current.Role != expected.Role || current.PaneID != expected.PaneID {
			return errRecordChanged
		}
		if current.Status != expected.Status && current.Status != records.AgentStale && current.Status != records.AgentStopped && current.Status != records.AgentExited {
			return errRecordChanged
		}
		current.Status = status
		current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return locked.WriteIn([]string{"agents"}, current.Key, current)
	})
}
