// Package remove wires the worktree remove command.
package remove

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	worktreelib "github.com/Harrison-Blair/hive/internal/worktree"
	"github.com/spf13/cobra"
)

// New creates the worktree remove command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "remove",
		Short: "remove the worktree of a task",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.Bool("force", false, "skip the dirty and merged guards")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "worktree remove", func(home string) (output.Envelope, string, error) {
			task, _ := cmd.Flags().GetString("task")
			force, _ := cmd.Flags().GetBool("force")
			envelope := worktreelib.Remove(cmd.Context(), home, worktreelib.RemoveOptions{Task: task, Force: force})
			return envelope, worktreelib.RemoveText(envelope), nil
		})
	}
	return command
}
