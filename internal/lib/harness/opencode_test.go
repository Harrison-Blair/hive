package harness

import (
	"reflect"
	"testing"
	"time"
)

// opencodeModel is the installed model that the live check requests. The value
// carries the provider and a slash.
const opencodeModel = "opencode-go/deepseek-v4.1-flash"

func TestOpenCodeArgs(t *testing.T) {
	profile, err := Lookup("opencode")
	if err != nil {
		t.Fatalf("Lookup(opencode) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(opencode) profile = nil")
	}
	if got, want := profile.Kind(), "opencode"; got != want {
		t.Fatalf("Kind() = %q; want %q", got, want)
	}
	cases := []struct {
		name  string
		model string
		extra []string
		want  []string
	}{
		{name: "no model", model: "", extra: nil, want: []string{}},
		{name: "model", model: opencodeModel, extra: nil, want: []string{"--model", opencodeModel}},
		{name: "extra only", model: "", extra: []string{"--auto"}, want: []string{"--auto"}},
		{name: "model and extra", model: opencodeModel, extra: []string{"--agent", "build"}, want: []string{"--model", opencodeModel, "--agent", "build"}},
		{name: "slash model", model: "anthropic/claude-sonnet-5-5", extra: nil, want: []string{"--model", "anthropic/claude-sonnet-5-5"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := profile.Args(test.model, test.extra)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Args(%q, %#v) = %#v; want %#v", test.model, test.extra, got, test.want)
			}
		})
	}
	modelArgs := profile.Args("anthropic/claude-sonnet-5-5", nil)
	if count := len(modelArgs); count != 2 {
		t.Fatalf("Args(provider/model) = %#v; want 2 arguments", modelArgs)
	}
	if modelArgs[0] != "--model" || modelArgs[1] != "anthropic/claude-sonnet-5-5" {
		t.Fatalf("Args(provider/model) = %#v; want the model as one unchanged argument", modelArgs)
	}
}

func TestOpenCodeTimeouts(t *testing.T) {
	profile, err := Lookup("opencode")
	if err != nil {
		t.Fatalf("Lookup(opencode) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(opencode) profile = nil")
	}
	if err := Check(profile); err != nil {
		t.Fatalf("Check(opencode) error = %v", err)
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

func TestOpenCodeProcessNames(t *testing.T) {
	profile, err := Lookup("opencode")
	if err != nil {
		t.Fatalf("Lookup(opencode) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(opencode) profile = nil")
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	if !names["opencode"] {
		t.Fatalf("ProcessNames() = %#v; want opencode", profile.ProcessNames())
	}
	for _, generic := range []string{"node", "python", "sh", "bash"} {
		if names[generic] {
			t.Fatalf("ProcessNames() = %#v; must reject %q", profile.ProcessNames(), generic)
		}
	}
}

func TestOpenCodeLookup(t *testing.T) {
	profile, err := Lookup("opencode")
	if err != nil {
		t.Fatalf("Lookup(opencode) error = %v; want nil", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(opencode) profile = nil; want a profile")
	}
	if profile.Kind() != "opencode" {
		t.Fatalf("Kind() = %q; want opencode", profile.Kind())
	}
	found := false
	for _, kind := range KnownKinds() {
		if kind == "opencode" {
			found = true
		}
	}
	if !found {
		t.Fatalf("KnownKinds() = %#v; want opencode", KnownKinds())
	}
}
