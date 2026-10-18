package task

import (
	"errors"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// ShowOptions holds the values of one task read.
type ShowOptions struct {
	Brief bool
}

// ShowResult holds one task record and its brief text.
type ShowResult struct {
	records.Task
	Brief string `json:"brief"`
}

// Show returns one task record and its brief without taking the lock.
func Show(home, id string, options ShowOptions) output.Envelope {
	if !records.ValidTaskID(id) {
		return failedEnvelope("validate", "task not found: "+id)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failedEnvelope("store", err.Error())
	}
	var record records.Task
	if err := registry.ReadIn([]string{tasksCollection}, id, &record); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failedEnvelope("read", "task not found: "+id)
		}
		return failedEnvelope("read", err.Error())
	}
	brief, err := readBrief(briefPath(home, id), nil, briefDeadline)
	if err != nil {
		return failedEnvelope("brief", err.Error())
	}
	return okEnvelope(ShowResult{Task: record, Brief: string(brief)})
}
