// Package spawn wires the agent spawn command.
package spawn

import (
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	agentspawn "github.com/Harrison-Blair/hive/internal/agent/spawn"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent spawn command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "spawn",
		Short: "start an agent in a new herdr pane",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("name", "", "agent name (required)")
	flags.String("harness", "", "harness: claude, codex, or pi (required)")
	flags.String("model", "", "model name (required)")
	flags.String("prompt", "", "first prompt text")
	flags.String("prompt-file", "", "file with the first prompt text")
	flags.String("project", "", "project ID: start in the checkout workspace of the project")
	flags.String("worktree", "", "worktree and branch name: start in a new worktree of --project")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		value := func(name string) string { text, _ := cmd.Flags().GetString(name); return text }
		return common.Run(cmd, "agent spawn", value("name"), func(home string, logger *slog.Logger) (output.Envelope, string, error) {
			cwd, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			envelope := agentspawn.Run(cmd.Context(), home, agentspawn.Options{
				Name:       value("name"),
				Harness:    value("harness"),
				Model:      value("model"),
				Prompt:     value("prompt"),
				PromptFile: value("prompt-file"),
				Project:    value("project"),
				Worktree:   value("worktree"),
				CallerPane: os.Getenv("HERDR_PANE_ID"),
				Cwd:        cwd,
				Guide:      common.Guide(cmd.Root()),
				Logger:     logger,
			})
			return envelope, agentspawn.Text(envelope), nil
		})
	}
	return command
}
