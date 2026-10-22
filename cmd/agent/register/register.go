// Package register wires the agent register command.
package register

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent register command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "register",
		Short: "register the current agent session",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("role", "", "agent role")
	command.Flags().String("project", "", "project ID")
	command.Flags().String("task", "", "task ID")
	command.Flags().String("harness", "", "harness kind")
	command.Flags().String("model", "", "model name")
	command.Flags().String("name", "", "agent name")
	command.Flags().String("pane", "", "herdr pane ID")
	command.Flags().Int("quiet-after", 10, "quiet minutes")
	command.Flags().Int64("pid", 0, "holder process ID")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		role, _ := cmd.Flags().GetString("role")
		project, _ := cmd.Flags().GetString("project")
		task, _ := cmd.Flags().GetString("task")
		harness, _ := cmd.Flags().GetString("harness")
		model, _ := cmd.Flags().GetString("model")
		name, _ := cmd.Flags().GetString("name")
		pane, _ := cmd.Flags().GetString("pane")
		quiet, _ := cmd.Flags().GetInt("quiet-after")
		pid, _ := cmd.Flags().GetInt64("pid")
		workspace, tab := "", ""
		if !cmd.Flags().Lookup("pane").Changed {
			pane, workspace, tab = paneFromEnv()
		}
		return common.Run(cmd, "agent register", func(home string) (output.Envelope, string, error) {
			envelope := agentlib.RegisterAndRename(home, agentlib.RegisterOptions{
				Role:          role,
				Project:       project,
				Task:          task,
				Harness:       harness,
				Model:         model,
				Name:          name,
				NameSet:       cmd.Flags().Lookup("name").Changed,
				PaneID:        pane,
				WorkspaceID:   workspace,
				TabID:         tab,
				QuietAfterMin: quiet,
				QuietAfterSet: cmd.Flags().Lookup("quiet-after").Changed,
				PID:           pid,
			})
			return envelope, agentlib.RegisterText(envelope), nil
		})
	}
	return command
}

func paneFromEnv() (pane, workspace, tab string) {
	return os.Getenv("HERDR_PANE_ID"), os.Getenv("HERDR_WORKSPACE_ID"), os.Getenv("HERDR_TAB_ID")
}
