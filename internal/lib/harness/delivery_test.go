package harness

import "testing"

// TestSystemPromptDelivery gives the delivery of each harness. Claude Code
// takes the profile and the guide with --append-system-prompt. Each other
// harness takes them at the start of the first prompt (D4). Pi uses this path
// because Herdr start arguments cannot carry multiline text.
func TestSystemPromptDelivery(t *testing.T) {
	cases := map[string]string{
		"claude":   "--append-system-prompt",
		"pi":       "",
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
