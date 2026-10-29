// Package send wires the agent send command.
package send

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	agentsend "github.com/Harrison-Blair/hive/internal/agent/send"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent send command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "send NAME TEXT",
		Short: "send a prompt to an agent",
		Args:  cobra.ExactArgs(2),
	}
	command.Flags().Bool("wait", false, "wait until the agent is idle or done again")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "agent send", func(home string) (output.Envelope, string, error) {
			logger, err := hivelog.Open(home, "cli")
			if err != nil {
				return output.Envelope{}, "", err
			}
			wait, _ := cmd.Flags().GetBool("wait")
			envelope := agentsend.Run(cmd.Context(), home, agentsend.Options{
				Name:       args[0],
				Text:       args[1],
				Wait:       wait,
				CallerPane: os.Getenv("HERDR_PANE_ID"),
				Logger:     logger,
			})
			return envelope, agentsend.Text(envelope), nil
		})
	}
	return command
}
