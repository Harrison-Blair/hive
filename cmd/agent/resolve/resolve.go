// Package resolve wires the agent resolve command.
package resolve

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/agent/resolve"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent resolve command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "resolve",
		Short: "settle an unresolved agent record",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("agent", "", "agent name")
	command.Flags().String("as", "", "new status: exited or live")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("agent")
		as, _ := cmd.Flags().GetString("as")
		if name == "" {
			return common.Usage(cmd, "agent resolve requires --agent")
		}
		return common.Run(cmd, "agent resolve", func(home string) (output.Envelope, string, error) {
			envelope := operation.Run(cmd.Context(), home, operation.Options{Agent: name, As: as})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
