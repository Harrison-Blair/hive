package list

import (
	"testing"

	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestListEmpty(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	envelope := tasklib.List(home, tasklib.ListOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("List() = %#v; want ok", envelope)
	}
	if text := Text(envelope); text != "no tasks\n" {
		t.Fatalf("ListText() = %q; want no tasks", text)
	}
}
