package task

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type finishSeam struct {
	*store.LockedStore
	updateCollection string
	writeCollection  string
	failUpdate       bool
	failWrite        bool
	failNext         bool
}

func (s *finishSeam) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	if s.failUpdate && strings.Join(collection, "/") == s.updateCollection {
		return errors.New("seam update failure")
	}
	return s.LockedStore.UpdateIn(collection, id, fn)
}

func (s *finishSeam) WriteIn(collection []string, id string, value any) error {
	if s.failWrite && strings.Join(collection, "/") == s.writeCollection {
		return errors.New("seam write failure")
	}
	return s.LockedStore.WriteIn(collection, id, value)
}

func (s *finishSeam) Next(name string) (uint64, error) {
	if s.failNext {
		return 0, errors.New("seam counter failure")
	}
	return s.LockedStore.Next(name)
}

func runFinishLocked(t *testing.T, home, command string, options TransitionOptions, seam *finishSeam) output.Envelope {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		writer := finishWriter(locked)
		if seam != nil {
			seam.LockedStore = locked
			writer = seam
		}
		envelope = finishLocked(writer, command, options)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

func readFinishAgent(t *testing.T, home, key string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func seedClaim(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

func finishInbox(t *testing.T, home, key string) []records.Inbox {
	t.Helper()
	found, err := inbox.All(home, key)
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func finishCounter(t *testing.T, home, key string) (uint64, bool) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var counter struct {
		Value uint64 `json:"value"`
	}
	if err := registry.Read("counters", "inbox:"+key, &counter); err != nil {
		return 0, false
	}
	return counter.Value, true
}

func effectResources(envelope output.Envelope) string {
	names := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		names = append(names, effect.Resource)
	}
	return strings.Join(names, ",")
}

func TestCompleteFinishWritesTaskCounterRecordAndMark(t *testing.T) {
	home, task := newSeed(t, "assigned")
	envelope := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-1"}, nil)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("finishLocked(complete) = %#v; want ok", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "completed" {
		t.Fatalf("task status = %q; want completed", got.Status)
	}
	if worker := readFinishAgent(t, home, "worker-1"); !worker.Finished {
		t.Fatalf("worker = %#v; want finished", worker)
	}
	inboxRecords := finishInbox(t, home, "coordinator")
	if len(inboxRecords) != 1 {
		t.Fatalf("inbox = %#v; want one record", inboxRecords)
	}
	record := inboxRecords[0]
	if record.Kind != records.InboxCompleted || record.Task != task.ID || record.Agent != "worker-1" || record.Text != "completed" {
		t.Fatalf("record = %#v; want completed by worker-1", record)
	}
	if value, ok := finishCounter(t, home, "coordinator"); !ok || value != 1 {
		t.Fatalf("counter = %d, %v; want 1", value, ok)
	}
	if got := effectResources(envelope); got != "task,counter,inbox,agent" {
		t.Fatalf("effects = %s; want task, counter, inbox, agent", got)
	}
}

func TestVerifyFinishWritesVerifiedOrFailRecord(t *testing.T) {
	cases := []struct {
		name       string
		fail       bool
		reason     string
		wantStatus string
		wantText   string
	}{
		{"pass", false, "", "verified", "verified"},
		{"fail", true, "differs", "assigned", "fail: differs"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, task := newSeed(t, "completed")
			envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: test.fail, Reason: test.reason}, nil)
			if envelope.Outcome != "ok" || envelope.Error != nil {
				t.Fatalf("finishLocked(verify) = %#v; want ok", envelope)
			}
			if got := readTask(t, home, task.ID); got.Status != test.wantStatus {
				t.Fatalf("task status = %q; want %s", got.Status, test.wantStatus)
			}
			if verifier := readFinishAgent(t, home, "verifier-1"); !verifier.Finished {
				t.Fatalf("verifier = %#v; want finished", verifier)
			}
			inboxRecords := finishInbox(t, home, "coordinator")
			if len(inboxRecords) != 1 {
				t.Fatalf("inbox = %#v; want one record", inboxRecords)
			}
			record := inboxRecords[0]
			if record.Kind != records.InboxVerified || record.Task != task.ID || record.Agent != "verifier-1" || record.Text != test.wantText {
				t.Fatalf("record = %#v; want verified with text %q", record, test.wantText)
			}
			if value, ok := finishCounter(t, home, "coordinator"); !ok || value != 1 {
				t.Fatalf("counter = %d, %v; want 1", value, ok)
			}
		})
	}
}

