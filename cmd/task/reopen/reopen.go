// Package reopen wires the task reopen command.
package reopen

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task reopen command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "reopen",
		Short: "reopen a verified task for repair",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.String("reason", "", "reason to reopen the task")
	_ = command.MarkFlagRequired("reason")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		taskID, _ := cmd.Flags().GetString("task")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		reason, _ := cmd.Flags().GetString("reason")
		return common.Run(cmd, "task reopen", func(home string) (output.Envelope, string, error) {
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			envelope := tasklib.Reopen(home, tasklib.TransitionOptions{
				Task: taskID, Agent: agent, PaneID: pane, Reason: reason,
			})
			return envelope, tasklib.TransitionText(envelope), nil
		})
	}
	return command
}
