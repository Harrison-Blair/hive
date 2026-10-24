package task

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
)

// EditOptions holds the values for one task edit.
type EditOptions struct {
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

// Edit changes one editable field of a task.
func Edit(home string, options EditOptions) output.Envelope {
	log := openOperationLog(home, options.Logger)
	defer log.close()
	actorID := options.Agent
	fail := func(phase, message string) output.Envelope {
		log.failure("task edit", "", options.Task, actorID, phase, message)
		return failedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+quoteIfControl(options.Task))
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
		if err := checkFeature(options.Feature); err != nil {
			return fail("validate", err.Error())
		}
	}
	deadline := options.Deadline
	if deadline <= 0 {
		deadline = briefDeadline
	}
	var brief []byte
	var err error
	if options.BriefSet {
		brief, err = readBrief(options.BriefPath, options.Input, deadline)
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
		if err := locked.ReadIn([]string{tasksCollection}, options.Task, &task); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = failedEnvelope("edit", "task not found: "+quoteIfControl(options.Task))
			} else {
				envelope = failedEnvelope("store", err.Error())
			}
			log.failure("task edit", "", options.Task, actorID, envelope.Phase, fmt.Sprint(envelope.Error))
			return nil
		}
		if options.BriefSet && task.Status != "created" && task.Status != "assigned" {
			message := fmt.Sprintf("task is %s: edit needs created or assigned", quoteIfControl(task.Status))
			envelope = failedEnvelope("edit", message)
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
			return nil
		}
		caller, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil {
			envelope = failedEnvelope("identity", err.Error())
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		actorID = caller.Key
		if err := checkControllingRole(task, caller); err != nil {
			envelope = failedEnvelope("identity", err.Error())
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		if options.NeedsUserSet || options.ClearNeedsUser {
			if options.NeedsUserSet {
				task.NeedsUser = options.NeedsUser
			} else {
				task.NeedsUser = ""
			}
			task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if err := locked.WriteIn([]string{tasksCollection}, task.ID, task); err != nil {
				envelope = failedEnvelope("store", err.Error())
				log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
				return nil
			}
			envelope = output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: []output.Effect{{Resource: "task", Path: taskRelPath(task.ID), Action: "updated"}},
				Error:   nil,
				Result:  task,
			}
			log.effect("task edit", task.Project, task.ID, "task", taskRelPath(task.ID), "updated")
			return nil
		}
		if options.FeatureSet {
			if task.Feature != "" {
				message := "task already has a feature: " + task.Feature
				envelope = failedEnvelope("edit", message)
				log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			task.Feature = options.Feature
			task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if err := locked.WriteIn([]string{tasksCollection}, task.ID, task); err != nil {
				envelope = failedEnvelope("store", err.Error())
				log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
				return nil
			}
			envelope = output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: []output.Effect{{Resource: "task", Path: taskRelPath(task.ID), Action: "updated"}},
				Error:   nil,
				Result:  task,
			}
			log.effect("task edit", task.Project, task.ID, "task", taskRelPath(task.ID), "updated")
			return nil
		}
		if err := checkBriefDir(home); err != nil {
			envelope = failedEnvelope("brief", err.Error())
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		oldBrief, err := readBrief(briefPath(home, task.ID), nil, deadline)
		if err != nil {
			envelope = failedEnvelope("brief", err.Error())
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		if err := writeBrief(home, task.ID, brief); err != nil {
			envelope = failedEnvelope("brief", err.Error())
			log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, err.Error())
			return nil
		}
		effects := []output.Effect{{Resource: "brief", Path: briefRelPath(task.ID), Action: "updated"}}
		if paths := TaskFeaturePaths(task); paths != nil {
			if err := writeWorktreeBrief(home, paths.Brief, brief); err != nil {
				message := "write feature brief: " + err.Error()
				envelope = output.Envelope{Outcome: "partial", Phase: "brief", Effects: effects, Error: message, Result: task}
				log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			effects = append(effects, output.Effect{Resource: "brief", Path: paths.Brief, Action: "updated"})
			log.effect("task edit", task.Project, task.ID, "brief", paths.Brief, "updated")
		}
		if paths := TaskFeaturePaths(task); paths != nil && task.Worktree != nil && task.Worktree.RemovedAt == "" {
			featurePath := filepath.Join(task.Worktree.Path, filepath.FromSlash(paths.Brief))
			if err := writeWorktreeBrief(task.Worktree.Path, paths.Brief, brief); err != nil {
				message := "copy feature brief: " + err.Error()
				envelope = output.Envelope{
					Outcome: "partial",
					Phase:   "brief",
					Effects: effects,
					Error:   message,
					Result:  task,
				}
				log.failure("task edit", task.Project, task.ID, actorID, envelope.Phase, message)
				return nil
			}
			effects = append(effects, output.Effect{Resource: "brief", Path: featurePath, Action: "copied"})
			log.effect("task edit", task.Project, task.ID, "brief", featurePath, "copied")
		}
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: effects,
			Error:   nil,
			Result:  task,
		}
		log.briefEdit(task.Project, task.ID, actorID, len(oldBrief), len(brief))
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	return envelope
}

func writeWorktreeBrief(worktreePath, relativePath string, data []byte) error {
	destination := filepath.Join(worktreePath, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create directory: %s", err)
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return fmt.Errorf("write file: %s", err)
	}
	return nil
}

// EditText renders the effects of a successful task edit.
func EditText(envelope output.Envelope) string {
	return effectsText(envelope.Effects)
}
