// Package version wires the `hive version` command.
package version

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/version"
	"github.com/spf13/cobra"
)

// New returns the version command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hive version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Version)
			return err
		},
	}
}