func TestCompleteFinishAndVerifyWriteFailures(t *testing.T) {
	commands := []struct {
		name       string
		command    string
		status     string
		options    TransitionOptions
		wantStatus string
	}{
		{"complete", "complete", "assigned", TransitionOptions{Agent: "worker-1"}, "completed"},
		{"verify", "verify", "completed", TransitionOptions{Agent: "verifier-1"}, "verified"},
		{"verify-fail", "verify", "completed", TransitionOptions{Agent: "verifier-1", Fail: true, Reason: "differs"}, "assigned"},
	}
	failures := []struct {
		name       string
		seam       func() *finishSeam
		outcome    string
		phase      string
		effects    int
		counter    bool
		record     bool
		marked     bool
		resource   string
		effectText string
	}{
		{"task", func() *finishSeam { return &finishSeam{failUpdate: true, updateCollection: "tasks"} }, "failed", "task", 0, false, false, false, "", ""},
		{"counter", func() *finishSeam { return &finishSeam{failNext: true} }, "partial", "counter", 1, false, false, false, "task", "task"},
		{"record", func() *finishSeam { return &finishSeam{failWrite: true, writeCollection: "inbox/coordinator"} }, "partial", "record", 2, true, false, false, "task,counter", "task,counter"},
		{"finished", func() *finishSeam { return &finishSeam{failUpdate: true, updateCollection: "agents"} }, "partial", "finished", 3, true, true, false, "task,counter,inbox", "task,counter,inbox"},
	}
	for _, command := range commands {
		for _, failure := range failures {
			t.Run(command.name+"/"+failure.name, func(t *testing.T) {
				home, task := newSeed(t, command.status)
				options := command.options
				options.Task = task.ID
				envelope := runFinishLocked(t, home, command.command, options, failure.seam())
				if envelope.Outcome != failure.outcome || envelope.Phase != failure.phase {
					t.Fatalf("finish = %#v; want %s/%s", envelope, failure.outcome, failure.phase)
				}
				if got := effectResources(envelope); got != failure.effectText {
					t.Fatalf("effects = %q; want %q", got, failure.effectText)
				}
				if len(envelope.Effects) != failure.effects {
					t.Fatalf("effect count = %d; want %d", len(envelope.Effects), failure.effects)
				}
				gotTask := readTask(t, home, task.ID)
				if failure.name == "task" {
					if gotTask.Status != command.status {
						t.Fatalf("task status = %q; want unchanged %s", gotTask.Status, command.status)
					}
				} else if gotTask.Status != command.wantStatus {
					t.Fatalf("task status = %q; want %s", gotTask.Status, command.wantStatus)
				}
				value, ok := finishCounter(t, home, "coordinator")
				if ok != failure.counter || (ok && value != 1) {
					t.Fatalf("counter = %d, %v; want present=%t", value, ok, failure.counter)
				}
				if got := len(finishInbox(t, home, "coordinator")); (got == 1) != failure.record {
					t.Fatalf("inbox records = %d; want present=%t", got, failure.record)
				}
				caller := command.options.Agent
				if marked := readFinishAgent(t, home, caller).Finished; marked != failure.marked {
					t.Fatalf("caller finished = %t; want %t", marked, failure.marked)
				}
				if envelope.Error == nil {
					t.Fatal("error = nil; want the failure text")
				}
			})
		}
	}
}

