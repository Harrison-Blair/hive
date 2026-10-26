// Package send wires the inbox send command.
package send

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/inbox/send"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the inbox send command. It reads the inherited --pane flag.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "send TEXT",
		Short: "append a message record",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().String("to", "", "inbox key")
	command.Flags().String("agent", "", "sending agent")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		to, _ := cmd.Flags().GetString("to")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "inbox send", func(home string) (output.Envelope, string, error) {
			envelope := operation.Run(home, inboxlib.AppendOptions{Inbox: to, Kind: "message", Agent: agent, Text: args[0]}, pane, os.Getenv("HERDR_PANE_ID"))
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
