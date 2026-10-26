// Package stop wires the agent stop command.
package stop

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/agent/stop"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent stop command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "stop",
		Short: "stop an agent that hive started",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("agent", "", "agent name")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("agent")
		if name == "" {
			return common.Usage(cmd, "agent stop requires --agent")
		}
		return common.Run(cmd, "agent stop", func(home string) (output.Envelope, string, error) {
			_, envelope := operation.Run(cmd.Context(), home, name)
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
