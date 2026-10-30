// Package guide wires the guide command.
package guide

import (
	"log/slog"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the guide command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "print the hive guide that each agent gets in its system prompt",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return common.Run(cmd, "guide", "", func(string, *slog.Logger) (output.Envelope, string, error) {
				text := common.Guide(cmd.Root())
				envelope := output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Result: guide.Result{Text: text}}
				return envelope, text, nil
			})
		},
	}
}
