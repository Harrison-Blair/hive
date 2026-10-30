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

// TestPiExtensionKeepsACharacterSplitAcrossWrites runs hive-inbox.ts under
// node and writes one line in two writes that split the ⬢ marker.
func TestPiExtensionKeepsACharacterSplitAcrossWrites(t *testing.T) {
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

	socket := filepath.Join(runtime, "hive-pi", strconv.Itoa(cmd.Process.Pid)+".sock")
	var conn net.Conn
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(50 * time.Millisecond) {
		if conn, err = net.Dial("unix", socket); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("no extension socket at %s: %v", socket, err)
		}
	}
	line := []byte(`{"text":"⬢ boss: héllo"}` + "\n")
	split := 10 // inside the 3 bytes of ⬢, which start at byte 9
	if _, err := conn.Write(line[:split]); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if _, err := conn.Write(line[split:]); err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()

	got := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			got <- scanner.Text()
		}
	}()
	select {
	case out := <-got:
		var message struct{ Text string }
		if err := json.Unmarshal([]byte(out), &message); err != nil || message.Text != "⬢ boss: héllo" {
			t.Fatalf("sendUserMessage got %s; want text ⬢ boss: héllo", out)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the extension sent no message")
	}
}
