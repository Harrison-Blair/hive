// Package list wires the agent list command.
package list

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	operation "github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent list command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "list agent records",
		Args:  cobra.NoArgs,
	}
	command.Flags().String("project", "", "project ID")
	command.Flags().String("profile", "", "agent profile")
	command.Flags().Bool("all", false, "include stale, stopped, and exited records")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		profile, _ := cmd.Flags().GetString("profile")
		all, _ := cmd.Flags().GetBool("all")
		return common.Run(cmd, "agent list", func(home string) (output.Envelope, string, error) {
			envelope := agentlib.List(home, agentlib.ListOptions{Project: project, Profile: profile, All: all})
			return envelope, operation.Text(envelope), nil
		})
	}
	return command
}
