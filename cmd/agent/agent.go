// Package agent defines the agent command and registers its operations.
package agent

import (
	"github.com/Harrison-Blair/hive/cmd/agent/graph"
	"github.com/Harrison-Blair/hive/cmd/agent/list"
	"github.com/Harrison-Blair/hive/cmd/agent/send"
	"github.com/Harrison-Blair/hive/cmd/agent/spawn"
	"github.com/Harrison-Blair/hive/cmd/agent/stop"
	"github.com/spf13/cobra"
)

// New creates the agent command tree.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "agent",
		Short: "manage agents in herdr panes",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(spawn.New(), stop.New(), send.New(), list.New(), graph.New())
	return command
}
