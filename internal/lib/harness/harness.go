// Package harness defines the native profile of one harness kind.
package harness

import (
	"context"
	"fmt"
	"log/slog"
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
	Trust() Trust
	// GuideArgs returns the start arguments that add guide to the system
	// prompt of the harness. Each value is one line, because herdr rejects a
	// start argument with a newline. It returns nil for an empty guide.
	GuideArgs(guide string) []string
	// Deliver hands text to the harness in pane through the native message
	// path of the harness. It never types into the pane. It fails with the
	// cause and the fix when the native path is not available. It logs each
	// herdr call and external process to logger.
	Deliver(ctx context.Context, paneID, text string, logger *slog.Logger) error
	// InboxArgs writes the files that the native message path needs below
	// home and returns its start arguments. It returns nil when the harness
	// needs none.
	InboxArgs(home string) ([]string, error)
}

// Trust is the trust dialog result of one harness (decision 13 of the
// general-agent spec). A method starts the harness in a new directory with no
// trust dialog, and trusts the directory only for that one process. The spawn
// of a general agent adds Args to the start arguments and Env to the
// environment of the harness process. An empty Args and Env with no NoMethod
// means that the harness shows no trust dialog.
type Trust struct {
	// Version is the harness version of the live check.
	Version string
	// Args are the start arguments of the method.
	Args []string
	// Env holds the KEY=VALUE pairs of the method.
	Env []string
	// NoMethod is the reason when no method exists. It is empty when a method
	// exists.
	NoMethod string
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
