// Package reply wires the chat reply command.
package reply

import (
	"io"

	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/chat/reply"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the chat reply command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "reply", Short: "save a coordinator reply", Args: cobra.NoArgs}
	command.Flags().String("to", "", "user message ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		to, _ := cmd.Flags().GetString("to")
		return common.Run(cmd, "chat reply", func(home string) (output.Envelope, string, error) {
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			text, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return output.Envelope{}, "", err
			}
			envelope := operation.Reply(home, operation.Options{Agent: agent, PaneID: pane, To: to, Text: string(text)})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
