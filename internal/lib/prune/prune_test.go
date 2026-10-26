package prune

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const testTime = "2026-01-02T03:04:05Z"

func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("# home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Open(home); err != nil {
		t.Fatal(err)
	}
	return home
}

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

func seedInbox(t *testing.T, home, key string, items ...records.Inbox) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		for _, item := range items {
			seq, err := locked.Next("inbox:" + key)
			if err != nil {
				return err
			}
			item.Seq = seq
			item.Inbox = key
			if item.Kind == "" {
				item.Kind = records.InboxMessage
			}
			if item.CreatedAt == "" {
				item.CreatedAt = testTime
			}
			if err := locked.WriteIn([]string{"inbox", key}, itoa(seq), item); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func itoa(value uint64) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	text := ""
	for value > 0 {
		text = string(digits[value%10]) + text
		value /= 10
	}
	return text
}

func agentExists(t *testing.T, home, key string) bool {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	err = registry.ReadIn([]string{"agents"}, key, &record)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.Fatalf("read agent %s: %v", key, err)
	}
	return true
}

func readInbox(t *testing.T, home, key string) []records.Inbox {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var found []records.Inbox
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		raw, err := locked.ListIn([]string{"inbox", key}, nil)
		if err != nil {
			return err
		}
		for _, data := range raw {
			var item records.Inbox
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}
			found = append(found, item)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func inboxExists(t *testing.T, home, key string) bool {
	t.Helper()
	info, err := os.Lstat(filepath.Join(home, ".hive", "state", "inbox", key))
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.Fatal(err)
	}
	return info.IsDir()
}

func readLogEvents(t *testing.T, home string) []map[string]any {
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
		for _, line := range splitLines(string(data)) {
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line %q: %v", line, err)
			}
			events = append(events, event)
		}
	}
	return events
}

func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, r := range text {
		if r == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
			continue
		}
		current += string(r)
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func pruneEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	found := []map[string]any{}
	for _, event := range readLogEvents(t, home) {
		if event["msg"] == "agent pruned" {
			found = append(found, event)
		}
	}
	return found
}

func requirePruneEvent(t *testing.T, home string) map[string]any {
	t.Helper()
	events := pruneEvents(t, home)
	if len(events) != 1 {
		t.Fatalf("agent pruned events = %d; want 1", len(events))
	}
	event := events[0]
	if event["component"] != "prune" {
		t.Fatalf("component = %v; want prune", event["component"])
	}
	return event
}

func TestPruneDeletesUnreferencedStoppedRecordAndInbox(t *testing.T) {
	home := newHome(t)
	seedAgent(t, home, records.Agent{
		Key: "worker-old", Name: "worker-old", Project: "demo", Status: records.AgentStopped,
		StartedAt: "2026-01-01T00:00:00Z",
	})
	seedAgent(t, home, records.Agent{
		Key: "owner-demo", Name: "owner-demo", Owns: "demo", Project: "demo",
		Status: records.AgentLive, StartedAt: "2026-01-01T00:00:00Z",
	})
	seedInbox(t, home, "worker-old", records.Inbox{Agent: "owner-demo", Text: "finish the task"})

	result, err := Run(context.Background(), home)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Pruned) != 1 || result.Pruned[0].Agent != "worker-old" {
		t.Fatalf("result = %#v; want one pruned worker-old", result)
	}
	if agentExists(t, home, "worker-old") {
		t.Fatalf("worker-old record remains")
	}
	if inboxExists(t, home, "worker-old") {
		t.Fatalf("worker-old inbox folder remains")
	}
	event := requirePruneEvent(t, home)
	if event["agent_id"] != "worker-old" || event["status"] != records.AgentStopped {
		t.Fatalf("event = %#v; want agent_id and status", event)
	}
	if event["inbox_records"] != float64(1) || event["moved_records"] != float64(1) || event["moved_to"] != "project:demo" {
		t.Fatalf("event = %#v; want one moved record to project:demo", event)
	}
}

