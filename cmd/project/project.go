// Package project defines the project command and registers its operations.
package project

import (
	"github.com/Harrison-Blair/hive/cmd/project/list"
	"github.com/Harrison-Blair/hive/cmd/project/register"
	"github.com/Harrison-Blair/hive/cmd/project/show"
	"github.com/spf13/cobra"
)

// New creates the project command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "project",
		Short: "manage registered projects",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(register.New(), list.New(), show.New())
	return command
}
