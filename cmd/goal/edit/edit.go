// Package edit wires the goal edit command.
package edit

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/goal/edit"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the goal edit command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "edit", Short: "edit a goal or wave", Args: cobra.NoArgs}
	flags := command.Flags()
	flags.String("goal", "", "goal ID")
	flags.String("add-wave", "", "new wave title")
	flags.String("done-when", "", "wave completion point")
	flags.String("wave", "", "wave ID")
	flags.String("add-task", "", "task ID to add")
	flags.String("needs-user", "", "wave stop text")
	flags.Bool("clear-needs-user", false, "clear wave stop text")
	flags.String("status", "", "goal status, or wave status with --wave")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "goal edit", func(home string) (output.Envelope, string, error) {
			goalID, _ := cmd.Flags().GetString("goal")
			addWave, _ := cmd.Flags().GetString("add-wave")
			doneWhen, _ := cmd.Flags().GetString("done-when")
			wave, _ := cmd.Flags().GetString("wave")
			addTask, _ := cmd.Flags().GetString("add-task")
			needsUser, _ := cmd.Flags().GetString("needs-user")
			clearNeedsUser, _ := cmd.Flags().GetBool("clear-needs-user")
			status, _ := cmd.Flags().GetString("status")
			agent, _ := cmd.Flags().GetString("agent")
			pane, _ := cmd.Flags().GetString("pane")
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			envelope := operation.Run(home, operation.Options{
				Goal: goalID, Agent: agent, PaneID: pane,
				AddWave: addWave, AddWaveSet: cmd.Flags().Changed("add-wave"), DoneWhen: doneWhen,
				Wave: wave, AddTask: addTask, AddTaskSet: cmd.Flags().Changed("add-task"),
				NeedsUser: needsUser, NeedsUserSet: cmd.Flags().Changed("needs-user"),
				ClearNeedsUser: clearNeedsUser, WaveStatus: status,
				WaveStatusSet: cmd.Flags().Changed("status") && wave != "", Status: status,
				StatusSet: cmd.Flags().Changed("status") && wave == "",
			})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
