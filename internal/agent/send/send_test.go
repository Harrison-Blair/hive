package send

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func sendRun(t *testing.T, home string, options Options) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Run(ctx, home, options)
}

func errorText(envelope output.Envelope) string {
	if envelope.Error == nil {
		return ""
	}
	return envelope.Error.(string)
}

// setup makes a target agent "kid" and a caller pane with the record "boss".
func setup(t *testing.T) (*agenttest.Herdr, string, string, string) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, bossPane := fake.AddWorkspace("mine", "", false)
	kidPane := fake.AddPane("w1", "")
	for _, record := range []agent.Record{{Name: "boss", Pane: bossPane.ID}, {Name: "kid", Pane: kidPane.ID}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	// The fake takes a prompt only for a pane that holds an agent.
	fake.StartAgent(kidPane.ID, "kid", "pi", "idle")
	return fake, home, bossPane.ID, kidPane.ID
}

func TestSendPrefixesTheCallerName(t *testing.T) {
	fake, home, bossPane, kidPane := setup(t)
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hello there", CallerPane: bossPane})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	pane, _ := fake.Pane(kidPane)
	if !slices.Equal(pane.Prompts, []string{"⬢ boss: hello there"}) {
		t.Fatalf("prompts = %q", pane.Prompts)
	}
	requests := fake.Requests("agent.prompt")
	if requests[0].Params["target"] != kidPane {
		t.Fatalf("target = %v, want pane %s", requests[0].Params["target"], kidPane)
	}
	if _, has := requests[0].Params["wait"]; has {
		t.Fatalf("prompt without --wait has a wait: %+v", requests[0].Params)
	}
}

func TestSendWithoutCallerRecordUsesUser(t *testing.T) {
	fake, home, _, kidPane := setup(t)
	for _, caller := range []string{"", "w9:p9"} {
		if envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: caller}); envelope.Outcome != "ok" {
			t.Fatalf("envelope = %+v", envelope)
		}
	}
	pane, _ := fake.Pane(kidPane)
	if !slices.Equal(pane.Prompts, []string{"⬢ user: hi", "⬢ user: hi"}) {
		t.Fatalf("prompts = %q", pane.Prompts)
	}
}

func TestSendUnknownName(t *testing.T) {
	fake, home, bossPane, _ := setup(t)
	envelope := sendRun(t, home, Options{Name: "ghost", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "unknown agent: ghost") {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := fake.Requests("agent.prompt"); len(got) != 0 {
		t.Fatalf("agent.prompt calls = %d", len(got))
	}
}

func TestSendGonePane(t *testing.T) {
	fake, home, bossPane, kidPane := setup(t)
	fake.ClosePane(kidPane)
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "gone") || !strings.Contains(errorText(envelope), kidPane) {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestSendBlockedTarget(t *testing.T) {
	fake, home, bossPane, kidPane := setup(t)
	fake.SetStatus(kidPane, "blocked")
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "blocked") || !strings.Contains(errorText(envelope), kidPane) {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestSendEmptyText(t *testing.T) {
	_, home, bossPane, _ := setup(t)
	envelope := sendRun(t, home, Options{Name: "kid", Text: " ", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "TEXT") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestSendWaitWaitsForIdleOrDone(t *testing.T) {
	fake, home, bossPane, _ := setup(t)
	fake.WaitStatus = "done"
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", Wait: true, CallerPane: bossPane})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	wait, ok := fake.Requests("agent.prompt")[0].Params["wait"].(map[string]any)
	if !ok {
		t.Fatalf("prompt has no wait")
	}
	until := []string{}
	for _, status := range wait["until"].([]any) {
		until = append(until, status.(string))
	}
	slices.Sort(until)
	if !slices.Equal(until, []string{"blocked", "done", "idle"}) {
		t.Fatalf("until = %q", until)
	}
	if result := envelope.Result.(Result); result.Status != "done" {
		t.Fatalf("result = %+v, want status done", result)
	}
}

func TestSendWaitEndsBlocked(t *testing.T) {
	fake, home, bossPane, kidPane := setup(t)
	fake.WaitStatus = "blocked"
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", Wait: true, CallerPane: bossPane})
	if envelope.Outcome == "ok" || !strings.Contains(errorText(envelope), "blocked") || !strings.Contains(errorText(envelope), kidPane) {
		t.Fatalf("envelope = %+v", envelope)
	}
}
