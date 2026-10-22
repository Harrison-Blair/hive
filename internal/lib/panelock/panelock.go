// Package panelock locks prompt submissions to one pane.
package panelock

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Lock acquires the lock for one pane ID.
func Lock(ctx context.Context, home, paneID string) (*store.NamedLock, error) {
	if paneID == "" {
		return nil, fmt.Errorf("pane ID is empty")
	}
	name := "pane-" + hex.EncodeToString([]byte(paneID))
	return store.LockNamed(ctx, home, name)
}
