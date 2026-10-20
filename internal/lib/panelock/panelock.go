// Package panelock locks prompt submissions to one pane.
package panelock

import (
	"context"
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Lock acquires the lock for one pane ID.
func Lock(ctx context.Context, home, paneID string) (*store.NamedLock, error) {
	if paneID == "" {
		return nil, fmt.Errorf("pane ID is empty")
	}
	name := "pane-" + strings.ReplaceAll(paneID, ":", "-")
	return store.LockNamed(ctx, home, name)
}
