package doctor

import (
	"context"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/prune"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func appendDoctorInbox(t *testing.T, home, key, text string) {
	t.Helper()
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: key, Kind: records.InboxMessage, Text: text}); envelope.Outcome != "ok" {
		t.Fatalf("Append(%s) = %#v", key, envelope)
	}
}

func TestDoctorOwnerlessInboxClearsAfterPrune(t *testing.T) {
	home := newHome(t)
	seedDoctorAgent(t, home, records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, Status: records.AgentLive})
	seedDoctorAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Status: records.AgentStopped})
	appendDoctorInbox(t, home, "worker-old", "unread")
	if check := checkOwnerlessInboxes(home); check.Status != StatusWarn {
		t.Fatalf("check before prune = %+v; want warn", check)
	}

	if _, err := prune.Run(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	check := checkOwnerlessInboxes(home)
	if check.Status != StatusOK || check.Detail != "0" {
		t.Fatalf("check after prune = %+v; want ok with 0", check)
	}
}

func TestDoctorListsUnreadRecordsWithoutLiveOwner(t *testing.T) {
	home := newHome(t)
	check := checkOwnerlessInboxes(home)
	if check.Name != "ownerless inbox records" || check.Status != StatusOK || check.Detail != "0" {
		t.Fatalf("check = %+v; want ok with 0", check)
	}

	seedDoctorAgent(t, home, records.Agent{Key: "manager-hive-g1", Name: "manager-hive", Role: "manager", Project: "hive", Status: records.AgentLive})
	seedDoctorAgent(t, home, records.Agent{Key: "t1-worker-1", Name: "t1-worker-1", Role: "worker", Project: "hive", Status: records.AgentStopped})
	appendDoctorInbox(t, home, "manager-hive", "owned")
	appendDoctorInbox(t, home, "t1-worker-1", "left one")
	appendDoctorInbox(t, home, "t1-worker-1", "left two")
	appendDoctorInbox(t, home, "coordinator", "no coordinator")
	read := inboxlib.Consume(home, "t1-worker-1")
	if read.Outcome != "ok" {
		t.Fatalf("Consume() = %#v", read)
	}
	appendDoctorInbox(t, home, "t1-worker-1", "unread")

	check = checkOwnerlessInboxes(home)
	if check.Status != StatusWarn || check.Detail != "2: coordinator/1, t1-worker-1/3" {
		t.Fatalf("check = %+v; want warn with the two unread records", check)
	}
	if line := Render(Result{Checks: []Check{check}}); line != "warn ownerless inbox records: 2: coordinator/1, t1-worker-1/3\n" {
		t.Fatalf("Render() = %q", line)
	}
}
