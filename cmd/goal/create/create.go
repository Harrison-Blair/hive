// Package create wires the goal create command.
package create

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	operation "github.com/Harrison-Blair/hive/internal/goal/create"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the goal create command.
func New() *cobra.Command {
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
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			envelope := operation.Run(home, operation.Options{
				Project: project, Title: title, Slug: slug, Grant: grant, Agent: agent, PaneID: pane,
			})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
