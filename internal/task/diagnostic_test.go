package task

import (
	"encoding/json"
	"strings"
	"testing"
)

func showLines(t *testing.T, home, id string) []string {
	t.Helper()
	envelope := Show(home, id, ShowOptions{})
	if envelope.Outcome != "ok" {
		t.Fatalf("Show(%q) = %#v; want ok", id, envelope)
	}
	return strings.Split(strings.TrimSuffix(ShowText(envelope), "\n"), "\n")
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
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Stop"})
	if envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "stop reason"}); envelope.Outcome != "ok" {
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
	home := newHome(t)
	registerProject(t, home, "one")
	plain := createOK(t, home, CreateOptions{Project: "one", Title: "Plain"})
	if envelope := Cancel(home, TransitionOptions{Task: plain.ID, Reason: "stop reason"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	control := createOK(t, home, CreateOptions{Project: "one", Title: "Control"})
	if envelope := Cancel(home, TransitionOptions{Task: control.ID, Reason: "stop\nreason\rtab\there"}); envelope.Outcome != "ok" {
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
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Omit"})

	data, err := json.Marshal(Show(home, task.ID, ShowOptions{}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "cancel_reason") {
		t.Fatalf("JSON = %s; want no cancel_reason field for an empty reason", data)
	}

	if envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "visible reason"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	data, err = json.Marshal(Show(home, task.ID, ShowOptions{}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"cancel_reason":"visible reason"`) {
		t.Fatalf("JSON = %s; want the stored cancel reason", data)
	}
}

func assertOneLineError(t *testing.T, label string, envelopeError any) string {
	t.Helper()
	message, ok := envelopeError.(string)
	if !ok {
		t.Fatalf("%s error = %#v; want a string", label, envelopeError)
	}
	if strings.ContainsAny(message, "\n\r\t") {
		t.Fatalf("%s error = %q; want one plain line", label, message)
	}
	return message
}

func TestShowInvalidIDErrorStaysOneLine(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	for _, id := range []string{"bad\nsecond-line", "bad\rsecond", "bad\tthird"} {
		envelope := Show(home, id, ShowOptions{})
		if envelope.Outcome != "failed" {
			t.Fatalf("Show(%q) = %#v; want failed", id, envelope)
		}
		message := assertOneLineError(t, "Show", envelope.Error)
		if !strings.HasPrefix(message, "task not found: ") {
			t.Fatalf("Show(%q) error = %q; want a task not found message", id, message)
		}
	}
}

func TestWrongActorErrorStaysOneLine(t *testing.T) {
	home, task := newSeed(t, "assigned")
	before := taskRecordBytes(t, home, task.ID)
	envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-2"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Complete() = %#v; want failed", envelope)
	}
	message := assertOneLineError(t, "Complete", envelope.Error)
	if !strings.HasPrefix(message, "not the acting agent: ") {
		t.Fatalf("Complete() error = %q; want a wrong actor message", message)
	}
	if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("wrong actor changed the task record")
	}
}

func TestTransitionInvalidIDErrorStaysOneLine(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := Complete(home, TransitionOptions{Task: "bad\nid", Agent: "worker-1"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Complete() = %#v; want failed", envelope)
	}
	message := assertOneLineError(t, "Complete", envelope.Error)
	if message != `task not found: "bad\nid"` {
		t.Fatalf("Complete() error = %q; want a quoted task ID", message)
	}
}

func TestCreateBriefPathErrorStaysOneLine(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := Create(home, CreateOptions{
		Project:   "one",
		Title:     "Step",
		BriefPath: "missing\nbrief.md",
		BriefSet:  true,
	})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %#v; want failed", envelope)
	}
	message := assertOneLineError(t, "Create", envelope.Error)
	if !strings.HasPrefix(message, "brief not found: ") {
		t.Fatalf("Create() error = %q; want a brief not found message", message)
	}
}
