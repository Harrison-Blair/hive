// Package list renders the project list of the parent project.List API.
package list

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

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
