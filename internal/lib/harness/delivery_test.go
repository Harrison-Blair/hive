package harness

import (
	"slices"
	"testing"
)

// TestGuideArgs gives the start arguments that carry the guide in the system
// prompt of each harness. Herdr rejects an argument with a newline, so each
// value is one line.
func TestGuideArgs(t *testing.T) {
	const guide = "a \"b\"\nc\\d"
	cases := map[string][]string{
		"claude": {"--append-system-prompt", `a "b"\nc\d`},
		"pi":     {"--append-system-prompt", `a "b"\nc\d`},
		"codex":  {"-c", `developer_instructions="a \"b\"\nc\\d"`},
	}
	for kind, want := range cases {
		t.Run(kind, func(t *testing.T) {
			profile, err := Lookup(kind)
			if err != nil {
				t.Fatalf("Lookup(%s) error = %v", kind, err)
			}
			if got := profile.GuideArgs(guide); !slices.Equal(got, want) {
				t.Fatalf("GuideArgs() = %q; want %q", got, want)
			}
			if got := profile.GuideArgs(""); got != nil {
				t.Fatalf("GuideArgs(\"\") = %q; want nil", got)
			}
		})
	}
}
