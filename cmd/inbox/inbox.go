package inbox

import (
	"io"

	"github.com/Harrison-Blair/hive/cmd/common"
	inboxlogic "github.com/Harrison-Blair/hive/internal/inbox"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the inbox command with its send child.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "inbox",
		Short: "read and send inbox records",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("role", "", "inbox key")
	command.Flags().Bool("all", false, "print each record and set nothing")
	command.AddCommand(newSend())
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		role, _ := cmd.Flags().GetString("role")
		all, _ := cmd.Flags().GetBool("all")
		var envelope output.Envelope
		original := cmd.OutOrStdout()
		tracker := &trackingWriter{writer: original}
		cmd.SetOut(tracker)
		err := common.Run(cmd, "inbox", func(home string) (output.Envelope, string, error) {
			envelope = inboxlogic.Read(home, role, all)
			return envelope, inboxlogic.ReadText(envelope), nil
		})
		cmd.SetOut(original)
		if tracker.err != nil {
			ids := inboxlogic.Consumed(envelope)
			if len(ids) > 0 {
				_ = inboxlogic.WriteReport(cmd.ErrOrStderr(), role, ids)
			}
		}
		return err
	}
	return command
}

func newSend() *cobra.Command {
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
		return common.Run(cmd, "inbox send", func(home string) (output.Envelope, string, error) {
			envelope := inboxlogic.Send(home, inboxlib.AppendOptions{Role: to, Kind: "message", Agent: agent, Text: args[0]})
			return envelope, inboxlogic.SendText(envelope), nil
		})
	}
	return command
}

type trackingWriter struct {
	writer io.Writer
	err    error
}

func (w *trackingWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err != nil && w.err == nil {
		w.err = err
	}
	return n, err
}
