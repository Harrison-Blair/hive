package task

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Verify records one independent check of a task.
func Verify(home string, options TransitionOptions) output.Envelope {
	if records.ValidTaskID(options.Task) {
		registry, err := store.Open(home)
		if err == nil {
			var task records.Task
			if err := registry.ReadIn([]string{tasksCollection}, options.Task, &task); err == nil && task.Worktree != nil {
				log := openOperationLog(home, options.Logger)
				defer log.close()
				if options.Logger == nil {
					options.Logger = log.logger
				}
				if err := checkVerifyWorktree(task.Worktree, options.Logger); err != nil {
					log.failure("task verify", task.Project, task.ID, options.Agent, "worktree", err.Error())
					return failedEnvelope("worktree", err.Error())
				}
			}
		}
	}
	return runFinish(home, "verify", options)
}

func checkVerifyWorktree(worktree *records.Worktree, logger *slog.Logger) error {
	ctx := context.Background()
	if logger != nil {
		ctx = run.WithLogger(ctx, logger)
	}
	dirty, err := gitx.IsDirty(ctx, worktree.Path)
	if err != nil {
		return fmt.Errorf("check worktree: %w", err)
	}
	if dirty {
		return fmt.Errorf("worktree is dirty: commit and push before verify")
	}
	remoteTip, exists, err := gitx.FetchBranchTip(ctx, worktree.Path, "origin", worktree.Branch)
	if err != nil {
		return fmt.Errorf("fetch origin %s: %w", worktree.Branch, err)
	}
	if !exists {
		return fmt.Errorf("branch not pushed: %s", worktree.Branch)
	}
	localTip, err := gitx.Tip(ctx, worktree.Path, "refs/heads/"+worktree.Branch)
	if err != nil {
		return fmt.Errorf("read branch %s: %w", worktree.Branch, err)
	}
	if localTip != remoteTip {
		return fmt.Errorf("branch not pushed: %s has commits not on origin", worktree.Branch)
	}
	return nil
}
