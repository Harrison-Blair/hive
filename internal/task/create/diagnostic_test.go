package create

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestCreateBriefPathErrorStaysOneLine(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	envelope := Run(home, Options{
		Project:   "one",
		Title:     "Step",
		BriefPath: "missing\nbrief.md",
		BriefSet:  true,
	})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %#v; want failed", envelope)
	}
	message := tasktest.AssertOneLineError(t, "Create", envelope.Error)
	if !strings.HasPrefix(message, "brief not found: ") {
		t.Fatalf("Create() error = %q; want a brief not found message", message)
	}
}
