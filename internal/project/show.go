package project

import (
	"errors"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Show returns one project record.
func Show(home, id string) output.Envelope {
	if !records.ValidID(id) {
		return failed("validate", invalidIDMessage(id))
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error())
	}
	var record records.Project
	if err := registry.Read("projects", id, &record); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failed("read", "project not found: "+id)
		}
		return failed("store", err.Error())
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}
