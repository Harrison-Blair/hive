// Package doctor wires the `hive doctor` command.
package doctor

import (
	"github.com/Harrison-Blair/hive/cmd/common"
	doctorlib "github.com/Harrison-Blair/hive/internal/doctor"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
)

// New returns the doctor command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the environment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return common.Run(cmd, "doctor", func(home string) (output.Envelope, string, error) {
				logger, err := hivelog.Open(home, "doctor")
				if err != nil {
					return output.Envelope{
						Outcome: "failed",
						Phase:   "log",
						Effects: []output.Effect{},
						Error:   err.Error(),
						Result:  nil,
					}, "", nil
				}
				result := doctorlib.Run(cmd.Context(), home, logger)
				outcome := "ok"
				if result.Failed() {
					outcome = "failed"
				}
				return output.Envelope{
					Outcome: outcome,
					Phase:   "complete",
					Effects: []output.Effect{},
					Error:   nil,
					Result:  result.Checks,
				}, doctorlib.Render(result), nil
			})
		},
	}
}
