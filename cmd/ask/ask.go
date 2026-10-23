// Package ask wires the ask command.
package ask

import (
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	asklib "github.com/Harrison-Blair/hive/internal/ask"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the ask command.
func New() *cobra.Command {
	command := &cobra.Command{Use: "ask", Short: "ask the user questions in the TUI", Args: cobra.NoArgs}
	command.Flags().String("agent", "", "calling agent name")
	command.Flags().String("pane", os.Getenv("HERDR_PANE_ID"), "herdr pane ID")
	command.Flags().String("file", "", "ask JSON file, or - for stdin")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		pane, _ := cmd.Flags().GetString("pane")
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			return common.Usage(cmd, "give the ask JSON with --file FILE or --file -")
		}
		return common.Run(cmd, "ask", func(home string) (output.Envelope, string, error) {
			if refused := common.CheckCaller(home, agent); refused != nil {
				return *refused, "", nil
			}
			var data []byte
			var err error
			if file == "-" {
				data, err = io.ReadAll(cmd.InOrStdin())
			} else {
				data, err = os.ReadFile(file)
			}
			if err != nil {
				return output.Envelope{}, "", err
			}
			envelope := asklib.Create(cmd.Context(), home, asklib.CreateOptions{Agent: agent, PaneID: pane, Data: data})
			return envelope, asklib.Text(envelope), nil
		})
	}
	return command
}
