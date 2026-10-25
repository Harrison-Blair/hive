package task

import (
	"context"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/prune"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Cancel moves one task to cancelled and stores the reason.
func Cancel(home string, options TransitionOptions) output.Envelope {
	envelope := runTransition(home, "cancel", options)
	task, ok := envelope.Result.(records.Task)
	if !ok || task.Status != "cancelled" {
		return envelope
	}
	if _, err := prune.Run(context.Background(), home); err != nil {
		log := openOperationLog(home, options.Logger)
		defer log.close()
		log.failure("task cancel", task.Project, task.ID, "", "prune", err.Error())
	}
	return envelope
}
