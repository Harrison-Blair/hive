// Package write wires the memory write command.
package write

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	scopelib "github.com/Harrison-Blair/hive/internal/lib/scope"
	memorylib "github.com/Harrison-Blair/hive/internal/memory"
	"github.com/spf13/cobra"
)

// New creates the memory write command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "write NAME",
		Short: "write one memory",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().String("scope", "", "memory scope")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "memory write", func(home string) (output.Envelope, string, error) {
			scopeValue, _ := cmd.Flags().GetString("scope")
			logger, _ := hivelog.Open(home, "memory")
			envelope := memorylib.Write(home, memorylib.WriteOptions{
				Scope:    scopeValue,
				Name:     args[0],
				Input:    cmd.InOrStdin(),
				Deadline: scopelib.DefaultDeadline,
				Logger:   logger,
			})
			return envelope, memorylib.WriteText(envelope), nil
		})
	}
	return command
}
