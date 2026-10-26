package complete

import (
	"encoding/json"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// TestTaskCompleteMailsTheProjectOwner proves AC2: task complete sends its
// record to the owner of the project of the task.
func TestTaskCompleteMailsTheProjectOwner(t *testing.T) {
	home := tasktest.SeedOwnerHome(t)
	envelope := Run(home, tasklib.TransitionOptions{Task: "1-step", Agent: "builder-1"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	found := readOwnerInbox(t, home, "owner-demo")
	if len(found) != 1 {
		t.Fatalf("owner inbox = %#v; want one completed record", found)
	}
	record := found[0]
	if record.Kind != records.InboxCompleted || record.Task != "1-step" || record.Agent != "builder-1" || record.Text != "completed" {
		t.Fatalf("record = %#v; want the completed record of 1-step", record)
	}
	if _, ok := tasktest.ReadOwnerRecord(t, home, "tasks", "1-step.json")["controlling_role"]; ok {
		t.Fatal("the task record still has controlling_role")
	}
}

// TestTaskCompleteUsesCoordinatorWithoutOwner proves AC2 with no live owner.
func TestTaskCompleteUsesCoordinatorWithoutOwner(t *testing.T) {
	home := tasktest.SeedOwnerHome(t)
	stopped := tasktest.ReadOwnerRecord(t, home, "agents", "owner-demo-g1.json")
	stopped["status"] = "stopped"
	tasktest.WriteOwnerState(t, home, []string{"agents", "owner-demo-g1.json"}, stopped)

	envelope := Run(home, tasklib.TransitionOptions{Task: "1-step", Agent: "builder-1"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	found := readOwnerInbox(t, home, "coordinator")
	if len(found) != 1 || found[0].Task != "1-step" {
		t.Fatalf("coordinator inbox = %#v; want the completed record", found)
	}
	if got := readOwnerInbox(t, home, "owner-demo"); len(got) != 0 {
		t.Fatalf("owner inbox = %#v; want no record for a stopped owner", got)
	}
}

func readOwnerInbox(t *testing.T, home, key string) []records.Inbox {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := registry.ListIn([]string{"inbox", key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := make([]records.Inbox, 0, len(raw))
	for _, data := range raw {
		var record records.Inbox
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		found = append(found, record)
	}
	return found
}
