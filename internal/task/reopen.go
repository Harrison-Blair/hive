package task

import (
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Reopen moves one verified task back to assigned.
func Reopen(home string, options TransitionOptions) output.Envelope {
	options.ResolvedActor = &records.Agent{}
	return runTransition(home, "reopen", options)
}
