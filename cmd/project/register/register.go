// Package register wires the project register command.
package register

import (
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	projectregister "github.com/Harrison-Blair/hive/internal/project/register"
	"github.com/spf13/cobra"
)

// New creates the project register command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "register PATH",
		Short: "register a Git repository",
		Args:  cobra.ExactArgs(1),
	}
	command.Flags().String("id", "", "project ID")
	command.Flags().String("base", "", "base branch")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		return common.Run(cmd, "project register", func(home string) (output.Envelope, string, error) {
			if args[0] == "" {
				return output.Envelope{}, "", &common.UsageError{Message: "PATH requires a value"}
			}
			workingDir, err := os.Getwd()
			if err != nil {
				return output.Envelope{}, "", err
			}
			id, _ := cmd.Flags().GetString("id")
			base, _ := cmd.Flags().GetString("base")
			envelope := projectregister.Run(home, projectregister.Options{
				Path:       args[0],
				ID:         id,
				IDSet:      cmd.Flags().Lookup("id").Changed,
				Base:       base,
				BaseSet:    cmd.Flags().Lookup("base").Changed,
				WorkingDir: workingDir,
			})
			return envelope, projectregister.Text(envelope), nil
		})
	}
	return command
}
