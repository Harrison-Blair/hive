package task

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ShowOptions holds the values of one task read.
type ShowOptions struct {
	Brief  bool
	Logger *slog.Logger
}

// ShowResult holds one task record and its brief text.
type ShowResult struct {
	records.Task
	Brief string         `json:"brief"`
	Agent *records.Agent `json:"agent,omitempty"`
}

// Show returns one task record and its brief without taking the lock.
func Show(home, id string, options ShowOptions) output.Envelope {
	log := openOperationLog(home, options.Logger)
	defer log.close()
	fail := func(phase, message string) output.Envelope {
		log.failure("task show", "", id, "", phase, message)
		return failedEnvelope(phase, message)
	}
	if !records.ValidTaskID(id) {
		return fail("validate", "task not found: "+quoteIfControl(id))
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var record records.Task
	var agent *records.Agent
	var readErr error
	if err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{tasksCollection}, id, &record); err != nil {
			readErr = err
			return nil
		}
		if record.ActingAgent == "" {
			return nil
		}
		found, ok, err := identity.FindHeld(locked, record.ActingAgent)
		if err != nil {
			readErr = err
			return nil
		}
		if ok {
			agent = &found
		}
		return nil
	}); err != nil {
		return fail("store", err.Error())
	}
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return fail("read", "task not found: "+quoteIfControl(id))
		}
		return fail("read", readErr.Error())
	}
	if err := checkBriefDir(home); err != nil {
		log.failure("task show", record.Project, id, "", "brief", err.Error())
		return failedEnvelope("brief", err.Error())
	}
	brief, err := readBrief(briefPath(home, id), nil, briefDeadline)
	if err != nil {
		log.failure("task show", record.Project, id, "", "brief", err.Error())
		return failedEnvelope("brief", err.Error())
	}
	return okEnvelope(ShowResult{Task: record, Brief: string(brief), Agent: agent})
}
