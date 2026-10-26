// Package send wires the hidden chat send command.
package send

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/chat/send"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the hidden chat send command. It refuses each call.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:    "send TEXT",
		Short:  "send a message to the coordinator",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
	}
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "chat send", func(home string) (output.Envelope, string, error) {
			workingDir, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			return operation.Refuse(home, pane, workingDir), "", nil
		})
	}
	return command
}
