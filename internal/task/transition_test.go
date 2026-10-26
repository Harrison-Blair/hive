package task

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// seedTask writes one task record.
func seedTask(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
}

// seedAgent writes one live agent record for the task fixture.
func seedAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

// newSeed makes a home with one task and its worker and verifier records.
func newSeed(t *testing.T, status string) (string, records.Task) {
	t.Helper()
	home := newHome(t)
	registerProject(t, home, "one")
	task := records.Task{
		ID:          "1-seed",
		Project:     "one",
		Title:       "Seed",
		ActingAgent: "worker-1",
		Status:      status,
		DependsOn:   []string{},
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-01T00:00:00Z",
	}
	seedTask(t, home, task)
	for i, name := range []string{"worker-1", "worker-2"} {
		seedAgent(t, home, records.Agent{
			Key: name, Name: name, Profile: records.ProfileBuilder, Project: "one", Task: task.ID,
			Status: records.AgentLive, PaneID: "w1:p" + string(rune('1'+i)), Generation: 1,
			StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		})
	}
	for i, name := range []string{"verifier-1", "verifier-2"} {
		seedAgent(t, home, records.Agent{
			Key: name, Name: name, Profile: records.ProfileVerifier, Project: "one", Task: task.ID,
			Status: records.AgentLive, PaneID: "w1:p" + string(rune('3'+i)), Generation: 1,
			StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		})
	}
	return home, task
}

func taskRecordBytes(t *testing.T, home, id string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "tasks", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readTask(t *testing.T, home, id string) records.Task {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, id, &task); err != nil {
		t.Fatal(err)
	}
	return task
}

type transitionCommand struct {
	key     string
	message string
	run     func(home, id string) output.Envelope
}

func transitionCommands() []transitionCommand {
	return []transitionCommand{
		{"assign", "assign", func(home, id string) output.Envelope {
			return Assign(home, TransitionOptions{Task: id, Agent: "worker-2"})
		}},
		{"complete", "complete", func(home, id string) output.Envelope {
			return Complete(home, TransitionOptions{Task: id, Agent: "worker-1"})
		}},
		{"verify", "verify", func(home, id string) output.Envelope {
			return Verify(home, TransitionOptions{Task: id, Agent: "verifier-1"})
		}},
		{"verify-fail", "verify", func(home, id string) output.Envelope {
			return Verify(home, TransitionOptions{Task: id, Agent: "verifier-1", Fail: true, Reason: "differs"})
		}},
		{"cancel", "cancel", func(home, id string) output.Envelope {
			return Cancel(home, TransitionOptions{Task: id})
		}},
	}
}

var validTransitions = map[string]map[string]string{
	"created":   {"assign": "assigned", "cancel": "cancelled"},
	"assigned":  {"assign": "assigned", "complete": "completed", "cancel": "cancelled"},
	"completed": {"verify": "verified", "verify-fail": "assigned", "cancel": "cancelled"},
	"verified":  {"cancel": "cancelled"},
	"cancelled": {},
}

func TestTransitionTable(t *testing.T) {
	statuses := []string{"created", "assigned", "completed", "verified", "cancelled"}
	valid, invalid := 0, 0
	for _, status := range statuses {
		for _, command := range transitionCommands() {
			t.Run(status+"/"+command.key, func(t *testing.T) {
				home, task := newSeed(t, status)
				before := taskRecordBytes(t, home, task.ID)
				envelope := command.run(home, task.ID)
				if want, ok := validTransitions[status][command.key]; ok {
					if envelope.Outcome != "ok" || envelope.Error != nil {
						t.Fatalf("%s on %s = %#v; want status %s", command.key, status, envelope, want)
					}
					got := readTask(t, home, task.ID)
					if got.Status != want {
						t.Fatalf("status = %q; want %q", got.Status, want)
					}
					return
				}
				if envelope.Outcome != "failed" || envelope.Error == nil {
					t.Fatalf("%s on %s = %#v; want a failure", command.key, status, envelope)
				}
				want := fmt.Sprintf("invalid transition: %s -> %s", status, command.message)
				message, _ := envelope.Error.(string)
				if !strings.HasPrefix(message, want) {
					t.Fatalf("error = %v; want the prefix %q", envelope.Error, want)
				}
				if (command.key == "complete" || strings.HasPrefix(command.key, "verify")) && status != "created" {
					for _, inspected := range []string{"task show", "agent show", "inbox"} {
						if !strings.Contains(message, inspected) {
							t.Fatalf("error = %q; want %q", message, inspected)
						}
					}
				} else if message != want {
					t.Fatalf("error = %v; want %q", envelope.Error, want)
				}
				if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
					t.Fatalf("record changed on an invalid transition:\nbefore %s\nafter  %s", before, after)
				}
			})
		}
	}
	for _, status := range statuses {
		for _, command := range transitionCommands() {
			if _, ok := validTransitions[status][command.key]; ok {
				valid++
			} else {
				invalid++
			}
		}
	}
	if valid != 9 || invalid != 16 {
		t.Fatalf("valid = %d, invalid = %d; want 9 and 16", valid, invalid)
	}
}

