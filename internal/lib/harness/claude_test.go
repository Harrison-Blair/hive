package harness

import (
	"reflect"
	"testing"
	"time"
)

func TestClaudeBypassPermissions(t *testing.T) {
	profile, err := Lookup("claude")
	if err != nil {
		t.Fatal(err)
	}
	args := profile.Args("", nil)
	if len(args) == 0 || args[0] != "--dangerously-skip-permissions" {
		t.Fatalf("Args() = %#v; want --dangerously-skip-permissions", args)
	}
}

func TestClaudeArgsAndTimeouts(t *testing.T) {
	profile, err := Lookup("claude")
	if err != nil {
		t.Fatalf("Lookup(claude) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(claude) profile = nil")
	}
	if got, want := profile.Kind(), "claude"; got != want {
		t.Fatalf("Kind() = %q; want %q", got, want)
	}
	const inbound = `{"crossSessionInbound":"accept"}`
	if got, want := profile.Args("", nil), []string{"--dangerously-skip-permissions", "--settings", inbound}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args empty = %#v; want %#v", got, want)
	}
	if got, want := profile.Args("claude-haiku-5-5", nil), []string{"--dangerously-skip-permissions", "--settings", inbound, "--model", "claude-haiku-5-5"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args model = %#v; want %#v", got, want)
	}
	if got, want := profile.Args("m", []string{"--one", "--two"}), []string{"--dangerously-skip-permissions", "--settings", inbound, "--model", "m", "--one", "--two"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Args extra = %#v; want %#v", got, want)
	}
	if got, want := profile.StartTimeout(), 120*time.Second; got != want {
		t.Fatalf("StartTimeout() = %s; want %s", got, want)
	}
	if got, want := profile.ReadyTimeout(), 120*time.Second; got != want {
		t.Fatalf("ReadyTimeout() = %s; want %s", got, want)
	}
	if got, want := profile.ProcessNames(), []string{"claude"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ProcessNames() = %#v; want %#v", got, want)
	}
}

func TestProcessNamesRejectGenericNames(t *testing.T) {
	profile, err := Lookup("claude")
	if err != nil {
		t.Fatalf("Lookup(claude) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(claude) profile = nil")
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	if !names["claude"] {
		t.Fatalf("ProcessNames() = %#v; want claude", profile.ProcessNames())
	}
	for _, generic := range []string{"node", "python", "sh", "bash"} {
		if names[generic] {
			t.Fatalf("ProcessNames() = %#v; must reject %q", profile.ProcessNames(), generic)
		}
	}
}
