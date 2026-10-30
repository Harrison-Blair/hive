package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentUpdateProcesses(t *testing.T) {
	home := t.TempDir()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WriteIn([]string{"tasks"}, "shared", map[string]int{"count": 0}); err != nil {
		t.Fatal(err)
	}
	const processes, iterations = 4, 50
	if err := runStoreHelpers(home, "update", processes, iterations); err != nil {
		t.Fatal(err)
	}
	var record struct {
		Count int `json:"count"`
	}
	if err := store.ReadIn([]string{"tasks"}, "shared", &record); err != nil {
		t.Fatal(err)
	}
	if record.Count != processes*iterations {
		t.Fatalf("updated count = %d; want %d", record.Count, processes*iterations)
	}
}

func TestUpdateInMissingRecordStartsEmpty(t *testing.T) {
	store := openStore(t, t.TempDir())
	var seen json.RawMessage
	err := store.UpdateIn([]string{"tasks"}, "fresh", func(raw json.RawMessage) (json.RawMessage, error) {
		seen = raw
		return json.RawMessage(`{"value":1}`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != nil {
		t.Fatalf("UpdateIn() callback got %q; want nil for a missing record", seen)
	}
	var record map[string]int
	if err := store.ReadIn([]string{"tasks"}, "fresh", &record); err != nil {
		t.Fatal(err)
	}
	if record["value"] != 1 {
		t.Fatalf("record value = %d; want 1", record["value"])
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
		case "update":
			err := store.UpdateIn([]string{"tasks"}, "shared", func(raw json.RawMessage) (json.RawMessage, error) {
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
		case "grouped-update":
			err := store.WithLock(context.Background(), func(locked *LockedStore) error {
				var record map[string]int
				if err := locked.ReadIn([]string{"tasks"}, "shared", &record); err != nil {
					return err
				}
				record["count"]++
				return locked.WriteIn([]string{"tasks"}, "shared", record)
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
	if err := store.WriteIn([]string{"tasks"}, "atomic", map[string]int{"value": 0}); err != nil {
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
			if err := store.ReadIn([]string{"tasks"}, "atomic", &record); err != nil {
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
		if err := store.WriteIn([]string{"tasks"}, "atomic", map[string]int{"value": value}); err != nil {
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
	entries, err := os.ReadDir(filepath.Join(home, "tasks"))
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
	if err := store.WriteIn([]string{"tasks"}, "locked", map[string]int{"count": 0}); err != nil {
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
				err := store.UpdateIn([]string{"tasks"}, "locked", func(raw json.RawMessage) (json.RawMessage, error) {
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
	if err := store.ReadIn([]string{"tasks"}, "locked", &record); err != nil {
		t.Fatal(err)
	}
	if record["count"] != workers*updates {
		t.Fatalf("updated count = %d; want %d", record["count"], workers*updates)
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
