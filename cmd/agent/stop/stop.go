// Package stop wires the agent stop command.
package stop

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentstop "github.com/Harrison-Blair/hive/internal/agent/stop"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
	"log/slog"
)

// New creates the agent stop command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "stop NAME",
		Short: "stop an agent that has a hive record",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().Bool("remove-worktree", false, "also remove the clean worktree checkout; the branch stays")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "agent stop", args[0], func(home string, logger *slog.Logger) (output.Envelope, string, error) {
			remove, _ := cmd.Flags().GetBool("remove-worktree")
			envelope := agentstop.Run(cmd.Context(), home, agentstop.Options{Name: args[0], RemoveWorktree: remove, Logger: logger})
			return envelope, agentstop.Text(envelope), nil
		})
	}
	return command
}
