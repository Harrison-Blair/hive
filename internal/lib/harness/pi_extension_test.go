package harness

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// fakePi loads the extension with a fake pi API and starts a session. Its
// sendUserMessage acts as AgentSession.prompt of pi 1.4.0 does: after an
// await, it rejects every message during a compaction, and a message without
// deliverAs during a run; else it starts a run of 300 ms or queues a
// follow-up. Each outcome is one JSON line on stdout. A stdin line "compact"
// runs a manual compaction of 500 ms.
const fakePi = `
const ext = await import(process.argv[2]);
const handlers = {};
let running = false;
let compacting = false;
const emit = (event) => handlers[event]?.({ type: event }, context);
const context = { isIdle: () => !running && !compacting };
const out = (event, text) => console.log(JSON.stringify({ event, text }));
ext.default({
	on: (event, handler) => { handlers[event] = handler; },
	sendUserMessage: async (text, options) => {
		await new Promise((resolve) => setTimeout(resolve, 20));
		if (compacting) return out("rejected", text);
		if (running) return out(options?.deliverAs ? "followUp" : "rejected", text);
		running = true;
		emit("agent_start");
		out("started", text);
		setTimeout(() => { running = false; emit("agent_settled"); }, 300);
	},
});
handlers.session_start({}, context);
process.stdin.on("data", () => {
	compacting = true;
	out("compacting", "");
	setTimeout(() => { compacting = false; emit("session_compact"); }, 500);
});
`

// event is one line of fakePi.
type event struct{ Event, Text string }

// runExtension runs hive-inbox.ts under node with fakePi. It returns the
// socket of the extension, the stdin of fakePi, and the channel of its events.
func runExtension(t *testing.T) (string, io.Writer, <-chan event) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	dir := t.TempDir()
	runtime := filepath.Join(dir, "run")
	extension := filepath.Join(dir, "hive-inbox.ts")
	runner := filepath.Join(dir, "fake-pi.mjs")
	if err := os.WriteFile(extension, inboxExtension, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runner, []byte(fakePi), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, runner, extension)
	cmd.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+runtime)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	events := make(chan event, 16)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var e event
			if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
				e.Event = "unreadable output: " + scanner.Text()
			}
			events <- e
		}
		close(events)
	}()
	socket := filepath.Join(runtime, "hive-pi", strconv.Itoa(cmd.Process.Pid)+".sock")
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(50 * time.Millisecond) {
		if _, err := os.Stat(socket); err == nil {
			return socket, stdin, events
		}
		if time.Now().After(deadline) {
			t.Fatalf("no extension socket at %s", socket)
		}
	}
}

// writeTo writes each part to socket in its own write, 200 ms apart, and
// closes the connection.
func writeTo(t *testing.T, socket string, parts ...string) {
	t.Helper()
	conn, err := net.Dial("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	for i, part := range parts {
		if i > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		if _, err := conn.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
}

func nextEvent(t *testing.T, events <-chan event) event {
	t.Helper()
	select {
	case e, ok := <-events:
		if !ok {
			t.Fatal("the extension process ended")
		}
		return e
	case <-time.After(10 * time.Second):
		t.Fatal("the extension sent no message")
	}
	return event{}
}

// nextText returns the text of the next event, and fails the test unless pi
// accepted it.
func nextText(t *testing.T, events <-chan event) string {
	t.Helper()
	e := nextEvent(t, events)
	if e.Event != "started" && e.Event != "followUp" {
		t.Fatalf("pi event = %+v; want an accepted message", e)
	}
	return e.Text
}

// TestPiExtensionKeepsACharacterSplitAcrossWrites writes one line in two
// writes that split the ⬢ marker.
func TestPiExtensionKeepsACharacterSplitAcrossWrites(t *testing.T) {
	socket, _, texts := runExtension(t)
	line := `{"text":"⬢ boss: héllo"}` + "\n"
	// ⬢ takes bytes 9 to 11.
	writeTo(t, socket, line[:10], line[10:])
	if got := nextText(t, texts); got != "⬢ boss: héllo" {
		t.Fatalf("sendUserMessage got %q; want ⬢ boss: héllo", got)
	}
}

// TestPiExtensionSurvivesBadRecords sends records cut before their newline, a
// line that is not JSON, and a line without text, and checks that the
// extension drops them, keeps running, and still delivers the valid lines.
func TestPiExtensionSurvivesBadRecords(t *testing.T) {
	socket, _, texts := runExtension(t)
	writeTo(t, socket, `{"text":"⬢ boss: cut off`)
	writeTo(t, socket, `{"text":"⬢ boss: no newline"}`)
	writeTo(t, socket, "not json\n", `{"other":1}`+"\n", `{"text":"⬢ boss: after"}`+"\n")
	writeTo(t, socket, `{"text":"⬢ boss: next"}`+"\n")
	for _, want := range []string{"⬢ boss: after", "⬢ boss: next"} {
		if got := nextText(t, texts); got != want {
			t.Fatalf("sendUserMessage got %q; want %q", got, want)
		}
	}
}

// TestPiExtensionSerializesMessagesToAnIdlePi sends three lines at once to an
// idle pi. The first starts a run, and the others wait for that run and go as
// follow-ups, so pi rejects none.
func TestPiExtensionSerializesMessagesToAnIdlePi(t *testing.T) {
	socket, _, events := runExtension(t)
	writeTo(t, socket, `{"text":"one"}`+"\n"+`{"text":"two"}`+"\n"+`{"text":"three"}`+"\n")
	for _, want := range []event{{"started", "one"}, {"followUp", "two"}, {"followUp", "three"}} {
		if got := nextEvent(t, events); got != want {
			t.Fatalf("pi event = %+v; want %+v", got, want)
		}
	}
}

// TestPiExtensionHoldsMessagesDuringCompaction sends a line while pi runs a
// manual compaction, which rejects every message, and checks that the
// extension delivers it after the compaction.
func TestPiExtensionHoldsMessagesDuringCompaction(t *testing.T) {
	socket, stdin, events := runExtension(t)
	if _, err := io.WriteString(stdin, "compact\n"); err != nil {
		t.Fatal(err)
	}
	if got := nextEvent(t, events); got.Event != "compacting" {
		t.Fatalf("pi event = %+v; want compacting", got)
	}
	writeTo(t, socket, `{"text":"held"}`+"\n")
	if got, want := nextEvent(t, events), (event{"started", "held"}); got != want {
		t.Fatalf("pi event = %+v; want %+v", got, want)
	}
}
