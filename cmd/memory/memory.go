// Package memory defines the memory command and registers its operations.
package memory

import (
	"github.com/Harrison-Blair/hive/cmd/memory/write"
	"github.com/spf13/cobra"
)

// New creates the memory command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "memory",
		Short: "manage agent memory",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(write.New())
	return command
}
