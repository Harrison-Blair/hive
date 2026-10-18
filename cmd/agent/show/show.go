// Package show wires the agent show command.
package show

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent show command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "show NAME",
		Short: "show one agent record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.Run(cmd, "agent show", func(home string) (output.Envelope, string, error) {
				envelope := agentlib.Show(home, args[0])
				return envelope, agentlib.ShowText(envelope), nil
			})
		},
	}
}
