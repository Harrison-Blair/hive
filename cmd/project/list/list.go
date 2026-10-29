// Package list wires the project list command.
package list

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
	projectlist "github.com/Harrison-Blair/hive/internal/project/list"
	"github.com/spf13/cobra"
	"log/slog"
)

// New creates the project list command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list registered projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return common.Run(cmd, "project list", "", func(home string, _ *slog.Logger) (output.Envelope, string, error) {
				envelope := projectlib.List(home)
				return envelope, projectlist.Text(envelope), nil
			})
		},
	}
}