func TestCompleteFinishRefusesWrongCallersAndLeavesState(t *testing.T) {
	type refusal struct {
		name    string
		status  string
		command string
		agent   string
		want    string
		seed    func(*testing.T, string)
	}
	cases := []refusal{
		{"worker of another task", "assigned", "complete", "worker-3", "not the acting agent: worker-3", nil},
		{"manager", "assigned", "complete", "manager-one", "not the acting agent: manager-one", func(t *testing.T, home string) {
			seedAgent(t, home, records.Agent{Key: "manager-one-g1", Name: "manager-one", Role: "manager", Profile: records.ProfilePlanner, Owns: "one", Project: "one", Status: records.AgentLive, PaneID: "w1:p8", Generation: 1})
			seedClaim(t, home, records.Claim{Key: "owner-one", Agent: "manager-one-g1", Generation: 1})
		}},
		{"verifier complete", "assigned", "complete", "verifier-1", "not the acting agent: verifier-1", nil},
		{"worker verify", "completed", "verify", "worker-2", "not a verifier of 1-seed: worker-2", nil},
		{"verifier of another task", "completed", "verify", "verifier-3", "not a verifier of 1-seed: verifier-3", nil},
		{"self verify", "completed", "verify", "worker-1", errSelfVerify.Error(), nil},
		{"unknown caller", "assigned", "complete", "nobody", "agent not found: nobody", nil},
		{"refused transition", "created", "complete", "worker-1", "invalid transition: created -> complete", nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, task := newSeed(t, test.status)
			seedAgent(t, home, records.Agent{Key: "worker-3", Name: "worker-3", Role: "worker", Project: "one", Task: "2-other", Status: records.AgentLive, PaneID: "w1:p9", Generation: 1})
			seedAgent(t, home, records.Agent{Key: "verifier-3", Name: "verifier-3", Role: "verifier", Project: "one", Task: "2-other", Status: records.AgentLive, PaneID: "w1:p10", Generation: 1})
			if test.seed != nil {
				test.seed(t, home)
			}
			envelope := runFinishLocked(t, home, test.command, TransitionOptions{Task: task.ID, Agent: test.agent}, nil)
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("finish = %#v; want a failure", envelope)
			}
			message, _ := envelope.Error.(string)
			if !strings.Contains(message, test.want) {
				t.Fatalf("error = %q; want text %q", message, test.want)
			}
			if got := readTask(t, home, task.ID); got.Status != test.status || got.ActingAgent != "worker-1" {
				t.Fatalf("task = %#v; want unchanged %s", got, test.status)
			}
			if value, ok := finishCounter(t, home, "coordinator"); ok {
				t.Fatalf("counter = %d; want no counter write", value)
			}
			if got := len(finishInbox(t, home, "coordinator")); got != 0 {
				t.Fatalf("inbox records = %d; want none", got)
			}
			for _, name := range []string{"worker-1", "verifier-1"} {
				if got := readFinishAgent(t, home, name); got.Finished {
					t.Fatalf("%s finished = true; want false", name)
				}
			}
		})
	}
}

