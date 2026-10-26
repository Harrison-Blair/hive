package task

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestFailureReportPrintsProjectInbox proves AC8: the failure report names the
// project inbox and --inbox.
func TestFailureReportPrintsProjectInbox(t *testing.T) {
	envelope := output.Envelope{Outcome: "partial", Phase: "inbox", Result: records.Task{ID: "1-step", Project: "demo"}}
	text := FailureReport(envelope, "1-step", "builder-1")
	if !strings.Contains(text, "inspect: hive inbox --inbox project:demo --all") {
		t.Fatalf("FailureReport() = %q; want inspect: hive inbox --inbox project:demo --all", text)
	}
}
