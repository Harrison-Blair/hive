package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	homeconfig "github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const noHerdrNotice = "not in herdr: no coordinator or supervisor started"

type coordinatorStarter func(ctx context.Context, launch *agent.CoordinatorLaunch, paneID, harnessKind, model string) error

func setupSession(homePath string, logger *slog.Logger) (string, error) {
	return setupSessionWith(homePath, logger, startCoordinator)
}

func setupSessionWith(homePath string, logger *slog.Logger, startCoordinatorFn coordinatorStarter) (string, error) {
	if os.Getenv("HIVE_NO_STARTUP") == "1" {
		for _, step := range []string{"current tab label", "coordinator", "supervisor"} {
			logStartupStep(logger, step, "skipped", nil, "reason", "no-startup")
		}
		return "", nil
	}
	if os.Getenv("HERDR_ENV") != "1" || os.Getenv("HERDR_PANE_ID") == "" {
		logStartupStep(logger, "herdr environment", "skipped", nil)
		return noHerdrNotice, nil
	}
	ctx := context.Background()
	paneID := os.Getenv("HERDR_PANE_ID")
	pane, err := herdr.PaneGet(ctx, paneID, herdr.WithLogger(logger))
	if err != nil {
		logStartupStep(logger, "current pane", "failed", err)
		return "", fmt.Errorf("current pane: %w", err)
	}
	logStartupStep(logger, "current pane", "ok", nil)

	stepErrors := []string{}
	var coordinatorPaneID string
	var coordinatorLaunch *agent.CoordinatorLaunch
	var coordinatorSettings homeconfig.Settings
	if err := herdr.TabRename(ctx, pane.TabID, "hive", herdr.WithLogger(logger)); err != nil {
		logStartupStep(logger, "current tab label", "failed", err)
		stepErrors = append(stepErrors, "current tab label: "+err.Error())
	} else {
		logStartupStep(logger, "current tab label", "ok", nil)
	}

	coordinatorLive, err := hasLiveCoordinator(ctx, homePath, logger)
	if err != nil {
		logStartupStep(logger, "coordinator", "failed", err)
		stepErrors = append(stepErrors, "coordinator: "+err.Error())
	} else if coordinatorLive {
		logStartupStep(logger, "coordinator", "kept", nil)
	} else {
		settings := homeconfig.Settings{Coordinator: homeconfig.CoordinatorConfig{Harness: "claude", Model: "claude-opus-5-5"}}
		configured, configErr := homeconfig.ReadSettings()
		if configErr == nil {
			settings = configured
		} else if !errors.Is(configErr, os.ErrNotExist) {
			logStartupStep(logger, "coordinator", "failed", configErr)
			stepErrors = append(stepErrors, "coordinator config: "+configErr.Error())
		}
		var launch *agent.CoordinatorLaunch
		if configErr == nil || errors.Is(configErr, os.ErrNotExist) {
			if startCoordinatorFn == nil {
				startErr := errors.New("coordinator starter is unavailable")
				logStartupStep(logger, "coordinator", "failed", startErr)
				stepErrors = append(stepErrors, "coordinator: "+startErr.Error())
			} else if claimed, claimErr := agent.ClaimCoordinatorStart(ctx, homePath, logger); claimErr != nil {
				logStartupStep(logger, "coordinator", "failed", claimErr)
				stepErrors = append(stepErrors, "coordinator: "+claimErr.Error())
			} else {
				launch = claimed
			}
		}
		if launch != nil {
			created, createErr := herdr.TabCreate(ctx, herdr.TabCreateParams{
				WorkspaceID: pane.WorkspaceID,
				Cwd:         homePath,
				Label:       "coordinator",
				Focus:       false,
				Env:         agent.StartupTabEnv(homePath, "coordinator"),
			}, herdr.WithLogger(logger))
			if createErr == nil && (created.Tab.WorkspaceID != pane.WorkspaceID || created.RootPane.WorkspaceID != pane.WorkspaceID || created.RootPane.TabID != created.Tab.TabID) {
				createErr = errors.New("tab create returned a different workspace")
			}
			if createErr != nil {
				if releaseErr := launch.Release(); releaseErr != nil {
					createErr = fmt.Errorf("%w; remove starting claim: %v", createErr, releaseErr)
				}
				logStartupStep(logger, "coordinator", "failed", createErr)
				stepErrors = append(stepErrors, "coordinator tab: "+createErr.Error())
			} else {
				coordinatorLaunch = launch
				coordinatorPaneID = created.RootPane.PaneID
				coordinatorSettings = settings
			}
		}
	}

	supervisorFresh, err := hasFreshSupervisorClaim(homePath)
	if err != nil {
		logStartupStep(logger, "supervisor", "failed", err)
		stepErrors = append(stepErrors, "supervisor: "+err.Error())
	} else if supervisorFresh {
		logStartupStep(logger, "supervisor", "kept", nil)
	} else {
		created, createErr := herdr.TabCreate(ctx, herdr.TabCreateParams{
			WorkspaceID: pane.WorkspaceID,
			Cwd:         homePath,
			Label:       "supervisor",
			Focus:       false,
			Env:         agent.StartupTabEnv(homePath, "supervisor"),
		}, herdr.WithLogger(logger))
		if createErr == nil && (created.Tab.WorkspaceID != pane.WorkspaceID || created.RootPane.WorkspaceID != pane.WorkspaceID || created.RootPane.TabID != created.Tab.TabID) {
			createErr = errors.New("tab create returned a different workspace")
		}
		if createErr != nil {
			logStartupStep(logger, "supervisor", "failed", createErr)
			stepErrors = append(stepErrors, "supervisor tab: "+createErr.Error())
		} else if err := herdr.PaneSendInput(ctx, created.RootPane.PaneID, "hive supervise", herdr.WithLogger(logger)); err != nil {
			logStartupStep(logger, "supervisor", "failed", err)
			stepErrors = append(stepErrors, "supervisor: "+err.Error())
		} else {
			logStartupStep(logger, "supervisor", "started", nil)
		}
	}
	if coordinatorPaneID != "" {
		if startErr := startCoordinatorFn(ctx, coordinatorLaunch, coordinatorPaneID, coordinatorSettings.Coordinator.Harness, coordinatorSettings.Coordinator.Model); startErr != nil {
			logStartupStep(logger, "coordinator", "failed", startErr)
			stepErrors = append(stepErrors, "coordinator: "+startErr.Error())
		} else {
			logStartupStep(logger, "coordinator", "started", nil)
		}
	}
	if len(stepErrors) != 0 {
		return "", errors.New(strings.Join(stepErrors, "; "))
	}
	return "", nil
}

