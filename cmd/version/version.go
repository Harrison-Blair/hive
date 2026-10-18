// Package version wires the `hive version` command.
package version

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/version"
	"github.com/spf13/cobra"
)

// New returns the version command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hive version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := version.Info()
			modified := ""
			if info.Modified {
				modified = ", modified"
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "hive %s (commit %s%s, %s)\n", info.Version, info.Commit, modified, info.Go)
			return err
		},
	}
}
