// Package edit wires the task edit command.
package edit

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task edit command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "edit",
		Short: "edit a task brief",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("task", "", "task ID")
	flags.String("brief", "", "brief file, or - for standard input")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		taskID, _ := cmd.Flags().GetString("task")
		briefPath, _ := cmd.Flags().GetString("brief")
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		return common.Run(cmd, "task edit", func(home string) (output.Envelope, string, error) {
			envelope := tasklib.EditBrief(home, tasklib.EditOptions{
				Task: taskID, Agent: agent, PaneID: pane,
				BriefPath: briefPath, Input: cmd.InOrStdin(),
			})
			return envelope, tasklib.EditText(envelope), nil
		})
	}
	return command
}
