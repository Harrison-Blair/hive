package store

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWalkSpecialNodes(t *testing.T) {
	tests := []struct {
		name       string
		collection []string
		filename   string
		socket     bool
		want       string
	}{
		{name: "root_fifo", filename: "fifo", want: "store path invalid"},
		{name: "root_socket", filename: "sock", socket: true, want: "store path invalid"},
		{name: "nested_fifo", collection: []string{"tasks"}, filename: "fifo", want: "store path invalid"},
		{name: "nested_socket", collection: []string{"tasks"}, filename: "sock", socket: true, want: "store path invalid"},
		{name: "json_leaf_fifo", collection: []string{"tasks"}, filename: "fifo.json", want: "record invalid"},
		{name: "json_leaf_socket", collection: []string{"tasks"}, filename: "sock.json", socket: true, want: "record invalid"},
		{name: "atomic_fifo", filename: ".tasks.json.tmp-123", want: "store path invalid"},
		{name: "atomic_socket", filename: ".tasks.json.tmp-456", socket: true, want: "store path invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := openStore(t, t.TempDir())
			dir := store.stateDir()
			if len(test.collection) != 0 {
				dir = filepath.Join(append([]string{dir}, test.collection...)...)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(dir, test.filename)
			makeSpecialNode(t, path, test.socket)

			err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
				t.Errorf("WalkRecords() returned %v from a special node", record)
				return nil
			})
			if err == nil || err.Error() != test.want+": "+path {
				t.Fatalf("WalkRecords() error = %v; want %s for %s", err, test.want, path)
			}
		})
	}
}

func TestWalkAcceptsRegularSkippedNames(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.stateDir(), "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "one.json"), []byte(`{"id":"one"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a record"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".one.json.tmp-123"), []byte(`{"id":"one"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var found []string
	if err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
		found = append(found, record.Name)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0] != "one.json" {
		t.Fatalf("WalkRecords() names = %q; want [one.json]", found)
	}
}

func makeSpecialNode(t *testing.T, path string, socket bool) {
	t.Helper()
	if socket {
		listener, err := net.Listen("unix", path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = listener.Close() })
		return
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
}
