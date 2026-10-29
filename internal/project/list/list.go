// Package list reads the project registry and renders the project list.
package list

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// List returns project records in ID order.
func List(home string) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return output.Failed("store", err.Error())
	}
	rawRecords, err := registry.List("projects")
	if err != nil {
		return output.Failed("store", err.Error())
	}
	projects := make([]records.Project, 0, len(rawRecords))
	for _, raw := range rawRecords {
		var record records.Project
		if err := json.Unmarshal(raw, &record); err != nil {
			return output.Failed("store", fmt.Sprintf("record invalid: projects: %v", err))
		}
		projects = append(projects, record)
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: projects}
}

// Text renders a project list.
func Text(envelope output.Envelope) string {
	projects, ok := envelope.Result.([]records.Project)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(projects) == 0 {
		return "no projects\n"
	}
	var lines []string
	for _, record := range projects {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s", record.ID, record.Path, record.Base))
	}
	return strings.Join(lines, "\n") + "\n"
}
