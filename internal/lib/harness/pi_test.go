package harness

import (
	"reflect"
	"testing"
	"time"
)

// piRoute is the approved route for the live check. The suffix keeps the
// requested thinking level. Pi normalizes the native DeepSeek level at start.
const piRoute = "opencode-go/deepseek-v4.1-flash"

func TestPiArgs(t *testing.T) {
	profile, err := Lookup("pi")
	if err != nil {
		t.Fatalf("Lookup(pi) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(pi) profile = nil")
	}
	if got, want := profile.Kind(), "pi"; got != want {
		t.Fatalf("Kind() = %q; want %q", got, want)
	}
	if got, want := profile.Args("", nil), []string{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args empty = %#v; want %#v", got, want)
	}
	if got, want := profile.Args(piRoute+":medium", nil), []string{"--model", piRoute + ":medium"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args medium = %#v; want %#v", got, want)
	}
	if got, want := profile.Args("anthropic/claude-sonnet-5-5:high", nil), []string{"--model", "anthropic/claude-sonnet-5-5:high"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args slash suffix = %#v; want %#v", got, want)
	}
	if got, want := profile.Args("m", []string{"--approve", "--one"}), []string{"--model", "m", "--approve", "--one"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args extra = %#v; want %#v", got, want)
	}
	for _, arg := range profile.Args(piRoute+":medium", nil) {
		if arg == "--provider" {
			t.Fatalf("Args = %#v; must not pass --provider", profile.Args(piRoute+":medium", nil))
		}
	}
}

func TestPiTimeouts(t *testing.T) {
	profile, err := Lookup("pi")
	if err != nil {
		t.Fatalf("Lookup(pi) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(pi) profile = nil")
	}
	if err := Check(profile); err != nil {
		t.Fatalf("Check(pi) error = %v", err)
	}
	start := profile.StartTimeout()
	if start <= 3*time.Second {
		t.Fatalf("StartTimeout() = %s; want more than 3s", start)
	}
	if start > 300*time.Second {
		t.Fatalf("StartTimeout() = %s; want at most 300s", start)
	}
	if got, want := profile.ReadyTimeout(), start; got != want {
		t.Fatalf("ReadyTimeout() = %s; want the start timeout %s", got, want)
	}
}

func TestPiProcessNames(t *testing.T) {
	profile, err := Lookup("pi")
	if err != nil {
		t.Fatalf("Lookup(pi) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(pi) profile = nil")
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	if !names["pi"] {
		t.Fatalf("ProcessNames() = %#v; want pi", profile.ProcessNames())
	}
	for _, generic := range []string{"node", "python", "sh", "bash"} {
		if names[generic] {
			t.Fatalf("ProcessNames() = %#v; must reject %q", profile.ProcessNames(), generic)
		}
	}
}

func TestPiLookup(t *testing.T) {
	profile, err := Lookup("pi")
	if err != nil {
		t.Fatalf("Lookup(pi) error = %v; want nil", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(pi) profile = nil; want a profile")
	}
	if profile.Kind() != "pi" {
		t.Fatalf("Kind() = %q; want pi", profile.Kind())
	}
	found := false
	for _, kind := range KnownKinds() {
		if kind == "pi" {
			found = true
		}
	}
	if !found {
		t.Fatalf("KnownKinds() = %#v; want pi", KnownKinds())
	}
}
