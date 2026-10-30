package store

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
	path := filepath.Join(home, "lock")
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

func TestWriteInLockTimeout(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.WriteIn([]string{"tasks"}, "sentinel", map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	release := holdStateLock(t, store.home)
	defer release()

	start := time.Now()
	err := store.WriteIn([]string{"tasks"}, "sentinel", map[string]int{"value": 2})
	elapsed := time.Since(start)
	if err == nil || err.Error() != "store lock timeout" {
		t.Fatalf("WriteIn() error = %v; want store lock timeout", err)
	}
	if elapsed < 1800*time.Millisecond {
		t.Fatalf("WriteIn() returned after %v; want about two seconds", elapsed)
	}
	if elapsed > 6*time.Second {
		t.Fatalf("WriteIn() returned after %v; want a bounded wait", elapsed)
	}
	var record map[string]int
	if err := store.ReadIn([]string{"tasks"}, "sentinel", &record); err != nil {
		t.Fatal(err)
	}
	if record["value"] != 1 {
		t.Fatalf("record value = %d; want 1 after a failed acquisition", record["value"])
	}
}

func TestWriteInWaitsForLockRelease(t *testing.T) {
	store := openStore(t, t.TempDir())
	release := holdStateLock(t, store.home)
	defer release()
	go func() {
		time.Sleep(300 * time.Millisecond)
		release()
	}()

	start := time.Now()
	if err := store.WriteIn([]string{"tasks"}, "one", map[string]int{"value": 1}); err != nil {
		t.Fatalf("WriteIn() error = %v; want no error", err)
	}
	if elapsed := time.Since(start); elapsed < 250*time.Millisecond {
		t.Fatalf("WriteIn() wrote after %v; want after the holder released the lock", elapsed)
	}
}

func TestWithLockCallbackErrorReleases(t *testing.T) {
	store := openStore(t, t.TempDir())
	sentinel := errors.New("callback failed")
	err := store.withLock(defaultLockWait, func() error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("withLock() error = %v; want the callback error", err)
	}

	ran := false
	if err := store.withLock(500*time.Millisecond, func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("second withLock() error = %v; want no error", err)
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
				t.Error("withLock() did not propagate the callback panic")
			}
		}()
		_ = store.withLock(defaultLockWait, func() error {
			panic("callback panic")
		})
	}()

	ran := false
	if err := store.withLock(500*time.Millisecond, func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("withLock() after panic error = %v; want no error", err)
	}
	if !ran {
		t.Fatal("callback did not run after the panic")
	}
}

func TestReadInDoesNotTakeTheLock(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.WriteIn([]string{"tasks"}, "visible", map[string]int{"value": 7}); err != nil {
		t.Fatal(err)
	}
	release := holdStateLock(t, store.home)
	defer release()

	var generic map[string]int
	if err := store.ReadIn([]string{"tasks"}, "visible", &generic); err != nil {
		t.Fatalf("ReadIn() error = %v; want no lock wait", err)
	}
	if generic["value"] != 7 {
		t.Fatalf("ReadIn() value = %d; want 7", generic["value"])
	}
}

func TestWriteInRejectsDirectoryLock(t *testing.T) {
	store := openStore(t, t.TempDir())
	lockPath := filepath.Join(store.home, "lock")
	if err := os.Mkdir(lockPath, 0o755); err != nil {
		t.Fatal(err)
	}
	err := store.WriteIn([]string{"tasks"}, "one", map[string]int{"value": 1})
	if err == nil || err.Error() != "store path invalid: "+lockPath {
		t.Fatalf("WriteIn() error = %v; want store path invalid for %s", err, lockPath)
	}
	if _, statErr := os.Lstat(filepath.Join(store.home, "tasks", "one.json")); !os.IsNotExist(statErr) {
		t.Fatalf("WriteIn() made a record with a directory at the lock path: %v", statErr)
	}
}

func TestWithLockWaiterIsNotStarved(t *testing.T) {
	store := openStore(t, t.TempDir())
	stop := make(chan struct{})
	holding := make(chan struct{})
	hogDone := make(chan error, 1)
	go func() {
		var once sync.Once
		for {
			select {
			case <-stop:
				hogDone <- nil
				return
			default:
			}
			err := store.withLock(defaultLockWait, func() error {
				once.Do(func() { close(holding) })
				time.Sleep(23 * time.Millisecond)
				return nil
			})
			if err != nil {
				hogDone <- err
				return
			}
		}
	}()
	<-holding

	// Each wait allows about four holds. A waiter that polls for the lock
	// misses the short gaps between holds and runs out of time.
	var err error
	var elapsed time.Duration
	for range 5 {
		start := time.Now()
		err = store.withLock(100*time.Millisecond, func() error { return nil })
		elapsed = time.Since(start)
		if err != nil {
			break
		}
	}
	close(stop)
	if hogErr := <-hogDone; hogErr != nil {
		t.Fatalf("holder withLock() error = %v", hogErr)
	}
	if err != nil {
		t.Fatalf("waiter withLock() error = %v after %v; want the lock at a release while another caller asks again at once", err, elapsed)
	}
}

func TestWithLockAbandonedWaitReleases(t *testing.T) {
	store := openStore(t, t.TempDir())
	baseFDs := openFDCount(t)
	baseGoroutines := runtime.NumGoroutine()
	release := holdStateLock(t, store.home)
	defer release()

	ran := false
	err := store.withLock(50*time.Millisecond, func() error {
		ran = true
		return nil
	})
	if !errors.Is(err, errLockTimeout) {
		t.Fatalf("withLock() error = %v; want errLockTimeout", err)
	}
	release()

	path := filepath.Join(store.home, "lock")
	deadline := time.Now().Add(2 * time.Second)
	for {
		free := lockIsFree(t, path)
		fds := openFDCount(t)
		goroutines := runtime.NumGoroutine()
		if free && fds == baseFDs && goroutines == baseGoroutines {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("after the abandoned wait: lock free = %v, descriptors = %d (want %d), goroutines = %d (want %d)",
				free, fds, baseFDs, goroutines, baseGoroutines)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if ran {
		t.Fatal("callback ran after the wait was abandoned")
	}
}

// openFDCount counts the open descriptors of the test process.
func openFDCount(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
}

// lockIsFree reports whether a new descriptor can take the lock at path now.
func lockIsFree(t *testing.T, path string) bool {
	t.Helper()
	file, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return false
	}
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	return true
}
