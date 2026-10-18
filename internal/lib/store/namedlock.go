package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// namedLockRetry is the wait between non-blocking lock attempts.
const namedLockRetry = 10 * time.Millisecond

// NamedLock holds one named flock file below the state locks directory. It
// covers one long operation that outlives a single store write.
type NamedLock struct {
	file *os.File
}

// LockNamed acquires one named lock. It waits until the lock is free or the
// context ends.
func LockNamed(ctx context.Context, home, name string) (*NamedLock, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateComponent("lock", name); err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".hive", "state", "locks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &NamedLock{file: file}, nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			_ = file.Close()
			return nil, err
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-time.After(namedLockRetry):
		}
	}
}

// Unlock releases the named lock.
func (l *NamedLock) Unlock() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if err != nil {
		return fmt.Errorf("unlock named lock: %w", err)
	}
	return closeErr
}
