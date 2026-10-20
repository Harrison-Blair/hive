package panelock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestLockUsesSanitizedPaneName(t *testing.T) {
	home := t.TempDir()
	first, err := Lock(context.Background(), home, "w1:p1")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Unlock()
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "locks", "pane-w1-p1")); err != nil {
		t.Fatalf("pane lock file: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := store.LockNamed(ctx, home, "pane-w1-p1"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second lock error = %v; want deadline while the same pane lock is held", err)
	}
}

func TestLockKeepsDifferentPanesSeparate(t *testing.T) {
	home := t.TempDir()
	first, err := Lock(context.Background(), home, "w1:p1")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Unlock()
	second, err := Lock(context.Background(), home, "w1:p2")
	if err != nil {
		t.Fatalf("lock for another pane: %v", err)
	}
	if err := second.Unlock(); err != nil {
		t.Fatal(err)
	}
}
