// Package list wires the goal list command.
package list

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/goal/list"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the goal list command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "list", Short: "list goals", Args: cobra.NoArgs}
	command.Flags().String("project", "", "project ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "goal list", func(home string) (output.Envelope, string, error) {
			project, _ := cmd.Flags().GetString("project")
			envelope := operation.Run(home, project)
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
