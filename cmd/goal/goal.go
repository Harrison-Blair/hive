// Package goal wires the goal command tree.
package goal

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/goal/create"
	"github.com/Harrison-Blair/hive/cmd/goal/edit"
	"github.com/Harrison-Blair/hive/cmd/goal/list"
	"github.com/Harrison-Blair/hive/cmd/goal/show"
	"github.com/spf13/cobra"
)

// New creates the goal command tree.
func New() *cobra.Command {
	command := &cobra.Command{Use: "goal", Short: "manage project goals", Args: cobra.NoArgs}
	command.PersistentFlags().String("agent", "", "acting agent name")
	command.PersistentFlags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.AddCommand(create.New(), edit.New(), show.New(), list.New())
	return command
}
