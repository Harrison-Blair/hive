package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentCounterProcesses(t *testing.T) {
	home := t.TempDir()
	const processes, iterations = 4, 50
	if err := runStoreHelpers(home, "counter", processes, iterations); err != nil {
		t.Fatal(err)
	}
	data, err := readRecordFile(filepath.Join(home, ".hive", "state", "counters", "shared.json"))
	if err != nil {
		t.Fatal(err)
	}
	var counter struct {
		Value uint64 `json:"value"`
	}
	if err := json.Unmarshal(data, &counter); err != nil {
		t.Fatal(err)
	}
	if counter.Value != processes*iterations {
		t.Fatalf("counter = %d; want %d", counter.Value, processes*iterations)
	}
}

func TestUpdateCounterStartsAtOne(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Next("task")
	if err != nil || first != 1 {
		t.Fatalf("first counter = %d, %v; want 1", first, err)
	}
	second, err := store.Next("task")
	if err != nil || second != 2 {
		t.Fatalf("second counter = %d, %v; want 2", second, err)
	}
}

func TestConcurrentUpdateProcesses(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Write("tasks", "shared", map[string]int{"count": 0}); err != nil {
		t.Fatal(err)
	}
	const processes, iterations = 4, 50
	if err := runStoreHelpers(home, "update", processes, iterations); err != nil {
		t.Fatal(err)
	}
	var record struct {
		Count int `json:"count"`
	}
	if err := store.Read("tasks", "shared", &record); err != nil {
		t.Fatal(err)
	}
	if record.Count != processes*iterations {
		t.Fatalf("updated count = %d; want %d", record.Count, processes*iterations)
	}
}

