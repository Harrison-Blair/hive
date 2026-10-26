// Package report wires the agent report command.
package report

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/agent/report"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent report command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "report",
		Short: "send the result of a general agent to its parent",
		Long:  "Read the result text from standard input, 4096 bytes or less, or name a file with --file. Only a general agent can report.",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("file", "", "result file")
	command.Flags().String("agent", "", "reporting agent")
	command.Flags().String("pane", "", "herdr pane ID of the caller")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		file, _ := cmd.Flags().GetString("file")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "agent report", func(home string) (output.Envelope, string, error) {
			envelope := operation.Run(home, operation.Options{
				Agent:     agent,
				Pane:      pane,
				HerdrPane: os.Getenv("HERDR_PANE_ID"),
				Input:     cmd.InOrStdin(),
				File:      file,
				FileSet:   cmd.Flags().Lookup("file").Changed,
			})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
