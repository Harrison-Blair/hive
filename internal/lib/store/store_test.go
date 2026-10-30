package store

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWriteInProcessesWaitForTheLock(t *testing.T) {
	home := t.TempDir()
	openStore(t, home)
	release := holdStateLock(t, home)
	defer release()
	const processes, iterations = 4, 20
	done := make(chan error, 1)
	go func() { done <- runStoreHelpers(home, "write", processes, iterations) }()

	// The helpers start and ask for the lock well within this hold, which is
	// shorter than the two second lock wait.
	time.Sleep(time.Second)
	if entries, err := os.ReadDir(filepath.Join(home, "tasks")); err == nil && len(entries) > 0 {
		t.Fatalf("helpers wrote %d entries while another holder had the lock", len(entries))
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(home, "tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != processes {
		t.Fatalf("tasks has %d entries; want one record per helper: %v", len(entries), entries)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(home, "tasks", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]int
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		if record["count"] != iterations {
			t.Fatalf("%s count = %d; want %d", entry.Name(), record["count"], iterations)
		}
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
	id := fmt.Sprintf("p%d", os.Getpid())
	for i := range iterations {
		switch mode {
		case "write":
			if err := store.WriteIn([]string{"tasks"}, id, map[string]int{"count": i + 1}); err != nil {
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
