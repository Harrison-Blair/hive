// Package remove wires the worktree remove command.
package remove

import (
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	worktreeremove "github.com/Harrison-Blair/hive/internal/worktree/remove"
	"github.com/spf13/cobra"
)

// New creates the worktree remove command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "remove BRANCH",
		Short: "remove a clean, idle checkout that hive made; the branch stays",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.Run(cmd, "worktree remove", "", func(home string, logger *slog.Logger) (output.Envelope, string, error) {
				if args[0] == "" {
					return output.Envelope{}, "", &common.UsageError{Message: "BRANCH requires a value"}
				}
				cwd, err := os.Getwd()
				if err != nil {
					return output.Envelope{}, "", err
				}
				envelope := worktreeremove.Run(cmd.Context(), home, worktreeremove.Options{Dir: cwd, Branch: args[0], CallerPane: os.Getenv("HERDR_PANE_ID"), Logger: logger})
				return envelope, worktreeremove.Text(envelope), nil
			})
		},
	}
}
