package harness

import (
	"reflect"
	"testing"
)

// The live check of task 15-harness-trust measured these results. Each
// method starts the harness in a new empty directory with no trust dialog.
func TestEachProfileGivesItsTrustResult(t *testing.T) {
	cases := []struct {
		kind     string
		version  string
		args     []string
		env      []string
		noMethod bool
	}{
		{kind: "claude", version: "2.1.307", env: []string{"CLAUDE_CODE_SANDBOXED=1"}},
		{kind: "codex", version: "0.160.1", noMethod: true},
		{kind: "pi", version: "1.3.0", args: []string{"--approve"}},
	}
	if got, want := len(cases), len(KnownKinds()); got != want {
		t.Fatalf("cases = %d; want one case for each of %d known kinds %v", got, want, KnownKinds())
	}
	for _, test := range cases {
		t.Run(test.kind, func(t *testing.T) {
			profile, err := Lookup(test.kind)
			if err != nil {
				t.Fatalf("Lookup(%s) error = %v", test.kind, err)
			}
			trust := profile.Trust()
			if trust.Version != test.version {
				t.Fatalf("Trust().Version = %q; want %q", trust.Version, test.version)
			}
			if !reflect.DeepEqual(trust.Args, test.args) {
				t.Fatalf("Trust().Args = %#v; want %#v", trust.Args, test.args)
			}
			if !reflect.DeepEqual(trust.Env, test.env) {
				t.Fatalf("Trust().Env = %#v; want %#v", trust.Env, test.env)
			}
			if got := trust.NoMethod != ""; got != test.noMethod {
				t.Fatalf("Trust().NoMethod = %q; want a reason: %t", trust.NoMethod, test.noMethod)
			}
		})
	}
}
