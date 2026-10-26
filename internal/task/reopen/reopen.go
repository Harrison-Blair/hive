// Package reopen moves one verified task back to assigned.
package reopen

import (
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Run moves one verified task back to assigned.
func Run(home string, options tasklib.TransitionOptions) output.Envelope {
	options.ResolvedActor = &records.Agent{}
	return tasklib.RunTransition(home, "reopen", options)
}
