// Package land wires the land command.
package land

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	landlib "github.com/Harrison-Blair/hive/internal/land"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the land command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "land",
		Short: "deliver a verified task branch",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.String("mode", "", "merge or pr")
	flags.String("grant", "", "grant citation")
	flags.String("evidence", "", "evidence URL or path")
	flags.Bool("draft", false, "open a draft pull request")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "land", func(home string) (output.Envelope, string, error) {
			task, _ := cmd.Flags().GetString("task")
			mode, _ := cmd.Flags().GetString("mode")
			grant, _ := cmd.Flags().GetString("grant")
			evidence, _ := cmd.Flags().GetString("evidence")
			draft, _ := cmd.Flags().GetBool("draft")
			envelope := landlib.Land(cmd.Context(), home, landlib.Options{
				Task:     task,
				Mode:     mode,
				Grant:    grant,
				Evidence: evidence,
				Draft:    draft,
			})
			return envelope, landlib.LandText(envelope), nil
		})
	}
	return command
}
