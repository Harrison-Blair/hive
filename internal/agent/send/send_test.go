package send

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
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

// setup makes a claude target agent "kid" with an inbox socket and a caller
// pane with the record "boss". It returns the lines that reach the socket.
func setup(t *testing.T) (*agenttest.Herdr, string, string, string, <-chan string) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	_, bossPane := fake.AddWorkspace("mine", "", false)
	kidPane := fake.AddPane("w1", "")
	for _, record := range []agent.Record{{Name: "boss", Pane: bossPane.ID, Harness: "claude"}, {Name: "kid", Pane: kidPane.ID, Harness: "claude"}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	profile, err := harness.Lookup("claude")
	if err != nil {
		t.Fatal(err)
	}
	fake.StartAgent(kidPane.ID, "kid", "claude", "idle")
	fake.SetArgs(kidPane.ID, profile.Args("m", nil))
	return fake, home, bossPane.ID, kidPane.ID, agenttest.ListenClaude(t, kidPane.PID)
}

// received returns the content of each message line that reached lines.
func received(t *testing.T, lines <-chan string) []string {
	t.Helper()
	var out []string
	for {
		select {
		case line := <-lines:
			const prefix, suffix = `{"type":"user","message":{"role":"user","content":"`, "\"}}\n"
			if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
				t.Fatalf("line = %q; want one user message line", line)
			}
			out = append(out, strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix))
		case <-time.After(200 * time.Millisecond):
			return out
		}
	}
}

// noTyping fails when send typed into a pane.
func noTyping(t *testing.T, fake *agenttest.Herdr) {
	t.Helper()
	for _, method := range []string{"agent.prompt", "pane.send_input", "pane.send_text"} {
		if got := fake.Requests(method); len(got) != 0 {
			t.Fatalf("%s calls = %d; send must not type into a pane", method, len(got))
		}
	}
}

func TestSendDeliversThroughTheHarnessSocket(t *testing.T) {
	fake, home, bossPane, kidPane, lines := setup(t)
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hello there", CallerPane: bossPane})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := received(t, lines); !slices.Equal(got, []string{"⬢ boss: hello there"}) {
		t.Fatalf("messages = %q", got)
	}
	noTyping(t, fake)
	if result := envelope.Result.(Result); result.Pane != kidPane || result.Sender != "boss" {
		t.Fatalf("result = %+v", result)
	}
}

func TestSendWithoutCallerRecordUsesUser(t *testing.T) {
	_, home, _, _, lines := setup(t)
	for _, caller := range []string{"", "w9:p9"} {
		if envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: caller}); envelope.Outcome != "ok" {
			t.Fatalf("envelope = %+v", envelope)
		}
	}
	if got := received(t, lines); !slices.Equal(got, []string{"⬢ user: hi", "⬢ user: hi"}) {
		t.Fatalf("messages = %q", got)
	}
}

func TestSendFailsClearlyWithoutTheNativePath(t *testing.T) {
	fake, home, bossPane, kidPane, lines := setup(t)
	fake.SetArgs(kidPane, []string{"--dangerously-skip-permissions"})
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || envelope.Phase != "deliver" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if text := errorText(envelope); !strings.Contains(text, "agent kid") || !strings.Contains(text, "restart") {
		t.Fatalf("error = %q; want the agent and the fix", text)
	}
	if got := received(t, lines); len(got) != 0 {
		t.Fatalf("messages = %q; want none", got)
	}
	noTyping(t, fake)
}

func TestSendUnknownName(t *testing.T) {
	fake, home, bossPane, _, _ := setup(t)
	envelope := sendRun(t, home, Options{Name: "ghost", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "unknown agent: ghost") {
		t.Fatalf("envelope = %+v", envelope)
	}
	noTyping(t, fake)
}

func TestSendGonePane(t *testing.T) {
	fake, home, bossPane, kidPane, _ := setup(t)
	fake.ClosePane(kidPane)
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "gone") || !strings.Contains(errorText(envelope), kidPane) {
		t.Fatalf("envelope = %+v", envelope)
	}
}

// A harness that queues messages behind a dialog takes the message.
func TestSendDeliversToABlockedTarget(t *testing.T) {
	fake, home, bossPane, kidPane, lines := setup(t)
	fake.SetStatus(kidPane, "blocked")
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", CallerPane: bossPane})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := received(t, lines); !slices.Equal(got, []string{"⬢ boss: hi"}) {
		t.Fatalf("messages = %q", got)
	}
}

func TestSendEmptyText(t *testing.T) {
	_, home, bossPane, _, _ := setup(t)
	envelope := sendRun(t, home, Options{Name: "kid", Text: " ", CallerPane: bossPane})
	if envelope.Outcome != "failed" || !strings.Contains(errorText(envelope), "TEXT") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func untilOf(t *testing.T, request agenttest.Request) []string {
	t.Helper()
	until := []string{}
	raw, _ := request.Params["until"].([]any)
	for _, status := range raw {
		until = append(until, status.(string))
	}
	slices.Sort(until)
	return until
}

func TestSendWaitWaitsForTheTurnToStartAndEnd(t *testing.T) {
	fake, home, bossPane, _, lines := setup(t)
	fake.WaitStatus = "done"
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", Wait: true, CallerPane: bossPane})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := received(t, lines); len(got) != 1 {
		t.Fatalf("messages = %q", got)
	}
	waits := fake.Requests("agent.wait")
	if len(waits) != 2 {
		t.Fatalf("agent.wait calls = %d; want 2", len(waits))
	}
	if got := untilOf(t, waits[0]); !slices.Equal(got, []string{"working"}) {
		t.Fatalf("first until = %q; want working", got)
	}
	if got := untilOf(t, waits[1]); !slices.Equal(got, []string{"blocked", "done", "idle"}) {
		t.Fatalf("second until = %q", got)
	}
	if result := envelope.Result.(Result); result.Status != "done" {
		t.Fatalf("result = %+v, want status done", result)
	}
}

func TestSendWaitEndsBlocked(t *testing.T) {
	fake, home, bossPane, kidPane, _ := setup(t)
	fake.SetStatus(kidPane, "blocked")
	envelope := sendRun(t, home, Options{Name: "kid", Text: "hi", Wait: true, CallerPane: bossPane})
	if envelope.Outcome != "partial" || !strings.Contains(errorText(envelope), "blocked") || !strings.Contains(errorText(envelope), kidPane) {
		t.Fatalf("envelope = %+v", envelope)
	}
}
