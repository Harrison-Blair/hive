// Package homeinit wires the `hive init` command.
package homeinit

import (
	"fmt"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/cmd/common"
	homeinitlib "github.com/Harrison-Blair/hive/internal/homeinit"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
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
	cwd, err := os.Getwd()
	if err != nil {
		return fail(cmd, jsonOutput, "home", err, 1)
	}
	markedHome, ok := home.FindMarkerAbove(cwd)
	if !ok {
		return fail(cmd, jsonOutput, "home", fmt.Errorf("not a hive home: no marker above %s", cwd), 1)
	}
	requestID, err := hivelog.NewRequestID()
	if err != nil {
		return fail(cmd, jsonOutput, "log", err, 1)
	}
	restoreRequestID := hivelog.SetRequestID(requestID)
	defer restoreRequestID()
	logger, err := common.Logger(markedHome, "")
	if err != nil {
		return fail(cmd, jsonOutput, "log", err, 1)
	}
	started := time.Now()
	logger.Debug("command start", "command", "init")
	result, runErr := homeinitlib.Initialize(cwd)
	if runErr != nil {
		logger.Debug("command end", "command", "init", "duration_ms", time.Since(started).Milliseconds(), "outcome", "failed", "error", runErr.Error())
		return fail(cmd, jsonOutput, "command", runErr, 1)
	}
	effects := []output.Effect{}
	text := fmt.Sprintf("Hive home already initialized at %s\n", result.Home)
	if result.ConfigSaved {
		effects = append(effects, output.Effect{Resource: "config", Path: result.ConfigPath, Action: "created"})
		text = fmt.Sprintf("Initialized hive home at %s\n", result.Home)
	}
	logger.Debug("command end", "command", "init", "duration_ms", time.Since(started).Milliseconds(), "outcome", "ok")
	envelope := output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: effects,
		Error:   nil,
		Result:  map[string]string{"home": result.Home},
	}
	if err := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, text); err != nil {
		return &common.StatusError{Status: 1}
	}
	return &common.StatusError{Status: 0}
}

func fail(cmd *cobra.Command, jsonOutput bool, phase string, err error, status int) error {
	envelope := output.Envelope{
		Outcome: "failed",
		Phase:   phase,
		Effects: []output.Effect{},
		Error:   err.Error(),
		Result:  nil,
	}
	if renderErr := output.Render(cmd.OutOrStdout(), cmd.ErrOrStderr(), jsonOutput, envelope, ""); renderErr != nil {
		return &common.StatusError{Status: status}
	}
	return &common.StatusError{Status: status}
}
