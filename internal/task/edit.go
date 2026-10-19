package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// EditOptions holds the values for one brief edit.
type EditOptions struct {
	Task      string
	Agent     string
	PaneID    string
	BriefPath string
	Input     io.Reader
	Deadline  time.Duration
	Logger    *slog.Logger
}

// EditBrief replaces the brief of one created or assigned task.
func EditBrief(home string, options EditOptions) output.Envelope {
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
	deadline := options.Deadline
	if deadline <= 0 {
		deadline = briefDeadline
	}
	brief, err := readBrief(options.BriefPath, options.Input, deadline)
	if err != nil {
		return fail("brief", err.Error())
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
		if task.Status != "created" && task.Status != "assigned" {
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
		envelope = output.Envelope{
			Outcome: "ok",
			Phase:   "complete",
			Effects: []output.Effect{{Resource: "brief", Path: briefRelPath(task.ID), Action: "updated"}},
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

// EditText renders the effect of a successful brief edit.
func EditText(envelope output.Envelope) string {
	return effectsText(envelope.Effects)
}
