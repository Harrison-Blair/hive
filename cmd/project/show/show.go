// Package show wires the project show command.
package show

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
	"github.com/spf13/cobra"
)

// New creates the project show command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "show ID",
		Short: "show one registered project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.Run(cmd, "project show", func(home string) (output.Envelope, string, error) {
				envelope := projectlib.Show(home, args[0])
				return envelope, projectlib.ShowText(envelope), nil
			})
		},
	}
}
