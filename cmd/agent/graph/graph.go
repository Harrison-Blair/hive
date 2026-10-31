// Package graph wires the agent graph command.
package graph

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	agentgraph "github.com/Harrison-Blair/hive/internal/agent/graph"
	agentlist "github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the agent graph command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "graph",
		Short: "show a live graph of the agents; --json prints one list snapshot",
		Args:  cobra.NoArgs,
	}
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
		if err != nil {
			return err
		}
		if jsonOutput {
			return common.Run(cmd, "agent graph", "", func(home string, logger *slog.Logger) (output.Envelope, string, error) {
				envelope := agentlist.Run(cmd.Context(), home, agentlist.Options{Logger: logger})
				return envelope, agentlist.Text(envelope), nil
			})
		}
		resolve := func() (string, error) {
			cwd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			return home.Resolve(os.Getenv("HIVE_HOME"), cwd)
		}
		fail := func(phase string, err error) error {
			return common.RenderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), false, phase, err.Error(), 1)
		}
		return common.Stream(cmd, "agent graph", resolve, fail, func(home string, logger *slog.Logger) error {
			if err := agentgraph.Run(home, logger, cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %v\n", err)
				return err
			}
			return nil
		})
	}
	return command
}
