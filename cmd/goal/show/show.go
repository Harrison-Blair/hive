// Package show wires the goal show command.
package show

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/goal/show"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the goal show command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "show ID", Short: "show a goal", Args: cobra.ExactArgs(1)}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "goal show", func(home string) (output.Envelope, string, error) {
			envelope := operation.Run(home, args[0])
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
