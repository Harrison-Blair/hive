package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CoordinatorPrompt is the startup prompt of the coordinator. Hive registers
// the coordinator, so the prompt does not tell it to register. It names the
// profile in place of the old skill (D4).
const CoordinatorPrompt = marker.Prefix + "You are the coordinator. Hive started you and wrote your record. Read profiles/coordinator.md and run the coordinator startup now."

// Start starts the configured harness in paneID under the starting claim.
// When the harness start fails, Start removes the claim. After the start, it
// writes the coordinator record, sets the claim live, and sends the startup
// prompt one time. When a step after the start fails, the record is
// unresolved (decisions 31 and 32).
func (l *CoordinatorLaunch) Start(ctx context.Context, paneID, harnessKind, model string) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	logger := l.logger
	if logger != nil {
		logger.Debug("coordinator harness start", "command", l.command, "home", l.home, "pane_id", paneID, "harness", harnessKind, "model", model)
	}
	started := false
	defer func() {
		if err == nil || started {
			return
		}
		if releaseErr := l.release(); releaseErr != nil {
			err = fmt.Errorf("%w; remove starting claim: %v", err, releaseErr)
		}
	}()
	profile, err := harness.Lookup(harnessKind)
	if err != nil {
		return err
	}
	if err := harness.Check(profile); err != nil {
		return err
	}
	instructions, message := instructionsText(l.home, records.ProfileCoordinator)
	if message != "" {
		return errors.New(message)
	}
	prompt := CoordinatorPrompt
	if profile.SystemPromptArg() == "" && instructions != "" {
		prompt = instructions + "\n\n" + prompt
	}
	systemArgs, file, err := prepareInstructions(l.home, profile, instructions)
	if err != nil {
		return err
	}
	// keepFile is true while the harness can still read the file. Claude Code
	// reads it before it draws its prompt, so a ready start with a checked
	// process identity has read it. A refused start cannot read it.
	keepFile := false
	if file != nil {
		defer func() {
			if !keepFile {
				if removeErr := file.remove(); removeErr == nil {
					return
				} else if logger != nil {
					logger.Error("instruction file remove", "command", l.command, "path", file.path, "error", removeErr.Error())
				}
			}
			if logger != nil {
				logger.Debug("instruction file kept", "command", l.command, "path", file.path, "cleanup", "remove the file after the owner settles the coordinator")
			}
			if err != nil {
				err = fmt.Errorf("%w; instruction file kept until the owner settles the coordinator: %s", err, file.path)
			}
		}()
	}
	if err := l.setPane(paneID); err != nil {
		return fmt.Errorf("set starting claim pane: %w", err)
	}
	if err := herdr.PaneWaitForShellPrompt(ctx, paneID, defaultShellReadyTimeout, herdr.WithLogger(logger)); err != nil {
		return fmt.Errorf("wait for coordinator shell prompt: %w", err)
	}
	if err := waitForShell(ctx, paneID, defaultShellReadyTimeout, logger); err != nil {
		return fmt.Errorf("coordinator pane: %w", err)
	}
	options := []herdr.Option{herdr.WithLogger(logger)}
	launched, err := herdr.AgentStart(ctx, herdr.AgentStartParams{
		Name:      "coordinator",
		Kind:      profile.Kind(),
		PaneID:    paneID,
		Args:      profile.Args(model, systemArgs),
		TimeoutMS: uint64(profile.StartTimeout().Milliseconds()),
	}, options...)
	if err != nil {
		var serverError *herdr.Error
		keepFile = !errors.As(err, &serverError)
		return fmt.Errorf("start coordinator harness: %w", err)
	}
	started = true
	now := time.Now().UTC().Format(time.RFC3339Nano)
	record := records.Agent{
		Key:           fmt.Sprintf("coordinator-g%d", l.generation),
		Name:          "coordinator",
		Harness:       profile.Kind(),
		Model:         model,
		PaneID:        launched.Agent.PaneID,
		WorkspaceID:   launched.Agent.WorkspaceID,
		TabID:         launched.Agent.TabID,
		TerminalID:    launched.Agent.TerminalID,
		Cwd:           l.home,
		Status:        records.AgentLive,
		Replaces:      l.replaces,
		QuietAfterMin: defaultQuietAfterMin,
		Generation:    l.generation,
		StartedAt:     now,
		UpdatedAt:     now,
		StartedByHive: true,
	}
	if failure := l.waitForHarness(ctx, profile, paneID, &record); failure != nil {
		keepFile = true
		record.Status = records.AgentUnresolved
		record.LastError = failure.Error()
		if commitErr := l.commit(record); commitErr != nil {
			return fmt.Errorf("%w; write coordinator record: %v", failure, commitErr)
		}
		return failure
	}
	if err := l.commit(record); err != nil {
		return err
	}
	if logger != nil {
		logger.Debug("coordinator record written", "command", l.command, "agent_id", record.Key, "pane_id", record.PaneID, "process_pid", record.ProcessPID)
	}
	if _, err := herdr.AgentPrompt(ctx, herdr.AgentPromptParams{Target: record.PaneID, Text: prompt}, options...); err != nil {
		failure := fmt.Errorf("send coordinator startup prompt: %w", err)
		if markErr := l.markUnresolved(record.Key, failure.Error()); markErr != nil {
			return fmt.Errorf("%w; mark %s unresolved: %v", failure, record.Key, markErr)
		}
		return failure
	}
	return nil
}

// waitForHarness waits until the harness is ready and records its process.
func (l *CoordinatorLaunch) waitForHarness(ctx context.Context, profile harness.Profile, paneID string, record *records.Agent) error {
	logger := l.logger
	waitOptions := []herdr.Option{herdr.WithLogger(logger), herdr.WithDeadline(profile.ReadyTimeout() + 5*time.Second)}
	ready, err := herdr.AgentWait(ctx, herdr.AgentWaitParams{
		Target:    paneID,
		Until:     SettledStatuses,
		TimeoutMS: uint64(profile.ReadyTimeout().Milliseconds()),
	}, waitOptions...)
	if err != nil {
		return fmt.Errorf("wait for coordinator harness: %w", err)
	}
	if ready.AgentStatus == "blocked" {
		return fmt.Errorf("coordinator harness is blocked during startup")
	}
	if _, err := waitForCoordinatorReady(ctx, paneID, profile.ReadyTimeout(), logger); err != nil {
		return fmt.Errorf("wait for coordinator harness readiness: %w", err)
	}
	info, infoErr := herdr.PaneProcessInfoGet(ctx, paneID, herdr.WithLogger(logger))
	found, reason := captureIdentity(profile, *record, info, infoErr)
	if reason != "" {
		return fmt.Errorf("coordinator process identity unresolved: %s", reason)
	}
	record.ProcessPID = found.pid
	record.ProcessName = found.name
	record.ProcessStart = found.start
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
