// Package cmd holds the hive root command and registers its subcommands.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/cmd/agent"
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/cmd/guide"
	"github.com/Harrison-Blair/hive/cmd/homeinit"
	"github.com/Harrison-Blair/hive/cmd/logs"
	"github.com/Harrison-Blair/hive/cmd/project"
	"github.com/Harrison-Blair/hive/cmd/version"
	"github.com/Harrison-Blair/hive/internal/coordinator"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// NewRootCmd builds a fresh hive command tree. Bare hive runs the coordinator
// picker in the home.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hive",
		Short:         "pick a harness and a model, then run the coordinator in this pane",
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
			return coordinator.Main(resolved, common.Guide(cmd.Root()), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.PersistentFlags().Bool("json", false, "write JSON output")
	root.AddCommand(version.New(), project.New(), logs.New(), homeinit.New(), agent.New(), guide.New())
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
	if found, err := root.ExecuteC(); err != nil {
		var statusErr *common.StatusError
		if errors.As(err, &statusErr) {
			return statusErr.Status
		}
		// A UsageError comes from a command that ran and logged it. Any other
		// error means that Cobra rejected the arguments or flags before the
		// command ran. Bare hive and hive version keep their own logging.
		var usage *common.UsageError
		if !errors.As(err, &usage) && found != nil && found != root && found.Name() != "version" {
			common.LogRejected(strings.TrimPrefix(found.CommandPath(), root.Name()+" "), err)
		}
		_ = output.Render(out, errOut, jsonRequested(args), output.Failed("usage", err.Error()), "")
		return 2
	}
	return 0
}

// jsonRequested reports whether args set --json before a "--". Cobra can
// reject the arguments before it parses the flags, so the flag value is not
// available then.
func jsonRequested(args []string) bool {
	requested := false
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--json" {
			requested = true
		} else if value, ok := strings.CutPrefix(arg, "--json="); ok {
			requested, _ = strconv.ParseBool(value)
		}
	}
	return requested
}