func TestStoreProcessHelper(t *testing.T) {
	mode := os.Getenv("HIVE_STORE_HELPER")
	if mode == "" {
		return
	}
	home := os.Getenv("HIVE_STORE_HOME")
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	iterations := 50
	if value := os.Getenv("HIVE_STORE_ITERATIONS"); value != "" {
		if _, err := fmt.Sscanf(value, "%d", &iterations); err != nil {
			t.Fatal(err)
		}
	}
	for range iterations {
		switch mode {
		case "counter":
			if _, err := store.Next("shared"); err != nil {
				t.Fatal(err)
			}
		case "update":
			err := store.Update("tasks", "shared", func(raw json.RawMessage) (json.RawMessage, error) {
				var record map[string]int
				if err := json.Unmarshal(raw, &record); err != nil {
					return nil, err
				}
				record["count"]++
				return json.Marshal(record)
			})
			if err != nil {
				t.Fatal(err)
			}
		case "grouped-counter":
			err := store.WithLock(context.Background(), func(locked *LockedStore) error {
				_, err := locked.Next("shared")
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unknown helper mode %q", mode)
		}
	}
}

func TestAtomicWriteReadersSeeCompleteRecord(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Write("tasks", "atomic", map[string]int{"value": 0}); err != nil {
		t.Fatal(err)
	}
	const writes = 200
	var wg sync.WaitGroup
	stop := make(chan struct{})
	readErr := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			var record map[string]int
			if err := store.Read("tasks", "atomic", &record); err != nil {
				select {
				case readErr <- err:
				default:
				}
				return
			}
			if record["value"] < 0 || record["value"] > writes {
				select {
				case readErr <- fmt.Errorf("unexpected record value %d", record["value"]):
				default:
				}
				return
			}
		}
	}()
	for value := 1; value <= writes; value++ {
		if err := store.Write("tasks", "atomic", map[string]int{"value": value}); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
	select {
	case err := <-readErr:
		t.Fatalf("reader saw invalid record: %v", err)
	default:
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "atomic.json" {
		t.Fatalf("record directory contains temporary files: %v", entries)
	}
}

func TestUpdateCallbackRunsUnderLock(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Write("tasks", "locked", map[string]int{"count": 0}); err != nil {
		t.Fatal(err)
	}
	const workers, updates = 8, 40
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range updates {
				err := store.Update("tasks", "locked", func(raw json.RawMessage) (json.RawMessage, error) {
					var record map[string]int
					if err := json.Unmarshal(raw, &record); err != nil {
						return nil, err
					}
					record["count"]++
					return json.Marshal(record)
				})
				if err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
	var record map[string]int
	if err := store.Read("tasks", "locked", &record); err != nil {
		t.Fatal(err)
	}
	if record["count"] != workers*updates {
		t.Fatalf("updated count = %d; want %d", record["count"], workers*updates)
	}
}

func TestReadRejectsOversizedInvalidAndNonRegularRecords(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".hive", "state", "projects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	large := filepath.Join(dir, "large.json")
	if err := os.WriteFile(large, []byte(`{"padding":"`+strings.Repeat("x", maxRecordSize)+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	truncated := filepath.Join(dir, "truncated.json")
	if err := os.WriteFile(truncated, []byte(`{"id":`), 0o644); err != nil {
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
	if err := os.Mkdir(filepath.Join(dir, "directory.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "target.json")
	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "symlink.json")); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]string{
		"large":     "record too large: " + large,
		"truncated": "record invalid: " + truncated,
		"array":     "record invalid: " + array,
		"empty":     "record invalid: " + empty,
		"directory": "record invalid: " + filepath.Join(dir, "directory.json"),
		"symlink":   "record invalid: " + filepath.Join(dir, "symlink.json"),
	} {
		t.Run(id, func(t *testing.T) {
			var record map[string]any
			if err := store.Read("projects", id, &record); err == nil || err.Error() != want {
				t.Fatalf("Read() error = %v; want %q", err, want)
			}
		})
	}
}

func TestReadAcceptsRecordOfExactlyOneMiB(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	prefix, suffix := `{"padding":"`, `"}`
	data := prefix + strings.Repeat("x", maxRecordSize-len(prefix)-len(suffix)) + suffix
	if len(data) != maxRecordSize {
		t.Fatalf("fixture size = %d; want %d", len(data), maxRecordSize)
	}
	path := filepath.Join(home, ".hive", "state", "projects", "exact.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	var record map[string]string
	if err := store.Read("projects", "exact", &record); err != nil {
		t.Fatalf("Read() rejected exactly 1 MiB: %v", err)
	}
}

func TestReadListUsesOnlyValidIDFileNames(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".hive", "state", "projects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"valid-id.json":            `{"id":"valid-id"}`,
		"bad_id.json":              `{"id":"bad_id"}`,
		"a..b.json":                `{"id":"a..b"}`,
		"-bad.json":                `{"id":"-bad"}`,
		"abcdefghijklmnopqrs.json": `{"id":"long"}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	records, err := store.List("projects")
	if err != nil || len(records) != 1 || !strings.Contains(string(records[0]), "valid-id") {
		t.Fatalf("List() = %q, %v; want only valid-id", records, err)
	}
}

func TestReadListReturnsNoRecordsAfterInvalidFile(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".hive", "state", "projects")
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
	records, err := store.List("projects")
	if err == nil || err.Error() != "record invalid: "+invalid || records != nil {
		t.Fatalf("List() = %q, %v; want no records and the first read error", records, err)
	}
}

func TestReadAndUpdateRejectUnsafeNames(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		kind, id string
	}{
		{"", "ok"}, {"../bad", "ok"}, {"kind", ""}, {"kind", "../bad"}, {"kind", "with/slash"},
	} {
		if err := store.Read(test.kind, test.id, &map[string]any{}); err == nil {
			t.Errorf("Read(%q, %q) succeeded", test.kind, test.id)
		}
	}
	if err := store.Write("kind", "", map[string]any{}); err == nil {
		t.Error("Write with an empty ID succeeded")
	}
	if err := store.Update("kind", "../bad", func(json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{"value":1}`), nil
	}); err == nil {
		t.Error("Update with a parent path succeeded")
	}
	for _, name := range []string{"", "..", "../bad", "with/slash"} {
		if _, err := store.Next(name); err == nil {
			t.Errorf("Next(%q) succeeded", name)
		}
	}
}

func runStoreHelpers(home, mode string, processes, iterations int) error {
	var wg sync.WaitGroup
	errCh := make(chan error, processes)
	for range processes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			command := exec.Command(os.Args[0], "-test.run=TestStoreProcessHelper")
			command.Env = append(os.Environ(),
				"HIVE_STORE_HELPER="+mode,
				"HIVE_STORE_HOME="+home,
				fmt.Sprintf("HIVE_STORE_ITERATIONS=%d", iterations),
			)
			if output, err := command.CombinedOutput(); err != nil {
				errCh <- fmt.Errorf("helper failed: %w: %s", err, output)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}
