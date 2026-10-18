package store

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestReadInPathTypesLinux(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(dir, "fifo.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "socket.json")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	for id, path := range map[string]string{"fifo": fifo, "socket": socket} {
		var record map[string]any
		if err := store.ReadIn([]string{"tasks"}, id, &record); err == nil || err.Error() != "record invalid: "+path {
			t.Errorf("ReadIn(%q) error = %v; want record invalid", id, err)
		}
		if err := store.WriteIn([]string{"tasks"}, id, map[string]string{"id": id}); err == nil || !strings.Contains(err.Error(), "record invalid") {
			t.Errorf("WriteIn(%q) error = %v; want record invalid", id, err)
		}
	}
}
