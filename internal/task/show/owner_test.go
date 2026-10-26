package show

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// TestTaskShowPrintsProjectMail proves AC8: task show prints mail: project:ID
// and no controlling_role line.
func TestTaskShowPrintsProjectMail(t *testing.T) {
	text := Text(tasklib.OKEnvelope(Result{Task: records.Task{ID: "1-step", Project: "demo"}}))
	if !strings.Contains(text, "mail: project:demo") {
		t.Fatalf("ShowText() = %q; want mail: project:demo", text)
	}
	if strings.Contains(text, "controlling_role:") {
		t.Fatalf("ShowText() = %q; want no controlling_role line", text)
	}
}