func startCoordinator(ctx context.Context, launch *agent.CoordinatorLaunch, paneID, harnessKind, model string) error {
	return launch.Start(ctx, paneID, harnessKind, model)
}

func hasLiveCoordinator(ctx context.Context, homePath string, logger *slog.Logger) (bool, error) {
	registry, err := store.Open(homePath)
	if err != nil {
		return false, err
	}
	recordsJSON, err := registry.ListIn([]string{"agents"}, nil)
	if err != nil {
		return false, err
	}
	for _, data := range recordsJSON {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return false, err
		}
		if record.Kind != records.AgentKindCoordinator || record.Status != records.AgentLive {
			continue
		}
		gone, err := coordinatorPaneGone(ctx, homePath, record, logger)
		if err != nil {
			return false, err
		}
		if !gone {
			return true, nil
		}
	}
	return false, nil
}

func coordinatorPaneGone(ctx context.Context, homePath string, record records.Agent, logger *slog.Logger) (bool, error) {
	pane, err := herdr.PaneGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if err != nil {
		var serverError *herdr.Error
		if errors.As(err, &serverError) && serverError.Code == herdr.CodePaneNotFound {
			logCoordinatorPaneCheck(logger, record.PaneID, "gone", nil)
			if err := markCoordinatorStale(homePath, record); err != nil {
				return false, fmt.Errorf("mark coordinator %s stale: %w", record.Key, err)
			}
			return true, nil
		}
		logCoordinatorPaneCheck(logger, record.PaneID, "failed", err)
		return false, fmt.Errorf("coordinator pane %s: %w", record.PaneID, err)
	}
	if pane.PaneID != record.PaneID {
		err := fmt.Errorf("pane lookup returned %s", pane.PaneID)
		logCoordinatorPaneCheck(logger, record.PaneID, "failed", err)
		return false, fmt.Errorf("coordinator pane %s: %w", record.PaneID, err)
	}
	logCoordinatorPaneCheck(logger, record.PaneID, "present", nil)
	return false, nil
}

func markCoordinatorStale(homePath string, expected records.Agent) error {
	registry, err := store.Open(homePath)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if current.Kind != records.AgentKindCoordinator || current.Status != records.AgentLive || current.PaneID != expected.PaneID {
			return nil
		}
		current.Status = records.AgentStale
		current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return locked.WriteIn([]string{"agents"}, current.Key, current)
	})
}

func logCoordinatorPaneCheck(logger *slog.Logger, paneID, outcome string, err error) {
	if logger == nil {
		return
	}
	attrs := []any{"component", "tui", "command", "bare hive", "step", "coordinator pane", "outcome", outcome, "pane_id", paneID}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	logger.Debug("coordinator pane check", attrs...)
}

func hasFreshSupervisorClaim(homePath string) (bool, error) {
	registry, err := store.Open(homePath)
	if err != nil {
		return false, err
	}
	var claim records.Claim
	err = registry.ReadIn([]string{"claims"}, "supervisor", &claim)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	alive, err := process.Alive(claim.PID, claim.PIDStart)
	return alive && err == nil, nil
}

func logStartupStep(logger *slog.Logger, step, outcome string, err error, extra ...any) {
	if logger == nil {
		return
	}
	attrs := append([]any{"component", "tui", "command", "bare hive", "step", step, "outcome", outcome}, extra...)
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	logger.Debug("startup step", attrs...)
}
