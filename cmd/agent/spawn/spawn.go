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
	command.Flags().String("role", "", "old name of --profile (wave A)")
	command.Flags().String("agent", "", "caller agent name or key")
	command.Flags().String("harness", "", "harness kind")
	command.Flags().String("pane", "", "destination herdr pane ID: hive starts the agent in this pane, and agent stop closes it; herdr then closes its tab or workspace when it holds no other pane")
	command.Flags().String("adopt", "", "adopt an existing workspace")
	command.Flags().String("model", "", "model name")
	command.Flags().String("profile", "", "profile name: the default is general")
	command.Flags().String("project", "", "project ID")
	command.Flags().String("task", "", "task ID")
	command.Flags().String("name", "", "agent name")
	command.Flags().Int("quiet-after", 10, "quiet minutes")
	command.Flags().StringArray("arg", nil, "native agent argument")
	command.Flags().String("prompt", "", "spawner text of the agent")
	command.Flags().String("prompt-file", "", "file with the spawner text of the agent")
	command.Flags().String("cwd", "", "working directory of an agent in a new tab")
	command.Flags().Bool("allow-trust-dialog", false, "start an agent that can show a trust dialog; keep its first prompt for the supervisor")
	command.Flags().Bool("owner", false, "make the agent the owner of its project")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		role, _ := cmd.Flags().GetString("role")
		profile, _ := cmd.Flags().GetString("profile")
		owner, _ := cmd.Flags().GetBool("owner")
		agent, _ := cmd.Flags().GetString("agent")
		harness, _ := cmd.Flags().GetString("harness")
		pane, _ := cmd.Flags().GetString("pane")
		adopt, _ := cmd.Flags().GetString("adopt")
		model, _ := cmd.Flags().GetString("model")
		project, _ := cmd.Flags().GetString("project")
		task, _ := cmd.Flags().GetString("task")
		name, _ := cmd.Flags().GetString("name")
		quiet, _ := cmd.Flags().GetInt("quiet-after")
		extra, _ := cmd.Flags().GetStringArray("arg")
		prompt, _ := cmd.Flags().GetString("prompt")
		promptFile, _ := cmd.Flags().GetString("prompt-file")
		cwd, _ := cmd.Flags().GetString("cwd")
		allowTrust, _ := cmd.Flags().GetBool("allow-trust-dialog")
		return common.Run(cmd, "agent spawn", func(home string) (output.Envelope, string, error) {
			envelope := agentlib.SpawnFromCLI(cmd.Context(), home, agentlib.SpawnOptions{
				Role:          role,
				Profile:       profile,
				ProfileSet:    cmd.Flags().Lookup("profile").Changed,
				Owner:         owner,
				Agent:         agent,
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

				Prompt:           prompt,
				PromptSet:        cmd.Flags().Lookup("prompt").Changed,
				PromptFile:       promptFile,
				PromptFileSet:    cmd.Flags().Lookup("prompt-file").Changed,
				Cwd:              cwd,
				CwdSet:           cmd.Flags().Lookup("cwd").Changed,
				AllowTrustDialog: allowTrust,
			})
			return envelope, agentlib.SpawnText(envelope), nil
		})
	}
	return command
}
