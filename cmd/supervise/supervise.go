// Package supervise wires the `hive supervise` command.
package supervise

import (
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	superviselib "github.com/Harrison-Blair/hive/internal/supervise"
	"github.com/spf13/cobra"
)

// New returns the supervise command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "supervise",
		Short: "run the supervisor of this home",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonOutput, err := cmd.Root().PersistentFlags().GetBool("json")
			if err != nil {
				return err
			}
			return common.Run(cmd, "supervise", func(home string) (output.Envelope, string, error) {
				writer, err := log.NewWriter(home, "supervise", time.Now)
				if err != nil {
					return output.Envelope{Outcome: "failed", Phase: "log", Effects: []output.Effect{}, Error: err.Error(), Result: nil}, "", nil
				}
				defer func() { _ = writer.Close() }()
				logger := slog.New(writer)
				ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
				defer stop()
				out := cmd.OutOrStdout()
				if jsonOutput {
					out = io.Discard
				}
				pid := int64(os.Getpid())
				start, _ := process.Start(pid)
				_, envelope := superviselib.Run(ctx, superviselib.Options{
					Home:        home,
					Logger:      logger,
					Out:         out,
					PaneID:      os.Getenv("HERDR_PANE_ID"),
					WorkspaceID: os.Getenv("HERDR_WORKSPACE_ID"),
					TabID:       os.Getenv("HERDR_TAB_ID"),
					PID:         pid,
					PIDStart:    start,
				})
				return envelope, "", nil
			})
		},
	}
}
