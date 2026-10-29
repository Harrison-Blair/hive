package list

import (
	"context"
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

func TestListReconcilesRecordsWithHerdr(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	fake.AddWorkspace("mine", "", false)
	live := fake.AddPane("w1", "")
	fake.StartAgent(live.ID, "alive", "pi", "working")
	gone := fake.AddPane("w1", "")
	for _, record := range []agent.Record{
		{Name: "alive", Pane: live.ID, Harness: "pi", Model: "m"},
		{Name: "dead", Pane: gone.ID, Harness: "claude", Model: "n", Parent: "alive"},
	} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	fake.ClosePane(gone.ID)

	envelope := listRun(t, home)
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	entries := envelope.Result.([]Entry)
	if len(entries) != 2 || entries[0].Name != "alive" || entries[0].Status != "working" || entries[1].Name != "dead" || entries[1].Status != "gone" {
		t.Fatalf("entries = %+v", entries)
	}
	text := Text(envelope)
	for _, want := range []string{"alive", "working", "dead", "gone", live.ID} {
		if !strings.Contains(text, want) {
			t.Fatalf("text = %q, want %q", text, want)
		}
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
