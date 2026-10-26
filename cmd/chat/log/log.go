// Package log wires the chat log command.
package log

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	operation "github.com/Harrison-Blair/hive/internal/chat/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the chat log command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "log", Short: "show chat history", Args: cobra.NoArgs}
	command.Flags().String("since", "", "include messages on or after this UTC date")
	command.Flags().Int("limit", 0, "show the most recent N messages")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		since, _ := cmd.Flags().GetString("since")
		limit, _ := cmd.Flags().GetInt("limit")
		return common.Run(cmd, "chat log", func(home string) (output.Envelope, string, error) {
			envelope := chatlib.Log(home, chatlib.LogOptions{Since: since, Limit: limit})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
