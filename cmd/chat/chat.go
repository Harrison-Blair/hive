// Package chat defines the chat command and registers its operations.
package chat

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/chat/log"
	"github.com/Harrison-Blair/hive/cmd/chat/reply"
	"github.com/Harrison-Blair/hive/cmd/chat/send"
	"github.com/spf13/cobra"
)

// New creates the chat command tree.
func New() *cobra.Command {
	command := &cobra.Command{Use: "chat", Short: "send, reply to, and read chat messages", Args: cobra.NoArgs}
	command.PersistentFlags().String("agent", "", "calling agent name")
	command.PersistentFlags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.AddCommand(send.New(), reply.New(), log.New())
	return command
}
