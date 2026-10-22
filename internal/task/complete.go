package task

import "github.com/Harrison-Blair/hive/internal/lib/output"

// Complete moves one task to completed for the acting agent.
func Complete(home string, options TransitionOptions) output.Envelope {
	return runFinish(home, "complete", options)
}
