package herdr

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestWithRedactHidesTheValueInTheRequestAndReplyLog(t *testing.T) {
	server := newFakeServer(t)
	// herdr writes the reply without HTML escapes: "<" stays "<".
	server.scriptResult("agent.start", `{"type":"agent_started","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"unknown","focused":false,"revision":1,"name":"kid","agent":"claude"},"argv":["claude","--append-system-prompt","say \"hi\" <now> c:\\d"]}`)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	secret := `say "hi" <now> c:\d`
	_, err := AgentStart(context.Background(),
		AgentStartParams{Name: "kid", Kind: "claude", PaneID: "w1:p1", Args: []string{"--append-system-prompt", secret}, TimeoutMS: 1000},
		WithLogger(logger), WithRedact(secret, "<guide 19 bytes>"))
	if err != nil {
		t.Fatalf("AgentStart error = %v", err)
	}
	text := logs.String()
	if strings.Count(text, `<guide 19 bytes>`) != 2 {
		t.Fatalf("logs = %s; want the placeholder in the request and the reply", text)
	}
	for _, leak := range []string{"say ", "<now>", `\u003cnow`} {
		if strings.Contains(text, leak) {
			t.Fatalf("logs = %s; want no %q", text, leak)
		}
	}
	if got := server.recorded("agent.start")[0].Params; !strings.Contains(got, `say \"hi\" \u003cnow\u003e c:\\d`) {
		t.Fatalf("params = %s; want the value sent unchanged", got)
	}
}

func TestWithReplyRedactHidesTheValuesThatTheReplyGives(t *testing.T) {
	server := newFakeServer(t)
	server.scriptResult("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","foreground_process_group_id":7,"foreground_processes":[{"argv":["claude","--append-system-prompt","the <guide>"],"cmdline":"claude --append-system-prompt the <guide>","name":"claude","pid":7}]}}`)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	find := func(reply []byte) map[string]string {
		if bytes.Contains(reply, []byte("the <guide>")) {
			return map[string]string{"the <guide>": "<guide 11 bytes>"}
		}
		return nil
	}
	info, err := PaneProcessInfo(context.Background(), "w1:p1", WithLogger(logger), WithReplyRedact(find))
	if err != nil {
		t.Fatalf("PaneProcessInfo error = %v", err)
	}
	if got := info.ForegroundProcesses[0].Argv[2]; got != "the <guide>" {
		t.Fatalf("argv = %q; want the value unchanged in the result", got)
	}
	text := logs.String()
	if strings.Contains(text, "the <guide>") || strings.Contains(text, `the <guide`) {
		t.Fatalf("logs = %s; want no guide", text)
	}
	if strings.Count(text, "<guide 11 bytes>") != 2 {
		t.Fatalf("logs = %s; want the placeholder in argv and cmdline", text)
	}
}
