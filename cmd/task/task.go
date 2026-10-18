// Package task defines the task command and registers its operations.
package task

import (
	"github.com/Harrison-Blair/hive/cmd/task/assign"
	"github.com/Harrison-Blair/hive/cmd/task/cancel"
	"github.com/Harrison-Blair/hive/cmd/task/complete"
	"github.com/Harrison-Blair/hive/cmd/task/create"
	"github.com/Harrison-Blair/hive/cmd/task/list"
	"github.com/Harrison-Blair/hive/cmd/task/show"
	"github.com/Harrison-Blair/hive/cmd/task/verify"
	"github.com/spf13/cobra"
)

// New creates the task command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "task",
		Short: "manage task records",
		Args:  cobra.NoArgs,
	}
	command.PersistentFlags().String("agent", "", "acting agent name")
	command.AddCommand(
		create.New(),
		assign.New(),
		list.New(),
		show.New(),
		complete.New(),
		verify.New(),
		cancel.New(),
	)
	return command
}
