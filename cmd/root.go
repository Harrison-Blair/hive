// Package cmd holds the hive root command and registers its subcommands.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/agent"
	"github.com/Harrison-Blair/hive/cmd/ask"
	"github.com/Harrison-Blair/hive/cmd/chat"
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/cmd/doctor"
	goalcmd "github.com/Harrison-Blair/hive/cmd/goal"
	"github.com/Harrison-Blair/hive/cmd/handoff"
	"github.com/Harrison-Blair/hive/cmd/homeinit"
	"github.com/Harrison-Blair/hive/cmd/inbox"
	"github.com/Harrison-Blair/hive/cmd/land"
	"github.com/Harrison-Blair/hive/cmd/logs"
	"github.com/Harrison-Blair/hive/cmd/memory"
	"github.com/Harrison-Blair/hive/cmd/project"
	"github.com/Harrison-Blair/hive/cmd/supervise"
	"github.com/Harrison-Blair/hive/cmd/task"
	"github.com/Harrison-Blair/hive/cmd/version"
	"github.com/Harrison-Blair/hive/cmd/worktree"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCmd builds a fresh hive command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hive",
		Short:         "hive",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
			if err != nil {
				return err
			}
			noStartup, err := cmd.Flags().GetBool("no-startup")
			if err != nil {
				return err
			}
			demo, err := cmd.Flags().GetString("demo")
			if err != nil {
				return err
			}
			size, err := cmd.Flags().GetString("size")
			if err != nil {
				return err
			}
			if size != "" && demo == "" {
				return &common.UsageError{Message: "--size needs --demo"}
			}
			if demo != "" {
				return runDemo(cmd, jsonOutput, demo, size)
			}
			if noStartup {
				// The TUI startup reads the variable, so the flag sets it.
				if err := os.Setenv("HIVE_NO_STARTUP", "1"); err != nil {
					return err
				}
			}
			return common.Run(cmd, "tui", func(resolvedHome string) (output.Envelope, string, error) {
				if jsonOutput {
					return output.Envelope{Outcome: "failed", Phase: "validate", Effects: []output.Effect{}, Error: "TUI requires text output", Result: nil}, "", nil
				}
				if err := tui.RunHome(resolvedHome, cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
					return output.Envelope{}, "", err
				}
				return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: nil}, "", nil
			})
		},
	}
	root.PersistentFlags().Bool("json", false, "write JSON output")
	root.Flags().Bool("no-startup", false, "open the TUI and start no coordinator, no supervisor, and no tab")
	root.Flags().String("demo", "", "open the TUI in demo mode from a fixture file")
	root.Flags().String("size", "", "draw the demo TUI at WxH cells")
	_ = root.Flags().MarkHidden("demo")
	_ = root.Flags().MarkHidden("size")
	root.AddCommand(version.New(), project.New(), logs.New(), doctor.New(), task.New(), agent.New(), inbox.New(), homeinit.New(), memory.New(), handoff.New(), worktree.New(), land.New(), supervise.New(), goalcmd.New(), chat.New(), ask.New())
	return root
}

// runDemo opens the TUI in demo mode. Demo mode implies --no-startup. It
// resolves the home without the shared command wrapper, so that each log
// event of the run carries demo=true.
func runDemo(cmd *cobra.Command, jsonOutput bool, fixture, size string) error {
	if jsonOutput {
		return &common.UsageError{Message: "TUI requires text output"}
	}
	if size != "" {
		if _, _, err := tui.ParseDemoSize(size); err != nil {
			return &common.UsageError{Message: err.Error()}
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	resolvedHome, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return err
	}
	return tui.RunDemo(resolvedHome, fixture, size, cmd.InOrStdin(), cmd.OutOrStdout())
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
