// Package common shares persistent command context and output handling.
package common

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// StatusError carries the process status for an error already rendered to a stream.
type StatusError struct {
	Status int
}

func (e *StatusError) Error() string { return fmt.Sprintf("command exited with status %d", e.Status) }

// UsageError reports an invalid command argument. The root command renders it
// as a usage error.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string { return e.Message }

// Run resolves the home, logs the command, and renders its result. Target is
// the agent that the command acts on, or empty. Execute gets the command logger.
func Run(cmd *cobra.Command, command, target string, execute func(string, *slog.Logger) (output.Envelope, string, error)) error {
	root := cmd.Root()
	jsonOutput, err := root.PersistentFlags().GetBool("json")
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return renderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, "home", err.Error(), 1)
	}
	resolvedHome, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return renderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, "home", err.Error(), 1)
	}
	requestID, err := hivelog.NewRequestID()
	if err != nil {
		return renderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, "log", err.Error(), 1)
	}
	restoreRequestID := hivelog.SetRequestID(requestID)
	defer restoreRequestID()
	logger, err := Logger(resolvedHome, target)
	if err != nil {
		return renderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, "log", err.Error(), 1)
	}
	started := time.Now()
	logger.Debug("command start", "command", command)
	envelope, text, executeErr := execute(resolvedHome, logger)
	if executeErr != nil {
		var usage *UsageError
		if errors.As(executeErr, &usage) {
			logger.Debug("command end", "command", command, "duration_ms", time.Since(started).Milliseconds(), "outcome", output.OutcomeFailed, "error", usage.Error())
			return usage
		}
		envelope = output.Failed("command", executeErr.Error())
		text = ""
	}
	attrs := []any{"command", command, "duration_ms", time.Since(started).Milliseconds(), "outcome", envelope.Outcome}
	if envelope.Outcome != output.OutcomeOK && envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("command end", attrs...)
	if err := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, text); err != nil {
		return &StatusError{Status: 1}
	}
	return &StatusError{Status: output.Status(envelope.Outcome)}
}

// Logger opens the command logger of the home. Each event carries agent_id,
// the caller, and target when target is not empty. The caller is the record
// name of HERDR_PANE_ID, or "user" when that pane has no record.
func Logger(home, target string) (*slog.Logger, error) {
	logger, err := hivelog.Open(home, "cli")
	if err != nil {
		return nil, err
	}
	caller, lookupErr := agent.CallerName(home, os.Getenv("HERDR_PANE_ID"))
	if caller == "" {
		caller = "user"
	}
	logger = logger.With("agent_id", caller)
	if target != "" {
		logger = logger.With("target", target)
	}
	if lookupErr != nil {
		logger.Debug("caller lookup failed", "error", lookupErr.Error())
	}
	return logger, nil
}

// LogRejected logs command start and a failed command end for a command that
// Cobra rejected before it ran, for example for a wrong argument count or an
// unknown flag. It logs nothing when the home does not resolve.
func LogRejected(command string, rejected error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	resolvedHome, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return
	}
	requestID, err := hivelog.NewRequestID()
	if err != nil {
		return
	}
	restoreRequestID := hivelog.SetRequestID(requestID)
	defer restoreRequestID()
	logger, err := Logger(resolvedHome, "")
	if err != nil {
		return
	}
	started := time.Now()
	logger.Debug("command start", "command", command)
	logger.Debug("command end", "command", command, "duration_ms", time.Since(started).Milliseconds(), "outcome", output.OutcomeFailed, "error", rejected.Error())
}

// Usage writes one usage line and returns status 2.
func Usage(cmd *cobra.Command, message string) error {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %s\n", message)
	return &StatusError{Status: 2}
}

func renderFailure(stdout, stderr io.Writer, jsonOutput bool, phase, message string, status int) error {
	if err := output.Render(stdout, stderr, jsonOutput, output.Failed(phase, message), ""); err != nil {
		return &StatusError{Status: status}
	}
	return &StatusError{Status: status}
}
