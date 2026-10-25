// Package land wires the land command.
package land

import (
	"os"

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
	flags.String("agent", "", "caller agent")
	flags.String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "land", func(home string) (output.Envelope, string, error) {
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			workingDir, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			task, _ := cmd.Flags().GetString("task")
			mode, _ := cmd.Flags().GetString("mode")
			grant, _ := cmd.Flags().GetString("grant")
			evidence, _ := cmd.Flags().GetString("evidence")
			draft, _ := cmd.Flags().GetBool("draft")
			envelope := landlib.Land(cmd.Context(), home, landlib.Options{
				Task:       task,
				Mode:       mode,
				Grant:      grant,
				Evidence:   evidence,
				Draft:      draft,
				Agent:      agent,
				Pane:       pane,
				WorkingDir: workingDir,
			})
			return envelope, landlib.LandText(envelope), nil
		})
	}
	return command
}
