package harness

import (
	"reflect"
	"testing"
	"time"
)

// cursorModel is the installed model that the live check requests. It carries a
// bracket parameter block, so the test proves that the value stays one
// unchanged argument.
const cursorModel = "gpt-5.6-sol-high[context=1m,effort=high]"

func TestCursorArgs(t *testing.T) {
	profile, err := Lookup("cursor")
	if err != nil {
		t.Fatalf("Lookup(cursor) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(cursor) profile = nil")
	}
	if got, want := profile.Kind(), "cursor"; got != want {
		t.Fatalf("Kind() = %q; want %q", got, want)
	}
	cases := []struct {
		name  string
		model string
		extra []string
		want  []string
	}{
		{name: "no model", model: "", extra: nil, want: []string{}},
		{name: "model", model: "composer-2.5", extra: nil, want: []string{"--model", "composer-2.5"}},
		{name: "extra only", model: "", extra: []string{"--force"}, want: []string{"--force"}},
		{name: "model and extra", model: "composer-2.5", extra: []string{"--force"}, want: []string{"--model", "composer-2.5", "--force"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := profile.Args(test.model, test.extra)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Args(%q, %#v) = %#v; want %#v", test.model, test.extra, got, test.want)
			}
		})
	}
	bracketArgs := profile.Args(cursorModel, nil)
	if count := len(bracketArgs); count != 2 {
		t.Fatalf("Args(bracket model) = %#v; want 2 arguments", bracketArgs)
	}
	if bracketArgs[0] != "--model" || bracketArgs[1] != cursorModel {
		t.Fatalf("Args(bracket model) = %#v; want the model as one unchanged argument", bracketArgs)
	}
}

func TestCursorTimeouts(t *testing.T) {
	profile, err := Lookup("cursor")
	if err != nil {
		t.Fatalf("Lookup(cursor) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(cursor) profile = nil")
	}
	if err := Check(profile); err != nil {
		t.Fatalf("Check(cursor) error = %v", err)
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

func TestCursorProcessNames(t *testing.T) {
	profile, err := Lookup("cursor")
	if err != nil {
		t.Fatalf("Lookup(cursor) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(cursor) profile = nil")
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	if !names["cursor-agent"] {
		t.Fatalf("ProcessNames() = %#v; want cursor-agent", profile.ProcessNames())
	}
	for _, generic := range []string{"node", "python", "sh", "bash"} {
		if names[generic] {
			t.Fatalf("ProcessNames() = %#v; must reject %q", profile.ProcessNames(), generic)
		}
	}
}

func TestCursorLookup(t *testing.T) {
	profile, err := Lookup("cursor")
	if err != nil {
		t.Fatalf("Lookup(cursor) error = %v; want nil", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(cursor) profile = nil; want a profile")
	}
	if profile.Kind() != "cursor" {
		t.Fatalf("Kind() = %q; want cursor", profile.Kind())
	}
	found := false
	for _, kind := range KnownKinds() {
		if kind == "cursor" {
			found = true
		}
	}
	if !found {
		t.Fatalf("KnownKinds() = %#v; want cursor", KnownKinds())
	}
}
