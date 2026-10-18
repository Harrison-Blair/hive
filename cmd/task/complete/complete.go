// Package complete wires the task complete command.
package complete

import (
	"fmt"
	"io"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task complete command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "complete",
		Short: "mark a task complete",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("task", "", "task ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		taskID, _ := cmd.Flags().GetString("task")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		var envelope output.Envelope
		original := cmd.OutOrStdout()
		tracker := &trackingWriter{writer: original}
		cmd.SetOut(tracker)
		err := common.Run(cmd, "task complete", func(home string) (output.Envelope, string, error) {
			envelope = tasklib.Complete(home, tasklib.TransitionOptions{Task: taskID, Agent: agent, PaneID: pane})
			return envelope, tasklib.TransitionText(envelope), nil
		})
		cmd.SetOut(original)
		if tracker.err != nil {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), tasklib.FailureReport(envelope, taskID, agent))
			return &common.StatusError{Status: 1}
		}
		return err
	}
	return command
}

type trackingWriter struct {
	writer io.Writer
	err    error
}

func (w *trackingWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil && w.err == nil {
		w.err = err
	}
	return n, err
}
