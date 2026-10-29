package project

import (
	"encoding/json"
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// List returns project records in ID order.
func List(home string) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return Failed("store", err.Error())
	}
	rawRecords, err := registry.List("projects")
	if err != nil {
		return Failed("store", err.Error())
	}
	projects := make([]records.Project, 0, len(rawRecords))
	for _, raw := range rawRecords {
		var record records.Project
		if err := json.Unmarshal(raw, &record); err != nil {
			return Failed("store", fmt.Sprintf("record invalid: projects: %v", err))
		}
		projects = append(projects, record)
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: projects}
}
