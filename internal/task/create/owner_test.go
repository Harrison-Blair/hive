package create

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// TestCreateWritesNoControllingRole proves AC2 for a new task record.
func TestCreateWritesNoControllingRole(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "demo")
	envelope := Run(home, Options{Project: "demo", Title: "Step"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %#v; want ok", envelope)
	}
	record := envelope.Result.(records.Task)
	if _, ok := tasktest.ReadOwnerRecord(t, home, "tasks", record.ID+".json")["controlling_role"]; ok {
		t.Fatal("the new task record has controlling_role")
	}
}
