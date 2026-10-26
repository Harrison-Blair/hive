package reopen

import (
	"bytes"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
	"github.com/Harrison-Blair/hive/internal/task/complete"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
	"github.com/Harrison-Blair/hive/internal/task/verify"
)

func seedController(t *testing.T, home, project string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	agent := records.Agent{
		Key: "manager-" + project, Name: "manager-" + project, Profile: records.ProfilePlanner,
		Owns: project, Project: project,
		Status: records.AgentLive, PaneID: "manager-pane", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"agents"}, agent.Key, agent); err != nil {
		t.Fatal(err)
	}
	claim := records.Claim{Key: "owner-" + project, Agent: agent.Key, Generation: agent.Generation}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

func verifiedTask(task records.Task, agent, at string) records.Task {
	task.Status = "verified"
	task.Verification = &records.Verification{Agent: agent, Result: "pass", At: at}
	return task
}

func TestReopenVerifiedByProjectOwner(t *testing.T) {
	home, task := tasktest.NewSeed(t, "verified")
	task = verifiedTask(task, "verifier-old", "2026-01-02T00:00:00Z")
	tasktest.SeedTask(t, home, task)
	seedController(t, home, task.Project)

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "manager-one", Reason: "merge conflict"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Reopen() = %#v; want ok", envelope)
	}
	got := tasktest.ReadTask(t, home, task.ID)
	if got.Status != "assigned" || got.ActingAgent != task.ActingAgent || got.Verification != nil {
		t.Fatalf("task = %#v; want assigned, same acting agent, and no current verification", got)
	}
	if len(got.VerificationHistory) != 1 {
		t.Fatalf("verification history = %#v; want one item", got.VerificationHistory)
	}
	item := got.VerificationHistory[0]
	if item.Verification == nil || *item.Verification != *task.Verification || item.Reason != "merge conflict" ||
		item.Caller != "manager-one" {
		t.Fatalf("history item = %#v; want the old check, reason, and caller", item)
	}
	if _, err := time.Parse(time.RFC3339Nano, item.At); err != nil {
		t.Fatalf("history time = %q; want RFC3339Nano: %v", item.At, err)
	}
}

func TestReopenByWorkerFails(t *testing.T) {
	home, task := tasktest.NewSeed(t, "verified")
	task = verifiedTask(task, "verifier-old", "2026-01-02T00:00:00Z")
	tasktest.SeedTask(t, home, task)
	seedController(t, home, task.Project)
	before := tasktest.TaskRecordBytes(t, home, task.ID)

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "worker-1", Reason: "repair"})
	if envelope.Outcome != "failed" || envelope.Error != "not the owner of project one or the coordinator: worker-1" {
		t.Fatalf("Reopen() = %#v; want owner refusal", envelope)
	}
	if after := tasktest.TaskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatal("task record changed after a worker reopen request")
	}
}

func TestReopenNonVerifiedStatusesFail(t *testing.T) {
	for _, status := range []string{"created", "assigned", "completed", "cancelled"} {
		t.Run(status, func(t *testing.T) {
			home, task := tasktest.NewSeed(t, status)
			seedController(t, home, task.Project)
			before := tasktest.TaskRecordBytes(t, home, task.ID)

			envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "manager-one", Reason: "repair"})
			if envelope.Outcome != "failed" || envelope.Error != "invalid transition: "+status+" -> reopen" {
				t.Fatalf("Reopen() = %#v; want invalid transition", envelope)
			}
			if after := tasktest.TaskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
				t.Fatal("task record changed after an invalid transition")
			}
		})
	}
}

func TestReopenRequiresReason(t *testing.T) {
	home, task := tasktest.NewSeed(t, "verified")
	task = verifiedTask(task, "verifier-old", "2026-01-02T00:00:00Z")
	tasktest.SeedTask(t, home, task)
	seedController(t, home, task.Project)
	before := tasktest.TaskRecordBytes(t, home, task.ID)

	envelope := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "manager-one"})
	if envelope.Outcome != "failed" || envelope.Error != "--reason is required" {
		t.Fatalf("Reopen() = %#v; want required reason error", envelope)
	}
	if after := tasktest.TaskRecordBytes(t, home, task.ID); !bytes.Equal(before, after) {
		t.Fatal("task record changed after a reopen request without a reason")
	}
}

func TestReopenKeepsVerificationHistoryAcrossReopens(t *testing.T) {
	home, task := tasktest.NewSeed(t, "verified")
	task = verifiedTask(task, "verifier-old", "2026-01-02T00:00:00Z")
	tasktest.SeedTask(t, home, task)
	seedController(t, home, task.Project)

	first := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "manager-one", Reason: "first repair"})
	if first.Outcome != "ok" {
		t.Fatalf("first Reopen() = %#v; want ok", first)
	}
	if envelope := complete.Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "worker-1"}); envelope.Outcome != "ok" {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	if envelope := verify.Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "verifier-1"}); envelope.Outcome != "ok" {
		t.Fatalf("Verify() = %#v; want ok", envelope)
	}
	secondVerification := tasktest.ReadTask(t, home, task.ID).Verification
	second := Run(home, tasklib.TransitionOptions{Task: task.ID, Agent: "manager-one", Reason: "second repair"})
	if second.Outcome != "ok" {
		t.Fatalf("second Reopen() = %#v; want ok", second)
	}

	got := tasktest.ReadTask(t, home, task.ID)
	if got.Verification != nil || len(got.VerificationHistory) != 2 {
		t.Fatalf("task = %#v; want no current check and two history items", got)
	}
	want := []struct {
		verification *records.Verification
		reason       string
	}{
		{task.Verification, "first repair"},
		{secondVerification, "second repair"},
	}
	for index, expected := range want {
		item := got.VerificationHistory[index]
		if item.Verification == nil || *item.Verification != *expected.verification || item.Reason != expected.reason || item.Caller != "manager-one" {
			t.Fatalf("history[%d] = %#v; want verification, reason %q, and caller", index, item, expected.reason)
		}
		if _, err := time.Parse(time.RFC3339Nano, item.At); err != nil {
			t.Fatalf("history[%d] time = %q; want RFC3339Nano: %v", index, item.At, err)
		}
	}
}
