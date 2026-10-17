// Package cmd holds the hive root command and registers its subcommands.
package cmd

import (
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/version"
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
			return tui.Run(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.AddCommand(version.New())
	return root
}

// Execute runs hive with the process arguments and standard streams.
func Execute() error {
	return ExecuteWithArgs(os.Args[1:], os.Stdin, os.Stdout)
}

// ExecuteWithArgs runs hive with the given arguments and streams.
func ExecuteWithArgs(args []string, in io.Reader, out io.Writer) error {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(out)
	return root.Execute()
}
