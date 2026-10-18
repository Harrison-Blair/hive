package harness

import "time"

// piProfile is the native profile of Pi.
type piProfile struct{}

func (piProfile) Kind() string { return "pi" }

func (piProfile) Args(model string, extra []string) []string { return nil }

func (piProfile) StartTimeout() time.Duration { return 0 }
func (piProfile) ReadyTimeout() time.Duration { return 0 }
func (piProfile) ProcessNames() []string      { return nil }