func TestPruneCandidateStatuses(t *testing.T) {
	for _, status := range []string{records.AgentStopped, records.AgentExited, records.AgentStale} {
		t.Run(status, func(t *testing.T) {
			home := newHome(t)
			seedAgent(t, home, records.Agent{
				Key: "worker-" + status, Name: "worker-" + status, Status: status,
				StartedAt: "2026-01-01T00:00:00Z",
			})
			if _, err := Run(context.Background(), home); err != nil {
				t.Fatal(err)
			}
			if agentExists(t, home, "worker-"+status) {
				t.Fatalf("%s record remains", status)
			}
		})
	}
}

func TestPruneKeepsReferencedRecords(t *testing.T) {
	t.Run("live child", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "parent", Name: "parent", Status: records.AgentStopped})
		seedAgent(t, home, records.Agent{Key: "child", Name: "child", Parent: "parent", Status: records.AgentLive})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if !agentExists(t, home, "parent") {
			t.Fatalf("stopped parent of a live child was pruned")
		}
	})

	t.Run("live successor replaces", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "old", Name: "old", Status: records.AgentStopped})
		seedAgent(t, home, records.Agent{Key: "new", Name: "new", Replaces: "old", Status: records.AgentLive})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if !agentExists(t, home, "old") {
			t.Fatalf("stopped record named by a live replaces was pruned")
		}
	})

	t.Run("four generation replaces chain from the parent", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "manager-old", Name: "manager", Status: records.AgentStopped})
		seedAgent(t, home, records.Agent{Key: "manager-middle", Name: "manager", Replaces: "manager-old", Status: records.AgentExited})
		seedAgent(t, home, records.Agent{Key: "manager-newer", Name: "manager", Replaces: "manager-middle", Status: records.AgentExited})
		seedAgent(t, home, records.Agent{Key: "manager-new", Name: "manager", Replaces: "manager-newer", Status: records.AgentLive})
		seedAgent(t, home, records.Agent{Key: "worker", Name: "worker", Parent: "manager-old", Status: records.AgentLive})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"manager-old", "manager-middle", "manager-newer", "manager-new"} {
			if !agentExists(t, home, key) {
				t.Fatalf("%s in the parent replaces chain was pruned", key)
			}
		}
	})

	t.Run("open task acting agent", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Status: records.AgentStopped})
		seedTask(t, home, records.Task{ID: "1-demo", Project: "demo", ActingAgent: "worker-1", Status: "assigned"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if !agentExists(t, home, "worker-1") {
			t.Fatalf("stopped acting agent of an open task was pruned")
		}
	})

	t.Run("open task agent with the task fact", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Task: "1-demo", Status: records.AgentStopped})
		seedTask(t, home, records.Task{ID: "1-demo", Project: "demo", ActingAgent: "other", Status: "assigned"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if !agentExists(t, home, "worker-1") {
			t.Fatalf("stopped agent with the task fact of an open task was pruned")
		}
	})

	t.Run("unresolved record", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Status: records.AgentUnresolved})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if !agentExists(t, home, "worker-1") {
			t.Fatalf("unresolved record was pruned")
		}
	})

	t.Run("unresolved facts do not keep other records", func(t *testing.T) {
		for _, relation := range []string{"parent", "replaces"} {
			t.Run(relation, func(t *testing.T) {
				home := newHome(t)
				seedAgent(t, home, records.Agent{Key: "old", Name: "old", Status: records.AgentStopped})
				uncertain := records.Agent{Key: "uncertain", Name: "uncertain", Status: records.AgentUnresolved}
				if relation == "parent" {
					uncertain.Parent = "old"
				} else {
					uncertain.Replaces = "old"
				}
				seedAgent(t, home, uncertain)
				if _, err := Run(context.Background(), home); err != nil {
					t.Fatal(err)
				}
				if agentExists(t, home, "old") {
					t.Fatalf("stopped record kept by the %s fact of an unresolved agent", relation)
				}
				if !agentExists(t, home, "uncertain") {
					t.Fatalf("unresolved record was pruned")
				}
			})
		}
	})

	t.Run("task status controls the reference", func(t *testing.T) {
		for _, test := range []struct {
			name     string
			status   string
			pushed   bool
			delivery bool
			kept     bool
		}{
			{name: "created keeps", status: "created", kept: true},
			{name: "assigned keeps", status: "assigned", kept: true},
			{name: "completed keeps", status: "completed", kept: true},
			{name: "cancelled clears", status: "cancelled", kept: false},
			{name: "verified without delivery keeps", status: "verified", kept: true},
			{name: "verified with an unpushed delivery keeps", status: "verified", delivery: true, kept: true},
			{name: "verified with a pushed delivery clears", status: "verified", delivery: true, pushed: true, kept: false},
		} {
			t.Run(test.name, func(t *testing.T) {
				home := newHome(t)
				seedAgent(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Task: "1-demo", Status: records.AgentStopped})
				task := records.Task{ID: "1-demo", Project: "demo", ActingAgent: "other", Status: test.status}
				if test.delivery {
					task.Delivery = &records.Delivery{Mode: "merge", Pushed: test.pushed}
				}
				seedTask(t, home, task)
				if _, err := Run(context.Background(), home); err != nil {
					t.Fatal(err)
				}
				if got := agentExists(t, home, "worker-1"); got != test.kept {
					t.Fatalf("record exists = %v; want %v", got, test.kept)
				}
			})
		}
	})
}

