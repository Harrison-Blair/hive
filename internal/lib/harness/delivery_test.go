package harness

import "testing"

// TestSystemPromptDelivery gives the delivery of each harness. Claude Code and
// Pi take the profile and the guide with --append-system-prompt. Each other
// harness takes them at the start of the first prompt (D4).
func TestSystemPromptDelivery(t *testing.T) {
	cases := map[string]string{
		"claude":   "--append-system-prompt",
		"pi":       "--append-system-prompt",
		"codex":    "",
		"opencode": "",
		"cursor":   "",
	}
	for kind, want := range cases {
		t.Run(kind, func(t *testing.T) {
			profile, err := Lookup(kind)
			if err != nil {
				t.Fatalf("Lookup(%s) error = %v", kind, err)
			}
			if got := profile.SystemPromptArg(); got != want {
				t.Fatalf("SystemPromptArg() = %q; want %q", got, want)
			}
		})
	}
}
