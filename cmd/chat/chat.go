// Package chat wires the chat commands.
package chat

import (
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the chat command tree.
func New() *cobra.Command {
	command := &cobra.Command{Use: "chat", Short: "send, reply to, and read chat messages", Args: cobra.NoArgs}
	command.PersistentFlags().String("agent", "", "calling agent name")
	command.PersistentFlags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.AddCommand(newSend(), newReply(), newLog())
	return command
}

func newSend() *cobra.Command {
	command := &cobra.Command{Use: "send TEXT", Short: "send a message to the coordinator", Args: cobra.ExactArgs(1)}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "chat send", func(home string) (output.Envelope, string, error) {
			workingDir, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			text := args[0]
			if text == "-" {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return output.Envelope{}, "", err
				}
				text = string(data)
			}
			envelope := chatlib.Send(cmd.Context(), home, chatlib.SendOptions{Agent: agent, PaneID: pane, WorkingDir: workingDir, Text: text})
			return envelope, chatlib.SendText(envelope), nil
		})
	}
	return command
}

func newReply() *cobra.Command {
	command := &cobra.Command{Use: "reply", Short: "save a coordinator reply", Args: cobra.NoArgs}
	command.Flags().String("to", "", "user message ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		to, _ := cmd.Flags().GetString("to")
		return common.Run(cmd, "chat reply", func(home string) (output.Envelope, string, error) {
			text, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return output.Envelope{}, "", err
			}
			envelope := chatlib.Reply(home, chatlib.ReplyOptions{Agent: agent, PaneID: pane, To: to, Text: string(text)})
			return envelope, chatlib.ReplyText(envelope), nil
		})
	}
	return command
}

func newLog() *cobra.Command {
	command := &cobra.Command{Use: "log", Short: "show chat history", Args: cobra.NoArgs}
	command.Flags().String("since", "", "include messages on or after this UTC date")
	command.Flags().Int("limit", 0, "show the most recent N messages")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		since, _ := cmd.Flags().GetString("since")
		limit, _ := cmd.Flags().GetInt("limit")
		return common.Run(cmd, "chat log", func(home string) (output.Envelope, string, error) {
			envelope := chatlib.Log(home, chatlib.LogOptions{Since: since, Limit: limit})
			return envelope, chatlib.LogText(envelope), nil
		})
	}
	return command
}
