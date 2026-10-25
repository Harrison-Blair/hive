package inbox

import (
	"strings"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// liveOwner returns one live owner record with the facts of D1 and D7.
func liveOwner(key, name, project, pane string) records.Agent {
	return records.Agent{
		Key: key, Name: name, Role: "manager", Profile: records.ProfilePlanner,
		Owns: project, Project: project, PaneID: pane,
		Status: records.AgentLive, Generation: 1,
	}
}

// TestSendProjectAddressWritesToLiveOwner proves AC1: project:ID goes to the
// inbox of the live owner.
func TestSendProjectAddressWritesToLiveOwner(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveOwner("owner-demo-g1", "owner-demo", "demo", "w1:p1"),
		&records.Claim{Key: "owner-demo", Agent: "owner-demo-g1", Generation: 1})

	envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "project:demo", Kind: records.InboxMessage, Agent: "builder-1", Text: "hello"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v; want ok", envelope)
	}
	record, ok := envelope.Result.(records.Inbox)
	if !ok || record.Role != "owner-demo" || record.Text != "hello" {
		t.Fatalf("Append() result = %#v; want the record in inbox owner-demo", envelope.Result)
	}
	found, err := inboxlib.All(home, "owner-demo")
	if err != nil || len(found) != 1 || found[0].Text != "hello" {
		t.Fatalf("All(owner-demo) = %#v, %v; want the appended record", found, err)
	}
	if effect := envelope.Effects[len(envelope.Effects)-1]; effect.Path != ".hive/state/inbox/owner-demo/1.json" {
		t.Fatalf("effect = %#v; want the owner inbox path", effect)
	}
}

// TestSendProjectAddressUsesCoordinatorWithoutLiveOwner proves AC1: a project
// with no live owner sends to the coordinator.
func TestSendProjectAddressUsesCoordinatorWithoutLiveOwner(t *testing.T) {
	home := t.TempDir()
	stopped := liveOwner("owner-demo-g1", "owner-demo", "demo", "w1:p1")
	stopped.Status = records.AgentStopped
	seedInboxAgent(t, home, stopped, nil)

	envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "project:demo", Kind: records.InboxMessage, Text: "hello"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v; want ok", envelope)
	}
	record, ok := envelope.Result.(records.Inbox)
	if !ok || record.Role != "coordinator" {
		t.Fatalf("Append() result = %#v; want the coordinator inbox", envelope.Result)
	}
	found, err := inboxlib.All(home, "coordinator")
	if err != nil || len(found) != 1 || found[0].Text != "hello" {
		t.Fatalf("All(coordinator) = %#v, %v; want the appended record", found, err)
	}
}

// TestSendRefusesInvalidProjectAddress shows an invalid address is refused
// before any write.
func TestSendRefusesInvalidProjectAddress(t *testing.T) {
	home := t.TempDir()
	envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "project:Bad ID", Kind: records.InboxMessage, Text: "hello"})
	if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), "invalid inbox key:") {
		t.Fatalf("Append() = %#v; want an invalid inbox key", envelope)
	}
}
