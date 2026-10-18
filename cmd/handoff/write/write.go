// Package write wires the handoff write command.
package write

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/handoff"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	scopelib "github.com/Harrison-Blair/hive/internal/lib/scope"
	"github.com/spf13/cobra"
)

// New creates the handoff write command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "write",
		Short: "write the handoff",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("scope", "", "handoff scope")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "handoff write", func(home string) (output.Envelope, string, error) {
			scopeValue, _ := cmd.Flags().GetString("scope")
			logger, _ := hivelog.Open(home, "handoff")
			envelope := handoff.Write(home, handoff.WriteOptions{
				Scope:    scopeValue,
				Input:    cmd.InOrStdin(),
				Deadline: scopelib.DefaultDeadline,
				Logger:   logger,
			})
			return envelope, handoff.WriteText(envelope), nil
		})
	}
	return command
}
