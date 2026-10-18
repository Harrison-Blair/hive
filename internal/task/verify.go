package task

import "github.com/Harrison-Blair/hive/internal/lib/output"

// Verify records one independent check of a task.
func Verify(home string, options TransitionOptions) output.Envelope {
	return runFinish(home, "verify", options)
}
