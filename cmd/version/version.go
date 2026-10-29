// Package version wires the `hive version` command.
package version

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/version"
	"github.com/spf13/cobra"
)

type versionResult struct {
	Version  string `json:"version"`
	Commit   string `json:"commit"`
	Modified bool   `json:"modified"`
	Go       string `json:"go"`
}

// New returns the version command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hive version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := version.Info()
			jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOutput {
				return output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), true, output.Envelope{
					Outcome: output.OutcomeOK,
					Phase:   "complete",
					Effects: []output.Effect{},
					Error:   nil,
					Result: versionResult{
						Version:  info.Version,
						Commit:   info.Commit,
						Modified: info.Modified,
						Go:       info.Go,
					},
				}, "")
			}
			modified := ""
			if info.Modified {
				modified = ", modified"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "hive %s (commit %s%s, %s)\n", info.Version, info.Commit, modified, info.Go)
			return err
		},
	}
}