func TestCompleteFinishAdvancedRefusalAndFreshRework(t *testing.T) {
	home, task := newSeed(t, "assigned")
	if envelope := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-1"}, nil); envelope.Outcome != "ok" {
		t.Fatalf("complete = %#v; want ok", envelope)
	}
	beforeInbox := len(finishInbox(t, home, "coordinator"))
	repeated := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-1"}, nil)
	if repeated.Outcome != "failed" {
		t.Fatalf("repeated complete = %#v; want failed", repeated)
	}
	message, _ := repeated.Error.(string)
	for _, want := range []string{"task show", "agent show", "inbox"} {
		if !strings.Contains(message, want) {
			t.Fatalf("repeated complete error = %q; want %q", message, want)
		}
	}
	if got := readTask(t, home, task.ID); got.Status != "completed" {
		t.Fatalf("task status = %q; want completed", got.Status)
	}
	if got := len(finishInbox(t, home, "coordinator")); got != beforeInbox {
		t.Fatalf("inbox records = %d; want %d", got, beforeInbox)
	}

	if envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true, Reason: "first"}, nil); envelope.Outcome != "ok" {
		t.Fatalf("verify fail = %#v; want ok", envelope)
	}
	beforeInbox = len(finishInbox(t, home, "coordinator"))
	retry := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true, Reason: "changed reason"}, nil)
	if retry.Outcome != "failed" {
		t.Fatalf("changed-reason retry = %#v; want failed", retry)
	}
	message, _ = retry.Error.(string)
	if !strings.Contains(message, "task show") || !strings.Contains(message, "agent show") || !strings.Contains(message, "inbox") {
		t.Fatalf("changed-reason retry error = %q; want the inspection commands", message)
	}
	if got := len(finishInbox(t, home, "coordinator")); got != beforeInbox {
		t.Fatalf("inbox records after retry = %d; want %d", got, beforeInbox)
	}
	if got := readTask(t, home, task.ID); got.Status != "assigned" {
		t.Fatalf("task status = %q; want assigned", got.Status)
	}

	if envelope := Assign(home, TransitionOptions{Task: task.ID, Agent: "worker-1"}); envelope.Outcome != "failed" || envelope.Error != "agent finished: worker-1: use a fresh agent" {
		t.Fatalf("same-agent assign = %#v; want the finished-agent refusal", envelope)
	}
	if got := readTask(t, home, task.ID); got.ActingAgent != "worker-1" {
		t.Fatalf("acting agent = %q; want unchanged worker-1", got.ActingAgent)
	}
	if envelope := Assign(home, TransitionOptions{Task: task.ID, Agent: "worker-2"}); envelope.Outcome != "ok" {
		t.Fatalf("fresh assign = %#v; want ok", envelope)
	}
	if envelope := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-2"}, nil); envelope.Outcome != "ok" {
		t.Fatalf("fresh complete = %#v; want ok", envelope)
	}
	if envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-2"}, nil); envelope.Outcome != "ok" {
		t.Fatalf("fresh verify = %#v; want ok", envelope)
	}
}

func TestVerifyFinishNotificationByteBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		reason  string
		wantLen int
		ok      bool
	}{
		{"4090-ascii", strings.Repeat("a", 4090), 4096, true},
		{"4091-ascii", strings.Repeat("a", 4091), 0, false},
		{"2045-utf8", strings.Repeat("\u00e9", 2045), 4096, true},
		{"2046-utf8", strings.Repeat("\u00e9", 2046), 0, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, task := newSeed(t, "completed")
			envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true, Reason: test.reason}, nil)
			if !test.ok {
				if envelope.Outcome != "failed" || envelope.Phase != "validate" {
					t.Fatalf("finish = %#v; want failed validate", envelope)
				}
				if len(envelope.Effects) != 0 {
					t.Fatalf("effects = %#v; want none", envelope.Effects)
				}
				if got := readTask(t, home, task.ID); got.Status != "completed" {
					t.Fatalf("task status = %q; want no task write", got.Status)
				}
				if got := len(finishInbox(t, home, "coordinator")); got != 0 {
					t.Fatalf("inbox records = %d; want none", got)
				}
				if readFinishAgent(t, home, "verifier-1").Finished {
					t.Fatal("verifier marked finished; want no mark")
				}
				return
			}
			if envelope.Outcome != "ok" {
				t.Fatalf("finish = %#v; want ok", envelope)
			}
			inboxRecords := finishInbox(t, home, "coordinator")
			if len(inboxRecords) != 1 {
				t.Fatalf("inbox = %#v; want one record", inboxRecords)
			}
			if len(inboxRecords[0].Text) != test.wantLen || !strings.HasPrefix(inboxRecords[0].Text, "fail: ") {
				t.Fatalf("text length = %d; want %d with fail prefix", len(inboxRecords[0].Text), test.wantLen)
			}
			if got := len(test.reason); got != test.wantLen-6 {
				t.Fatalf("reason length = %d; want %d", got, test.wantLen-6)
			}
		})
	}
}

