package show

import (
	"encoding/json"
	"strings"
	"testing"

	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/cancel"
	"github.com/Harrison-Blair/hive/internal/task/create"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func showLines(t *testing.T, home, id string) []string {
	t.Helper()
	envelope := Run(home, id, Options{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show(%q) = %#v; want ok", id, envelope)
	}
	return strings.Split(strings.TrimSuffix(Text(envelope), "\n"), "\n")
}

func cancelReasonLine(t *testing.T, lines []string) (string, bool) {
	t.Helper()
	for _, line := range lines {
		if strings.HasPrefix(line, "cancel_reason:") {
			return line, true
		}
	}
	return "", false
}

func TestShowTextRendersCancelReason(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := createOK(t, home, create.Options{Project: "one", Title: "Stop"})
	if envelope := cancel.Run(home, tasklib.TransitionOptions{Task: task.ID, Reason: "stop reason"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}

	lines := showLines(t, home, task.ID)
	line, ok := cancelReasonLine(t, lines)
	if !ok {
		t.Fatalf("ShowText() = %q; want a cancel_reason field", strings.Join(lines, "\n"))
	}
	if line != "cancel_reason: stop reason" {
		t.Fatalf("cancel reason line = %q; want the stored reason", line)
	}
}

func TestShowTextKeepsCancelReasonOnOneLine(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	plain := createOK(t, home, create.Options{Project: "one", Title: "Plain"})
	if envelope := cancel.Run(home, tasklib.TransitionOptions{Task: plain.ID, Reason: "stop reason"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	control := createOK(t, home, create.Options{Project: "one", Title: "Control"})
	if envelope := cancel.Run(home, tasklib.TransitionOptions{Task: control.ID, Reason: "stop\nreason\rtab\there"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}

	plainLines := showLines(t, home, plain.ID)
	controlLines := showLines(t, home, control.ID)
	if len(controlLines) != len(plainLines) {
		t.Fatalf("control reason produced %d lines; want %d", len(controlLines), len(plainLines))
	}
	line, ok := cancelReasonLine(t, controlLines)
	if !ok {
		t.Fatalf("ShowText() = %q; want a cancel_reason field", strings.Join(controlLines, "\n"))
	}
	if line != "cancel_reason: stop reason tab here" {
		t.Fatalf("control reason line = %q; want one safe line", line)
	}
}

func TestShowJSONOmitsEmptyCancelReason(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := createOK(t, home, create.Options{Project: "one", Title: "Omit"})

	data, err := json.Marshal(Run(home, task.ID, Options{}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "cancel_reason") {
		t.Fatalf("JSON = %s; want no cancel_reason field for an empty reason", data)
	}

	if envelope := cancel.Run(home, tasklib.TransitionOptions{Task: task.ID, Reason: "visible reason"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	data, err = json.Marshal(Run(home, task.ID, Options{}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"cancel_reason":"visible reason"`) {
		t.Fatalf("JSON = %s; want the stored cancel reason", data)
	}
}

func TestShowInvalidIDErrorStaysOneLine(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	for _, id := range []string{"bad\nsecond-line", "bad\rsecond", "bad\tthird"} {
		envelope := Run(home, id, Options{})
		if envelope.Outcome != "failed" {
			t.Fatalf("Show(%q) = %#v; want failed", id, envelope)
		}
		message := tasktest.AssertOneLineError(t, "Show", envelope.Error)
		if !strings.HasPrefix(message, "task not found: ") {
			t.Fatalf("Show(%q) error = %q; want a task not found message", id, message)
		}
	}
}
