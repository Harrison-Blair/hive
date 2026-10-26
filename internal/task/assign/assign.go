// Package assign moves one task to assigned for an agent.
package assign

import (
	"github.com/Harrison-Blair/hive/internal/lib/output"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Run moves one task to assigned and sets the acting agent.
func Run(home string, options tasklib.TransitionOptions) output.Envelope {
	return tasklib.RunTransition(home, "assign", options)
}
