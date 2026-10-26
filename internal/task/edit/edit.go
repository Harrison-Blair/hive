// Package edit changes one editable field of a task.
package edit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Options holds the values for one task edit.
type Options struct {
	Task           string
	Agent          string
	PaneID         string
	Feature        string
	FeatureSet     bool
	BriefSet       bool
	NeedsUser      string
	NeedsUserSet   bool
	ClearNeedsUser bool
	BriefPath      string
	Input          io.Reader
	Deadline       time.Duration
	Logger         *slog.Logger
}

// Run changes one editable field of a task.
func Run(home string, options Options) output.Envelope {
	log := tasklib.OpenLog(home, options.Logger)
	defer log.Close()
	actorID := options.Agent
	fail := func(phase, message string) output.Envelope {
		log.Failure("task edit", "", options.Task, actorID, phase, message)
		return tasklib.FailedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+tasklib.QuoteIfControl(options.Task))
	}
	if options.FeatureSet && options.BriefSet {
		return fail("validate", "task edit accepts --brief or --feature, not both")
	}
	operations := 0
	for _, selected := range []bool{options.BriefSet, options.FeatureSet, options.NeedsUserSet, options.ClearNeedsUser} {
		if selected {
			operations++
		}
	}
	if operations != 1 {
		return fail("validate", "task edit requires exactly one of --brief, --feature, --needs-user, or --clear-needs-user")
	}
	if options.NeedsUserSet {
		if len([]rune(options.NeedsUser)) > 200 {
			return fail("validate", "needs_user must be at most 200 characters")
		}
		if strings.ContainsAny(options.NeedsUser, "\r\n\u0085\u2028\u2029") {
			return fail("validate", "needs_user must be one line")
		}
	}
	if options.FeatureSet {
		if err := tasklib.CheckFeature(options.Feature); err != nil {
			return fail("validate", err.Error())
		}
	}
	deadline := options.Deadline
	if deadline <= 0 {
		deadline = tasklib.BriefDeadline
	}
	var brief []byte
	var err error
	if options.BriefSet {
		brief, err = tasklib.ReadBrief(options.BriefPath, options.Input, deadline)
		if err != nil {
			return fail("brief", err.Error())
		}
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var task records.Task
		if err := locked.ReadIn([]string{tasklib.TasksCollection}, options.Task, &task); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = tasklib.FailedEnvelope("edit", "task not found: "+tasklib.QuoteIfControl(options.Task))
			} else {
				envelope = tasklib.FailedEnvelope("store", err.Error())
			}
			log.Failure("task edit", "", options.Task, actorID, envelope.Phase, fmt.Sprint(envelope.Error))
			return nil
		}
		if options.BriefSet && task.Status != "created" && task.Status != "assigned" {
			message := fmt.Sprintf("task is %s: edit needs created or assigned", tasklib.QuoteIfControl(task.Status))
			envelope = tasklib.FailedEnvelope("edit", message)
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
			return nil
		}
		caller, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil {
			envelope = tasklib.FailedEnvelope("identity", err.Error())
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		actorID = caller.Key
		if err := tasklib.CheckOwner(task, caller); err != nil {
			envelope = tasklib.FailedEnvelope("identity", err.Error())
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		if options.NeedsUserSet || options.ClearNeedsUser {
			if options.NeedsUserSet {
				task.NeedsUser = options.NeedsUser
			} else {
				task.NeedsUser = ""
			}
			task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if err := locked.WriteIn([]string{tasklib.TasksCollection}, task.ID, task); err != nil {
				envelope = tasklib.FailedEnvelope("store", err.Error())
				log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
				return nil
			}
			envelope = output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: []output.Effect{{Resource: "task", Path: tasklib.TaskRelPath(task.ID), Action: "updated"}},
				Error:   nil,
				Result:  task,
			}
			log.Effect("task edit", task.Project, task.ID, "task", tasklib.TaskRelPath(task.ID), "updated")
			return nil
		}
		if options.FeatureSet {
			if task.Feature != "" {
				message := "task already has a feature: " + task.Feature
				envelope = tasklib.FailedEnvelope("edit", message)
				log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			task.Feature = options.Feature
			task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if err := locked.WriteIn([]string{tasklib.TasksCollection}, task.ID, task); err != nil {
				envelope = tasklib.FailedEnvelope("store", err.Error())
				log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
				return nil
			}
			envelope = output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: []output.Effect{{Resource: "task", Path: tasklib.TaskRelPath(task.ID), Action: "updated"}},
				Error:   nil,
				Result:  task,
			}
			log.Effect("task edit", task.Project, task.ID, "task", tasklib.TaskRelPath(task.ID), "updated")
			return nil
		}
		if err := tasklib.CheckBriefDir(home); err != nil {
			envelope = tasklib.FailedEnvelope("brief", err.Error())
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		oldBrief, err := tasklib.ReadBrief(tasklib.BriefPath(home, task.ID), nil, deadline)
		if err != nil {
			envelope = tasklib.FailedEnvelope("brief", err.Error())
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		if err := tasklib.WriteBrief(home, task.ID, brief); err != nil {
			envelope = tasklib.FailedEnvelope("brief", err.Error())
			log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		effects := []output.Effect{{Resource: "brief", Path: tasklib.BriefRelPath(task.ID), Action: "updated"}}
		if paths := tasklib.TaskFeaturePaths(task); paths != nil {
			if err := tasklib.WriteWorktreeBrief(home, paths.Brief, brief); err != nil {
				message := "write feature brief: " + err.Error()
				envelope = output.Envelope{Outcome: "partial", Phase: "brief", Effects: effects, Error: message, Result: task}
				log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			effects = append(effects, output.Effect{Resource: "brief", Path: paths.Brief, Action: "updated"})
			log.Effect("task edit", task.Project, task.ID, "brief", paths.Brief, "updated")
		}
		if paths := tasklib.TaskFeaturePaths(task); paths != nil && task.Worktree != nil && task.Worktree.RemovedAt == "" {
			featurePath := filepath.Join(task.Worktree.Path, filepath.FromSlash(paths.Brief))
			if err := tasklib.WriteWorktreeBrief(task.Worktree.Path, paths.Brief, brief); err != nil {
				message := "copy feature brief: " + err.Error()
				envelope = output.Envelope{
					Outcome: "partial",
					Phase:   "brief",
					Effects: effects,
					Error:   message,
					Result:  task,
				}
				log.Failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			effects = append(effects, output.Effect{Resource: "brief", Path: featurePath, Action: "copied"})
			log.Effect("task edit", task.Project, task.ID, "brief", featurePath, "copied")
		}
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: effects,
			Error:   nil,
			Result:  task,
		}
		logBriefEdit(log, task.Project, task.ID, actorID, len(oldBrief), len(brief))
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	return envelope
}

// Text renders the effects of a successful task edit.
func Text(envelope output.Envelope) string {
	return tasklib.EffectsText(envelope.Effects)
}

// logBriefEdit writes one completed brief replacement without its text.
func logBriefEdit(l tasklib.Log, projectID, taskID, agentID string, oldSize, newSize int) {
	if l.Logger == nil {
		return
	}
	fields := l.Fields("task edit", projectID, taskID, agentID)
	fields = append(fields, "old_size", oldSize, "new_size", newSize)
	l.Logger.Debug("task brief edit", fields...)
}
