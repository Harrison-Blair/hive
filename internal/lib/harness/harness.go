// Package harness defines the native profile of one harness kind.
package harness

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Profile gives the native arguments, the timeouts, and the process names of
// one harness kind.
type Profile interface {
	Kind() string
	Args(model string, extra []string) []string
	StartTimeout() time.Duration
	ReadyTimeout() time.Duration
	ProcessNames() []string
}

// Check validates the bounds of one profile. A start timeout below the ready
// timeout lets herdr drop the name while hive still waits.
func Check(profile Profile) error {
	if profile == nil {
		return fmt.Errorf("harness profile is nil")
	}
	start := profile.StartTimeout()
	if start <= 3*time.Second {
		return fmt.Errorf("start timeout must be more than 3s: %s", start)
	}
	if start > 300*time.Second {
		return fmt.Errorf("start timeout must be at most 300s: %s", start)
	}
	if start < profile.ReadyTimeout() {
		return fmt.Errorf("start timeout %s is below ready timeout %s", start, profile.ReadyTimeout())
	}
	return nil
}

// Lookup returns the profile of one harness kind.
func Lookup(kind string) (Profile, error) {
	profile, found := profiles[kind]
	if !found {
		return nil, fmt.Errorf("unknown harness: %s (known: %s)", kind, strings.Join(KnownKinds(), ", "))
	}
	return profile, nil
}

// KnownKinds returns the sorted harness kinds in the registry.
func KnownKinds() []string {
	kinds := make([]string, 0, len(profiles))
	for kind := range profiles {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
