// Package spawn wires the agent spawn command.
package spawn

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent spawn command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "spawn",
		Short: "spawn an agent in a pane",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("role", "", "agent role")
	command.Flags().String("harness", "", "harness kind")
	command.Flags().String("pane", "", "herdr pane ID")
	command.Flags().String("adopt", "", "adopt an existing workspace")
	command.Flags().String("model", "", "model name")
	command.Flags().String("project", "", "project ID")
	command.Flags().String("task", "", "task ID")
	command.Flags().String("name", "", "agent name")
	command.Flags().Int("quiet-after", 10, "quiet minutes")
	command.Flags().StringArray("arg", nil, "native agent argument")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		role, _ := cmd.Flags().GetString("role")
		harness, _ := cmd.Flags().GetString("harness")
		pane, _ := cmd.Flags().GetString("pane")
		adopt, _ := cmd.Flags().GetString("adopt")
		model, _ := cmd.Flags().GetString("model")
		project, _ := cmd.Flags().GetString("project")
		task, _ := cmd.Flags().GetString("task")
		name, _ := cmd.Flags().GetString("name")
		quiet, _ := cmd.Flags().GetInt("quiet-after")
		extra, _ := cmd.Flags().GetStringArray("arg")
		return common.Run(cmd, "agent spawn", func(home string) (output.Envelope, string, error) {
			envelope := agentlib.Spawn(cmd.Context(), home, agentlib.SpawnOptions{
				Role:          role,
				Project:       project,
				Task:          task,
				Harness:       harness,
				Model:         model,
				Name:          name,
				NameSet:       cmd.Flags().Lookup("name").Changed,
				PaneID:        pane,
				PaneSet:       cmd.Flags().Lookup("pane").Changed,
				Adopt:         adopt,
				QuietAfterMin: quiet,
				QuietAfterSet: cmd.Flags().Lookup("quiet-after").Changed,
				ExtraArgs:     extra,
			})
			return envelope, agentlib.SpawnText(envelope), nil
		})
	}
	return command
}
