package land

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// landMerge merges the task branch into the base branch and pushes the base.
func landMerge(ctx context.Context, registry *store.Store, logger *slog.Logger, options Options, task records.Task, project records.Project) output.Envelope {
	repo := project.Path
	base := project.Base
	branch := task.Worktree.Branch
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, project.ID, task.ID, phase, message)
		return failedEnvelope(phase, message)
	}
	retry := task.Delivery != nil && task.Delivery.Commit != ""
	commit := ""
	pushed := false
	effects := []output.Effect{}
	if retry {
		commit = task.Delivery.Commit
		pushed = task.Delivery.Pushed
	} else {
		current, err := gitx.CurrentBranch(ctx, repo)
		if err != nil {
			return fail("precondition", "read current branch: "+oneLine(err.Error()))
		}
		if current != base {
			return fail("precondition", "base branch not checked out: "+base)
		}
		dirty, err := gitx.IsDirty(ctx, repo)
		if err != nil {
			return fail("precondition", "read repository state: "+oneLine(err.Error()))
		}
		if dirty {
			return fail("precondition", "repository is dirty")
		}
		mergeCtx, cancel := context.WithTimeout(ctx, bound(options.MergeTimeout, mergeTimeout))
		result, mergeErr := gitx.MergeNoFF(mergeCtx, repo, branch)
		cancel()
		if mergeErr != nil {
			return mergeFailure(ctx, logger, project, task, oneLine(mergeErr.Error()))
		}
		if result.ExitCode != 0 {
			return mergeFailure(ctx, logger, project, task, gitOutput(result))
		}
		commit, err = gitx.Tip(ctx, repo, "HEAD")
		if err != nil {
			return fail("merge", "read merge commit: "+oneLine(err.Error()))
		}
		effects = append(effects, output.Effect{Resource: "merge", Path: commit, Action: "merged"})
		logEffect(logger, project.ID, task.ID, "merge", commit, "merged")
	}

	if pushed {
		return deliveryEnvelope(registry, logger, task, project, *task.Delivery, "ok", "complete", "", "", effects)
	}
	hasOrigin, err := gitx.HasRemote(ctx, repo, "origin")
	if err != nil {
		return fail("push", "read remotes: "+oneLine(err.Error()))
	}
	if !hasOrigin {
		delivery := records.Delivery{Mode: "merge", Target: base, Grant: options.Grant, Commit: commit, Pushed: false, Evidence: options.Evidence, At: now()}
		return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", "no origin remote: the merge is local", effects)
	}
	pushCtx, cancel := context.WithTimeout(ctx, bound(options.PushTimeout, pushTimeout))
	pushResult, pushErr := gitx.Push(pushCtx, repo, "origin", base)
	cancel()
	if pushErr != nil || pushResult.ExitCode != 0 {
		var message string
		if pushErr != nil {
			message = "push failed: " + oneLine(pushErr.Error())
		} else {
			message = "push failed: " + gitOutput(pushResult)
		}
		delivery := records.Delivery{Mode: "merge", Target: base, Grant: options.Grant, Commit: commit, Pushed: false, Evidence: options.Evidence, At: now()}
		return deliveryEnvelope(registry, logger, task, project, delivery, "partial", "push", message, "", effects)
	}
	effects = append(effects, output.Effect{Resource: "push", Path: base, Action: "pushed"})
	logEffect(logger, project.ID, task.ID, "push", base, "pushed")
	delivery := records.Delivery{Mode: "merge", Target: base, Grant: options.Grant, Commit: commit, Pushed: true, Evidence: options.Evidence, At: now()}
	return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", "", effects)
}

// mergeFailure aborts an unfinished merge and reports the merge error.
func mergeFailure(ctx context.Context, logger *slog.Logger, project records.Project, task records.Task, gitText string) output.Envelope {
	repo := project.Path
	inProgress, inProgressErr := gitx.MergeInProgress(ctx, repo)
	if inProgressErr != nil || !inProgress {
		logFailure(logger, project.ID, task.ID, "merge", gitText)
		return failedEnvelope("merge", gitText)
	}
	files, filesErr := conflictFiles(ctx, repo)
	if filesErr == nil && len(files) > 0 {
		abortResult, abortErr := gitx.MergeAbort(ctx, repo)
		if abortErr != nil || abortResult.ExitCode != 0 {
			message := "merge conflict: " + strings.Join(files, ", ") + "; merge abort failed: " + abortText(abortErr, abortResult)
			logFailure(logger, project.ID, task.ID, "merge", message)
			return failedEnvelopeWithEffects("merge", message, []output.Effect{{Resource: "merge", Action: "in progress"}})
		}
		message := "merge conflict: " + strings.Join(files, ", ")
		logFailure(logger, project.ID, task.ID, "merge", message)
		return failedEnvelope("merge", message)
	}
	abortResult, abortErr := gitx.MergeAbort(ctx, repo)
	if abortErr != nil || abortResult.ExitCode != 0 {
		message := gitText + "; merge abort failed: " + abortText(abortErr, abortResult)
		logFailure(logger, project.ID, task.ID, "merge", message)
		return failedEnvelopeWithEffects("merge", message, []output.Effect{{Resource: "merge", Action: "in progress"}})
	}
	logFailure(logger, project.ID, task.ID, "merge", gitText)
	return failedEnvelope("merge", gitText)
}

// conflictFiles lists the unmerged paths of one checkout.
func conflictFiles(ctx context.Context, repo string) ([]string, error) {
	result, err := run.Command(ctx, repo, "git", "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("git diff exited with code %d", result.ExitCode)
	}
	files := []string{}
	for _, line := range strings.Split(result.Stdout, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files, nil
}

// gitOutput returns the bounded one-line output of one git command.
func gitOutput(result *run.Result) string {
	if result == nil {
		return "git exited with no result"
	}
	text := strings.TrimSpace(result.Stderr)
	if text == "" {
		text = strings.TrimSpace(result.Stdout)
	}
	if text == "" {
		return fmt.Sprintf("git exited with code %d", result.ExitCode)
	}
	return oneLine(text)
}

func abortText(err error, result *run.Result) string {
	if err != nil {
		return oneLine(err.Error())
	}
	return gitOutput(result)
}
