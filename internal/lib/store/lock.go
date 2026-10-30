package store

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// errLockTimeout reports that the bounded lock wait ended.
var errLockTimeout = errors.New("store lock timeout")

const defaultLockWait = 2 * time.Second

// withLock runs fn with the store lock. It waits at most wait for the lock.
func (s *Store) withLock(wait time.Duration, fn func() error) error {
	if err := s.checkLock(); err != nil {
		return err
	}
	file, err := acquireLock(s.lockPath(), wait)
	if err != nil {
		return err
	}
	defer func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}()
	return fn()
}

// acquireLock waits for the lock in the kernel, so the lock goes to a waiter
// when the holder releases it. The wait ends after wait. An abandoned wait
// closes the file: the kernel closes the descriptor and drops the lock when
// the blocked flock returns, and the goroutine then ends.
func acquireLock(path string, wait time.Duration) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	conn, err := file.SyscallConn()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	locked := make(chan error, 1)
	go func() {
		var lockErr error
		// Control holds a reference to the descriptor, so a Close during
		// the wait does not free the descriptor number before flock returns.
		err := conn.Control(func(fd uintptr) {
			lockErr = syscall.Flock(int(fd), syscall.LOCK_EX)
		})
		if err == nil {
			err = lockErr
		}
		locked <- err
	}()
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case err := <-locked:
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		return file, nil
	case <-timer.C:
		_ = file.Close()
		return nil, errLockTimeout
	}
}
