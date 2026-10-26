package task

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestCompleteRefusesCallerThatIsNotTheActingAgent proves the D8 rule for a
// caller with the task fact that is not the acting agent.
func TestCompleteRefusesCallerThatIsNotTheActingAgent(t *testing.T) {
	home, task := newSeed(t, "assigned")
	before := taskRecordBytes(t, home, task.ID)
	envelope := RunFinish(home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-2"})
	if envelope.Outcome != "failed" || envelope.Error != "not the acting agent of 1-seed: worker-2" {
		t.Fatalf("Complete() = %#v; want the exact D8 refusal", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("the refusal changed the task record")
	}
}

// TestCompleteRefusesActingNameWithoutTheTaskFact proves the second half of
// the D8 complete rule: the caller's task fact must be the task.
func TestCompleteRefusesActingNameWithoutTheTaskFact(t *testing.T) {
	home, task := newSeed(t, "assigned")
	seedAgent(t, home, records.Agent{
		Key: "worker-1", Name: "worker-1", Profile: records.ProfileBuilder, Project: "one", Task: "2-other",
		Status: records.AgentLive, PaneID: "w1:p1", Generation: 1,
	})
	before := taskRecordBytes(t, home, task.ID)
	envelope := RunFinish(home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "failed" || envelope.Error != "not the acting agent of 1-seed: worker-1" {
		t.Fatalf("Complete() = %#v; want the exact D8 refusal", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("the refusal changed the task record")
	}
}

// TestVerifyAllowsAnotherBoundAgent proves AC2: an agent with the same task
// fact that is not the acting agent can verify the task.
func TestVerifyAllowsAnotherBoundAgent(t *testing.T) {
	home, task := newSeed(t, "completed")
	envelope := RunFinish(home, "verify", TransitionOptions{Task: task.ID, Agent: "worker-2"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Verify() = %#v; want ok for another bound agent", envelope)
	}
	got := readTask(t, home, task.ID)
	if got.Status != "verified" || got.Verification == nil || got.Verification.Agent != "worker-2" {
		t.Fatalf("task = %#v; want verified by worker-2", got)
	}
}

// TestVerifyRefusesCallerNotBoundToTheTask proves the D8 verify rule and its
// exact refusal text for a caller without the task fact.
func TestVerifyRefusesCallerNotBoundToTheTask(t *testing.T) {
	home, task := newSeed(t, "completed")
	seedAgent(t, home, records.Agent{
		Key: "verifier-3", Name: "verifier-3", Profile: records.ProfileVerifier, Project: "one", Task: "2-other",
		Status: records.AgentLive, PaneID: "w1:p3", Generation: 1,
	})
	before := taskRecordBytes(t, home, task.ID)
	envelope := RunFinish(home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-3"})
	if envelope.Outcome != "failed" || envelope.Error != "not bound to 1-seed: verifier-3" {
		t.Fatalf("Verify() = %#v; want the exact D8 refusal", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("the refusal changed the task record")
	}
}

// TestVerifyRefusesTheActingAgent proves the self-verify refusal text of D8.
func TestVerifyRefusesTheActingAgent(t *testing.T) {
	home, task := newSeed(t, "completed")
	before := taskRecordBytes(t, home, task.ID)
	envelope := RunFinish(home, "verify", TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "failed" || envelope.Error != "self verify refused: worker-1 is the acting agent of 1-seed" {
		t.Fatalf("Verify() = %#v; want the exact self-verify refusal", envelope)
	}
	if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
		t.Fatal("the refusal changed the task record")
	}
	if found := finishInbox(t, home, "coordinator"); len(found) != 0 {
		t.Fatalf("inbox = %#v; want no verified record", found)
	}
}

// TestVerifyBoundBuilderCannotCompleteAfterItsVerify keeps the D8 facts
// meaningful: a builder that verifies is finished and cannot then complete.
func TestVerifyBoundBuilderCannotCompleteAfterItsVerify(t *testing.T) {
	home, task := newSeed(t, "completed")
	if envelope := RunFinish(home, "verify", TransitionOptions{Task: task.ID, Agent: "worker-2"}); envelope.Outcome != "ok" {
		t.Fatalf("Verify() = %#v; want ok", envelope)
	}
	refused := RunFinish(home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-2"})
	if refused.Outcome != "failed" || !strings.Contains(refused.Error.(string), "agent finished: worker-2") {
		t.Fatalf("Complete() = %#v; want the finished-agent refusal", refused)
	}
}
