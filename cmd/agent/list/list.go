// Package list wires the agent list command.
package list

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentlist "github.com/Harrison-Blair/hive/internal/agent/list"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent list command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "list the agent records with their herdr status",
		Args:  cobra.NoArgs,
	}
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "agent list", func(home string) (output.Envelope, string, error) {
			logger, err := hivelog.Open(home, "cli")
			if err != nil {
				return output.Envelope{}, "", err
			}
			envelope := agentlist.Run(cmd.Context(), home, agentlist.Options{Logger: logger})
			return envelope, agentlist.Text(envelope), nil
		})
	}
	return command
}
