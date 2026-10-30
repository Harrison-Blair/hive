package harness

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// fakePi loads the extension with a fake pi API, starts a session, and prints
// each message of pi.sendUserMessage as one JSON line.
const fakePi = `
const ext = await import(process.argv[2]);
const handlers = {};
ext.default({
	on: (event, handler) => { handlers[event] = handler; },
	sendUserMessage: (text, options) => console.log(JSON.stringify({ text, options })),
});
handlers.session_start({}, { isIdle: () => true });
`

// runExtension runs hive-inbox.ts under node with a fake pi. It returns the
// socket of the extension and the channel of the texts that the extension
// passes to pi.sendUserMessage.
func runExtension(t *testing.T) (string, <-chan string) {
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
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	texts := make(chan string, 8)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var message struct{ Text string }
			if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
				message.Text = "unreadable output: " + scanner.Text()
			}
			texts <- message.Text
		}
		close(texts)
	}()
	socket := filepath.Join(runtime, "hive-pi", strconv.Itoa(cmd.Process.Pid)+".sock")
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(50 * time.Millisecond) {
		if _, err := os.Stat(socket); err == nil {
			return socket, texts
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

func nextText(t *testing.T, texts <-chan string) string {
	t.Helper()
	select {
	case text, ok := <-texts:
		if !ok {
			t.Fatal("the extension process ended")
		}
		return text
	case <-time.After(10 * time.Second):
		t.Fatal("the extension sent no message")
	}
	return ""
}

// TestPiExtensionKeepsACharacterSplitAcrossWrites writes one line in two
// writes that split the ⬢ marker.
func TestPiExtensionKeepsACharacterSplitAcrossWrites(t *testing.T) {
	socket, texts := runExtension(t)
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
	socket, texts := runExtension(t)
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
