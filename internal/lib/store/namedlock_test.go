package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func lockHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestNamedLockSerializesOneName(t *testing.T) {
	home := lockHome(t)
	ctx := context.Background()
	first, err := LockNamed(ctx, home, "land-one")
	if err != nil {
		t.Fatalf("first LockNamed() error = %v", err)
	}
	acquired := make(chan error, 1)
	go func() {
		second, err := LockNamed(ctx, home, "land-one")
		if err == nil {
			_ = second.Unlock()
		}
		acquired <- err
	}()
	select {
	case err := <-acquired:
		t.Fatalf("second LockNamed() returned with %v while the first lock was held", err)
	case <-time.After(150 * time.Millisecond):
	}
	if err := first.Unlock(); err != nil {
		t.Fatalf("first Unlock() error = %v", err)
	}
	select {
	case err := <-acquired:
		if err != nil {
			t.Fatalf("second LockNamed() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second LockNamed() did not return after the first release")
	}
}

func TestNamedLockDoesNotBlockAnotherName(t *testing.T) {
	home := lockHome(t)
	ctx := context.Background()
	first, err := LockNamed(ctx, home, "land-one")
	if err != nil {
		t.Fatalf("first LockNamed() error = %v", err)
	}
	defer func() { _ = first.Unlock() }()
	other, err := LockNamed(ctx, home, "land-two")
	if err != nil {
		t.Fatalf("LockNamed() error = %v", err)
	}
	if err := other.Unlock(); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}
}

func TestNamedLockStopsOnContext(t *testing.T) {
	home := lockHome(t)
	first, err := LockNamed(context.Background(), home, "land-one")
	if err != nil {
		t.Fatalf("first LockNamed() error = %v", err)
	}
	defer func() { _ = first.Unlock() }()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = LockNamed(ctx, home, "land-one")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("LockNamed() error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("LockNamed() waited %v, want two seconds or less", elapsed)
	}
}

func TestNamedLockRejectsInvalidName(t *testing.T) {
	home := lockHome(t)
	if _, err := LockNamed(context.Background(), home, "../escape"); err == nil {
		t.Fatal("LockNamed() error = nil, want an invalid name error")
	}
}
