package task

import (
	"errors"
	"log/slog"
	"os"

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
	Brief string `json:"brief"`
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
		return fail("validate", "task not found: "+id)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var record records.Task
	if err := registry.ReadIn([]string{tasksCollection}, id, &record); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("read", "task not found: "+id)
		}
		return fail("read", err.Error())
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
	return okEnvelope(ShowResult{Task: record, Brief: string(brief)})
}
