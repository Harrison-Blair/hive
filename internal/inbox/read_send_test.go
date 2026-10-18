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
	return records.Agent{Key: key, Name: name, Role: role, Project: project, Status: records.AgentLive, PaneID: pane, Generation: 1}
}

func TestReadDefaultsRoleFromManagerCaller(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("manager-one-g1", "manager-one", "manager", "one", "w1:p1"), &records.Claim{Key: "manager-one", Agent: "manager-one-g1", Generation: 1})
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "manager-one", Kind: records.InboxMessage, Text: "hello"}); envelope.Outcome != "ok" {
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

func TestReadRejectsWorkerWithoutInbox(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("worker-1", "worker-1", "worker", "one", "w1:p1"), nil)
	envelope := Read(home, ReadOptions{PaneID: "w1:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "role has no inbox: worker" {
		t.Fatalf("Read() = %#v; want role has no inbox", envelope)
	}
}

func TestReadUnknownCallerAsksForRole(t *testing.T) {
	envelope := Read(t.TempDir(), ReadOptions{})
	if envelope.Outcome != "failed" {
		t.Fatalf("Read() = %#v; want failed", envelope)
	}
	message, _ := envelope.Error.(string)
	if message != "caller unknown: pass --role or pass --agent" {
		t.Fatalf("error = %q; want the role hint", message)
	}
}

func TestReadExplicitRoleKeepsWorkingWithoutCaller(t *testing.T) {
	home := t.TempDir()
	if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	envelope := Read(home, ReadOptions{Role: "coordinator", RoleSet: true})
	if envelope.Outcome != "ok" {
		t.Fatalf("Read(explicit) = %#v; want ok", envelope)
	}
}

func TestSendDefaultsAgentAndKeepsExplicitSender(t *testing.T) {
	home := t.TempDir()
	seedInboxAgent(t, home, liveInboxAgent("worker-1", "worker-1", "worker", "one", "w1:p1"), nil)
	envelope := Send(home, inboxlib.AppendOptions{Role: "manager-one", Kind: records.InboxMessage, Text: "default"}, "w1:p1")
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok", envelope)
	}
	found, err := inboxlib.All(home, "manager-one")
	if err != nil || len(found) != 1 || found[0].Agent != "worker-1" {
		t.Fatalf("records = %#v, %v; want the caller as sender", found, err)
	}
	explicit := Send(home, inboxlib.AppendOptions{Role: "manager-one", Kind: records.InboxMessage, Agent: "sender-x", Text: "explicit"}, "")
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
	envelope := Send(home, inboxlib.AppendOptions{Role: "manager-one", Kind: records.InboxMessage, Text: "hello"}, "")
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
		{"claims", "manager-one", "wrong-claim"},
	} {
		t.Run(mismatch.collection, func(t *testing.T) {
			home := t.TempDir()
			seedInboxAgent(t, home, liveInboxAgent("manager-one-g1", "manager-one", "manager", "one", "w1:p1"), &records.Claim{Key: "manager-one", Agent: "manager-one-g1", Generation: 1})
			if envelope := inboxlib.Append(home, inboxlib.AppendOptions{Role: "manager-one", Kind: records.InboxMessage, Text: "unread"}); envelope.Outcome != "ok" {
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
	envelope := Send(home, inboxlib.AppendOptions{Role: "../x", Kind: "message", Agent: "explicit", Text: "body"}, "")
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
	envelope = Read(unsupported, ReadOptions{PaneID: "w1:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "role has no inbox: worker" {
		t.Fatalf("Read(worker) = %#v; want role has no inbox", envelope)
	}
	events = inboxLogEvents(t, unsupported)
	if len(events) != 1 || events[0]["msg"] != "inbox operation failed" {
		t.Fatalf("log events = %#v; want one inbox error", events)
	}
	if message, _ := events[0]["error"].(string); !strings.Contains(message, "role has no inbox: worker") {
		t.Fatalf("error event = %#v; want the role refusal", events[0])
	}
}
