package harness

import (
	"reflect"
	"testing"
	"time"
)

// codexModel is the installed model that the live check requests.
const codexModel = "gpt-6.1-sol"

func TestCodexBypassApprovalsAndSandbox(t *testing.T) {
	profile, err := Lookup("codex")
	if err != nil {
		t.Fatal(err)
	}
	args := profile.Args("", nil)
	for _, arg := range args {
		if arg == "--dangerously-bypass-approvals-and-sandbox" {
			return
		}
	}
	t.Fatalf("Args() = %#v; want --dangerously-bypass-approvals-and-sandbox", args)
}

func TestCodexArgs(t *testing.T) {
	profile, err := Lookup("codex")
	if err != nil {
		t.Fatalf("Lookup(codex) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(codex) profile = nil")
	}
	if got, want := profile.Kind(), "codex"; got != want {
		t.Fatalf("Kind() = %q; want %q", got, want)
	}
	cases := []struct {
		name  string
		model string
		extra []string
		want  []string
	}{
		{name: "no model", model: "", extra: nil, want: []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox"}},
		{name: "model", model: codexModel, extra: nil, want: []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--model", codexModel}},
		{name: "extra only", model: "", extra: []string{"--approve-for-me"}, want: []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--approve-for-me"}},
		{name: "model and extra", model: codexModel, extra: []string{"--sandbox", "workspace-write"}, want: []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--model", codexModel, "--sandbox", "workspace-write"}},
		{name: "colon model", model: "gpt-6.1-sol:high", extra: nil, want: []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--model", "gpt-6.1-sol:high"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := profile.Args(test.model, test.extra)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Args(%q, %#v) = %#v; want %#v", test.model, test.extra, got, test.want)
			}
			if len(got) == 0 || got[0] != "--no-daemon" {
				t.Fatalf("Args(%q, %#v) = %#v; want --no-daemon first", test.model, test.extra, got)
			}
		})
	}
	modelArgs := profile.Args(codexModel, nil)
	if count := len(modelArgs); count != 4 {
		t.Fatalf("Args(codex model) = %#v; want 4 arguments", modelArgs)
	}
	if modelArgs[2] != "--model" || modelArgs[3] != codexModel {
		t.Fatalf("Args(codex model) = %#v; want the model as one unchanged argument", modelArgs)
	}
}

func TestCodexTimeouts(t *testing.T) {
	profile, err := Lookup("codex")
	if err != nil {
		t.Fatalf("Lookup(codex) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(codex) profile = nil")
	}
	if err := Check(profile); err != nil {
		t.Fatalf("Check(codex) error = %v", err)
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

func TestCodexProcessNames(t *testing.T) {
	profile, err := Lookup("codex")
	if err != nil {
		t.Fatalf("Lookup(codex) error = %v", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(codex) profile = nil")
	}
	names := map[string]bool{}
	for _, name := range profile.ProcessNames() {
		names[name] = true
	}
	if !names["codex"] {
		t.Fatalf("ProcessNames() = %#v; want codex", profile.ProcessNames())
	}
	for _, generic := range []string{"node", "python", "sh", "bash"} {
		if names[generic] {
			t.Fatalf("ProcessNames() = %#v; must reject %q", profile.ProcessNames(), generic)
		}
	}
}

func TestCodexLookup(t *testing.T) {
	profile, err := Lookup("codex")
	if err != nil {
		t.Fatalf("Lookup(codex) error = %v; want nil", err)
	}
	if profile == nil {
		t.Fatalf("Lookup(codex) profile = nil; want a profile")
	}
	if profile.Kind() != "codex" {
		t.Fatalf("Kind() = %q; want codex", profile.Kind())
	}
	found := false
	for _, kind := range KnownKinds() {
		if kind == "codex" {
			found = true
		}
	}
	if !found {
		t.Fatalf("KnownKinds() = %#v; want codex", KnownKinds())
	}
}
