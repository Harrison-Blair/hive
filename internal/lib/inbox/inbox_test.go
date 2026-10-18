package inbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func appendMessage(t *testing.T, home, role, text string) output.Envelope {
	t.Helper()
	envelope := Append(home, AppendOptions{Role: role, Kind: records.InboxMessage, Agent: "worker-x", Text: text})
	if envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	return envelope
}

func resultOf(t *testing.T, envelope output.Envelope) Result {
	t.Helper()
	result, ok := envelope.Result.(Result)
	if !ok {
		t.Fatalf("result = %#v; want inbox.Result", envelope.Result)
	}
	return result
}

func TestAppendGivesSequenceOneTwoThree(t *testing.T) {
	home := t.TempDir()
	for i := 1; i <= 3; i++ {
		envelope := appendMessage(t, home, "coordinator", fmt.Sprintf("message %d", i))
		if len(envelope.Effects) != 2 || envelope.Effects[0].Action != "updated" || envelope.Effects[1].Action != "created" {
			t.Fatalf("effects = %#v; want counter updated and inbox created", envelope.Effects)
		}
	}
	found, err := Unread(home, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 3 {
		t.Fatalf("unread = %d; want 3", len(found))
	}
	for i, record := range found {
		if record.Seq != uint64(i+1) {
			t.Fatalf("record %d seq = %d", i, record.Seq)
		}
		if record.Kind != records.InboxMessage || record.ReadAt != "" {
			t.Fatalf("record = %#v", record)
		}
	}
	if count, err := UnreadCount(home, "coordinator"); err != nil || count != 3 {
		t.Fatalf("UnreadCount() = %d, %v; want 3", count, err)
	}
	counts, err := UnreadCountAll(home)
	if err != nil || counts["coordinator"] != 3 {
		t.Fatalf("UnreadCountAll() = %#v, %v; want coordinator 3", counts, err)
	}
}

func TestNumericSequenceIDsListAndOrder(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	ids := []uint64{1, 2, 10, ^uint64(0)}
	for _, seq := range ids {
		record := records.Inbox{Seq: seq, Role: "coordinator", Kind: records.InboxMessage, Text: strconv.FormatUint(seq, 10)}
		if err := st.WriteIn([]string{"inbox", "coordinator"}, strconv.FormatUint(seq, 10), record); err != nil {
			t.Fatal(err)
		}
	}
	found, err := All(home, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != len(ids) {
		t.Fatalf("all = %d; want %d", len(found), len(ids))
	}
	for i, seq := range ids {
		if found[i].Seq != seq {
			t.Fatalf("record %d seq = %d; want %d", i, found[i].Seq, seq)
		}
	}
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "ok" {
		t.Fatalf("Consume() = %#v", envelope)
	}
	result := resultOf(t, envelope)
	for i, seq := range ids {
		if result.Records[i].Seq != seq {
			t.Fatalf("consumed %d seq = %d; want %d", i, result.Records[i].Seq, seq)
		}
		if result.Records[i].ReadAt == "" {
			t.Fatalf("record %d has no read_at", seq)
		}
	}
	if remaining, err := Unread(home, "coordinator"); err != nil || len(remaining) != 0 {
		t.Fatalf("unread after consume = %#v, %v; want empty", remaining, err)
	}
}

func TestConsumeMarksAndCountsAgree(t *testing.T) {
	home := t.TempDir()
	for i := 0; i < 3; i++ {
		appendMessage(t, home, "coordinator", "hello")
	}
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "ok" || len(envelope.Effects) != 3 {
		t.Fatalf("Consume() = %#v; want ok with three consumed effects", envelope)
	}
	for _, effect := range envelope.Effects {
		if effect.Resource != "inbox" || effect.Action != "consumed" {
			t.Fatalf("effect = %#v", effect)
		}
	}
	if count, err := UnreadCount(home, "coordinator"); err != nil || count != 0 {
		t.Fatalf("UnreadCount() = %d, %v; want 0", count, err)
	}
	if all, err := All(home, "coordinator"); err != nil || len(all) != 3 {
		t.Fatalf("All() = %#v, %v; want three marked records", all, err)
	}
}

func TestParallelAppendsLoseNoRecord(t *testing.T) {
	home := t.TempDir()
	const writers, each = 10, 5
	var wg sync.WaitGroup
	errs := make(chan string, writers*each)
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < each; j++ {
				envelope := Append(home, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: fmt.Sprintf("%d-%d", i, j)})
				if envelope.Outcome != "ok" {
					errs <- fmt.Sprintf("%v", envelope.Error)
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for message := range errs {
		t.Fatal(message)
	}
	all, err := All(home, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != writers*each {
		t.Fatalf("records = %d; want %d", len(all), writers*each)
	}
	seen := map[uint64]bool{}
	for _, record := range all {
		if seen[record.Seq] {
			t.Fatalf("duplicate sequence %d", record.Seq)
		}
		seen[record.Seq] = true
	}
}

func TestParallelConsumersGetDisjointRecords(t *testing.T) {
	home := t.TempDir()
	const total = 100
	for i := 0; i < total; i++ {
		appendMessage(t, home, "coordinator", "hello")
	}
	var wg sync.WaitGroup
	results := make(chan output.Envelope, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- Consume(home, "coordinator")
		}()
	}
	wg.Wait()
	close(results)
	seen := map[uint64]bool{}
	for envelope := range results {
		if envelope.Outcome != "ok" {
			t.Fatalf("Consume() = %#v", envelope)
		}
		for _, record := range resultOf(t, envelope).Records {
			if seen[record.Seq] {
				t.Fatalf("sequence %d printed twice", record.Seq)
			}
			seen[record.Seq] = true
		}
	}
	if len(seen) != total {
		t.Fatalf("consumed %d records; want %d", len(seen), total)
	}
}

func TestInvalidKeysFailBeforeFileOperation(t *testing.T) {
	for _, key := range []string{"../x", "A", "", strings.Repeat("a", 42)} {
		t.Run(key, func(t *testing.T) {
			home := t.TempDir()
			envelope := Append(home, AppendOptions{Role: key, Text: "hello"})
			if envelope.Outcome != "failed" || envelope.Error != "invalid inbox key: "+key {
				t.Fatalf("Append() = %#v; want invalid key", envelope)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid key created state: %v", err)
			}
			if _, err := Unread(home, key); err == nil {
				t.Fatal("Unread() accepted an invalid key")
			}
			if envelope := Consume(home, key); envelope.Outcome != "failed" {
				t.Fatalf("Consume() = %#v; want invalid key", envelope)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid key created state: %v", err)
			}
		})
	}
}

func TestKeyLengthBoundaries(t *testing.T) {
	accepted := strings.Repeat("a", 41)
	if envelope := Append(t.TempDir(), AppendOptions{Role: accepted, Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v; want the literal 41-byte key accepted", envelope)
	}
	refused := strings.Repeat("a", 42)
	home := t.TempDir()
	envelope := Append(home, AppendOptions{Role: refused, Text: "hello"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Append() = %#v; want the literal 42-byte key refused", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("42-byte key created state: %v", err)
	}
}

func TestTextByteBoundaries(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"4095-ascii", strings.Repeat("a", 4095), true},
		{"4096-ascii", strings.Repeat("a", 4096), true},
		{"4097-ascii", strings.Repeat("a", 4097), false},
		{"4096-utf8", strings.Repeat("\u00e9", 2048), true},
		{"4097-utf8", "a" + strings.Repeat("\u00e9", 2048), false},
		{"8192-ascii", strings.Repeat("a", 8192), false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			envelope := Append(home, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: test.text})
			if test.want {
				if envelope.Outcome != "ok" {
					t.Fatalf("Append() = %#v; want %d bytes accepted", envelope, len(test.text))
				}
				return
			}
			if envelope.Outcome != "failed" {
				t.Fatalf("Append() = %#v; want %d bytes refused", envelope, len(test.text))
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("refused text created state: %v", err)
			}
		})
	}
}

func TestInvalidKindFails(t *testing.T) {
	home := t.TempDir()
	envelope := Append(home, AppendOptions{Role: "coordinator", Kind: "note", Text: "hello"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Append() = %#v; want invalid kind", envelope)
	}
}

func TestConsumeNoUnreadRecords(t *testing.T) {
	home := t.TempDir()
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 {
		t.Fatalf("Consume() = %#v; want ok with no effect", envelope)
	}
	if result := resultOf(t, envelope); len(result.Records) != 0 {
		t.Fatalf("records = %#v; want empty", result.Records)
	}
}

func markNth(calls *int, failAt int, now string) MarkFunc {
	return func(locked *store.LockedStore, record records.Inbox) (records.Inbox, error) {
		*calls++
		if *calls == failAt {
			return record, errors.New("seam mark failure")
		}
		var updated records.Inbox
		err := locked.UpdateIn([]string{"inbox", record.Role}, strconv.FormatUint(record.Seq, 10), func(raw json.RawMessage) (json.RawMessage, error) {
			if err := json.Unmarshal(raw, &updated); err != nil {
				return nil, err
			}
			updated.ReadAt = now
			return json.Marshal(updated)
		})
		return updated, err
	}
}

func runConsumeLocked(t *testing.T, home, key string, mark MarkFunc) output.Envelope {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = consumeLocked(locked, key, mark)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

func readInbox(t *testing.T, home, key string, seq uint64) records.Inbox {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Inbox
	if err := st.ReadIn([]string{"inbox", key}, strconv.FormatUint(seq, 10), &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func TestConsumeFirstMarkFailureLeavesEachRecordUnread(t *testing.T) {
	home := t.TempDir()
	appendMessage(t, home, "coordinator", "one")
	appendMessage(t, home, "coordinator", "two")
	calls := 0
	envelope := runConsumeLocked(t, home, "coordinator", markNth(&calls, 1, "2026-01-01T00:00:00Z"))
	if envelope.Outcome != "failed" || envelope.Phase != "consume" || len(envelope.Effects) != 0 {
		t.Fatalf("consumeLocked() = %#v; want failed consume with no effect", envelope)
	}
	if record := readInbox(t, home, "coordinator", 1); record.ReadAt != "" {
		t.Fatalf("record 1 read_at = %q; want empty", record.ReadAt)
	}
	if record := readInbox(t, home, "coordinator", 2); record.ReadAt != "" {
		t.Fatalf("record 2 read_at = %q; want empty", record.ReadAt)
	}
}

func TestConsumePartialAfterFirstMark(t *testing.T) {
	home := t.TempDir()
	appendMessage(t, home, "coordinator", "one")
	appendMessage(t, home, "coordinator", "two")
	calls := 0
	envelope := runConsumeLocked(t, home, "coordinator", markNth(&calls, 2, "2026-01-01T00:00:00Z"))
	if envelope.Outcome != "partial" || envelope.Phase != "consume" {
		t.Fatalf("consumeLocked() = %#v; want partial consume", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0] != (output.Effect{Resource: "inbox", Path: ".hive/state/inbox/coordinator/1.json", Action: "consumed"}) {
		t.Fatalf("effects = %#v; want one consumed effect", envelope.Effects)
	}
	result := resultOf(t, envelope)
	if len(result.Records) != 1 || result.Records[0].Seq != 1 {
		t.Fatalf("records = %#v; want first marked record", result.Records)
	}
	if len(result.Remaining) != 1 || result.Remaining[0].Seq != 2 {
		t.Fatalf("remaining = %#v; want second unread record", result.Remaining)
	}
	if record := readInbox(t, home, "coordinator", 1); record.ReadAt == "" {
		t.Fatal("record 1 is unread; want read")
	}
	if record := readInbox(t, home, "coordinator", 2); record.ReadAt != "" {
		t.Fatal("record 2 is read; want unread")
	}
}

type appendSeam struct {
	*store.LockedStore
	failWrite bool
	failNext  bool
	writes    int
	nexts     int
}

func (s *appendSeam) Next(name string) (uint64, error) {
	s.nexts++
	if s.failNext {
		return 0, errors.New("seam counter failure")
	}
	return s.LockedStore.Next(name)
}

func (s *appendSeam) WriteIn(collection []string, id string, value any) error {
	s.writes++
	if s.failWrite {
		return errors.New("seam record failure")
	}
	return s.LockedStore.WriteIn(collection, id, value)
}

func runAppendLocked(t *testing.T, home string, seam *appendSeam, options AppendOptions) output.Envelope {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		seam.LockedStore = locked
		envelope = appendLocked(locked, seam, options, func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

func TestAppendRecordFailureLeavesCounterGap(t *testing.T) {
	home := t.TempDir()
	seam := &appendSeam{failWrite: true}
	envelope := runAppendLocked(t, home, seam, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: "hello"})
	if envelope.Outcome != "partial" || envelope.Phase != "record" {
		t.Fatalf("appendLocked() = %#v; want partial record", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Resource != "counter" || envelope.Effects[0].Action != "updated" {
		t.Fatalf("effects = %#v; want one counter effect", envelope.Effects)
	}
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var counter struct {
		Value uint64 `json:"value"`
	}
	if err := st.Read("counters", "inbox:coordinator", &counter); err != nil || counter.Value != 1 {
		t.Fatalf("counter = %#v, %v; want 1", counter, err)
	}
	if all, err := All(home, "coordinator"); err != nil || len(all) != 0 {
		t.Fatalf("records = %#v, %v; want none", all, err)
	}
	appendMessage(t, home, "coordinator", "retry")
	all, err := All(home, "coordinator")
	if err != nil || len(all) != 1 || all[0].Seq != 2 {
		t.Fatalf("retry records = %#v, %v; want sequence 2", all, err)
	}
}

func TestAppendCounterFailureHasNoEffect(t *testing.T) {
	home := t.TempDir()
	seam := &appendSeam{failNext: true}
	envelope := runAppendLocked(t, home, seam, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: "hello"})
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("appendLocked() = %#v; want failed with no effect", envelope)
	}
}

func TestConsumeMarksReadAtAndPreservesOtherFields(t *testing.T) {
	home := t.TempDir()
	appendMessage(t, home, "coordinator", "hello")
	before := readInbox(t, home, "coordinator", 1)
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "ok" {
		t.Fatalf("Consume() = %#v", envelope)
	}
	after := readInbox(t, home, "coordinator", 1)
	if after.ReadAt == "" {
		t.Fatal("record 1 read_at is empty; want a completed mark")
	}
	if after.Seq != before.Seq || after.Role != before.Role || after.Kind != before.Kind ||
		after.Task != before.Task || after.Agent != before.Agent || after.Text != before.Text ||
		after.CreatedAt != before.CreatedAt || after.NudgedAt != before.NudgedAt {
		t.Fatalf("mark changed a field other than read_at: before %#v after %#v", before, after)
	}
}

func TestConsumeRefusesCrossRoleStoredRecord(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"coordinator", "supervisor"} {
		record := records.Inbox{Seq: 1, Role: "supervisor", Kind: records.InboxMessage, Agent: "worker-x", Text: key}
		if err := st.WriteIn([]string{"inbox", key}, "1", record); err != nil {
			t.Fatal(err)
		}
	}
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("Consume() = %#v; want failed with no effect", envelope)
	}
	message, _ := envelope.Error.(string)
	if !strings.Contains(message, ".hive/state/inbox/coordinator/1.json") {
		t.Fatalf("error = %q; want the selected record path", message)
	}
	for _, key := range []string{"coordinator", "supervisor"} {
		if record := readInbox(t, home, key, 1); record.ReadAt != "" {
			t.Fatalf("%s record read_at = %q; want no mark", key, record.ReadAt)
		}
	}
}

func TestConsumeLockedKeepsTheRecordPathWithoutHome(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Inbox{Seq: 1, Role: "supervisor", Kind: records.InboxMessage, Text: "cross-role"}
	if err := st.WriteIn([]string{"inbox", "coordinator"}, "1", record); err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = consumeLocked(locked, "coordinator", markRead)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("consumeLocked() = %#v; want failed with no effect", envelope)
	}
	path := ".hive/state/inbox/coordinator/1.json"
	message, _ := envelope.Error.(string)
	if !strings.Contains(message, path) {
		t.Fatalf("error = %q; want the record path %s", message, path)
	}
}

func TestConsumeRefusesSequencePathMismatch(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Inbox{Seq: 1, Role: "coordinator", Kind: records.InboxMessage, Text: "mismatch"}
	if err := st.WriteIn([]string{"inbox", "coordinator"}, "2", record); err != nil {
		t.Fatal(err)
	}
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("Consume() = %#v; want failed with no effect", envelope)
	}
	message, _ := envelope.Error.(string)
	if !strings.Contains(message, ".hive/state/inbox/coordinator/2.json") {
		t.Fatalf("error = %q; want the mismatched record path", message)
	}
	if record := readInbox(t, home, "coordinator", 2); record.ReadAt != "" {
		t.Fatalf("mismatched record read_at = %q; want no mark", record.ReadAt)
	}
}

func TestConsumeDecodeErrorKeepsRecordPath(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".hive", "state", "inbox", "coordinator", "1.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"seq":"wrong","role":"coordinator"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	envelope := Consume(home, "coordinator")
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("Consume() = %#v; want failed with no effect", envelope)
	}
	message, _ := envelope.Error.(string)
	if !strings.Contains(message, path) {
		t.Fatalf("error = %q; want the record path %s", message, path)
	}
}

func TestUnreadCountAllRejectsLinkedCollections(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		home := t.TempDir()
		if _, err := store.Open(home); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(home, "target-inbox")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		root := filepath.Join(home, ".hive", "state", "inbox")
		if err := os.Symlink(target, root); err != nil {
			t.Fatal(err)
		}
		counts, err := UnreadCountAll(home)
		if err == nil || counts != nil || !strings.Contains(err.Error(), root) {
			t.Fatalf("UnreadCountAll() = %#v, %v; want a path error for %s", counts, err, root)
		}
	})
	t.Run("role", func(t *testing.T) {
		home := t.TempDir()
		if _, err := store.Open(home); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(home, "target-role")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		role := filepath.Join(home, ".hive", "state", "inbox", "coordinator")
		if err := os.MkdirAll(filepath.Dir(role), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, role); err != nil {
			t.Fatal(err)
		}
		counts, err := UnreadCountAll(home)
		if err == nil || counts != nil || !strings.Contains(err.Error(), role) {
			t.Fatalf("UnreadCountAll() = %#v, %v; want a path error for %s", counts, err, role)
		}
	})
}

func readInboxLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	dir := filepath.Join(home, ".hive", "logs")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
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
			events = append(events, event)
		}
	}
	return events
}

func findInboxLogEvent(events []map[string]any, message string) map[string]any {
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	return nil
}

func TestInboxLogsTransitionsAndErrors(t *testing.T) {
	home := t.TempDir()
	appendMessage(t, home, "coordinator", "secret message text")
	if envelope := Consume(home, "coordinator"); envelope.Outcome != "ok" {
		t.Fatalf("Consume() = %#v", envelope)
	}
	if envelope := Append(home, AppendOptions{Role: "../x", Text: "secret message text"}); envelope.Outcome != "failed" {
		t.Fatalf("invalid key Append() = %#v", envelope)
	}
	events := readInboxLogEvents(t, home)
	appendEvent := findInboxLogEvent(events, "inbox record appended")
	if appendEvent == nil || appendEvent["role"] != "coordinator" || appendEvent["sequence"] != float64(1) || appendEvent["agent"] != "worker-x" {
		t.Fatalf("append event = %#v; want role, sequence 1, and agent", appendEvent)
	}
	consumeEvent := findInboxLogEvent(events, "inbox record consumed")
	if consumeEvent == nil || consumeEvent["role"] != "coordinator" || consumeEvent["sequence"] != float64(1) {
		t.Fatalf("consume event = %#v; want role and sequence 1", consumeEvent)
	}
	failureEvent := findInboxLogEvent(events, "inbox operation failed")
	if failureEvent == nil {
		t.Fatalf("events = %#v; want an inbox failure event", events)
	}
	message, _ := failureEvent["error"].(string)
	if !strings.Contains(message, "invalid inbox key") {
		t.Fatalf("failure event = %#v; want the invalid key error", failureEvent)
	}
	for _, event := range events {
		data, _ := json.Marshal(event)
		if strings.Contains(string(data), "secret message text") {
			t.Fatalf("log event contains message text: %s", data)
		}
	}
}

func TestAppendUsesTheLockedCounter(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	seam := &appendSeam{}
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		seam.LockedStore = locked
		envelope := appendLocked(locked, seam, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: "hello"}, time.Now)
		if envelope.Outcome != "ok" {
			t.Fatalf("appendLocked() = %#v", envelope)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seam.nexts != 1 {
		t.Fatalf("Next() calls = %d; want the grouped counter once", seam.nexts)
	}
}
