package inbox

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func liveInboxAgent(key, name, role, project, pane string) records.Agent {
	record := records.Agent{Key: key, Name: name, Project: project, Status: records.AgentLive, PaneID: pane, Generation: 1}
	switch role {
	case "coordinator":
		record.Kind = records.AgentKindCoordinator
		record.Profile = records.ProfileCoordinator
	case "manager":
		record.Profile = records.ProfilePlanner
		record.Owns = project
	case "worker":
		record.Profile = records.ProfileBuilder
	case "verifier":
		record.Profile = records.ProfileVerifier
	}
	return record
}

func TestReadDefaultsRoleFromManagerCaller(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("manager-one-g1", "manager-one", "manager", "one", "w1:p1"), &records.Claim{Key: "owner-one", Agent: "manager-one-g1", Generation: 1})
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{PaneID: "w1:p1"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Read() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(inboxlib.Result)
	if !ok || len(result.Records) != 1 || result.Records[0].Text != "hello" || result.Records[0].ReadAt == "" {
		t.Fatalf("Read() = %#v; want one consumed manager record", envelope.Result)
	}
}

func TestReadDefaultsWorkerToNameInbox(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("worker-1", "worker-1", "worker", "one", "w1:p1"), nil)
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "worker-1", Kind: records.InboxMessage, Text: "for worker"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{PaneID: "w1:p1"})
	result, ok := envelope.Result.(inboxlib.Result)
	if envelope.Outcome != "ok" || !ok || len(result.Records) != 1 || result.Records[0].Text != "for worker" || result.Records[0].ReadAt == "" {
		t.Fatalf("Read() = %#v; want one consumed record of inbox worker-1", envelope)
	}
}

func TestReadUnknownCallerAsksForAgentOrPane(t *testing.T) {
	envelope := Read(t.TempDir(), ReadOptions{})
	if envelope.Outcome != "failed" {
		t.Fatalf("Read() = %#v; want failed", envelope)
	}
	if envelope.Error != "caller unknown: pass --agent or run inside herdr" {
		t.Fatalf("error = %q; want the caller hint", envelope.Error)
	}
}

func TestReadExplicitRoleKeepsWorkingForOwner(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("manager-one-g1", "manager-one", "manager", "one", "w1:p1"), &records.Claim{Key: "owner-one", Agent: "manager-one-g1", Generation: 1})
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Inbox: "manager-one", InboxSet: true, Agent: "manager-one"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Read(explicit) = %#v; want ok", envelope)
	}
}

func TestReadExplicitRoleRejectsUnknownCallerWithoutConsuming(t *testing.T) {
	home := t.TempDir()
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-hive", Kind: records.InboxMessage, Text: "unread"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Inbox: "manager-hive", InboxSet: true})
	if envelope.Outcome != "failed" || envelope.Error != "caller unknown: pass --agent or run inside herdr" {
		t.Fatalf("Read() = %#v; want an unknown-caller refusal", envelope)
	}
	found, err := inboxlib.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All() = %#v, %v; want one unread record", found, err)
	}
}

func TestCoordinatorReadManagerInboxPrintsWithoutConsuming(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("coordinator-g1", "coordinator", "coordinator", "", "w0:p1"), &records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-hive", Kind: records.InboxMessage, Text: "manager message"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Inbox: "manager-hive", InboxSet: true, PaneID: "w0:p1"})
	if envelope.Outcome != "ok" || !strings.Contains(ReadText(envelope), "manager message") {
		t.Fatalf("Read() = %#v, text %q; want the manager message", envelope, ReadText(envelope))
	}
	found, err := inboxlib.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All() = %#v, %v; want one unread record", found, err)
	}
}

func TestVerifierCannotReadManagerInbox(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("verifier-1", "verifier-1", "verifier", "hive", "w2:p1"), nil)
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-hive", Kind: records.InboxMessage, Text: "private"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Inbox: "manager-hive", InboxSet: true, PaneID: "w2:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "inbox manager-hive belongs to the agent manager-hive" {
		t.Fatalf("Read() = %#v; want the manager ownership error", envelope)
	}
	found, err := inboxlib.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All() = %#v, %v; want one unread record", found, err)
	}
}

func TestVerifierCannotReadAllManagerInbox(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("verifier-1", "verifier-1", "verifier", "hive", "w2:p1"), nil)
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-hive", Kind: records.InboxMessage, Text: "private"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Inbox: "manager-hive", InboxSet: true, All: true, PaneID: "w2:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "inbox manager-hive belongs to the agent manager-hive" {
		t.Fatalf("Read(--all) = %#v; want the manager ownership error", envelope)
	}
	found, err := inboxlib.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All() = %#v, %v; want one unread record", found, err)
	}
}

