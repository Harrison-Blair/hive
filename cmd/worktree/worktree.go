// Package worktree defines the worktree command and registers its operations.
package worktree

import (
	"github.com/Harrison-Blair/hive/cmd/worktree/create"
	"github.com/Harrison-Blair/hive/cmd/worktree/remove"
	"github.com/spf13/cobra"
)

// New creates the worktree command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "worktree",
		Short: "manage task worktrees",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(create.New(), remove.New())
	return command
}