func TestCompleteFinishCancelKeepsExistingReasonBehavior(t *testing.T) {
	home, task := newSeed(t, "completed")
	if envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: "verifier-1", Fail: true, Reason: "first"}, nil); envelope.Outcome != "ok" {
		t.Fatalf("verify fail = %#v; want ok", envelope)
	}
	reason := strings.Repeat("x", 5000)
	if envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: reason}); envelope.Outcome != "ok" {
		t.Fatalf("cancel = %#v; want ok without a new reason cap", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "cancelled" || got.CancelReason != reason {
		t.Fatalf("task = %#v; want cancelled with the plain reason", got)
	}
	if envelope := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: "worker-1"}, nil); envelope.Outcome != "failed" {
		t.Fatalf("complete after cancel = %#v; want failed", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "cancelled" {
		t.Fatalf("task status = %q; want cancelled", got.Status)
	}
}

func TestCompleteFinishConcurrentAttempts(t *testing.T) {
	home, task := newSeed(t, "assigned")
	var wg sync.WaitGroup
	results := make(chan output.Envelope, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-1"})
		}()
	}
	wg.Wait()
	close(results)
	ok, failed := 0, 0
	for envelope := range results {
		switch envelope.Outcome {
		case "ok":
			ok++
		case "failed":
			failed++
		default:
			t.Fatalf("finish = %#v; want ok or failed", envelope)
		}
	}
	if ok != 1 || failed != 1 {
		t.Fatalf("ok = %d, failed = %d; want one each", ok, failed)
	}
	if got := len(finishInbox(t, home, "coordinator")); got != 1 {
		t.Fatalf("inbox records = %d; want one", got)
	}
	if value, ok := finishCounter(t, home, "coordinator"); !ok || value != 1 {
		t.Fatalf("counter = %d, %v; want 1", value, ok)
	}
	if !readFinishAgent(t, home, "worker-1").Finished {
		t.Fatal("worker is not finished")
	}
}

func TestCompleteFinishConcurrentInboxReaders(t *testing.T) {
	home, task := newSeed(t, "assigned")
	const readers, reads = 4, 10
	var wg sync.WaitGroup
	errs := make(chan error, readers)
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range reads {
				if _, err := inbox.All(home, "coordinator"); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-1"})
		if envelope.Outcome != "ok" {
			errs <- errors.New("the finish did not complete")
		}
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if got := readTask(t, home, task.ID); got.Status != "completed" {
		t.Fatalf("task status = %q; want completed", got.Status)
	}
	inboxRecords := finishInbox(t, home, "coordinator")
	if len(inboxRecords) != 1 || inboxRecords[0].Kind != records.InboxCompleted || inboxRecords[0].ReadAt != "" {
		t.Fatalf("inbox = %#v; want one coherent completed record", inboxRecords)
	}
}

func TestCompleteFinishResolvedAliasRoutes(t *testing.T) {
	routes := []struct {
		name    string
		agent   string
		pane    string
		history bool
	}{
		{"key", "worker-key", "", false},
		{"name", "worker-name", "", false},
		{"pane", "", "w1:p7", false},
		{"history", "", "w1:p0", true},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			home, task := newSeed(t, "assigned")
			record := records.Agent{
				Key: "worker-key", Name: "worker-name", Role: "worker", Project: "one", Task: task.ID,
				Status: records.AgentLive, PaneID: "w1:p7", Generation: 1,
			}
			if route.history {
				record.PaneHistory = []string{"w1:p0"}
			}
			seedAgent(t, home, record)
			task.ActingAgent = "worker-key"
			seedTask(t, home, task)
			envelope := runFinishLocked(t, home, "complete", TransitionOptions{Task: task.ID, Agent: route.agent, PaneID: route.pane}, nil)
			if envelope.Outcome != "ok" || envelope.Error != nil {
				t.Fatalf("complete through %s = %#v; want ok", route.name, envelope)
			}
			if got := readTask(t, home, task.ID); got.Status != "completed" || got.ActingAgent != "worker-key" {
				t.Fatalf("task = %#v; want completed with the stored acting agent", got)
			}
			if !readFinishAgent(t, home, "worker-key").Finished {
				t.Fatal("the resolved worker is not finished")
			}
			inboxRecords := finishInbox(t, home, "coordinator")
			if len(inboxRecords) != 1 || inboxRecords[0].Agent != "worker-name" || inboxRecords[0].Text != "completed" {
				t.Fatalf("inbox = %#v; want the stored name in the notification", inboxRecords)
			}
		})
	}
}

