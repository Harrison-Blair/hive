package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

const (
	defaultShellReadyTimeout = 10 * time.Second
	shellReadyPollInterval   = 200 * time.Millisecond
)

func waitForShell(ctx context.Context, paneID string, timeout time.Duration, logger *slog.Logger) error {
	if ctx == nil {
		ctx = context.Background()
	}
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return &herdr.Error{Code: herdr.CodeTimeout, Message: fmt.Sprintf("pane %s did not become an available shell within %s", paneID, timeout)}
		}
		pane, paneErr := herdr.PaneGet(ctx, paneID, herdr.WithDeadline(remaining), herdr.WithLogger(logger))
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if paneErr == nil && pane.PaneID == paneID {
			if pane.Agent != "" || (pane.AgentStatus != "" && pane.AgentStatus != "unknown") {
				return &herdr.Error{Code: "agent_pane_busy", Message: "pane is busy"}
			}
			if pane.AgentStatus == "unknown" && pane.Agent == "" {
				remaining = time.Until(deadline)
				if remaining > 0 {
					processInfo, processErr := herdr.PaneProcessInfoGet(ctx, paneID, herdr.WithDeadline(remaining), herdr.WithLogger(logger))
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if processErr == nil && processInfo.PaneID == paneID && processInfo.ShellPID != nil && processInfo.ForegroundProcessGroupID != nil && *processInfo.ShellPID == *processInfo.ForegroundProcessGroupID {
						return nil
					}
				}
			}
		}
		remaining = time.Until(deadline)
		if remaining <= 0 {
			return &herdr.Error{Code: herdr.CodeTimeout, Message: fmt.Sprintf("pane %s did not become an available shell within %s", paneID, timeout)}
		}
		pause := shellReadyPollInterval
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
			return ctx.Err()
		case <-timer.C:
		}
	}
}
