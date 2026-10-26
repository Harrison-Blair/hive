// Package cancel moves one task to cancelled and prunes the stopped records.
package cancel

import (
	"context"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/prune"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Run moves one task to cancelled and stores the reason.
func Run(home string, options tasklib.TransitionOptions) output.Envelope {
	envelope := tasklib.RunTransition(home, "cancel", options)
	task, ok := envelope.Result.(records.Task)
	if !ok || task.Status != "cancelled" {
		return envelope
	}
	if _, err := prune.Run(context.Background(), home); err != nil {
		log := tasklib.OpenLog(home, options.Logger)
		defer log.Close()
		log.Failure("task cancel", task.Project, task.ID, "", "prune", err.Error())
	}
	return envelope
}
