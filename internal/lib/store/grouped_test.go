package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// withGroupedReadLock allows race-instrumented scheduling time; the test checks
// version consistency, not lock acquisition speed.
func withGroupedReadLock(store *Store, fn func(*LockedStore) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return store.WithLock(ctx, fn)
}

func TestGroupedReadSeesEqualVersions(t *testing.T) {
	store := openStore(t, t.TempDir())
	for _, id := range []string{"a", "b"} {
		if err := store.WriteIn([]string{"tasks"}, id, map[string]int{"version": 0}); err != nil {
			t.Fatal(err)
		}
	}
	const rounds = 200
	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for version := 1; version <= rounds; version++ {
			err := withGroupedReadLock(store, func(locked *LockedStore) error {
				if err := locked.WriteIn([]string{"tasks"}, "a", map[string]int{"version": version}); err != nil {
					return err
				}
				return locked.WriteIn([]string{"tasks"}, "b", map[string]int{"version": version})
			})
			if err != nil {
				errCh <- err
				return
			}
		}
	}()
	for range rounds {
		err := withGroupedReadLock(store, func(locked *LockedStore) error {
			var a, b map[string]int
			if err := locked.ReadIn([]string{"tasks"}, "a", &a); err != nil {
				return err
			}
			if err := locked.ReadIn([]string{"tasks"}, "b", &b); err != nil {
				return err
			}
			if a["version"] != b["version"] {
				return fmt.Errorf("reader saw versions %d and %d", a["version"], b["version"])
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
	var final map[string]int
	if err := store.ReadIn([]string{"tasks"}, "a", &final); err != nil {
		t.Fatal(err)
	}
	if final["version"] != rounds {
		t.Fatalf("final version = %d; want %d", final["version"], rounds)
	}
}

func TestGroupedPartialWriteStays(t *testing.T) {
	store := openStore(t, t.TempDir())
	sentinel := errors.New("stop the group")
	err := store.WithLock(context.Background(), func(locked *LockedStore) error {
		if err := locked.WriteIn([]string{"tasks"}, "a", map[string]int{"value": 1}); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithLock() error = %v; want the callback error", err)
	}
	var a map[string]int
	if err := store.ReadIn([]string{"tasks"}, "a", &a); err != nil {
		t.Fatal(err)
	}
	if a["value"] != 1 {
		t.Fatalf("record a value = %d; want the completed prefix", a["value"])
	}

	err = store.WithLock(context.Background(), func(locked *LockedStore) error {
		if err := locked.WriteIn([]string{"tasks"}, "a", map[string]int{"value": 2}); err != nil {
			return err
		}
		if err := locked.WriteIn([]string{"tasks"}, "b", map[string]int{"value": 3}); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("second WithLock() error = %v; want the callback error", err)
	}
	for id, want := range map[string]int{"a": 2, "b": 3} {
		var record map[string]int
		if err := store.ReadIn([]string{"tasks"}, id, &record); err != nil {
			t.Fatal(err)
		}
		if record["value"] != want {
			t.Fatalf("record %s value = %d; want %d", id, record["value"], want)
		}
	}
}

func TestGroupedCounterProcesses(t *testing.T) {
	home := t.TempDir()
	const processes, iterations = 4, 50
	if err := runStoreHelpers(home, "grouped-counter", processes, iterations); err != nil {
		t.Fatal(err)
	}
	data, err := readRecordFile(filepath.Join(home, "counters", "shared.json"))
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
