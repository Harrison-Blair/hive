// Package show wires the task show command.
package show

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task show command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "show ID",
		Short: "show one task record",
		Args:  cobra.ExactArgs(1),
	}
	flags := command.Flags()
	flags.Bool("brief", false, "print the brief text only")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "task show", func(home string) (output.Envelope, string, error) {
			brief, _ := cmd.Flags().GetBool("brief")
			envelope := tasklib.Show(home, args[0], tasklib.ShowOptions{Brief: brief})
			text := tasklib.ShowText(envelope)
			if brief {
				text = tasklib.ShowBriefText(envelope)
			}
			return envelope, text, nil
		})
	}
	return command
}
