// Package handoff defines the handoff command and registers its operations.
package handoff

import (
	"github.com/Harrison-Blair/hive/cmd/handoff/write"
	"github.com/spf13/cobra"
)

// New creates the handoff command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "handoff",
		Short: "manage a role handoff",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(write.New())
	return command
}
