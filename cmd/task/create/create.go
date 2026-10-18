// Package create wires the task create command.
package create

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/spf13/cobra"
)

// New creates the task create command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "create a task record and its brief",
		Args:  cobra.NoArgs,
	}
	flags := command.Flags()
	flags.String("project", "", "project ID")
	flags.String("title", "", "task title")
	flags.String("slug", "", "task slug")
	flags.String("depends-on", "", "comma-separated dependency IDs")
	flags.String("brief", "", "brief file, or - for standard input")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return common.Run(cmd, "task create", func(home string) (output.Envelope, string, error) {
			project, _ := cmd.Flags().GetString("project")
			title, _ := cmd.Flags().GetString("title")
			slug, _ := cmd.Flags().GetString("slug")
			dependsOn, _ := cmd.Flags().GetString("depends-on")
			briefPath, _ := cmd.Flags().GetString("brief")
			envelope := tasklib.Create(home, tasklib.CreateOptions{
				Project:    project,
				Title:      title,
				Slug:       slug,
				SlugSet:    cmd.Flags().Changed("slug"),
				DependsOn:  dependsOn,
				DependsSet: cmd.Flags().Changed("depends-on"),
				BriefPath:  briefPath,
				BriefSet:   cmd.Flags().Changed("brief"),
				Input:      cmd.InOrStdin(),
			})
			return envelope, tasklib.CreateText(envelope), nil
		})
	}
	return command
}
