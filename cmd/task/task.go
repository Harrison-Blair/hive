// Package task defines the task command and registers its operations.
package task

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/task/assign"
	"github.com/Harrison-Blair/hive/cmd/task/cancel"
	"github.com/Harrison-Blair/hive/cmd/task/complete"
	"github.com/Harrison-Blair/hive/cmd/task/create"
	"github.com/Harrison-Blair/hive/cmd/task/edit"
	"github.com/Harrison-Blair/hive/cmd/task/list"
	"github.com/Harrison-Blair/hive/cmd/task/reopen"
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
	command.PersistentFlags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.AddCommand(
		create.New(),
		edit.New(),
		assign.New(),
		list.New(),
		show.New(),
		complete.New(),
		verify.New(),
		cancel.New(),
		reopen.New(),
	)
	return command
}
