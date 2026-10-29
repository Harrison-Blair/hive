// Package logs wires the logs command.
package logs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	loglib "github.com/Harrison-Blair/hive/internal/logs"
	"github.com/spf13/cobra"
)

// New creates the logs command.
func New() *cobra.Command {
	opts := &loglib.Options{}
	command := &cobra.Command{
		Use:   "logs",
		Short: "read the daily log files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts)
		},
	}
	flags := command.Flags()
	flags.StringVar(&opts.Since, "since", "", "duration or RFC 3339 cutoff")
	flags.StringVar(&opts.Level, "level", "", "minimum level")
	flags.StringVar(&opts.Component, "component", "", "component")
	flags.StringVar(&opts.Agent, "agent", "", "agent name: matches agent_id or target")
	flags.StringVar(&opts.Request, "request", "", "request ID")
	flags.IntVar(&opts.Limit, "limit", 200, "maximum matches to print")
	flags.BoolVar(&opts.Follow, "follow", false, "print new events as they arrive")
	return command
}

func run(cmd *cobra.Command, opts *loglib.Options) error {
	jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
	if err != nil {
		return err
	}
	opts.JSON = jsonOutput
	resolve := func() (string, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	}
	fail := func(_ string, err error) error {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %v\n", err)
		return &common.StatusError{Status: 1}
	}
	return common.Stream(cmd, "logs", resolve, fail, func(resolvedHome string, _ *slog.Logger) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
		defer stop()
		err := loglib.Run(ctx, resolvedHome, *opts, cmd.OutOrStdout(), cmd.ErrOrStderr())
		var option *loglib.OptionError
		if errors.As(err, &option) {
			return &common.UsageError{Message: option.Message}
		}
		return err
	})
}