func TestPruneMovesUnreadMail(t *testing.T) {
	t.Run("project address", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "owner-demo", Name: "owner-demo", Owns: "demo", Project: "demo", Status: records.AgentLive})
		seedAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Project: "demo", Status: records.AgentStopped})
		seedInbox(t, home, "worker-old", records.Inbox{Agent: "owner-demo", Kind: records.InboxMessage, Task: "1-demo", Text: "unread"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		moved := readInbox(t, home, "owner-demo")
		if len(moved) != 1 {
			t.Fatalf("owner inbox = %#v; want one moved record", moved)
		}
		if moved[0].Agent != "owner-demo" || moved[0].Text != "unread" || moved[0].Task != "1-demo" || moved[0].CreatedAt != testTime {
			t.Fatalf("moved record = %#v; want the sender, task, text, and time", moved[0])
		}
		if moved[0].Seq == 0 || moved[0].ReadAt != "" {
			t.Fatalf("moved record = %#v; want a new id and no read time", moved[0])
		}
		event := requirePruneEvent(t, home)
		if event["moved_to"] != "project:demo" || event["moved_records"] != float64(1) {
			t.Fatalf("event = %#v; want project:demo and one moved record", event)
		}
	})

	t.Run("parent inbox", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "boss", Name: "boss", Status: records.AgentLive})
		seedAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Parent: "boss", Status: records.AgentStopped})
		seedInbox(t, home, "worker-old", records.Inbox{Agent: "boss", Text: "unread"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if moved := readInbox(t, home, "boss"); len(moved) != 1 || moved[0].Text != "unread" {
			t.Fatalf("parent inbox = %#v; want one moved record", moved)
		}
		event := requirePruneEvent(t, home)
		if event["moved_to"] != "boss" {
			t.Fatalf("event = %#v; want moved_to boss", event)
		}
	})

	t.Run("coordinator fallback", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Status: records.AgentStopped})
		seedInbox(t, home, "worker-old", records.Inbox{Agent: "someone", Text: "unread"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if moved := readInbox(t, home, "coordinator"); len(moved) != 1 || moved[0].Text != "unread" {
			t.Fatalf("coordinator inbox = %#v; want one moved record", moved)
		}
		event := requirePruneEvent(t, home)
		if event["moved_to"] != "coordinator" {
			t.Fatalf("event = %#v; want moved_to coordinator", event)
		}
	})

	t.Run("read mail stays and no record moves", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Status: records.AgentStopped})
		seedInbox(t, home, "worker-old", records.Inbox{Agent: "someone", Text: "read", ReadAt: testTime})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if moved := readInbox(t, home, "coordinator"); len(moved) != 0 {
			t.Fatalf("coordinator inbox = %#v; want no move", moved)
		}
		if inboxExists(t, home, "worker-old") {
			t.Fatalf("worker-old inbox folder remains")
		}
		event := requirePruneEvent(t, home)
		if event["inbox_records"] != float64(1) || event["moved_records"] != float64(0) || event["moved_to"] != "" {
			t.Fatalf("event = %#v; want one record and no move", event)
		}
	})

	t.Run("pruned parent receives the child mail then routes it on", func(t *testing.T) {
		home := newHome(t)
		seedAgent(t, home, records.Agent{Key: "parent", Name: "parent", Status: records.AgentStopped})
		seedAgent(t, home, records.Agent{Key: "child", Name: "child", Parent: "parent", Status: records.AgentStopped})
		seedInbox(t, home, "child", records.Inbox{Agent: "parent", Text: "child unread"})
		if _, err := Run(context.Background(), home); err != nil {
			t.Fatal(err)
		}
		if agentExists(t, home, "parent") || agentExists(t, home, "child") {
			t.Fatalf("parent or child remains")
		}
		if inboxExists(t, home, "parent") || inboxExists(t, home, "child") {
			t.Fatalf("an inbox folder remains")
		}
		moved := readInbox(t, home, "coordinator")
		if len(moved) != 1 || moved[0].Text != "child unread" {
			t.Fatalf("coordinator inbox = %#v; want the child mail", moved)
		}
	})
}

