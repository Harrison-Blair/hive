package harness

import (
	"strings"
	"testing"
	"time"
)

type testProfile struct {
	kind  string
	args  []string
	start time.Duration
	ready time.Duration
	names []string
}

func (p testProfile) Kind() string                               { return p.kind }
func (p testProfile) Args(model string, extra []string) []string { return p.args }
func (p testProfile) StartTimeout() time.Duration                { return p.start }
func (p testProfile) ReadyTimeout() time.Duration                { return p.ready }
func (p testProfile) ProcessNames() []string                     { return p.names }

func TestCheckBounds(t *testing.T) {
	cases := []struct {
		name  string
		start time.Duration
		ready time.Duration
		want  string
	}{
		{name: "three seconds", start: 3 * time.Second, ready: 2 * time.Second, want: "more than 3"},
		{name: "three hundred one seconds", start: 301 * time.Second, ready: 2 * time.Second, want: "at most 300"},
		{name: "below ready", start: 10 * time.Second, ready: 11 * time.Second, want: "ready"},
		{name: "valid equal", start: 10 * time.Second, ready: 10 * time.Second, want: ""},
		{name: "valid above", start: 120 * time.Second, ready: 10 * time.Second, want: ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := Check(testProfile{kind: "demo", start: test.start, ready: test.ready})
			if test.want == "" {
				if err != nil {
					t.Fatalf("Check() error = %v; want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Check() error = nil; want text %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %q; want text %q", err, test.want)
			}
		})
	}
}

func TestLookupKnownAndUnknown(t *testing.T) {
	profile, err := Lookup("claude")
	if err != nil {
		t.Fatalf("Lookup(claude) error = %v; want nil", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(claude) profile = nil; want a profile")
	}
	if profile.Kind() != "claude" {
		t.Fatalf("Kind() = %q; want claude", profile.Kind())
	}
	for _, name := range profile.ProcessNames() {
		switch name {
		case "node", "python", "sh", "bash":
			t.Fatalf("ProcessNames() = %q; generic name %q must not identify the harness", profile.ProcessNames(), name)
		}
	}
	if !contains(profile.ProcessNames(), "claude") {
		t.Fatalf("ProcessNames() = %q; want claude", profile.ProcessNames())
	}
	_, err = Lookup("nope")
	if err == nil {
		t.Fatalf("Lookup(nope) error = nil; want an unknown harness error")
	}
	if !strings.Contains(err.Error(), "unknown harness: nope") {
		t.Fatalf("Lookup(nope) error = %q; want unknown harness: nope", err)
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Fatalf("Lookup(nope) error = %q; want the known kind list", err)
	}
}

func TestEachRegisteredProfilePassesCheck(t *testing.T) {
	for _, kind := range []string{"claude"} {
		profile, err := Lookup(kind)
		if err != nil {
			t.Fatalf("Lookup(%s) error = %v", kind, err)
		}
		if profile == nil {
			t.Fatalf("Lookup(%s) profile = nil", kind)
		}
		if err := Check(profile); err != nil {
			t.Fatalf("Check(%s) error = %v", kind, err)
		}
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
