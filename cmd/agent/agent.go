// Package agent defines the agent command and registers its operations.
package agent

import (
	"github.com/Harrison-Blair/hive/cmd/agent/list"
	"github.com/Harrison-Blair/hive/cmd/agent/register"
	"github.com/Harrison-Blair/hive/cmd/agent/send"
	"github.com/Harrison-Blair/hive/cmd/agent/show"
	"github.com/Harrison-Blair/hive/cmd/agent/spawn"
	"github.com/Harrison-Blair/hive/cmd/agent/stop"
	"github.com/spf13/cobra"
)

// New creates the agent command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "agent",
		Short: "manage agent sessions",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(register.New(), list.New(), show.New(), spawn.New(), send.New(), stop.New())
	return command
}