func TestVerifyFinishResolvedAliasRoutes(t *testing.T) {
	for _, route := range []struct {
		name  string
		agent string
		pane  string
	}{
		{"key", "verifier-key", ""},
		{"name", "verifier-name", ""},
		{"pane", "", "w1:p8"},
	} {
		t.Run(route.name, func(t *testing.T) {
			home, task := newSeed(t, "completed")
			seedAgent(t, home, records.Agent{
				Key: "verifier-key", Name: "verifier-name", Role: "verifier", Project: "one", Task: task.ID,
				Status: records.AgentLive, PaneID: "w1:p8", Generation: 1,
			})
			envelope := runFinishLocked(t, home, "verify", TransitionOptions{Task: task.ID, Agent: route.agent, PaneID: route.pane}, nil)
			if envelope.Outcome != "ok" || envelope.Error != nil {
				t.Fatalf("verify through %s = %#v; want ok", route.name, envelope)
			}
			got := readTask(t, home, task.ID)
			if got.Status != "verified" || got.Verification == nil || got.Verification.Agent != "verifier-name" {
				t.Fatalf("task = %#v; want the stored verifier name", got)
			}
			if !readFinishAgent(t, home, "verifier-key").Finished {
				t.Fatal("the resolved verifier is not finished")
			}
		})
	}
}

func TestAdvancedPartialRetriesIncludeInspection(t *testing.T) {
	cases := []struct {
		name    string
		status  string
		command string
		agent   string
		fail    bool
		reason  string
	}{
		{"complete", "completed", "complete", "worker-1", false, ""},
		{"verify pass", "verified", "verify", "verifier-1", false, ""},
		{"verify fail", "assigned", "verify", "verifier-1", true, "changed reason"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, task := newSeed(t, test.status)
			before := taskRecordBytes(t, home, task.ID)
			envelope := runFinishLocked(t, home, test.command, TransitionOptions{Task: task.ID, Agent: test.agent, Fail: test.fail, Reason: test.reason}, nil)
			if envelope.Outcome != "failed" || envelope.Phase != "task" || len(envelope.Effects) != 0 {
				t.Fatalf("retry = %#v; want a no-effect task refusal", envelope)
			}
			message, _ := envelope.Error.(string)
			for _, want := range []string{"invalid transition:", "task show", "agent show", "inbox"} {
				if !strings.Contains(message, want) {
					t.Fatalf("error = %q; want %q", message, want)
				}
			}
			if after := taskRecordBytes(t, home, task.ID); string(after) != string(before) {
				t.Fatal("advanced retry changed the task record")
			}
			if value, ok := finishCounter(t, home, "coordinator"); ok {
				t.Fatalf("counter = %d; want no counter write", value)
			}
			if got := len(finishInbox(t, home, "coordinator")); got != 0 {
				t.Fatalf("inbox records = %d; want none", got)
			}
			if readFinishAgent(t, home, test.agent).Finished {
				t.Fatal("advanced retry marked the caller finished")
			}
		})
	}
}
