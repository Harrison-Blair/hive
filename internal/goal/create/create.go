// Package create makes an active goal with no waves.
package create

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	goallib "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Options holds the values for one goal creation.
type Options struct {
	Project string
	Title   string
	Slug    string
	Grant   string
	Agent   string
	PaneID  string
}

// Run makes an active goal with no waves.
func Run(home string, options Options) output.Envelope {
	log := goallib.OpenLog(home)
	defer log.Close()
	actorID := options.Agent
	fail := func(phase, message string) output.Envelope {
		log.Failure("goal create", options.Project, "", actorID, phase, message)
		return goallib.Failed(phase, message)
	}
	if !records.ValidID(options.Project) {
		return fail("validate", "project requires a valid ID")
	}
	if strings.TrimSpace(options.Title) == "" {
		return fail("validate", "title requires a value")
	}
	if strings.TrimSpace(options.Grant) == "" {
		return fail("validate", "grant requires a value")
	}
	if len(options.Slug) > 178 || !slugPattern.MatchString(options.Slug) {
		return fail("validate", "slug must be kebab-case and fit in a goal ID")
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var project records.Project
		if err := locked.ReadIn([]string{"projects"}, options.Project, &project); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = fail("project", "project not found: "+options.Project)
			} else {
				envelope = fail("store", err.Error())
			}
			return nil
		}
		if project.ID != options.Project {
			envelope = fail("project", "project record ID does not match its path")
			return nil
		}
		caller, err := goallib.Authorize(locked, options.Agent, options.PaneID, project.ID)
		if err != nil {
			envelope = fail("identity", err.Error())
			return nil
		}
		actorID = caller.Key
		number, err := locked.Next("goals")
		if err != nil {
			envelope = fail("counter", err.Error())
			return nil
		}
		counterEffect := output.Effect{Resource: "counter", Path: ".hive/state/counters/goals.json", Action: "updated"}
		log.Effect("goal create", project.ID, "", actorID, counterEffect.Resource, counterEffect.Path, counterEffect.Action)
		now := time.Now().UTC().Format(time.RFC3339Nano)
		record := records.Goal{
			ID:        fmt.Sprintf("g%d-%s", number, options.Slug),
			Project:   project.ID,
			Title:     options.Title,
			Grant:     options.Grant,
			Status:    "active",
			Waves:     []records.GoalWave{},
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := locked.WriteIn([]string{goallib.GoalsCollection}, record.ID, record); err != nil {
			log.Failure("goal create", project.ID, record.ID, actorID, "goal", err.Error())
			envelope = output.Envelope{
				Outcome: "partial", Phase: "goal", Effects: []output.Effect{counterEffect},
				Error: err.Error(), Result: nil,
			}
			return nil
		}
		goalEffect := output.Effect{Resource: "goal", Path: goallib.GoalPath(record.ID), Action: "created"}
		log.Effect("goal create", project.ID, record.ID, actorID, goalEffect.Resource, goalEffect.Path, goalEffect.Action)
		log.Created(project.ID, record.ID, actorID)
		envelope = output.Envelope{
			Outcome: "ok", Phase: "complete",
			Effects: []output.Effect{counterEffect, goalEffect},
			Result:  record,
		}
		return nil
	})
	if err != nil {
		return fail("store", err.Error())
	}
	return envelope
}

// Text renders the new goal ID.
func Text(envelope output.Envelope) string {
	if record, ok := envelope.Result.(records.Goal); ok && envelope.Error == nil {
		return fmt.Sprintf("id: %s\n", record.ID)
	}
	return goallib.EffectsText(envelope.Effects)
}
