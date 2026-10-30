// Package guide wires the guide command.
package guide

import (
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the guide command. It logs the command only when the home
// directory exists, so it never creates a home just to print the guide.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "print the hive guide that each agent gets in its system prompt",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			text := common.Guide(cmd.Root())
			envelope := output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Result: guide.Result{Text: text}}
			cwd, err := os.Getwd()
			resolved := ""
			if err == nil {
				resolved, err = home.Resolve(os.Getenv("HIVE_HOME"), cwd)
			}
			if err == nil {
				_, err = os.Stat(resolved)
			}
			if err != nil {
				jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
				if err := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, text); err != nil {
					return &common.StatusError{Status: 1}
				}
				return nil
			}
			return common.Run(cmd, "guide", "", func(string, *slog.Logger) (output.Envelope, string, error) {
				return envelope, text, nil
			})
		},
	}
}
