// Package common shares persistent command context and output handling.
package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

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

// UsageError reports an invalid command argument.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string { return e.Message }

// Run resolves the home, logs the command, and renders its result.
func Run(cmd *cobra.Command, command string, execute func(string) (output.Envelope, string, error)) error {
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
	logger, err := hivelog.Open(resolvedHome, "cli")
	if err != nil {
		return renderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, "log", err.Error(), 1)
	}
	started := time.Now()
	logger.Debug("command start", "command", command)
	envelope, text, executeErr := execute(resolvedHome)
	if executeErr != nil {
		var usage *UsageError
		if errors.As(executeErr, &usage) {
			logger.Debug("command end", "command", command, "duration_ms", time.Since(started).Milliseconds(), "outcome", "failed")
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %s\n", usage.Error())
			return &StatusError{Status: 2}
		}
		envelope = output.Envelope{Outcome: "failed", Phase: "command", Effects: []output.Effect{}, Error: executeErr.Error(), Result: nil}
		text = ""
	}
	logger.Debug("command end", "command", command, "duration_ms", time.Since(started).Milliseconds(), "outcome", envelope.Outcome)
	if err := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, text); err != nil {
		return &StatusError{Status: 1}
	}
	return &StatusError{Status: output.Status(envelope.Outcome)}
}

// Usage writes one usage line and returns status 2.
func Usage(cmd *cobra.Command, message string) error {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %s\n", message)
	return &StatusError{Status: 2}
}

func renderFailure(stdout, stderr io.Writer, jsonOutput bool, phase, message string, status int) error {
	envelope := output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
	if err := output.Render(stdout, stderr, jsonOutput, envelope, ""); err != nil {
		return &StatusError{Status: status}
	}
	return &StatusError{Status: status}
}
