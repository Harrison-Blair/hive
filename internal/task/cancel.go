package task

import "github.com/Harrison-Blair/hive/internal/lib/output"

// Cancel moves one task to cancelled and stores the reason.
func Cancel(home string, options TransitionOptions) output.Envelope {
	return runTransition(home, "cancel", options)
}
