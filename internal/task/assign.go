package task

import "github.com/Harrison-Blair/hive/internal/lib/output"

// Assign moves one task to assigned and sets the acting agent.
func Assign(home string, options TransitionOptions) output.Envelope {
	return runTransition(home, "assign", options)
}
