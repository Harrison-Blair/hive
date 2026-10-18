package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// openStore opens a store in home and fails the test on error.
func openStore(t *testing.T, home string) *Store {
	t.Helper()
	store, err := Open(home)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

// holdStateLock takes the store lock without the store API. The returned
// function releases it. The function is safe to call more than one time.
func holdStateLock(t *testing.T, home string) func() {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "lock")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
			_ = file.Close()
		})
	}
}

func TestWithLockDefaultTimeout(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	defer release()

	ran := false
	start := time.Now()
	err := store.WithLock(context.Background(), func(*LockedStore) error {
		ran = true
		return nil
	})
	elapsed := time.Since(start)
	if ran {
		t.Fatal("callback ran while another holder had the lock")
	}
	if !errors.Is(err, ErrLockTimeout) {
		t.Fatalf("WithLock() error = %v; want ErrLockTimeout", err)
	}
	if elapsed < 1800*time.Millisecond {
		t.Fatalf("WithLock() returned after %v; want about two seconds", elapsed)
	}
	if elapsed > 6*time.Second {
		t.Fatalf("WithLock() returned after %v; want a bounded wait", elapsed)
	}
}

func TestWithLockContextDeadline(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	ran := false
	start := time.Now()
	err := store.WithLock(ctx, func(*LockedStore) error {
		ran = true
		return nil
	})
	elapsed := time.Since(start)
	if ran {
		t.Fatal("callback ran while another holder had the lock")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WithLock() error = %v; want context.DeadlineExceeded", err)
	}
	if elapsed < 80*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("WithLock() returned after %v; want about 100ms", elapsed)
	}
}

func TestWithLockCanceledContext(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ran := false
	err := store.WithLock(ctx, func(*LockedStore) error {
		ran = true
		return nil
	})
	if ran {
		t.Fatal("callback ran with a canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithLock() error = %v; want context.Canceled", err)
	}
}

func TestWithLockLongDeadlineWaits(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	released := make(chan struct{})
	go func() {
		time.Sleep(2200 * time.Millisecond)
		release()
		close(released)
	}()
	defer func() {
		select {
		case <-released:
		default:
			release()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ran := false
	start := time.Now()
	err := store.WithLock(ctx, func(*LockedStore) error {
		ran = true
		return nil
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("WithLock() error = %v; want no error", err)
	}
	if !ran {
		t.Fatal("callback did not run after the lock became free")
	}
	if elapsed < 2*time.Second {
		t.Fatalf("WithLock() acquired after %v; want after two seconds", elapsed)
	}
}

func TestWithLockFailureChangesNoRecord(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.Write("tasks", "sentinel", map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	release := holdStateLock(t, store.home)
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := store.WithLock(ctx, func(locked *LockedStore) error {
		return locked.Write("tasks", "sentinel", map[string]int{"value": 2})
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WithLock() error = %v; want context.DeadlineExceeded", err)
	}
	var record map[string]int
	if err := store.Read("tasks", "sentinel", &record); err != nil {
		t.Fatal(err)
	}
	if record["value"] != 1 {
		t.Fatalf("record value = %d; want 1 after a failed acquisition", record["value"])
	}
}

func TestWithLockCallbackErrorReleases(t *testing.T) {
	store := openStore(t, t.TempDir())
	sentinel := errors.New("callback failed")
	err := store.WithLock(context.Background(), func(*LockedStore) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithLock() error = %v; want the callback error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	ran := false
	if err := store.WithLock(ctx, func(*LockedStore) error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("second WithLock() error = %v; want no error", err)
	}
	if !ran {
		t.Fatal("second callback did not run after the first error")
	}
}

func TestWithLockPanicReleases(t *testing.T) {
	store := openStore(t, t.TempDir())
	func() {
		defer func() {
			if recovered := recover(); recovered == nil {
				t.Error("WithLock() did not propagate the callback panic")
			}
		}()
		_ = store.WithLock(context.Background(), func(*LockedStore) error {
			panic("callback panic")
		})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	ran := false
	if err := store.WithLock(ctx, func(*LockedStore) error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("WithLock() after panic error = %v; want no error", err)
	}
	if !ran {
		t.Fatal("callback did not run after the panic")
	}
}

func TestWithLockHandleAfterRelease(t *testing.T) {
	store := openStore(t, t.TempDir())
	var handle *LockedStore
	if err := store.WithLock(context.Background(), func(locked *LockedStore) error {
		handle = locked
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := handle.Read("tasks", "one", &map[string]any{}); !errors.Is(err, ErrLockReleased) {
		t.Fatalf("released handle Read() error = %v; want ErrLockReleased", err)
	}
	var nilHandle *LockedStore
	if err := nilHandle.Read("tasks", "one", &map[string]any{}); !errors.Is(err, ErrLockReleased) {
		t.Fatalf("nil handle Read() error = %v; want ErrLockReleased", err)
	}
}

func TestWithLockNestedFails(t *testing.T) {
	store := openStore(t, t.TempDir())
	var nestedErr error
	nestedRan := false
	err := store.WithLock(context.Background(), func(locked *LockedStore) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		nestedErr = store.WithLock(ctx, func(*LockedStore) error {
			nestedRan = true
			return nil
		})
		return locked.Write("tasks", "outer", map[string]int{"value": 1})
	})
	if err != nil {
		t.Fatalf("outer WithLock() error = %v", err)
	}
	if nestedRan {
		t.Fatal("nested callback ran while the outer callback held the lock")
	}
	if !errors.Is(nestedErr, context.DeadlineExceeded) && !errors.Is(nestedErr, ErrLockTimeout) {
		t.Fatalf("nested WithLock() error = %v; want a bounded failure", nestedErr)
	}
	var record map[string]int
	if err := store.Read("tasks", "outer", &record); err != nil {
		t.Fatal(err)
	}
	if record["value"] != 1 {
		t.Fatalf("outer record value = %d; want 1", record["value"])
	}
}

func TestWithLockLegacyMutationTimeout(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	defer release()

	checks := []struct {
		name string
		run  func() error
	}{
		{"Write", func() error {
			return store.Write("tasks", "one", map[string]int{"value": 1})
		}},
		{"Update", func() error {
			return store.Update("tasks", "one", func(json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(`{"value":1}`), nil
			})
		}},
		{"Next", func() error {
			_, err := store.Next("counter")
			return err
		}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			start := time.Now()
			err := check.run()
			elapsed := time.Since(start)
			if !errors.Is(err, ErrLockTimeout) {
				t.Fatalf("%s() error = %v; want ErrLockTimeout", check.name, err)
			}
			if elapsed < 1800*time.Millisecond {
				t.Fatalf("%s() returned after %v; want about two seconds", check.name, elapsed)
			}
		})
	}
}

func TestWithLockReadsDoNotTakeTheLock(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.Write("tasks", "visible", map[string]int{"value": 7}); err != nil {
		t.Fatal(err)
	}
	release := holdStateLock(t, store.home)
	defer release()

	var record map[string]int
	if err := store.Read("tasks", "visible", &record); err != nil {
		t.Fatalf("Read() error = %v; want no lock wait", err)
	}
	if record["value"] != 7 {
		t.Fatalf("Read() value = %d; want 7", record["value"])
	}
	var generic map[string]int
	if err := store.ReadIn([]string{"tasks"}, "visible", &generic); err != nil {
		t.Fatalf("ReadIn() error = %v; want no lock wait", err)
	}
	if generic["value"] != 7 {
		t.Fatalf("ReadIn() value = %d; want 7", generic["value"])
	}
}
