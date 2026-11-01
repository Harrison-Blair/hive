package agent

import (
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// layoutWait bounds the wait for the layout lock. A spawn holds the lock
// through the start of its harness.
const layoutWait = 2 * time.Minute

// Hold runs fn with the layout lock of home. Spawn, hive worktree create and
// remove, and bare hive hold it while they choose an agent name, a checkout,
// or a workspace and save the result, so two of them never act on one name,
// checkout, or workspace at the same time. fn must not call Hold.
func Hold(home string, fn func() error) error {
	records, err := store.Open(home)
	if err != nil {
		return err
	}
	return records.Hold("layout", layoutWait, fn)
}
