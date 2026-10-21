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
	if ctx == nil {
		ctx = context.Background()
	}
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
	if err := herdr.PaneWaitForShellPrompt(ctx, paneID, defaultShellReadyTimeout, herdr.WithLogger(logger)); err != nil {
		return fmt.Errorf("wait for coordinator shell prompt: %w", err)
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
	if _, err := waitForCoordinatorReady(ctx, started.Agent.PaneID, profile.ReadyTimeout(), logger); err != nil {
		return fmt.Errorf("wait for coordinator harness readiness: %w", err)
	}
	prompt := marker.Prefix + "You are the coordinator. Run the coordinator startup in .agents/skills/hive-coordinator/SKILL.md now."
	if _, err := herdr.AgentPrompt(ctx, herdr.AgentPromptParams{Target: started.Agent.PaneID, Text: prompt}, options...); err != nil {
		return fmt.Errorf("send coordinator startup prompt: %w", err)
	}
	return nil
}

const coordinatorReadyPollInterval = 250 * time.Millisecond

func waitForCoordinatorReady(ctx context.Context, paneID string, timeout time.Duration, logger *slog.Logger) (info herdr.AgentInfo, err error) {
	started := time.Now()
	defer func() {
		if logger == nil {
			return
		}
		attrs := []any{"pane_id", paneID, "duration_ms", time.Since(started).Milliseconds()}
		if err != nil {
			attrs = append(attrs, "outcome", "failed", "error", err.Error())
		} else {
			attrs = append(attrs, "outcome", "ready")
		}
		logger.Debug("coordinator harness readiness wait", attrs...)
	}()
	if ctx == nil {
		ctx = context.Background()
	}
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return herdr.AgentInfo{}, &herdr.Error{Code: herdr.CodeTimeout, Message: fmt.Sprintf("coordinator agent %s did not become ready within %s", paneID, timeout)}
		}
		info, err = herdr.AgentGet(ctx, paneID, herdr.WithDeadline(remaining), herdr.WithLogger(logger))
		if err != nil {
			return herdr.AgentInfo{}, err
		}
		if info.Interactive && !info.LaunchPending {
			return info, nil
		}
		remaining = time.Until(deadline)
		if remaining <= 0 {
			continue
		}
		pause := coordinatorReadyPollInterval
		if remaining < pause {
			pause = remaining
		}
		timer := time.NewTimer(pause)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return herdr.AgentInfo{}, ctx.Err()
		case <-timer.C:
		}
	}
}
