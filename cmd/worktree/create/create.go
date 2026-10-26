// Package create wires the worktree create command.
package create

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	operation "github.com/Harrison-Blair/hive/internal/worktree/create"
	"github.com/spf13/cobra"
)

// New creates the worktree create command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "create the worktree of a task",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.Bool("focus", false, "focus the new workspace")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "worktree create", func(home string) (output.Envelope, string, error) {
			task, _ := cmd.Flags().GetString("task")
			focus, _ := cmd.Flags().GetBool("focus")
			envelope := operation.Run(cmd.Context(), home, operation.Options{Task: task, Focus: focus})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
