// Package verify wires the task verify command.
package verify

import (
	"fmt"
	"io"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task verify command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "verify",
		Short: "record one independent check of a task",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.Bool("fail", false, "record a failed check")
	flags.String("reason", "", "reason of a failed check")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		taskID, _ := cmd.Flags().GetString("task")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		fail, _ := cmd.Flags().GetBool("fail")
		reason, _ := cmd.Flags().GetString("reason")
		var envelope output.Envelope
		original := cmd.OutOrStdout()
		tracker := &trackingWriter{writer: original}
		cmd.SetOut(tracker)
		err := common.Run(cmd, "task verify", func(home string) (output.Envelope, string, error) {
			envelope = tasklib.Verify(home, tasklib.TransitionOptions{
				Task: taskID, Agent: agent, PaneID: pane, Fail: fail, Reason: reason,
			})
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