func TestTransitionMissingTask(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := Complete(home, TransitionOptions{Task: "9-missing", Agent: "worker-1"})
	if envelope.Outcome != "failed" || envelope.Error != "task not found: 9-missing" {
		t.Fatalf("Complete() = %#v; want a missing task error", envelope)
	}
}

func TestTransitionCompleteWrongAgent(t *testing.T) {
	home, task := newSeed(t, "assigned")
	before := taskRecordBytes(t, home, task.ID)
	envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-2"})
	if envelope.Outcome != "failed" || envelope.Error != "not the acting agent of 1-seed: worker-2" {
		t.Fatalf("Complete() = %#v; want a wrong agent error", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatalf("record changed on a wrong agent")
	}
}

func TestTransitionCompleteWithoutAgent(t *testing.T) {
	home, task := newSeed(t, "assigned")
	envelope := Complete(home, TransitionOptions{Task: task.ID})
	if envelope.Outcome != "failed" || envelope.Error != "caller unknown: pass --agent" {
		t.Fatalf("Complete() = %#v; want a caller error", envelope)
	}
}

func TestTransitionAssignWithoutAgent(t *testing.T) {
	home, task := newSeed(t, "created")
	envelope := Assign(home, TransitionOptions{Task: task.ID})
	if envelope.Outcome != "failed" || envelope.Error != "caller unknown: pass --agent" {
		t.Fatalf("Assign() = %#v; want a caller error", envelope)
	}
}

func TestTransitionAssignReplacesAgent(t *testing.T) {
	home, task := newSeed(t, "assigned")
	envelope := Assign(home, TransitionOptions{Task: task.ID, Agent: "worker-9"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Assign() = %#v; want ok", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "assigned" || got.ActingAgent != "worker-9" {
		t.Fatalf("record = %#v; want assigned with worker-9", got)
	}
}

func TestVerifySelf(t *testing.T) {
	home, task := newSeed(t, "completed")
	before := taskRecordBytes(t, home, task.ID)
	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "failed" || envelope.Error != "self verify refused: worker-1 is the acting agent of 1-seed" {
		t.Fatalf("Verify() = %#v; want a self-verification refusal", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatalf("record changed on a self-verification")
	}
}

func TestVerifyFailWithoutReason(t *testing.T) {
	home, task := newSeed(t, "completed")
	before := taskRecordBytes(t, home, task.ID)
	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true})
	if envelope.Outcome != "failed" || envelope.Error != "--fail needs --reason" {
		t.Fatalf("Verify() = %#v; want a reason error", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatalf("record changed on a failed verification without a reason")
	}
}

func TestVerifyReplacement(t *testing.T) {
	home, task := newSeed(t, "assigned")
	task.Verification = &records.Verification{
		Agent: "verifier-old", Result: "fail", Reason: "old", At: "2026-01-01T00:00:00Z",
	}
	seedTask(t, home, task)

	if envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-1"}); envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	failed := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true, Reason: "V3 output differs"})
	if failed.Outcome != "ok" {
		t.Fatalf("Verify(fail) = %#v; want ok", failed)
	}
	got := readTask(t, home, task.ID)
	if got.Status != "assigned" {
		t.Fatalf("status = %q; want assigned", got.Status)
	}
	if got.Verification == nil || got.Verification.Agent != "verifier-1" ||
		got.Verification.Result != "fail" || got.Verification.Reason != "V3 output differs" ||
		got.Verification.At == "" {
		t.Fatalf("verification = %#v; want the failed check of verifier-1", got.Verification)
	}

	if envelope := Assign(home, TransitionOptions{Task: task.ID, Agent: "worker-2"}); envelope.Outcome != "ok" {
		t.Fatalf("Assign() = %#v; want the fresh worker", envelope)
	}
	if envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-2"}); envelope.Outcome != "ok" {
		t.Fatalf("second Complete() = %#v; want the fresh worker", envelope)
	}
	passed := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-2"})
	if passed.Outcome != "ok" {
		t.Fatalf("Verify() = %#v; want ok", passed)
	}
	got = readTask(t, home, task.ID)
	if got.Status != "verified" {
		t.Fatalf("status = %q; want verified", got.Status)
	}
	if got.Verification == nil || got.Verification.Agent != "verifier-2" ||
		got.Verification.Result != "pass" || got.Verification.Reason != "" {
		t.Fatalf("verification = %#v; want the pass of verifier-2 replacing the old one", got.Verification)
	}

	before := taskRecordBytes(t, home, task.ID)
	again := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-2"})
	if again.Outcome != "failed" {
		t.Fatalf("Verify() = %#v; want a refused repeated finish", again)
	}
	message, _ := again.Error.(string)
	for _, want := range []string{"task show", "agent show", "inbox"} {
		if !strings.Contains(message, want) {
			t.Fatalf("Verify() error = %q; want %q", message, want)
		}
	}
	if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatalf("record changed on the second verification of a verified task")
	}
}