func TestPruneAppendsEventAndKeepsExistingLog(t *testing.T) {
	home := newHome(t)
	logs := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(logs, "2000-01-01.jsonl")
	oldLine := `{"time":"2000-01-01T00:00:00Z","level":"DEBUG","msg":"keep me","component":"old"}`
	if err := os.WriteFile(oldPath, []byte(oldLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seedAgent(t, home, records.Agent{Key: "worker-old", Name: "worker-old", Status: records.AgentStopped})
	if _, err := Run(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != oldLine+"\n" {
		t.Fatalf("old log = %q; want it unchanged", string(data))
	}
	requirePruneEvent(t, home)
}

func TestPruneScratchHomeLifecycle(t *testing.T) {
	home := newHome(t)
	seedAgent(t, home, records.Agent{Key: "owner-demo", Name: "owner-demo", Owns: "demo", Project: "demo", Status: records.AgentLive})
	seedAgent(t, home, records.Agent{Key: "manager-old", Name: "manager", Status: records.AgentStopped})
	seedAgent(t, home, records.Agent{Key: "manager-new", Name: "manager", Replaces: "manager-old", Status: records.AgentLive})
	seedAgent(t, home, records.Agent{Key: "worker-child", Name: "worker-child", Parent: "manager-new", Status: records.AgentLive})
	seedAgent(t, home, records.Agent{Key: "worker-busy", Name: "worker-busy", Status: records.AgentStopped})
	seedAgent(t, home, records.Agent{Key: "worker-lost", Name: "worker-lost", Parent: "worker-stray", Status: records.AgentUnresolved})
	seedAgent(t, home, records.Agent{Key: "worker-stray", Name: "worker-stray", Project: "demo", Status: records.AgentStopped})
	seedTask(t, home, records.Task{ID: "1-demo", Project: "demo", ActingAgent: "worker-busy", Status: "assigned"})
	seedInbox(t, home, "worker-stray", records.Inbox{Agent: "owner-demo", Text: "still open"})
	seedInbox(t, home, "worker-busy", records.Inbox{Agent: "owner-demo", Text: "busy mail"})

	logs := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(logs, "2000-01-01.jsonl")
	if err := os.WriteFile(oldPath, []byte(`{"time":"2000-01-01T00:00:00Z","level":"DEBUG","msg":"before"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := readLogEvents(t, home)

	if _, err := Run(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"owner-demo", "manager-old", "manager-new", "worker-child", "worker-busy", "worker-lost"} {
		if !agentExists(t, home, key) {
			t.Fatalf("%s was pruned; want it kept", key)
		}
	}
	if agentExists(t, home, "worker-stray") || inboxExists(t, home, "worker-stray") {
		t.Fatalf("worker-stray or its inbox remains")
	}
	moved := readInbox(t, home, "owner-demo")
	if len(moved) != 1 {
		t.Fatalf("owner inbox = %#v; want the stray record", moved)
	}
	if kept := readInbox(t, home, "worker-busy"); len(kept) != 1 {
		t.Fatalf("worker-busy inbox = %#v; want its unread mail kept", kept)
	}
	after := readLogEvents(t, home)
	if len(after) != len(before)+1 {
		t.Fatalf("log events = %d; want %d with one new event", len(after), len(before)+1)
	}
	if after[len(after)-1]["msg"] != "agent pruned" {
		t.Fatalf("new log event = %#v; want agent pruned", after[len(after)-1])
	}
}
