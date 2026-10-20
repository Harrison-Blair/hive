package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
)

// StartCoordinator starts the configured harness and sends its startup prompt.
func StartCoordinator(ctx context.Context, home, paneID, harnessKind, model string, logger *slog.Logger) error {
	if logger != nil {
		logger.Debug("coordinator harness start", "command", "bare hive", "home", home, "pane_id", paneID, "harness", harnessKind, "model", model)
	}
	profile, err := harness.Lookup(harnessKind)
	if err != nil {
		return err
	}
	if err := harness.Check(profile); err != nil {
		return err
	}
	if err := waitForShell(ctx, paneID, defaultShellReadyTimeout, logger); err != nil {
		return fmt.Errorf("coordinator pane: %w", err)
	}
	options := []herdr.Option{herdr.WithLogger(logger)}
	started, err := herdr.AgentStart(ctx, herdr.AgentStartParams{
		Name:      "coordinator",
		Kind:      profile.Kind(),
		PaneID:    paneID,
		Args:      profile.Args(model, nil),
		TimeoutMS: uint64(profile.StartTimeout().Milliseconds()),
	}, options...)
	if err != nil {
		return fmt.Errorf("start coordinator harness: %w", err)
	}
	waitOptions := append(options, herdr.WithDeadline(profile.ReadyTimeout()+5*time.Second))
	ready, err := herdr.AgentWait(ctx, herdr.AgentWaitParams{
		Target:    paneID,
		Until:     settledStatuses,
		TimeoutMS: uint64(profile.ReadyTimeout().Milliseconds()),
	}, waitOptions...)
	if err != nil {
		return fmt.Errorf("wait for coordinator harness: %w", err)
	}
	if ready.AgentStatus == "blocked" {
		return fmt.Errorf("coordinator harness is blocked during startup")
	}
	prompt := marker.Prefix + "You are the coordinator. Run the coordinator startup in .agents/skills/hive-coordinator/SKILL.md now."
	if _, err := herdr.AgentPrompt(ctx, herdr.AgentPromptParams{Target: started.Agent.PaneID, Text: prompt}, options...); err != nil {
		return fmt.Errorf("send coordinator startup prompt: %w", err)
	}
	return nil
}
