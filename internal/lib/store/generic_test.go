package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreNamesAcceptValidComponents(t *testing.T) {
	store := openStore(t, t.TempDir())
	for _, id := range []string{"12-spawn-worker", strings.Repeat("a", 40)} {
		if err := store.WriteIn([]string{"tasks"}, id, map[string]string{"id": id}); err != nil {
			t.Fatalf("WriteIn(%q) error = %v", id, err)
		}
		var record map[string]string
		if err := store.ReadIn([]string{"tasks"}, id, &record); err != nil {
			t.Fatalf("ReadIn(%q) error = %v", id, err)
		}
		if record["id"] != id {
			t.Fatalf("ReadIn(%q) id = %q; want %q", id, record["id"], id)
		}
	}
}

func TestListInKeepsLongStoreSegment(t *testing.T) {
	store := openStore(t, t.TempDir())
	id := strings.Repeat("k", 201)
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(`{"id":"`+id+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := store.ListIn([]string{"tasks"}, nil)
	if err != nil {
		t.Fatalf("ListIn() error = %v", err)
	}
	if len(records) != 1 || !strings.Contains(string(records[0]), id) {
		t.Fatalf("ListIn() records = %q; want the long store segment", records)
	}
}

func TestStoreNamesRejectInvalidComponents(t *testing.T) {
	store := openStore(t, t.TempDir())
	invalid := []string{
		"", ".", "..", "../x", "a/b", `a\b`, "a\x01b", "Upper", strings.Repeat("a", 201),
	}
	for _, name := range invalid {
		if err := store.ReadIn([]string{"tasks"}, name, &map[string]any{}); err == nil || !strings.Contains(err.Error(), "invalid store") {
			t.Errorf("ReadIn(%q) error = %v; want invalid store", name, err)
		}
		if err := store.WriteIn([]string{"tasks"}, name, map[string]string{"id": "x"}); err == nil || !strings.Contains(err.Error(), "invalid store") {
			t.Errorf("WriteIn(%q) error = %v; want invalid store", name, err)
		}
	}
	collections := [][]string{
		nil,
		{},
		{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
		{"Bad"},
		{""},
		{"a", ".."},
		{strings.Repeat("a", 201)},
	}
	for _, collection := range collections {
		if err := store.ReadIn(collection, "one", &map[string]any{}); err == nil || !strings.Contains(err.Error(), "invalid store") {
			t.Errorf("ReadIn(%q) error = %v; want invalid store", collection, err)
		}
		if err := store.WriteIn(collection, "one", map[string]string{"id": "one"}); err == nil || !strings.Contains(err.Error(), "invalid store") {
			t.Errorf("WriteIn(%q) error = %v; want invalid store", collection, err)
		}
		if _, err := store.ListIn(collection, nil); err == nil || !strings.Contains(err.Error(), "invalid store") {
			t.Errorf("ListIn(%q) error = %v; want invalid store", collection, err)
		}
	}
}

func TestWriteInNestedInbox(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.WriteIn([]string{"inbox", "coordinator"}, "1", map[string]string{"body": "hello"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.home, ".hive", "state", "inbox", "coordinator", "1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("nested record was not written: %v", err)
	}
	var record map[string]string
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if record["body"] != "hello" {
		t.Fatalf("record body = %q; want hello", record["body"])
	}
}

func TestListInMissingCollection(t *testing.T) {
	store := openStore(t, t.TempDir())
	records, err := store.ListIn([]string{"missing"}, nil)
	if err != nil {
		t.Fatalf("ListIn() error = %v", err)
	}
	if records == nil {
		t.Fatal("ListIn() returned a nil slice; want an empty slice")
	}
	if len(records) != 0 {
		t.Fatalf("ListIn() returned %d records; want 0", len(records))
	}
}

func TestListInIncludeSelectsOne(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.WriteIn([]string{"tasks"}, "alpha", map[string]string{"id": "alpha"}); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteIn([]string{"tasks"}, "beta", map[string]string{"id": "beta"}); err != nil {
		t.Fatal(err)
	}
	var seen []string
	records, err := store.ListIn([]string{"tasks"}, func(filename string) bool {
		seen = append(seen, filename)
		return filename == "beta.json"
	})
	if err != nil {
		t.Fatalf("ListIn() error = %v", err)
	}
	if len(records) != 1 || !strings.Contains(string(records[0]), `"beta"`) {
		t.Fatalf("ListIn() records = %q; want only beta", records)
	}
	if len(seen) != 2 || seen[0] != "alpha.json" || seen[1] != "beta.json" {
		t.Fatalf("include saw %q; want the file names with the .json suffix", seen)
	}
}

func TestListInSkipsInvalidStoreSegments(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"valid.json":     `{"id":"valid"}`,
		"Bad.json":       `{"id":"bad-uppercase"}`,
		"a..b.json":      `{"id":"bad-dots"}`,
		".hidden.json":   `{"id":"hidden"}`,
		"has space.json": `{"id":"bad-space"}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	records, err := store.ListIn([]string{"tasks"}, nil)
	if err != nil {
		t.Fatalf("ListIn() error = %v", err)
	}
	var got []string
	for _, raw := range records {
		var record struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &record); err != nil {
			t.Fatal(err)
		}
		got = append(got, record.ID)
	}
	want := []string{"hidden", "bad-uppercase", "bad-space", "valid"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ListIn() IDs = %q; want %q", got, want)
	}
}

func TestReadInRecordInvalid(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	array := filepath.Join(dir, "array.json")
	if err := os.WriteFile(array, []byte(`[]`), 0o644); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(dir, "directory.json")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(store.home, "target.json")
	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(dir, "symlink.json")
	if err := os.Symlink(target, symlink); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]string{
		"empty":     "record invalid: " + empty,
		"array":     "record invalid: " + array,
		"directory": "record invalid: " + directory,
		"symlink":   "record invalid: " + symlink,
	} {
		var record map[string]any
		if err := store.ReadIn([]string{"tasks"}, id, &record); err == nil || err.Error() != want {
			t.Errorf("ReadIn(%q) error = %v; want %q", id, err, want)
		}
	}
}

const (
	sizeBelowLimit = 1048575
	sizeAtLimit    = 1048576
	sizeAboveLimit = 1048577
)

func recordOfSize(t *testing.T, size int) []byte {
	t.Helper()
	prefix, suffix := `{"padding":"`, `"}`
	data := prefix + strings.Repeat("x", size-len(prefix)-len(suffix)) + suffix
	if len(data) != size {
		t.Fatalf("fixture size = %d; want %d", len(data), size)
	}
	return []byte(data)
}

func TestReadInSizeBoundary(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id      string
		size    int
		wantErr bool
	}{
		{"below", sizeBelowLimit, false},
		{"exact", sizeAtLimit, false},
		{"above", sizeAboveLimit, true},
	} {
		path := filepath.Join(dir, test.id+".json")
		if err := os.WriteFile(path, recordOfSize(t, test.size), 0o644); err != nil {
			t.Fatal(err)
		}
		var record map[string]string
		err := store.ReadIn([]string{"tasks"}, test.id, &record)
		if test.wantErr {
			if err == nil || err.Error() != "record too large: "+path {
				t.Errorf("ReadIn(%q) error = %v; want record too large", test.id, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("ReadIn(%q) rejected %d bytes: %v", test.id, test.size, err)
		}
	}
}

func TestListInBadSelectedFileReturnsNoRecords(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(`{"id":"alpha"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(dir, "beta.json")
	if err := os.WriteFile(invalid, []byte(`{"id":`), 0o644); err != nil {
		t.Fatal(err)
	}
	records, err := store.ListIn([]string{"tasks"}, nil)
	if err == nil || err.Error() != "record invalid: "+invalid {
		t.Fatalf("ListIn() error = %v; want the first read error", err)
	}
	if records != nil {
		t.Fatalf("ListIn() records = %q; want no records", records)
	}
}

func TestReadInPathTypes(t *testing.T) {
	t.Run("state root symlink", func(t *testing.T) {
		home := t.TempDir()
		store := openStore(t, home)
		target := t.TempDir()
		state := filepath.Join(home, ".hive", "state")
		if err := os.RemoveAll(state); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, state); err != nil {
			t.Fatal(err)
		}
		if err := store.ReadIn([]string{"tasks"}, "one", &map[string]any{}); err == nil || !strings.Contains(err.Error(), "store path invalid") {
			t.Fatalf("ReadIn() error = %v; want store path invalid", err)
		}
	})

	t.Run("collection ancestor symlink", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		target := t.TempDir()
		link := filepath.Join(store.home, ".hive", "state", "inbox")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		err := store.WriteIn([]string{"inbox", "coordinator"}, "1", map[string]string{"body": "x"})
		if err == nil || !strings.Contains(err.Error(), "store path invalid") {
			t.Fatalf("WriteIn() error = %v; want store path invalid", err)
		}
		if _, statErr := os.Lstat(filepath.Join(target, "coordinator", "1.json")); !os.IsNotExist(statErr) {
			t.Fatalf("WriteIn() wrote through a linked ancestor: %v", statErr)
		}
	})

	t.Run("record leaf symlink", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		dir := filepath.Join(store.home, ".hive", "state", "tasks")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(store.home, "target.json")
		if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "linked.json")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		err := store.WriteIn([]string{"tasks"}, "linked", map[string]string{"id": "linked"})
		if err == nil || !strings.Contains(err.Error(), "record invalid") {
			t.Fatalf("WriteIn() error = %v; want record invalid", err)
		}
		data, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(data) != `{"id":"target"}` {
			t.Fatalf("WriteIn() changed the symlink target to %q", data)
		}
	})

	t.Run("dangling record leaf symlink", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		dir := filepath.Join(store.home, ".hive", "state", "tasks")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		missing := filepath.Join(store.home, "missing.json")
		link := filepath.Join(dir, "dangling.json")
		if err := os.Symlink(missing, link); err != nil {
			t.Fatal(err)
		}
		err := store.WriteIn([]string{"tasks"}, "dangling", map[string]string{"id": "dangling"})
		if err == nil || !strings.Contains(err.Error(), "record invalid") {
			t.Fatalf("WriteIn() error = %v; want record invalid", err)
		}
		if _, statErr := os.Lstat(missing); !os.IsNotExist(statErr) {
			t.Fatalf("WriteIn() created the dangling symlink target: %v", statErr)
		}
	})

	t.Run("record leaf directory", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		dir := filepath.Join(store.home, ".hive", "state", "tasks", "directory.json")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		err := store.WriteIn([]string{"tasks"}, "directory", map[string]string{"id": "directory"})
		if err == nil || !strings.Contains(err.Error(), "record invalid") {
			t.Fatalf("WriteIn() error = %v; want record invalid", err)
		}
	})

	t.Run("lock symlink", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		target := filepath.Join(store.home, "lock-target")
		if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
			t.Fatal(err)
		}
		lockPath := filepath.Join(store.home, ".hive", "state", "lock")
		if err := os.Symlink(target, lockPath); err != nil {
			t.Fatal(err)
		}
		err := store.WriteIn([]string{"tasks"}, "one", map[string]string{"id": "one"})
		if err == nil || !strings.Contains(err.Error(), "store path invalid") {
			t.Fatalf("WriteIn() error = %v; want store path invalid", err)
		}
		data, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(data) != "keep" {
			t.Fatalf("WriteIn() changed the lock symlink target to %q", data)
		}
	})
}