func TestSendDefaultsAgentAndKeepsExplicitSender(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("worker-1", "worker-1", "worker", "one", "w1:p1"), nil)
	envelope := Send(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "default"}, "w1:p1", "")
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok", envelope)
	}
	found, err := inboxlib.All(home, "manager-one")
	if err != nil || len(found) != 1 || found[0].Agent != "worker-1" {
		t.Fatalf("records = %#v, %v; want the caller as sender", found, err)
	}
	explicit := Send(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Agent: "sender-x", Text: "explicit"}, "", "")
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
	envelope := Send(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "hello"}, "", "")
	if envelope.Outcome != "failed" || envelope.Error != "caller unknown: pass --agent" {
		t.Fatalf("Send() = %#v; want caller unknown", envelope)
	}
	if found, err := inboxlib.All(home, "manager-one"); err != nil || len(found) != 0 {
		t.Fatalf("records = %#v, %v; want none", found, err)
	}
}

func TestReadRejectsPathKeyMismatchWithoutConsume(t *testing.T) {
	for _, mismatch := range []struct{ collection, oldName, newName string }{
		{"agents", "manager-one-g1", "wrong-holder"},
		{"claims", "owner-one", "wrong-claim"},
	} {
		t.Run(mismatch.collection, func(t *testing.T) {
			home := t.TempDir()
			seedInboxAgent(t, home, liveInboxAgent("manager-one-g1", "manager-one", "manager", "one", "w1:p1"), &records.Claim{Key: "owner-one", Agent: "manager-one-g1", Generation: 1})
			if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Inbox: "manager-one", Kind: records.InboxMessage, Text: "unread"}); envelope.Outcome != "ok" {
				t.Fatalf("Append() = %#v", envelope)
			}
			path := filepath.Join(home, ".hive", "state", mismatch.collection, mismatch.oldName+".json")
			if err := os.Rename(path, filepath.Join(filepath.Dir(path), mismatch.newName+".json")); err != nil {
				t.Fatal(err)
			}
			envelope := Read(home, ReadOptions{PaneID: "w1:p1"})
			if envelope.Outcome == "ok" {
				t.Fatalf("Read() = %#v; want a refusal", envelope)
			}
			found, err := inboxlib.All(home, "manager-one")
			if err != nil || len(found) != 1 {
				t.Fatalf("records = %#v, %v; want the unread record", found, err)
			}
			if found[0].ReadAt != "" {
				t.Fatalf("read_at = %q; want no consume effect", found[0].ReadAt)
			}
		})
	}
}

func TestSendInvalidExplicitDoesNotCreateState(t *testing.T) {
	home := t.TempDir()
	envelope := Send(home, inboxlib.AppendOptions{Inbox: "../x", Kind: "message", Agent: "explicit", Text: "body"}, "", "")
	if envelope.Outcome != "failed" || envelope.Phase != "validate" || envelope.Error != "invalid inbox key: ../x" {
		t.Fatalf("Send() = %#v; want failed validate", envelope)
	}
	if _, err := os.Lstat(filepath.Join(home, ".hive", "state")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("explicit invalid send created state, err=%v", err)
	}
}

func inboxLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	events := []map[string]any{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line is not JSON: %v; line %q", err, line)
			}
			if event["component"] == "inbox" {
				events = append(events, event)
			}
		}
	}
	return events
}

func TestReadRefusalsLogOneErrorEvent(t *testing.T) {
	unknown := t.TempDir()
	envelope := Read(unknown, ReadOptions{})
	if envelope.Outcome != "failed" {
		t.Fatalf("Read(unknown) = %#v; want failed", envelope)
	}
	events := inboxLogEvents(t, unknown)
	if len(events) != 1 || events[0]["msg"] != "inbox operation failed" {
		t.Fatalf("log events = %#v; want one inbox error", events)
	}
	if message, _ := events[0]["error"].(string); !strings.Contains(message, "caller unknown") {
		t.Fatalf("error event = %#v; want the caller refusal", events[0])
	}
	if _, ok := events[0]["role"]; ok {
		t.Fatalf("error event = %#v; want no fabricated role", events[0])
	}

	unsupported := t.TempDir()
	seedInboxAgent(t, unsupported, liveInboxAgent("worker-1", "worker-1", "worker", "one", "w1:p1"), nil)
	envelope = Read(unsupported, ReadOptions{Inbox: "manager-one", InboxSet: true, PaneID: "w1:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "inbox manager-one belongs to the agent manager-one" {
		t.Fatalf("Read(worker) = %#v; want the ownership refusal", envelope)
	}
	events = inboxLogEvents(t, unsupported)
	if len(events) != 1 || events[0]["msg"] != "inbox operation failed" {
		t.Fatalf("log events = %#v; want one inbox error", events)
	}
	if message, _ := events[0]["error"].(string); !strings.Contains(message, "belongs to the agent manager-one") {
		t.Fatalf("error event = %#v; want the ownership refusal", events[0])
	}
}
