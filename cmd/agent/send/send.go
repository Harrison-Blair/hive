// Package send wires the agent send command.
package send

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent send command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "send TEXT",
		Short: "send a prompt to an agent",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().String("agent", "", "agent name")
	command.Flags().Bool("wait", false, "wait for a settled status")
	command.Flags().Int("timeout", 600, "wait timeout in seconds")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("agent")
		if name == "" {
			return common.Usage(cmd, "agent send requires --agent")
		}
		wait, _ := cmd.Flags().GetBool("wait")
		timeout, _ := cmd.Flags().GetInt("timeout")
		return common.Run(cmd, "agent send", func(home string) (output.Envelope, string, error) {
			envelope := agentlib.Send(cmd.Context(), home, agentlib.SendOptions{
				Agent:      name,
				Text:       args[0],
				Wait:       wait,
				Timeout:    timeout,
				TimeoutSet: cmd.Flags().Lookup("timeout").Changed,
			})
			return envelope, agentlib.SendText(envelope), nil
		})
	}
	return command
}
