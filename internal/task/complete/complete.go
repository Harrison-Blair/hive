// Package complete moves one task to completed for its acting agent.
package complete

import (
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Run moves one task to completed for the acting agent.
func Run(home string, options tasklib.TransitionOptions) output.Envelope {
	return tasklib.RunFinish(home, "complete", options)
}
