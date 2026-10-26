// Package show reads one project record.
package show

import (
	"errors"
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
)

// Run returns one project record.
func Run(home, id string) output.Envelope {
	if !records.ValidID(id) {
		return projectlib.Failed("validate", projectlib.InvalidIDMessage(id))
	}
	registry, err := store.Open(home)
	if err != nil {
		return projectlib.Failed("store", err.Error())
	}
	var record records.Project
	if err := registry.Read("projects", id, &record); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return projectlib.Failed("read", "project not found: "+id)
		}
		return projectlib.Failed("store", err.Error())
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}

// Text renders one project record.
func Text(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Project)
	if !ok || envelope.Error != nil {
		return ""
	}
	return fmt.Sprintf("id: %s\npath: %s\nbase: %s\nregistered_at: %s\n", record.ID, record.Path, record.Base, record.RegisteredAt)
}
