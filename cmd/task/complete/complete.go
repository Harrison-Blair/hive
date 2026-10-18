// Package complete wires the task complete command.
package complete

import (
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
		return common.Run(cmd, "task complete", func(home string) (output.Envelope, string, error) {
			taskID, _ := cmd.Flags().GetString("task")
			agent, _ := cmd.Flags().GetString("agent")
			envelope := tasklib.Complete(home, tasklib.TransitionOptions{Task: taskID, Agent: agent})
			return envelope, tasklib.TransitionText(envelope), nil
		})
	}
	return command
}
