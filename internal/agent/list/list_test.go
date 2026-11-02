package list

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func listRun(t *testing.T, home string) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Run(ctx, home, Options{})
}

func TestListShowsTheStatusOfLiveRecords(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	live := fake.AddPane("w1", "")
	fake.StartAgent(live.ID, "alive", "pi", "working")
	if err := agent.Save(home, agent.Record{Name: "alive", Pane: live.ID, Harness: "pi", Model: "m"}); err != nil {
		t.Fatal(err)
	}

	envelope := listRun(t, home)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	entries := envelope.Result.([]Entry)
	if len(entries) != 1 || entries[0].Name != "alive" || entries[0].Status != "working" {
		t.Fatalf("entries = %+v", entries)
	}
	text := Text(envelope)
	for _, want := range []string{"alive", "working", live.ID} {
		if !strings.Contains(text, want) {
			t.Fatalf("text = %q, want %q", text, want)
		}
	}
}

// now is the clock of the list tests.
var now = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func listAt(t *testing.T, home string, logger *slog.Logger) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Run(ctx, home, Options{Logger: logger, Now: func() time.Time { return now }})
}

func save(t *testing.T, home string, record agent.Record) {
	t.Helper()
	if err := agent.Save(home, record); err != nil {
		t.Fatal(err)
	}
}

func names(envelope output.Envelope) []string {
	var out []string
	for _, entry := range envelope.Result.([]Entry) {
		out = append(out, entry.Name)
	}
	return out
}

func removedEffect(envelope output.Envelope, name string) bool {
	for _, effect := range envelope.Effects {
		if effect == (output.Effect{Resource: "record", Path: "agents/" + name + ".json", Action: "removed"}) {
			return true
		}
	}
	return false
}

func recordExists(t *testing.T, home, name string) bool {
	t.Helper()
	_, err := agent.Load(home, name)
	if err != nil && !errors.Is(err, agent.ErrNotFound) {
		t.Fatal(err)
	}
	return err == nil
}

func TestListRemovesTheRecordOfAGonePane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	gone := fake.AddPane("w1", "")
	save(t, home, agent.Record{Name: "dead", Pane: gone.ID, Harness: "claude", StartedAt: now})
	fake.ClosePane(gone.ID)
	var logs bytes.Buffer

	envelope := listAt(t, home, slog.New(slog.NewJSONHandler(&logs, nil)))
	if envelope.Outcome != "ok" || len(names(envelope)) != 0 || !removedEffect(envelope, "dead") || len(envelope.Effects) != 1 {
		t.Fatalf("envelope = %+v", envelope)
	}
	if recordExists(t, home, "dead") {
		t.Fatal("record dead stays")
	}
	if !strings.Contains(logs.String(), `"target":"dead"`) || !strings.Contains(logs.String(), `"reason":"pane gone"`) {
		t.Fatalf("logs = %s; want an event with target dead and reason pane gone", logs.String())
	}
	if Text(envelope) != "no agents\n" {
		t.Fatalf("text = %q", Text(envelope))
	}
}

func TestListRemovesAnOldRecordOfAShellPane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, pane := fake.AddWorkspace("mine", "", false)
	fake.SetProcess(pane.ID, "bash")
	save(t, home, agent.Record{Name: "exited", Pane: pane.ID, Harness: "claude", StartedAt: now.Add(-61 * time.Second)})
	var logs bytes.Buffer

	envelope := listAt(t, home, slog.New(slog.NewJSONHandler(&logs, nil)))
	if envelope.Outcome != "ok" || len(names(envelope)) != 0 || !removedEffect(envelope, "exited") {
		t.Fatalf("envelope = %+v", envelope)
	}
	if recordExists(t, home, "exited") {
		t.Fatal("record exited stays")
	}
	if !strings.Contains(logs.String(), `"target":"exited"`) || !strings.Contains(logs.String(), `"reason":"no harness process"`) {
		t.Fatalf("logs = %s; want an event with target exited and reason no harness process", logs.String())
	}
}

func TestListKeepsAYoungRecordOfAShellPane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, pane := fake.AddWorkspace("mine", "", false)
	fake.SetProcess(pane.ID, "hive")
	save(t, home, agent.Record{Name: "starting", Pane: pane.ID, Harness: "claude", StartedAt: now.Add(-59 * time.Second)})

	envelope := listAt(t, home, nil)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 || strings.Join(names(envelope), ",") != "starting" || !recordExists(t, home, "starting") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestListKeepsALiveHarnessWithStatusUnknown(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, pane := fake.AddWorkspace("mine", "", false)
	fake.StartAgent(pane.ID, "kid", "opencode", "unknown")
	fake.SetProcess(pane.ID, "claude")
	save(t, home, agent.Record{Name: "kid", Pane: pane.ID, Harness: "claude", StartedAt: now.Add(-time.Hour)})

	envelope := listAt(t, home, nil)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 || strings.Join(names(envelope), ",") != "kid" || !recordExists(t, home, "kid") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestListKeepsARecordThatANewAgentReplaced(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	gone := fake.AddPane("w1", "")
	fresh := fake.AddPane("w1", "")
	save(t, home, agent.Record{Name: "kid", Pane: gone.ID, Harness: "claude", StartedAt: now.Add(-time.Hour)})
	fake.ClosePane(gone.ID)
	// A spawn takes the name after list read the old record.
	fake.After("pane.get", func() {
		save(t, home, agent.Record{Name: "kid", Pane: fresh.ID, Harness: "claude", StartedAt: now})
	})

	envelope := listAt(t, home, nil)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got, err := agent.Load(home, "kid"); err != nil || got.Pane != fresh.ID {
		t.Fatalf("Load = %+v, %v; want the record of %s", got, err, fresh.ID)
	}
}

func TestListKeepsTheRecordWhenProcessInfoFails(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, pane := fake.AddWorkspace("mine", "", false)
	save(t, home, agent.Record{Name: "kid", Pane: pane.ID, Harness: "claude", StartedAt: now.Add(-time.Hour)})
	fake.Fail("pane.process_info", "server_error", "boom")

	envelope := listAt(t, home, nil)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 || strings.Join(names(envelope), ",") != "kid" || !recordExists(t, home, "kid") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestListJSONHasHarnessSessionID(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	withSession := fake.AddPane("w1", "")
	fake.StartAgent(withSession.ID, "a", "claude", "idle")
	fake.SetSession(withSession.ID, "6f1c-claude")
	without := fake.AddPane("w1", "")
	fake.StartAgent(without.ID, "b", "pi", "idle")
	for _, record := range []agent.Record{{Name: "a", Pane: withSession.ID}, {Name: "b", Pane: without.ID}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}

	envelope := listRun(t, home)
	data, err := json.Marshal(envelope.Result)
	if err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0]["harness_session_id"] != "6f1c-claude" || got[1]["harness_session_id"] != "" {
		t.Fatalf("entries = %s", data)
	}
}

func TestListEmpty(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	envelope := listRun(t, home)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if entries := envelope.Result.([]Entry); entries == nil || len(entries) != 0 {
		t.Fatalf("entries = %#v, want an empty list", entries)
	}
}

func TestListFailsWhenHerdrFails(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: "w1:p1"}); err != nil {
		t.Fatal(err)
	}
	fake.Fail("pane.get", "server_error", "boom")
	envelope := listRun(t, home)
	if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), "boom") {
		t.Fatalf("envelope = %+v", envelope)
	}
}
