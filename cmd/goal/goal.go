// Package goal wires the goal command tree.
package goal

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	goalservice "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the goal command tree.
func New() *cobra.Command {
	command := &cobra.Command{Use: "goal", Short: "manage project goals", Args: cobra.NoArgs}
	command.PersistentFlags().String("agent", "", "acting agent name")
	command.PersistentFlags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.AddCommand(createCommand(), editCommand(), showCommand(), listCommand())
	return command
}

func createCommand() *cobra.Command {
	command := &cobra.Command{Use: "create", Short: "create a goal", Args: cobra.NoArgs}
	flags := command.Flags()
	flags.String("project", "", "project ID")
	flags.String("title", "", "goal title")
	flags.String("slug", "", "goal slug")
	flags.String("grant", "", "grant line text")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "goal create", func(home string) (output.Envelope, string, error) {
			project, _ := cmd.Flags().GetString("project")
			title, _ := cmd.Flags().GetString("title")
			slug, _ := cmd.Flags().GetString("slug")
			grant, _ := cmd.Flags().GetString("grant")
			agent, _ := cmd.Flags().GetString("agent")
			pane, _ := cmd.Flags().GetString("pane")
			envelope := goalservice.Create(home, goalservice.CreateOptions{
				Project: project, Title: title, Slug: slug, Grant: grant, Agent: agent, PaneID: pane,
			})
			return envelope, goalservice.CreateText(envelope), nil
		})
	}
	return command
}

func editCommand() *cobra.Command {
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
			envelope := goalservice.Edit(home, goalservice.EditOptions{
				Goal: goalID, Agent: agent, PaneID: pane,
				AddWave: addWave, AddWaveSet: cmd.Flags().Changed("add-wave"), DoneWhen: doneWhen,
				Wave: wave, AddTask: addTask, AddTaskSet: cmd.Flags().Changed("add-task"),
				NeedsUser: needsUser, NeedsUserSet: cmd.Flags().Changed("needs-user"),
				ClearNeedsUser: clearNeedsUser, WaveStatus: status,
				WaveStatusSet: cmd.Flags().Changed("status") && wave != "", Status: status,
				StatusSet: cmd.Flags().Changed("status") && wave == "",
			})
			return envelope, goalservice.EditText(envelope), nil
		})
	}
	return command
}

func showCommand() *cobra.Command {
	command := &cobra.Command{Use: "show ID", Short: "show a goal", Args: cobra.ExactArgs(1)}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "goal show", func(home string) (output.Envelope, string, error) {
			envelope := goalservice.Show(home, args[0])
			return envelope, goalservice.ShowText(envelope), nil
		})
	}
	return command
}

func listCommand() *cobra.Command {
	command := &cobra.Command{Use: "list", Short: "list goals", Args: cobra.NoArgs}
	command.Flags().String("project", "", "project ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "goal list", func(home string) (output.Envelope, string, error) {
			project, _ := cmd.Flags().GetString("project")
			envelope := goalservice.List(home, project)
			return envelope, goalservice.ListText(envelope), nil
		})
	}
	return command
}
