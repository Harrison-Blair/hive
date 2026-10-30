package project

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// recordName is the name of the record in the directory of each project:
// <home>/projects/<id>/project.json.
const recordName = "project"

func collection(id string) []string {
	return []string{"projects", id}
}

// RecordPath returns the path of the record of id relative to the home.
func RecordPath(id string) string {
	return filepath.Join(append(collection(id), recordName+".json")...)
}

// Read reads the record of project id.
func Read(registry *store.Store, id string, record *records.Project) error {
	return registry.ReadIn(collection(id), recordName, record)
}

// Write writes the record of one project.
func Write(registry *store.Store, record records.Project) error {
	return registry.WriteIn(collection(record.ID), recordName, record)
}

// List reads the record of each project in ID order.
func List(registry *store.Store) ([]records.Project, error) {
	raws, err := registry.ListEach([]string{"projects"}, recordName)
	if err != nil {
		return nil, err
	}
	projects := make([]records.Project, 0, len(raws))
	for _, raw := range raws {
		var record records.Project
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, fmt.Errorf("record invalid: projects: %v", err)
		}
		projects = append(projects, record)
	}
	return projects, nil
}