func TestTransitionCancelStoresReason(t *testing.T) {
	home, task := newSeed(t, "assigned")
	envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "user stopped the work"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	got := readTask(t, home, task.ID)
	if got.Status != "cancelled" || got.CancelReason != "user stopped the work" {
		t.Fatalf("record = %#v; want cancelled with the reason", got)
	}
}

func TestTransitionCancelVerifiedTask(t *testing.T) {
	t.Run("without delivery", func(t *testing.T) {
		home, task := newSeed(t, "verified")
		envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "stop work"})
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Cancel() = %#v; want ok", envelope)
		}
		got := readTask(t, home, task.ID)
		if got.Status != "cancelled" || got.CancelReason != "stop work" {
			t.Fatalf("record = %#v; want cancelled with the reason", got)
		}
	})

	t.Run("with an unpushed delivery", func(t *testing.T) {
		home, task := newSeed(t, "verified")
		task.Delivery = &records.Delivery{Mode: "merge", Target: "main", Commit: "abc123"}
		seedTask(t, home, task)
		envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "stop work"})
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("Cancel() = %#v; want ok", envelope)
		}
		got := readTask(t, home, task.ID)
		if got.Status != "cancelled" || got.CancelReason != "stop work" {
			t.Fatalf("record = %#v; want cancelled with the reason", got)
		}
	})

	t.Run("with a pushed delivery", func(t *testing.T) {
		home, task := newSeed(t, "verified")
		task.Delivery = &records.Delivery{Mode: "merge", Target: "main", Commit: "abc123", Pushed: true}
		seedTask(t, home, task)
		before := taskRecordBytes(t, home, task.ID)
		envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "stop work"})
		if envelope.Outcome != "failed" || envelope.Error != "task is landed: cancel refused" {
			t.Fatalf("Cancel() = %#v; want landed-task refusal", envelope)
		}
		if after := taskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
			t.Fatalf("record changed when cancel refused")
		}
	})
}

func TestTransitionRecordInvalidID(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := Complete(home, TransitionOptions{Task: "not-a-task", Agent: "worker-1"})
	if envelope.Outcome != "failed" || envelope.Error != "task not found: not-a-task" {
		t.Fatalf("Complete() = %#v; want a plain not found error", envelope)
	}
}
