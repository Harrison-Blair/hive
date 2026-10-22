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
	retry := task.Delivery != nil && task.Delivery.Commit != ""
	delivery := records.Delivery{Mode: "merge", Target: base, Grant: options.Grant, Evidence: options.Evidence}
	if retry {
		delivery = resumeDelivery(*task.Delivery, "merge", base, options)
	}
	effects := []output.Effect{}
	note := ""
	// fail reports a failure after an effect. A completed effect stays visible
	// in the envelope and the record.
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, project.ID, task.ID, phase, message)
		if len(effects) == 0 {
			return failedEnvelope(phase, message)
		}
		delivery.At = now()
		return deliveryEnvelope(registry, logger, task, project, delivery, "partial", phase, message, "", effects)
	}
	if !retry {
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
		alreadyMerged, err := gitx.IsMerged(ctx, repo, base, branch)
		if err != nil {
			return fail("precondition", "check whether task branch is on base: "+oneLine(err.Error()))
		}
		var commit string
		action := "merged"
		if alreadyMerged {
			commit, err = gitx.Tip(ctx, repo, branch)
			if err != nil {
				return fail("merge", "read task branch tip: "+oneLine(err.Error()))
			}
			action = "already-merged"
			note = "branch was already on the base"
		} else {
			mergeCtx, cancel := context.WithTimeout(ctx, bound(options.MergeTimeout, mergeTimeout))
			result, mergeErr := gitx.MergeNoFF(mergeCtx, repo, branch)
			cancel()
			if mergeErr != nil {
				return cleanupMerge(ctx, logger, project, task, oneLine(mergeErr.Error()))
			}
			if result.ExitCode != 0 {
				return cleanupMerge(ctx, logger, project, task, gitOutput(result))
			}
			commit, err = gitx.Tip(ctx, repo, "HEAD")
			if err != nil {
				return fail("merge", "read merge commit: "+oneLine(err.Error()))
			}
		}
		effects = append(effects, output.Effect{Resource: "merge", Path: commit, Action: action})
		logEffect(logger, project.ID, task.ID, "merge", commit, action)
		delivery.Commit = commit
		delivery.Pushed = false
		delivery.At = now()
		if _, err := storeDelivery(registry, task.ID, project.ID, delivery); err != nil {
			logFailure(logger, project.ID, task.ID, "record", err.Error())
			return failedEnvelopeWithEffects("record", "persist delivery: "+oneLine(err.Error()), effects)
		}
	}

	if delivery.Pushed {
		return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", "", effects)
	}
	hasOrigin, err := gitx.HasRemote(ctx, repo, "origin")
	if err != nil {
		return fail("push", "read remotes: "+oneLine(err.Error()))
	}
	if !hasOrigin {
		delivery.Pushed = false
		delivery.At = now()
		noOriginNote := "no origin remote: the merge is local"
		if note != "" {
			note += "; " + noOriginNote
		} else {
			note = noOriginNote
		}
		return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", note, effects)
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
		delivery.Pushed = false
		delivery.At = now()
		return deliveryEnvelope(registry, logger, task, project, delivery, "partial", "push", message, note, effects)
	}
	effects = append(effects, output.Effect{Resource: "push", Path: base, Action: "pushed"})
	logEffect(logger, project.ID, task.ID, "push", base, "pushed")
	delivery.Pushed = true
	delivery.At = now()
	return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", note, effects)
}

// cleanupMerge aborts an unfinished merge with a fresh bounded context. The
// caller context can be expired when the merge deadline fires, so the cleanup
// must not depend on it.
func cleanupMerge(ctx context.Context, logger *slog.Logger, project records.Project, task records.Task, gitText string) output.Envelope {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
	defer cancel()
	return mergeFailure(cleanupCtx, logger, project, task, gitText)
}

// mergeFailure aborts an unfinished merge and reports the merge error.
func mergeFailure(ctx context.Context, logger *slog.Logger, project records.Project, task records.Task, gitText string) output.Envelope {
	repo := project.Path
	inProgress, inProgressErr := gitx.MergeInProgress(ctx, repo)
	if inProgressErr == nil && inProgress {
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
	}
	if message, ok := restoreWorktree(ctx, repo); !ok {
		logFailure(logger, project.ID, task.ID, "merge", message)
		return failedEnvelopeWithEffects("merge", message, []output.Effect{{Resource: "merge", Action: "in progress"}})
	}
	logFailure(logger, project.ID, task.ID, "merge", gitText)
	return failedEnvelope("merge", gitText)
}

// restoreWorktree resets a checkout that a killed merge left staged. The
// merge preconditions require a clean repository, so any remaining change is
// from the merge. It reports a failure message when the repository stays
// dirty.
func restoreWorktree(ctx context.Context, repo string) (string, bool) {
	dirty, err := gitx.IsDirty(ctx, repo)
	if err != nil {
		return "merge cleanup: " + oneLine(err.Error()), false
	}
	if !dirty {
		return "", true
	}
	result, err := run.Command(ctx, repo, "git", "reset", "--hard", "HEAD")
	if err != nil {
		return "merge cleanup: " + oneLine(err.Error()), false
	}
	if result.ExitCode != 0 {
		return "merge cleanup: " + gitOutput(result), false
	}
	return "", true
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
