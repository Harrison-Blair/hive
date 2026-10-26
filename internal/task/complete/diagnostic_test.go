package complete

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestWrongActorErrorStaysOneLine(t *testing.T) {
	home, task := tasktest.NewSeed(t, "assigned")
	tasktest.SeedAgent(t, home, records.Agent{
		Key: "worker-key", Name: "bad\nactor", Profile: records.ProfileBuilder, Project: "one", Task: task.ID,
		Status: records.AgentLive, PaneID: "w1:p9", Generation: 1,
	})
	before := tasktest.TaskRecordBytes(t, home, task.ID)
	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "worker-key"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Complete() = %#v; want failed", envelope)
	}
	message := tasktest.AssertOneLineError(t, "Complete", envelope.Error)
	if !strings.HasPrefix(message, "not the acting agent of 1-seed: ") {
		t.Fatalf("Complete() error = %q; want a wrong actor message", message)
	}
	if !strings.Contains(message, `"bad\nactor"`) {
		t.Fatalf("Complete() error = %q; want the quoted control-bearing record name", message)
	}
	if after := tasktest.TaskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("wrong actor changed the task record")
	}
}

func TestTransitionInvalidIDErrorStaysOneLine(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	envelope := Run(home, tasklib.TransitionOptions{Task: "bad\nid", Agent: "worker-1"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Complete() = %#v; want failed", envelope)
	}
	message := tasktest.AssertOneLineError(t, "Complete", envelope.Error)
	if message != `task not found: "bad\nid"` {
		t.Fatalf("Complete() error = %q; want a quoted task ID", message)
	}
}
