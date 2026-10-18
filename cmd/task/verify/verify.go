// Package verify wires the task verify command.
package verify

import (
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
		return common.Run(cmd, "task verify", func(home string) (output.Envelope, string, error) {
			taskID, _ := cmd.Flags().GetString("task")
			agent, _ := cmd.Flags().GetString("agent")
			fail, _ := cmd.Flags().GetBool("fail")
			reason, _ := cmd.Flags().GetString("reason")
			envelope := tasklib.Verify(home, tasklib.TransitionOptions{
				Task: taskID, Agent: agent, Fail: fail, Reason: reason,
			})
			return envelope, tasklib.TransitionText(envelope), nil
		})
	}
	return command
}
