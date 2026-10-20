package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	homeconfig "github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const noHerdrNotice = "not in herdr: no coordinator or supervisor started"

type coordinatorStarter func(context.Context, string, string, string, string, *slog.Logger) error

func setupSession(homePath string, logger *slog.Logger) (string, error) {
	return setupSessionWith(homePath, logger, startCoordinator)
}

func setupSessionWith(homePath string, logger *slog.Logger, startCoordinatorFn coordinatorStarter) (string, error) {
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
	var coordinatorSettings homeconfig.Settings
	if err := herdr.TabRename(ctx, pane.TabID, "hive", herdr.WithLogger(logger)); err != nil {
		logStartupStep(logger, "current tab label", "failed", err)
		stepErrors = append(stepErrors, "current tab label: "+err.Error())
	} else {
		logStartupStep(logger, "current tab label", "ok", nil)
	}

	coordinatorLive, err := hasLiveCoordinator(homePath)
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
		if configErr == nil || errors.Is(configErr, os.ErrNotExist) {
			created, createErr := herdr.TabCreate(ctx, herdr.TabCreateParams{
				WorkspaceID: pane.WorkspaceID,
				Cwd:         homePath,
				Label:       "coordinator",
				Focus:       false,
				Env:         startupTabEnv(homePath),
			}, herdr.WithLogger(logger))
			if createErr != nil {
				logStartupStep(logger, "coordinator", "failed", createErr)
				stepErrors = append(stepErrors, "coordinator tab: "+createErr.Error())
			} else if created.Tab.WorkspaceID != pane.WorkspaceID || created.RootPane.WorkspaceID != pane.WorkspaceID || created.RootPane.TabID != created.Tab.TabID {
				createErr = errors.New("tab create returned a different workspace")
				logStartupStep(logger, "coordinator", "failed", createErr)
				stepErrors = append(stepErrors, "coordinator tab: "+createErr.Error())
			} else if startCoordinatorFn == nil {
				startErr := errors.New("coordinator starter is unavailable")
				logStartupStep(logger, "coordinator", "failed", startErr)
				stepErrors = append(stepErrors, "coordinator: "+startErr.Error())
			} else {
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
			Env:         startupTabEnv(homePath),
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
		if startErr := startCoordinatorFn(ctx, homePath, coordinatorPaneID, coordinatorSettings.Coordinator.Harness, coordinatorSettings.Coordinator.Model, logger); startErr != nil {
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

func startupTabEnv(homePath string) map[string]string {
	env := map[string]string{"HIVE_HOME": homePath}
	for _, key := range []string{"XDG_CONFIG_HOME", "PATH"} {
		if value := os.Getenv(key); value != "" {
			env[key] = value
		}
	}
	return env
}

func startCoordinator(ctx context.Context, homePath, paneID, harnessKind, model string, logger *slog.Logger) error {
	return agent.StartCoordinator(ctx, homePath, paneID, harnessKind, model, logger)
}

func hasLiveCoordinator(homePath string) (bool, error) {
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
		if record.Role == "coordinator" && record.Status == records.AgentLive {
			return true, nil
		}
	}
	return false, nil
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

func logStartupStep(logger *slog.Logger, step, outcome string, err error) {
	if logger == nil {
		return
	}
	attrs := []any{"component", "tui", "command", "bare hive", "step", step, "outcome", outcome}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	logger.Debug("startup step", attrs...)
}
