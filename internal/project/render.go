package project

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// RegisterText renders the effects of project registration.
func RegisterText(envelope output.Envelope) string {
	var lines []string
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// ListText renders a project list.
func ListText(envelope output.Envelope) string {
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

// ShowText renders one project record.
func ShowText(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Project)
	if !ok || envelope.Error != nil {
		return ""
	}
	return fmt.Sprintf("id: %s\npath: %s\nbase: %s\nregistered_at: %s\n", record.ID, record.Path, record.Base, record.RegisteredAt)
}
