// Package assign wires the task assign command.
package assign

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	operation "github.com/Harrison-Blair/hive/internal/task/assign"
	"github.com/spf13/cobra"
)

// New creates the task assign command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "assign",
		Short: "assign a task to an agent",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("task", "", "task ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "task assign", func(home string) (output.Envelope, string, error) {
			taskID, _ := cmd.Flags().GetString("task")
			agent, _ := cmd.Flags().GetString("agent")
			envelope := operation.Run(home, tasklib.TransitionOptions{Task: taskID, Agent: agent})
			return envelope, tasklib.TransitionText(envelope), nil
		})
	}
	return command
}
