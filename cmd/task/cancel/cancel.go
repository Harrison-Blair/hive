// Package cancel wires the task cancel command.
package cancel

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task cancel command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "cancel",
		Short: "cancel a task",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.String("reason", "", "reason of the cancellation")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "task cancel", func(home string) (output.Envelope, string, error) {
			taskID, _ := cmd.Flags().GetString("task")
			reason, _ := cmd.Flags().GetString("reason")
			envelope := tasklib.Cancel(home, tasklib.TransitionOptions{Task: taskID, Reason: reason})
			return envelope, tasklib.TransitionText(envelope), nil
		})
	}
	return command
}
