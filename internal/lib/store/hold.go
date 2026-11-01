package store

import (
	"fmt"
	"path/filepath"
	"time"
)

// Hold runs fn with the lock file <home>/<name>.lock. That lock is separate
// from the store lock, so fn can write records. Hold waits at most wait for
// the lock.
func (s *Store) Hold(name string, wait time.Duration, fn func() error) error {
	if err := validateComponent("lock", name); err != nil {
		return err
	}
	path := filepath.Join(s.stateDir(), name+".lock")
	if err := checkLockFile(path); err != nil {
		return err
	}
	file, err := acquireLock(path, wait)
	if err != nil {
		return fmt.Errorf("lock %s: %w", path, err)
	}
	defer unlock(file)
	return fn()
}
