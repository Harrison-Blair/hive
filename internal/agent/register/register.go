// Package register registers the current agent session for the agent register command.
package register

import (
	"context"
	"log/slog"
	"time"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Run registers one agent and best-effort labels its tab. It is
// the CLI entry, so it refuses each name that hive code registers.
func Run(home string, options agentlib.RegisterOptions) output.Envelope {
	if message := agentlib.ReservedNameError(options.Name, options.NameSet); message != "" {
		envelope := agentlib.Failed("validate", message)
		agentlib.LogRegistration(home, options, envelope)
		return envelope
	}
	envelope := agentlib.Register(home, options)
	if envelope.Outcome != "ok" {
		return envelope
	}
	result, ok := envelope.Result.(agentlib.RegisterResult)
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
		agentlib.LogTabLabel(home, logger, "agent register", result.Agent.Key, "", "", err)
		return envelope
	}
	effect := agentlib.RenameAgentTab(context.Background(), home, logger, pane.TabID, result.Agent.Name, "agent register", result.Agent.Key)
	envelope.Effects = append(envelope.Effects, effect)
	return envelope
}

// Text renders the effects of one registration.
func Text(envelope output.Envelope) string {
	return agentlib.EffectsText(envelope)
}
