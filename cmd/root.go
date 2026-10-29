// Package cmd holds the hive root command and registers its subcommands.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/cmd/homeinit"
	"github.com/Harrison-Blair/hive/cmd/logs"
	"github.com/Harrison-Blair/hive/cmd/project"
	"github.com/Harrison-Blair/hive/cmd/version"
	"github.com/Harrison-Blair/hive/internal/coordinator"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/spf13/cobra"
)

// NewRootCmd builds a fresh hive command tree. Bare hive runs the coordinator
// picker in the home.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hive",
		Short:         "hive",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			resolved, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive: %s\n", err)
				return &common.StatusError{Status: 1}
			}
			return coordinator.Main(resolved, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.PersistentFlags().Bool("json", false, "write JSON output")
	root.AddCommand(version.New(), project.New(), logs.New(), homeinit.New())
	return root
}

// Execute runs hive with the process arguments and standard streams.
func Execute() int {
	return executeWithStreams(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}

// ExecuteWithArgs runs hive with the given arguments and streams.
func ExecuteWithArgs(args []string, in io.Reader, out io.Writer) error {
	if status := executeWithStreams(args, in, out, out); status != 0 {
		return fmt.Errorf("hive exited with status %d", status)
	}
	return nil
}

func executeWithStreams(args []string, in io.Reader, out, errOut io.Writer) int {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	if err := root.Execute(); err != nil {
		var statusErr *common.StatusError
		if errors.As(err, &statusErr) {
			return statusErr.Status
		}
		var usage *common.UsageError
		if errors.As(err, &usage) {
			_, _ = fmt.Fprintf(errOut, "hive: %s\n", usage.Error())
			return 2
		}
		_, _ = fmt.Fprintf(errOut, "hive: %s\n", err)
		return 2
	}
	return 0
}
