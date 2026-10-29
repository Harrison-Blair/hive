// Package homeinit wires the `hive init` command.
package homeinit

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	homeinitlib "github.com/Harrison-Blair/hive/internal/homeinit"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New creates the init command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "write the global hive home config",
		Args:  cobra.NoArgs,
		RunE:  run,
	}
}

func run(cmd *cobra.Command, _ []string) error {
	jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
	if err != nil {
		return err
	}
	resolve := func() (string, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return homeinitlib.FindHome(cwd)
	}
	fail := func(phase string, err error) error {
		return common.RenderFailure(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, phase, err.Error(), 1)
	}
	return common.Stream(cmd, "init", resolve, fail, func(markedHome string, _ *slog.Logger) error {
		envelope := homeinitlib.Initialize(markedHome)
		if err := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, homeinitlib.Text(envelope)); err != nil {
			return err
		}
		if envelope.Outcome != output.OutcomeOK {
			return errors.New(fmt.Sprint(envelope.Error))
		}
		return nil
	})
}
