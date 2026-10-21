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
	if record.Status != records.AgentLive && record.Status != records.AgentStale && record.Status != records.AgentStopped && record.Status != records.AgentExited {
		return RestartResult{}, fmt.Errorf("cannot restart %s with status %s", record.Name, record.Status)
	}
	workspaceID, err := restartWorkspace(ctx, record, logger)
	if err != nil {
		return RestartResult{}, err
	}
	settings := homeconfig.Settings{Coordinator: homeconfig.CoordinatorConfig{Harness: "claude", Model: "claude-opus-5-5"}}
	if record.Role == "coordinator" {
		configured, configErr := homeconfig.ReadSettings()
		if configErr == nil {
			settings = configured
		} else if !errors.Is(configErr, os.ErrNotExist) {
			return RestartResult{}, fmt.Errorf("read coordinator config: %w", configErr)
		}
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
	if record.Role == "coordinator" {
		if err := StartCoordinator(ctx, home, created.RootPane.PaneID, settings.Coordinator.Harness, settings.Coordinator.Model, logger); err != nil {
			return result, fmt.Errorf("start coordinator: %w", err)
		}
	} else if err := herdr.PaneSendInput(ctx, created.RootPane.PaneID, "hive supervise", herdr.WithLogger(logger)); err != nil {
		return result, fmt.Errorf("start supervisor: %w", err)
	}
	return result, nil
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
