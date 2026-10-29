// Package spawn wires the agent spawn command.
package spawn

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	agentspawn "github.com/Harrison-Blair/hive/internal/agent/spawn"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
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
		return common.Run(cmd, "agent spawn", func(home string) (output.Envelope, string, error) {
			logger, err := hivelog.Open(home, "cli")
			if err != nil {
				return output.Envelope{}, "", err
			}
			cwd, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			value := func(name string) string { text, _ := cmd.Flags().GetString(name); return text }
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
				Logger:     logger,
			})
			return envelope, agentspawn.Text(envelope), nil
		})
	}
	return command
}
