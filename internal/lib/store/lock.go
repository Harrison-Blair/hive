package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// ErrLockTimeout reports that the bounded lock wait ended.
	ErrLockTimeout = errors.New("store lock timeout")
	// ErrLockReleased reports that a LockedStore method ran after its callback.
	ErrLockReleased = errors.New("store lock released")
)

const (
	defaultLockWait = 2 * time.Second
	lockRetry       = 10 * time.Millisecond
)

// LockedStore is a store handle that uses the lock of one WithLock callback.
type LockedStore struct {
	store    *Store
	released atomic.Bool
}

// WithLock runs fn with the store lock. It waits at most two seconds unless a
// context deadline gives another bound.
func (s *Store) WithLock(ctx context.Context, fn func(*LockedStore) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.withLock(ctx, fn)
}

func (s *Store) withLock(ctx context.Context, fn func(*LockedStore) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.checkHive(); err != nil {
		return err
	}
	if err := s.checkRoot(); err != nil {
		return err
	}
	if err := s.checkLock(); err != nil {
		return err
	}
	file, err := acquireLock(ctx, s.lockPath())
	if err != nil {
		return err
	}
	locked := &LockedStore{store: s}
	defer func() {
		locked.released.Store(true)
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}()
	return fn(locked)
}

func acquireLock(ctx context.Context, path string) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	waitUntil := time.Now().Add(defaultLockWait)
	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		waitUntil, _ = ctx.Deadline()
	}
	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = file.Close()
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		remaining := time.Until(waitUntil)
		if remaining <= 0 {
			_ = file.Close()
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if hasDeadline {
				return nil, context.DeadlineExceeded
			}
			return nil, ErrLockTimeout
		}
		timer := time.NewTimer(min(remaining, lockRetry))
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (l *LockedStore) available() error {
	if l == nil || l.released.Load() {
		return ErrLockReleased
	}
	return nil
}

// Read reads one legacy record with the held lock.
func (l *LockedStore) Read(kind, id string, target any) error {
	if err := l.available(); err != nil {
		return err
	}
	return l.store.readUnlocked(kind, id, target)
}

// Write writes one legacy record with the held lock.
func (l *LockedStore) Write(kind, id string, value any) error {
	if err := l.available(); err != nil {
		return err
	}
	data, err := marshalRecord(value)
	if err != nil {
		return err
	}
	return l.store.writeData(kind, id, data)
}

// Update runs one legacy read-change-write operation with the held lock.
func (l *LockedStore) Update(kind, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	if err := l.available(); err != nil {
		return err
	}
	return l.store.updateUnlocked(kind, id, fn)
}

// List reads legacy records with the held lock.
func (l *LockedStore) List(kind string) ([]json.RawMessage, error) {
	if err := l.available(); err != nil {
		return nil, err
	}
	if err := validateSegment("kind", kind); err != nil {
		return nil, err
	}
	return l.store.listUnlocked(kind)
}

// Next increments a named counter with the held lock.
func (l *LockedStore) Next(name string) (uint64, error) {
	if err := l.available(); err != nil {
		return 0, err
	}
	return l.store.nextUnlocked(name)
}

// ReadIn reads one nested record with the held lock.
func (l *LockedStore) ReadIn(collection []string, id string, target any) error {
	if err := l.available(); err != nil {
		return err
	}
	return l.store.readInUnlocked(collection, id, target)
}

// WriteIn writes one nested record with the held lock.
func (l *LockedStore) WriteIn(collection []string, id string, value any) error {
	if err := l.available(); err != nil {
		return err
	}
	data, err := marshalRecord(value)
	if err != nil {
		return err
	}
	return l.store.writeInData(collection, id, data)
}

// UpdateIn runs one nested read-change-write operation with the held lock.
func (l *LockedStore) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	if err := l.available(); err != nil {
		return err
	}
	return l.store.updateInUnlocked(collection, id, fn)
}

// ListIn reads one nested collection with the held lock.
func (l *LockedStore) ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error) {
	if err := l.available(); err != nil {
		return nil, err
	}
	return l.store.listInUnlocked(collection, include)
}
