// Package create wires the worktree create command.
package create

import (
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	worktreecreate "github.com/Harrison-Blair/hive/internal/worktree/create"
	"github.com/spf13/cobra"
)

// New creates the worktree create command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "create BRANCH",
		Short: "print the checkout of BRANCH; make one if none",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().String("base", "", "start commit of a new branch (default: HEAD)")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "worktree create", "", func(home string, logger *slog.Logger) (output.Envelope, string, error) {
			if args[0] == "" {
				return output.Envelope{}, "", &common.UsageError{Message: "BRANCH requires a value"}
			}
			cwd, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			base, _ := cmd.Flags().GetString("base")
			envelope := worktreecreate.Run(cmd.Context(), home, worktreecreate.Options{Dir: cwd, Branch: args[0], Base: base, Logger: logger})
			return envelope, worktreecreate.Text(envelope), nil
		})
	}
	return command
}
