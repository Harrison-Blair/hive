package send

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func seedInboxAgent(t *testing.T, home string, record records.Agent, claim *records.Claim) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
	if claim != nil {
		if err := st.WriteIn([]string{"claims"}, claim.Key, *claim); err != nil {
			t.Fatal(err)
		}
	}
}

// liveWorker returns one live builder record.
func liveWorker(key, name, project, pane string) records.Agent {
	return records.Agent{Key: key, Name: name, Project: project, Status: records.AgentLive, PaneID: pane, Generation: 1, Profile: records.ProfileBuilder}
}

func TestSendDefaultsAgentAndKeepsExplicitSender(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveWorker("worker-1", "worker-1", "one", "w1:p1"), nil)
	envelope := Run(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "default"}, "w1:p1", "")
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok", envelope)
	}
	found, err := inboxlib.All(home, "manager-one")
	if err != nil || len(found) != 1 || found[0].Agent != "worker-1" {
		t.Fatalf("records = %#v, %v; want the caller as sender", found, err)
	}
	explicit := Run(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Agent: "sender-x", Text: "explicit"}, "", "")
	if explicit.Outcome != "ok" {
		t.Fatalf("Send(explicit) = %#v; want ok", explicit)
	}
	found, err = inboxlib.All(home, "manager-one")
	if err != nil || len(found) != 2 || found[1].Agent != "sender-x" {
		t.Fatalf("records = %#v, %v; want the explicit sender", found, err)
	}
}

func TestSendUnknownCallerWithoutAgentFails(t *testing.T) {
	home := t.TempDir()
	envelope := Run(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "hello"}, "", "")
	if envelope.Outcome != "failed" || envelope.Error != "caller unknown: pass --agent" {
		t.Fatalf("Send() = %#v; want caller unknown", envelope)
	}
	if found, err := inboxlib.All(home, "manager-one"); err != nil || len(found) != 0 {
		t.Fatalf("records = %#v, %v; want none", found, err)
	}
}

func TestSendInvalidExplicitDoesNotCreateState(t *testing.T) {
	home := t.TempDir()
	envelope := Run(home, inboxlib.AppendOptions{Inbox: "../x", Kind: "message", Agent: "explicit", Text: "body"}, "", "")
	if envelope.Outcome != "failed" || envelope.Phase != "validate" || envelope.Error != "invalid inbox key: ../x" {
		t.Fatalf("Send() = %#v; want failed validate", envelope)
	}
	if _, err := os.Lstat(filepath.Join(home, ".hive", "state")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("explicit invalid send created state, err=%v", err)
	}
}
