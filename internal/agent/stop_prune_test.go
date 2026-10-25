package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestStopPrunesUnreferencedAgentAndInbox(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := spawnedRecord(sleep)
	record.Key, record.Name = "worker-1", "worker-1"
	record.Kind, record.Profile, record.Role = "", records.ProfileBuilder, "worker"
	record.StartedByHive = true
	writeStopRecord(t, home, record)
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: record.Name, Agent: "manager", Text: "unread"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	fake := newFakeHerdr(t)
	scriptStop(t, fake, sleep, "", `{"error":{"code":"workspace_not_found","message":"workspace w1 not found"}}`)

	if _, envelope := Stop(context.Background(), home, record); envelope.Outcome != "ok" {
		t.Fatalf("Stop() = %#v; want ok", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", record.Key+".json")); !os.IsNotExist(err) {
		t.Fatalf("agent record still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "inbox", record.Name)); !os.IsNotExist(err) {
		t.Fatalf("inbox folder still exists: %v", err)
	}
	found, err := inboxlib.All(home, records.AgentKindCoordinator)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Text != "unread" {
		t.Fatalf("coordinator inbox = %#v; want the moved record", found)
	}
}
