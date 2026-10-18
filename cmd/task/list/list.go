// Package list wires the task list command.
package list

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task list command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "list task records",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("project", "", "project ID")
	flags.String("status", "", "task status")
	flags.Bool("ready", false, "keep only created tasks with verified dependencies")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "task list", func(home string) (output.Envelope, string, error) {
			project, _ := cmd.Flags().GetString("project")
			status, _ := cmd.Flags().GetString("status")
			ready, _ := cmd.Flags().GetBool("ready")
			envelope := tasklib.List(home, tasklib.ListOptions{Project: project, Status: status, Ready: ready})
			return envelope, tasklib.ListText(envelope), nil
		})
	}
	return command
}
