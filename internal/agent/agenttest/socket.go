package agenttest

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// ListenClaude sets XDG_RUNTIME_DIR to a new directory and listens on the
// inbox socket of the Claude Code process pid in it. It returns the channel of
// the lines that arrive, one line for each connection.
func ListenClaude(t testing.TB, pid int) <-chan string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)
	if err := os.MkdirAll(filepath.Join(dir, "cc-socks"), 0o700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", filepath.Join(dir, "cc-socks", strconv.Itoa(pid)+".sock"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	lines := make(chan string, 8)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			line, _ := bufio.NewReader(conn).ReadString('\n')
			_ = conn.Close()
			lines <- line
		}
	}()
	return lines
}
