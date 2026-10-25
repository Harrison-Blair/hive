package chat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

// TestTUISendStillWritesUserRecordAndPromptsCoordinator proves AC5: the path
// that the TUI uses keeps its user record and its coordinator prompt.
func TestTUISendStillWritesUserRecordAndPromptsCoordinator(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p2", Text: "from the TUI"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(SendResult)
	if !ok || result.ID != "c1" || result.Delivery != "sent" {
		t.Fatalf("result = %#v; want c1 sent", envelope.Result)
	}
	logEnvelope := Log(home, LogOptions{})
	logResult, ok := logEnvelope.Result.(LogResult)
	if logEnvelope.Outcome != "ok" || !ok || len(logResult.Entries) != 1 {
		t.Fatalf("Log() = %#v; want one user record", logEnvelope)
	}
	if entry := logResult.Entries[0]; entry.From != "user" || entry.Text != "from the TUI" || entry.Delivery != "sent" {
		t.Fatalf("entry = %#v; want the user message with sent delivery", entry)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt calls = %d; want one", len(prompts))
	}
	var params herdr.AgentPromptParams
	if err := json.Unmarshal(prompts[0].Params, &params); err != nil {
		t.Fatal(err)
	}
	if params.Target != "w1:p1" || params.Text != "⬢ user: from the TUI" {
		t.Fatalf("prompt = %#v; want one marked submission to the coordinator pane", params)
	}
}
